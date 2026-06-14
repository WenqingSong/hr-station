package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ─── Backfill Types ──────────────────────────────────────────────────────

// BackfillScanItem represents a single target discovered during a scan.
type BackfillScanItem struct {
	TargetType   string `json:"target_type"`
	TargetID     int64  `json:"target_id"`
	TargetName   string `json:"target_name,omitempty"`
	Status       string `json:"status"` // ready / missing / stale / failed
	ProfileID    int64  `json:"profile_id,omitempty"`
	ContentHash  string `json:"content_hash,omitempty"`
	ErrorMsg     string `json:"error_message,omitempty"`
}

// BackfillOptions controls a backfill run.
type BackfillOptions struct {
	TargetType   string // "resume" | "job" | "all"
	TargetID     int64  // 0 means all, >0 means only this target
	DryRun       bool
	Limit        int
	Concurrency  int
	RebuildStale bool
}

// BackfillResult summarises a backfill run.
type BackfillResult struct {
	TargetType string             `json:"target_type"`
	Total      int                `json:"total"`
	Ready      int                `json:"ready"`
	Missing    int                `json:"missing"`
	Stale      int                `json:"stale"`
	Failed     int                `json:"failed"`
	Scheduled  int                `json:"scheduled"`
	Generated  int                `json:"generated"`
	Skipped    int                `json:"skipped"`
	GenFailed  int                `json:"gen_failed"`
	Duration   string             `json:"duration"`
	Items      []BackfillScanItem `json:"items,omitempty"`
}

// ─── BackfillService ─────────────────────────────────────────────────────

// SemanticProfileBackfillService provides SEP profile warmup / backfill.
// It scans existing profile coverage and generates missing ones,
// with concurrency control, dry-run, and idempotent re-execution.
type SemanticProfileBackfillService struct {
	gate        *SemanticProfileGate
	semanticRepo *repository.SemanticProfileRepository
	resumeRepo  *repository.ResumeRepository
	jobRepo     *repository.JobRepository
}

// NewSemanticProfileBackfillService creates a new backfill service.
func NewSemanticProfileBackfillService(gate *SemanticProfileGate, semanticRepo *repository.SemanticProfileRepository, resumeRepo *repository.ResumeRepository, jobRepo *repository.JobRepository) *SemanticProfileBackfillService {
	return &SemanticProfileBackfillService{
		gate:         gate,
		semanticRepo: semanticRepo,
		resumeRepo:   resumeRepo,
		jobRepo:      jobRepo,
	}
}

// ─── Scan ─────────────────────────────────────────────────────────────────

// Scan discovers the current SEP profile coverage for the given target type.
// It never generates profiles and never writes to the database.
func (s *SemanticProfileBackfillService) Scan(ctx context.Context, targetType string) ([]BackfillScanItem, error) {
	switch targetType {
	case "resume":
		return s.scanResumes(ctx)
	case "job":
		return s.scanJobs(ctx)
	default:
		return nil, fmt.Errorf("unknown target type: %s", targetType)
	}
}

func (s *SemanticProfileBackfillService) scanResumes(ctx context.Context) ([]BackfillScanItem, error) {
	pairs, err := s.resumeRepo.ListAllIDPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}
	var items []BackfillScanItem
	for _, pair := range pairs {
		item := BackfillScanItem{TargetType: "resume", TargetID: int64(pair.ID), TargetName: fmt.Sprintf("resume-%d", pair.ID)}
		sp, spErr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, int64(pair.ID))
		if spErr != nil || sp == nil {
			item.Status = "missing"
		} else {
			item.ProfileID = sp.ID
			switch sp.Status {
			case model.SemanticProfileStatusVerified:
				item.Status = "ready"
			case model.SemanticProfileStatusFailed:
				item.Status = "failed"
				item.ErrorMsg = sp.ErrorMessage
			case model.SemanticProfileStatusStale:
				item.Status = "stale"
			default:
				item.Status = "missing"
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *SemanticProfileBackfillService) scanJobs(ctx context.Context) ([]BackfillScanItem, error) {
	jobs, err := s.jobRepo.ListRecommendCandidates(ctx, 100)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	var items []BackfillScanItem
	for _, j := range jobs {
		item := BackfillScanItem{TargetType: "job", TargetID: int64(j.ID), TargetName: j.Title}
		sp, spErr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, int64(j.ID))
		if spErr != nil || sp == nil {
			item.Status = "missing"
		} else {
			item.ProfileID = sp.ID
			switch sp.Status {
			case model.SemanticProfileStatusVerified:
				item.Status = "ready"
			case model.SemanticProfileStatusFailed:
				item.Status = "failed"
				item.ErrorMsg = sp.ErrorMessage
			case model.SemanticProfileStatusStale:
				item.Status = "stale"
			default:
				item.Status = "missing"
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// ─── BackfillMissing ─────────────────────────────────────────────────────

// BackfillMissing scans and generates missing/stale/failed profiles.
// In dry-run mode it only scans and reports what WOULD be generated.
func (s *SemanticProfileBackfillService) BackfillMissing(ctx context.Context, opts BackfillOptions) (*BackfillResult, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}
	if opts.Limit <= 0 {
		opts.Limit = 100
	}
	startTime := time.Now()
	targets := s.expandTargets(opts.TargetType)

	var allResults []BackfillResult
	for _, t := range targets {
		items, err := s.Scan(ctx, t)
		if err != nil {
			return nil, fmt.Errorf("scan %s: %w", t, err)
		}
		// Filter by target-id if specified
		if opts.TargetID > 0 {
			var filtered []BackfillScanItem
			for _, it := range items {
				if it.TargetID == opts.TargetID {
					filtered = append(filtered, it)
				}
			}
			items = filtered
		}
		result := s.summariseScan(t, items)
		result.Items = items

		if opts.DryRun {
			allResults = append(allResults, result)
			continue
		}

		// Collect items that need generation
		var toGen []BackfillScanItem
		for _, it := range items {
			if it.Status == "ready" {
				continue // always skip ready
			}
			if it.Status == "stale" && !opts.RebuildStale {
				continue // skip stale unless explicitly requested
			}
			toGen = append(toGen, it)
		}
		if len(toGen) > opts.Limit {
			toGen = toGen[:opts.Limit]
		}

		result.Scheduled = len(toGen)

		// Generate with concurrency control
		sem := make(chan struct{}, opts.Concurrency)
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i := range toGen {
			wg.Add(1)
			go func(item *BackfillScanItem) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				var genErr error
				switch item.TargetType {
				case "resume":
					r, e := s.gate.EnsureResumeProfile(ctx, item.TargetID)
					if e != nil || (r != nil && r.Status != model.SemanticProfileStatusVerified) {
						if e != nil {
							genErr = e
						} else {
							genErr = fmt.Errorf("status=%s", r.Status)
						}
					}
				case "job":
					r, e := s.gate.EnsureJobProfile(ctx, item.TargetID)
					if e != nil || (r != nil && r.Status != model.SemanticProfileStatusVerified) {
						if e != nil {
							genErr = e
						} else {
							genErr = fmt.Errorf("status=%s", r.Status)
						}
					}
				}

				mu.Lock()
				defer mu.Unlock()
				if genErr != nil {
					result.GenFailed++
					item.ErrorMsg = genErr.Error()
					log.Printf("[SEP-Backfill] %s %d: generation failed: %v", item.TargetType, item.TargetID, genErr)
				} else {
					result.Generated++
					item.Status = "ready"
					log.Printf("[SEP-Backfill] %s %d: generation done", item.TargetType, item.TargetID)
				}
			}(&toGen[i])
		}
		wg.Wait()

		result.Duration = time.Since(startTime).Round(time.Second).String()
		result.Skipped = result.Ready // ready items are always skipped
		allResults = append(allResults, result)
	}

	// Merge multiple target types into one summary
	merged := s.mergeResults(allResults, targets)
	merged.Items = s.collectItems(allResults)
	return merged, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func (s *SemanticProfileBackfillService) expandTargets(targetType string) []string {
	switch targetType {
	case "all":
		return []string{"resume", "job"}
	default:
		return []string{targetType}
	}
}

func (s *SemanticProfileBackfillService) summariseScan(targetType string, items []BackfillScanItem) BackfillResult {
	r := BackfillResult{TargetType: targetType, Total: len(items)}
	for _, it := range items {
		switch it.Status {
		case "ready":
			r.Ready++
		case "missing":
			r.Missing++
		case "stale":
			r.Stale++
		case "failed":
			r.Failed++
		}
	}
	return r
}

func (s *SemanticProfileBackfillService) mergeResults(results []BackfillResult, targets []string) *BackfillResult {
	merged := &BackfillResult{TargetType: "all"}
	for _, r := range results {
		merged.Total += r.Total
		merged.Ready += r.Ready
		merged.Missing += r.Missing
		merged.Stale += r.Stale
		merged.Failed += r.Failed
		merged.Scheduled += r.Scheduled
		merged.Generated += r.Generated
		merged.GenFailed += r.GenFailed
		merged.Skipped += r.Skipped
		if r.Duration != "" {
			merged.Duration = r.Duration
		}
	}
	if len(results) == 1 {
		merged.TargetType = results[0].TargetType
	}
	return merged
}

func (s *SemanticProfileBackfillService) collectItems(results []BackfillResult) []BackfillScanItem {
	var items []BackfillScanItem
	for _, r := range results {
		items = append(items, r.Items...)
	}
	if len(items) > 50 {
		items = items[:50]
	}
	return items
}
