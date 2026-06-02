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
	Keyword            string
	Direction          string
	City               string
	RecruitType        string
	Category           string
	BusinessGroup      string
	Skill              string
	ApplicationProject string
	ApplicationSubtype string
	ApplicationType    string
	ExternalSource     string
	Limit              int
	Offset             int
}

// OrgInfo represents an organization in API responses.
type OrgInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// JobResponse represents a job in API responses (J3 expanded).
type JobResponse struct {
	ID                  uint64          `json:"id"`
	ExternalID          *string         `json:"external_id,omitempty"`
	Title               string          `json:"title"`
	Department          string          `json:"department"`
	City                string          `json:"city"`
	Locations           json.RawMessage `json:"locations,omitempty"`
	Organizations       json.RawMessage `json:"organizations,omitempty"`
	Direction           string          `json:"direction"`
	Category            string          `json:"category"`
	JobType             string          `json:"job_type"`
	RecruitType         string          `json:"recruit_type"`
	ApplicationType     string          `json:"application_type,omitempty"`
	ApplicationProject  string          `json:"application_project,omitempty"`
	ApplicationSubtype  string          `json:"application_subtype,omitempty"`
	Slogan              string          `json:"slogan,omitempty"`
	Summary             string          `json:"summary,omitempty"`
	Description         string          `json:"description"`
	Requirements        string          `json:"requirements"`
	Responsibilities    json.RawMessage `json:"responsibilities,omitempty"`
	MustHave            json.RawMessage `json:"must_have,omitempty"`
	NiceToHave          json.RawMessage `json:"nice_to_have,omitempty"`
	BonusPoints         json.RawMessage `json:"bonus_points,omitempty"`
	Notes               json.RawMessage `json:"notes,omitempty"`
	SkillKeywords       json.RawMessage `json:"skill_keywords"`
	Skills              json.RawMessage `json:"skills,omitempty"`
	Status              string          `json:"status"`
	ExternalSource      string          `json:"external_source"`
	ExternalInnerID     *string         `json:"external_inner_id,omitempty"`
	SourceURL           *string         `json:"source_url,omitempty"`
	SourceName          string          `json:"source_name,omitempty"`
	DataQuality         string          `json:"data_quality"`
	InterviewLocations  json.RawMessage `json:"interview_locations,omitempty"`
	IsFeatured          bool            `json:"is_featured"`
	SyncedAt            *string         `json:"synced_at,omitempty"`
	CollectedAt         *string         `json:"collected_at,omitempty"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
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

// JobFacetsResponse represents aggregated facet counts (J3 expanded).
type JobFacetsResponse struct {
	Directions          []FacetItem `json:"directions"`
	Cities              []FacetItem `json:"cities"`
	RecruitTypes        []FacetItem `json:"recruit_types"`
	Categories          []FacetItem `json:"categories"`
	ApplicationProjects []FacetItem `json:"application_projects,omitempty"`
	ApplicationSubtypes []FacetItem `json:"application_subtypes,omitempty"`
	JobTypes            []FacetItem `json:"job_types,omitempty"`
	BusinessGroups      []FacetItem `json:"business_groups,omitempty"`
	Skills              []FacetItem `json:"skills,omitempty"`
}

var ValidExternalSources = map[string]bool{
	"local_mock":   true,
	"tencent_join": true,
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

func NewJobService(jobRepo *repository.JobRepository) *JobService {
	return &JobService{jobRepo: jobRepo}
}

func (s *JobService) List(ctx context.Context, req JobListRequest) (*JobListResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	query := repository.JobQuery{
		Keyword:            req.Keyword,
		Direction:          req.Direction,
		City:               req.City,
		RecruitType:        req.RecruitType,
		Category:           req.Category,
		BusinessGroup:      req.BusinessGroup,
		Skill:              req.Skill,
		ApplicationProject: req.ApplicationProject,
		ApplicationSubtype: req.ApplicationSubtype,
		ApplicationType:    req.ApplicationType,
		ExternalSource:     req.ExternalSource,
		Limit:              req.Limit,
		Offset:             req.Offset,
	}

	jobs, total, err := s.jobRepo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	items := make([]*JobResponse, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, toJobResponse(j))
	}

	return &JobListResponse{Items: items, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}

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

func (s *JobService) GetFacets(ctx context.Context) (*JobFacetsResponse, error) {
	f, err := s.jobRepo.GetFacets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get facets: %w", err)
	}

	resp := &JobFacetsResponse{}

	for _, d := range f.Directions {
		resp.Directions = append(resp.Directions, FacetItem{Value: d.Value, Label: d.Value, Count: d.Count})
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
	for _, p := range f.ApplicationProjects {
		resp.ApplicationProjects = append(resp.ApplicationProjects, FacetItem{Value: p.Value, Label: p.Value, Count: p.Count})
	}
	for _, s := range f.ApplicationSubtypes {
		resp.ApplicationSubtypes = append(resp.ApplicationSubtypes, FacetItem{Value: s.Value, Label: s.Value, Count: s.Count})
	}
	for _, j := range f.JobTypes {
		resp.JobTypes = append(resp.JobTypes, FacetItem{Value: j.Value, Label: j.Value, Count: j.Count})
	}
	for _, b := range f.BusinessGroups {
		resp.BusinessGroups = append(resp.BusinessGroups, FacetItem{Value: b.Value, Label: b.Value, Count: b.Count})
	}
	for _, s := range f.Skills {
		resp.Skills = append(resp.Skills, FacetItem{Value: s.Value, Label: s.Value, Count: s.Count})
	}

	return resp, nil
}

func toJobResponse(j *model.Job) *JobResponse {
	skillKeywords := emptyJSONArray()
	if len(j.SkillKeywordsJSON) > 0 {
		skillKeywords = j.SkillKeywordsJSON
	}

	var syncedAtStr, collectedAtStr *string
	if j.SyncedAt != nil {
		s := j.SyncedAt.Format("2006-01-02T15:04:05Z07:00")
		syncedAtStr = &s
	}
	if j.CollectedAt != nil {
		s := j.CollectedAt.Format("2006-01-02T15:04:05Z07:00")
		collectedAtStr = &s
	}

	status := j.Status
	if status == "" {
		status = "open"
	}

	return &JobResponse{
		ID:                 j.ID,
		ExternalID:         j.ExternalID,
		Title:              j.Title,
		Department:         j.Department,
		City:               j.City,
		Locations:          jsonOrEmpty(j.LocationsJSON),
		Organizations:      jsonOrEmpty(j.OrganizationsJSON),
		Direction:          j.Direction,
		Category:           j.Category,
		JobType:            j.JobType,
		RecruitType:        j.JobType,
		ApplicationType:    j.ApplicationType,
		ApplicationProject: j.ApplicationProject,
		ApplicationSubtype: j.ApplicationSubtype,
		Slogan:             j.Slogan,
		Summary:            j.Summary,
		Description:        j.Description,
		Requirements:       j.Requirements,
		Responsibilities:   jsonOrEmpty(j.ResponsibilitiesJSON),
		MustHave:           jsonOrEmpty(j.MustHaveJSON),
		NiceToHave:         jsonOrEmpty(j.NiceToHaveJSON),
		BonusPoints:        jsonOrEmpty(j.BonusPointsJSON),
		Notes:              jsonOrEmpty(j.NotesJSON),
		SkillKeywords:      skillKeywords,
		Skills:             jsonOrEmpty(j.SkillsJSON),
		Status:             status,
		ExternalSource:     j.ExternalSource,
		ExternalInnerID:    j.ExternalInnerID,
		SourceURL:          j.SourceURL,
		SourceName:         j.SourceName,
		DataQuality:        j.DataQuality,
		InterviewLocations: jsonOrEmpty(j.InterviewLocationsJSON),
		IsFeatured:         j.IsFeatured,
		SyncedAt:           syncedAtStr,
		CollectedAt:        collectedAtStr,
		CreatedAt:          j.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          j.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func jsonOrEmpty(b []byte) json.RawMessage {
	if len(b) > 0 {
		return json.RawMessage(b)
	}
	return emptyJSONArray()
}

func emptyJSONArray() json.RawMessage {
	return json.RawMessage("[]")
}
