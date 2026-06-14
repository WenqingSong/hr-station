package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"offer-catcher/server/internal/model"
)

// ─── SemanticProfileGate ──────────────────────────────────────────────────

// SemanticProfileGateResult is the summary of a gate check or generation run.
type SemanticProfileGateResult struct {
	TargetType   string `json:"target_type"`
	TargetID     int64  `json:"target_id"`
	ContentHash  string `json:"content_hash"`
	ProfileID    int64  `json:"profile_id"`
	Summary      string `json:"summary"`
	Status       string `json:"status"`
	TagsCount    int    `json:"tags_count"`
	Generated    bool   `json:"generated"`
	Reused       bool   `json:"reused"`
	MarkedStale  bool   `json:"marked_stale"`
	ErrorMessage string `json:"error_message,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
}

// SemanticProfileGate checks profile existence, computes hashes, and triggers generation.
type SemanticProfileGate struct {
	service        *SemanticProfileService
	profiler       *SemanticProfileProfiler
	enableProfiler bool // if false, gate only checks existence (read-only mode)

	// async fields — in-flight dedup for background profile generation
	asyncMu       sync.Mutex
	asyncInFlight map[string]struct{}
}

func asyncKey(targetType string, targetID int64) string {
	return fmt.Sprintf("%s:%d", targetType, targetID)
}

// NewSemanticProfileGate creates a new gate.
func NewSemanticProfileGate(service *SemanticProfileService, profiler *SemanticProfileProfiler) *SemanticProfileGate {
	return &SemanticProfileGate{
		service:        service,
		profiler:       profiler,
		enableProfiler: true,
		asyncInFlight:  make(map[string]struct{}),
	}
}

// ─── Resume Gate ────────────────────────────────────────────────────────

// EnsureResumeProfile ensures a verified resume profile exists, generating one if needed.
func (g *SemanticProfileGate) EnsureResumeProfile(ctx context.Context, resumeID int64) (*SemanticProfileGateResult, error) {
	return g.ensureProfile(ctx, model.SemanticTargetResume, resumeID, nil)
}

// EnsureJobProfile ensures a verified job profile exists, generating one if needed.
func (g *SemanticProfileGate) EnsureJobProfile(ctx context.Context, jobID int64) (*SemanticProfileGateResult, error) {
	return g.ensureProfile(ctx, model.SemanticTargetJob, jobID, nil)
}

// EnsureProfilesForRecommendation ensures both resume and all job profiles are ready.
func (g *SemanticProfileGate) EnsureProfilesForRecommendation(ctx context.Context, resumeID int64, jobIDs []int64) ([]*SemanticProfileGateResult, error) {
	var results []*SemanticProfileGateResult

	resumeResult, err := g.ensureProfile(ctx, model.SemanticTargetResume, resumeID, nil)
	if err != nil {
		return results, fmt.Errorf("ensure profiles: resume %d: %w", resumeID, err)
	}
	results = append(results, resumeResult)

	for _, jobID := range jobIDs {
		jobResult, err := g.ensureProfile(ctx, model.SemanticTargetJob, jobID, nil)
		if err != nil {
			results = append(results, &SemanticProfileGateResult{
				TargetType:   model.SemanticTargetJob,
				TargetID:     jobID,
				Status:       model.SemanticProfileStatusFailed,
				ErrorMessage: err.Error(),
			})
			continue
		}
		results = append(results, jobResult)
	}

	return results, nil
}

// ─── Async Background Generation (Non-blocking) ──────────────────────────

// EnsureResumeProfileAsync schedules a background resume SEP profile generation.
// It never blocks the calling goroutine. If an in-flight generation for the same
// resume already exists, the call is silently dropped (dedup).
func (g *SemanticProfileGate) EnsureResumeProfileAsync(resumeID int64) {
	key := asyncKey(model.SemanticTargetResume, resumeID)
	g.scheduleAsync(key, func(ctx context.Context) (*SemanticProfileGateResult, error) {
		return g.EnsureResumeProfile(ctx, resumeID)
	})
}

// EnsureJobProfileAsync schedules a background job SEP profile generation.
// It never blocks the calling goroutine. If an in-flight generation for the same
// job already exists, the call is silently dropped (dedup).
func (g *SemanticProfileGate) EnsureJobProfileAsync(jobID int64) {
	key := asyncKey(model.SemanticTargetJob, jobID)
	g.scheduleAsync(key, func(ctx context.Context) (*SemanticProfileGateResult, error) {
		return g.EnsureJobProfile(ctx, jobID)
	})
}

// scheduleAsync is the shared async dispatch with in-flight deduplication.
func (g *SemanticProfileGate) scheduleAsync(key string, fn func(ctx context.Context) (*SemanticProfileGateResult, error)) {
	g.asyncMu.Lock()
	if _, inFlight := g.asyncInFlight[key]; inFlight {
		g.asyncMu.Unlock()
		return // already running, skip duplicate
	}
	g.asyncInFlight[key] = struct{}{}
	g.asyncMu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[SEP-Gate-Async] %s: panic recovered: %v", key, r)
			}
			g.asyncMu.Lock()
			delete(g.asyncInFlight, key)
			g.asyncMu.Unlock()
		}()

		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		start := time.Now()
		result, err := fn(bgCtx)
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("[SEP-Gate-Async] %s: generation failed after %v: %v", key, elapsed, err)
		} else if result != nil {
			log.Printf("[SEP-Gate-Async] %s: generation done in %v (status=%s tags=%d reused=%v)",
				key, elapsed, result.Status, result.TagsCount, result.Reused)
		}
	}()
}

// ─── Internal ──────────────────────────────────────────────────────────────

func (g *SemanticProfileGate) ensureProfile(ctx context.Context, targetType string, targetID int64, _ []int64) (*SemanticProfileGateResult, error) {
	var hash string
	var err error

	switch targetType {
	case model.SemanticTargetResume:
		hash, err = g.service.ComputeResumeContentHash(ctx, targetID)
	case model.SemanticTargetJob:
		hash, err = g.service.ComputeJobContentHash(ctx, targetID)
	default:
		return nil, fmt.Errorf("ensure profile: unknown target type %q", targetType)
	}
	if err != nil {
		return &SemanticProfileGateResult{
			TargetType:   targetType,
			TargetID:     targetID,
			Status:       model.SemanticProfileStatusFailed,
			ErrorMessage: fmt.Sprintf("hash compute: %v", err),
		}, fmt.Errorf("ensure profile: %s %d: %w", targetType, targetID, err)
	}

	// Try to find existing verified profile with matching hash
	existing, tags, err := g.service.GetVerifiedProfile(ctx, targetType, targetID, hash)
	if err != nil {
		log.Printf("[SEP-Gate] %s %d: lookup error: %v", targetType, targetID, err)
	}
	if existing != nil && existing.Status == model.SemanticProfileStatusVerified {
		// Integrity check: verified profile must have >= 5 tags
		if len(tags) < 5 {
			log.Printf("[SEP-Gate] %s %d: verified profile %d has only %d tags (<5), marking failed for regeneration",
				targetType, targetID, existing.ID, len(tags))
			_ = g.service.UpdateProfileStatus(ctx, existing.ID, model.SemanticProfileStatusFailed,
				"SemanticProfileGate: verified profile has insufficient tags (<5), marked failed for regeneration")
			// Fall through to generation
		} else {
			return &SemanticProfileGateResult{
				TargetType:  targetType,
				TargetID:    targetID,
				ContentHash: hash,
				ProfileID:   existing.ID,
				Status:      existing.Status,
				TagsCount:   len(tags),
				Summary:     truncateStr(existing.SummarySentence, 100),
				Reused:      true,
			}, nil
		}
	}

	// Need to generate
	if !g.enableProfiler || g.profiler == nil {
		return &SemanticProfileGateResult{
			TargetType:   targetType,
			TargetID:     targetID,
			ContentHash:  hash,
			Status:       model.SemanticProfileStatusFailed,
			ErrorMessage: "profiler not available (read-only gate mode)",
		}, fmt.Errorf("ensure profile: %s %d: profiler not available", targetType, targetID)
	}

	log.Printf("[SEP-Gate] %s %d: generating new profile (hash=%s...)", targetType, targetID, hash[:16])

	// Mark stale existing profiles first
	_ = g.service.MarkStale(ctx, targetType, targetID)

	var profile *model.SemanticProfile
	switch targetType {
	case model.SemanticTargetResume:
		profile, tags, err = g.profiler.GenerateResumeSemanticProfile(ctx, 0, targetID)
	case model.SemanticTargetJob:
		profile, tags, err = g.profiler.GenerateJobSemanticProfile(ctx, targetID)
	}

	if err != nil {
		log.Printf("[SEP-Gate] %s %d: generation failed: %v", targetType, targetID, err)
		return &SemanticProfileGateResult{
			TargetType:   targetType,
			TargetID:     targetID,
			ContentHash:  hash,
			Status:       model.SemanticProfileStatusFailed,
			ErrorMessage: err.Error(),
		}, fmt.Errorf("ensure profile: %s %d: generation: %w", targetType, targetID, err)
	}

	return &SemanticProfileGateResult{
		TargetType:  targetType,
		TargetID:    targetID,
		ContentHash: hash,
		ProfileID:   profile.ID,
		Status:      profile.Status,
		TagsCount:   len(tags),
		Summary:     truncateStr(profile.SummarySentence, 100),
		Generated:   true,
	}, nil
}

// truncateStr truncates a string to maxLen runes (not bytes), adding "..." if truncated.
// Uses []rune to avoid cutting multi-byte UTF-8 characters (e.g. Chinese) mid-byte.
func truncateStr(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}
