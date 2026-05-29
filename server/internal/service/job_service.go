package service

import (
	"context"
	"encoding/json"
	"fmt"

	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// JobListRequest holds the parameters for listing jobs.
type JobListRequest struct {
	Keyword        string
	City           string
	Category       string
	ExternalSource string
	Limit          int
	Offset         int
}

// JobResponse represents a job in API responses.
type JobResponse struct {
	ID              uint64          `json:"id"`
	Title           string          `json:"title"`
	Department      string          `json:"department"`
	City            string          `json:"city"`
	Category        string          `json:"category"`
	JobType         string          `json:"job_type"`
	Description     string          `json:"description"`
	Requirements    string          `json:"requirements"`
	SkillKeywords   json.RawMessage `json:"skill_keywords"`
	ExternalSource  string          `json:"external_source"`
	ExternalID      *string         `json:"external_id,omitempty"`
	ExternalInnerID *string         `json:"external_inner_id,omitempty"`
	SourceURL       *string         `json:"source_url,omitempty"`
	DataQuality     string          `json:"data_quality"`
	IsFeatured      bool            `json:"is_featured"`
	SyncedAt        *string         `json:"synced_at,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

// JobListResponse represents the response for listing jobs.
type JobListResponse struct {
	Items  []*JobResponse `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// ValidExternalSources defines allowed values for external_source filter.
var ValidExternalSources = map[string]bool{
	"local_mock":   true,
	"tencent_join": true,
}

// JobService handles job business logic.
type JobService struct {
	jobRepo *repository.JobRepository
}

// NewJobService creates a new JobService.
func NewJobService(jobRepo *repository.JobRepository) *JobService {
	return &JobService{jobRepo: jobRepo}
}

// List returns a paginated list of jobs matching the request filters.
func (s *JobService) List(ctx context.Context, req JobListRequest) (*JobListResponse, error) {
	// Validate external_source if provided
	if req.ExternalSource != "" && !ValidExternalSources[req.ExternalSource] {
		return nil, fmt.Errorf("invalid external_source")
	}

	// Apply defaults and bounds
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	query := repository.JobQuery{
		Keyword:        req.Keyword,
		City:           req.City,
		Category:       req.Category,
		ExternalSource: req.ExternalSource,
		Limit:          req.Limit,
		Offset:         req.Offset,
	}

	jobs, total, err := s.jobRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	items := make([]*JobResponse, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, toJobResponse(j))
	}

	return &JobListResponse{
		Items:  items,
		Total:  total,
		Limit:  req.Limit,
		Offset: req.Offset,
	}, nil
}

// GetDetail returns a single job by ID.
func (s *JobService) GetDetail(ctx context.Context, id uint64) (*JobResponse, error) {
	job, err := s.jobRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job detail: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}
	return toJobResponse(job), nil
}

// toJobResponse converts a model.Job to a JobResponse.
// Note: raw_json is intentionally excluded from API responses.
func toJobResponse(j *model.Job) *JobResponse {
	var skillKeywords json.RawMessage
	if len(j.SkillKeywordsJSON) > 0 {
		skillKeywords = json.RawMessage(j.SkillKeywordsJSON)
	} else {
		skillKeywords = json.RawMessage("[]")
	}

	var syncedAtStr *string
	if j.SyncedAt != nil {
		s := j.SyncedAt.Format("2006-01-02T15:04:05Z07:00")
		syncedAtStr = &s
	}

	return &JobResponse{
		ID:              j.ID,
		Title:           j.Title,
		Department:      j.Department,
		City:            j.City,
		Category:        j.Category,
		JobType:         j.JobType,
		Description:     j.Description,
		Requirements:    j.Requirements,
		SkillKeywords:   skillKeywords,
		ExternalSource:  j.ExternalSource,
		ExternalID:      j.ExternalID,
		ExternalInnerID: j.ExternalInnerID,
		SourceURL:       j.SourceURL,
		DataQuality:     j.DataQuality,
		IsFeatured:      j.IsFeatured,
		SyncedAt:        syncedAtStr,
		CreatedAt:       j.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       j.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
