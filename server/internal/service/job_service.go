package service

import (
	"context"
	"encoding/json"
	"fmt"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// JobListRequest holds the parameters for listing jobs.
type JobListRequest struct {
	Keyword        string
	Direction      string
	City           string
	RecruitType    string
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
	Direction       string          `json:"direction"`
	Category        string          `json:"category"`
	JobType         string          `json:"job_type"`
	RecruitType     string          `json:"recruit_type"`
	Description     string          `json:"description"`
	Requirements    string          `json:"requirements"`
	SkillKeywords   json.RawMessage `json:"skill_keywords"`
	Status          string          `json:"status"`
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

// FacetItem represents a single facet entry.
type FacetItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// JobFacetsResponse represents aggregated facet counts.
type JobFacetsResponse struct {
	Directions   []FacetItem `json:"directions"`
	Cities       []FacetItem `json:"cities"`
	RecruitTypes []FacetItem `json:"recruit_types"`
	Categories   []FacetItem `json:"categories"`
}

// ValidExternalSources defines allowed values for external_source filter.
var ValidExternalSources = map[string]bool{
	"local_mock":   true,
	"tencent_join": true,
}

// --- label maps for facet output ---

var directionLabels = map[string]string{
	"软件开发":       "软件开发",
	"算法与大数据":   "算法与大数据",
	"产品策划与运营": "产品策划/运营",
	"设计":          "设计",
}

var recruitTypeLabels = map[string]string{
	"校招": "校园招聘",
	"实习": "实习生",
	"提前批": "提前批",
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
	if req.ExternalSource != "" && !ValidExternalSources[req.ExternalSource] {
		return nil, fmt.Errorf("invalid external_source")
	}

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
		Direction:      req.Direction,
		City:           req.City,
		RecruitType:    req.RecruitType,
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
		return nil, apperror.ErrJobNotFound
	}
	return toJobResponse(job), nil
}

// GetFacets returns aggregated facet counts for job filtering.
func (s *JobService) GetFacets(ctx context.Context) (*JobFacetsResponse, error) {
	f, err := s.jobRepo.GetFacets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get facets: %w", err)
	}

	resp := &JobFacetsResponse{}

	for _, d := range f.Directions {
		label := d.Value
		if l, ok := directionLabels[d.Value]; ok {
			label = l
		}
		resp.Directions = append(resp.Directions, FacetItem{Value: d.Value, Label: label, Count: d.Count})
	}
	for _, c := range f.Cities {
		resp.Cities = append(resp.Cities, FacetItem{Value: c.Value, Label: c.Value, Count: c.Count})
	}
	for _, r := range f.RecruitTypes {
		label := r.Value
		if l, ok := recruitTypeLabels[r.Value]; ok {
			label = l
		}
		resp.RecruitTypes = append(resp.RecruitTypes, FacetItem{Value: r.Value, Label: label, Count: r.Count})
	}
	for _, c := range f.Categories {
		resp.Categories = append(resp.Categories, FacetItem{Value: c.Value, Label: c.Value, Count: c.Count})
	}

	return resp, nil
}

// toJobResponse converts a model.Job to a JobResponse.
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

	status := j.Status
	if status == "" {
		status = "active"
	}

	return &JobResponse{
		ID:              j.ID,
		Title:           j.Title,
		Department:      j.Department,
		City:            j.City,
		Direction:       j.Direction,
		Category:        j.Category,
		JobType:         j.JobType,
		RecruitType:     j.JobType, // API 层 recruit_type 映射到 job_type
		Description:     j.Description,
		Requirements:    j.Requirements,
		SkillKeywords:   skillKeywords,
		Status:          status,
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
