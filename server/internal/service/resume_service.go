package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// CreateResumeRequest represents a request to create a new resume version.
// All fields are optional.
type CreateResumeRequest struct {
	Name           string          `json:"name"`
	Phone          string          `json:"phone"`
	Email          string          `json:"email"`
	Location       string          `json:"location"`
	TargetCity     string          `json:"target_city"`
	TargetRole     string          `json:"target_role"`
	Education      json.RawMessage `json:"education"`
	Experience     json.RawMessage `json:"experience"`
	Project        json.RawMessage `json:"project"`
	Skills         json.RawMessage `json:"skills"`
	SelfEvaluation string          `json:"self_evaluation"`
	RawText        string          `json:"raw_text"`
}

// ResumeVersionResponse represents a resume version in API responses.
type ResumeVersionResponse struct {
	ID             uint64          `json:"id"`
	VersionNo      int             `json:"version_no"`
	Name           string          `json:"name"`
	Phone          string          `json:"phone"`
	Email          string          `json:"email"`
	Location       string          `json:"location"`
	TargetCity     string          `json:"target_city"`
	TargetRole     string          `json:"target_role"`
	Education      json.RawMessage `json:"education"`
	Experience     json.RawMessage `json:"experience"`
	Project        json.RawMessage `json:"project"`
	Skills         json.RawMessage `json:"skills"`
	SelfEvaluation string          `json:"self_evaluation"`
	RawText        string          `json:"raw_text"`
	CreatedAt      string          `json:"created_at"`
}

// VersionListItem represents a simplified version entry in the version list.
type VersionListItem struct {
	ID         uint64 `json:"id"`
	VersionNo  int    `json:"version_no"`
	Name       string `json:"name"`
	TargetCity string `json:"target_city"`
	TargetRole string `json:"target_role"`
	CreatedAt  string `json:"created_at"`
}

// VersionListResponse represents the response for listing versions.
type VersionListResponse struct {
	Versions []VersionListItem `json:"versions"`
	Total    int               `json:"total"`
}

// ResumeService handles resume business logic.
type ResumeService struct {
	resumeRepo *repository.ResumeRepository
}

// NewResumeService creates a new ResumeService.
func NewResumeService(resumeRepo *repository.ResumeRepository) *ResumeService {
	return &ResumeService{resumeRepo: resumeRepo}
}

// CreateVersion creates a new resume version for the given user.
// Each call creates a new version; historical versions are read-only.
func (s *ResumeService) CreateVersion(ctx context.Context, userID uint64, req CreateResumeRequest) (*ResumeVersionResponse, error) {
	// Get next version number
	versionNo, err := s.resumeRepo.GetNextVersionNo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version number: %w", err)
	}

	// Normalize JSON fields: default to empty arrays if not provided
	educationJSON := normalizeJSONArray(req.Education)
	experienceJSON := normalizeJSONArray(req.Experience)
	projectJSON := normalizeJSONArray(req.Project)
	skillsJSON := normalizeJSONArray(req.Skills)

	resume := &model.ResumeVersion{
		UserID:         userID,
		VersionNo:      versionNo,
		Name:           req.Name,
		Phone:          req.Phone,
		Email:          req.Email,
		Location:       req.Location,
		TargetCity:     req.TargetCity,
		TargetRole:     req.TargetRole,
		EducationJSON:  educationJSON,
		ExperienceJSON: experienceJSON,
		ProjectJSON:    projectJSON,
		SkillsJSON:     skillsJSON,
		SelfEvaluation: req.SelfEvaluation,
		RawText:        req.RawText,
	}

	if err := s.resumeRepo.CreateVersion(ctx, resume); err != nil {
		return nil, fmt.Errorf("failed to create resume version: %w", err)
	}

	// Reload to get database-generated fields (created_at)
	created, err := s.resumeRepo.FindByID(ctx, resume.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload created resume: %w", err)
	}

	return toResumeVersionResponse(created), nil
}

// GetCurrent returns the latest resume version for the given user.
// Returns nil response if the user has no resume.
func (s *ResumeService) GetCurrent(ctx context.Context, userID uint64) (*ResumeVersionResponse, error) {
	resume, err := s.resumeRepo.GetCurrentByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current resume: %w", err)
	}
	if resume == nil {
		return nil, nil
	}
	return toResumeVersionResponse(resume), nil
}

// ListVersions returns all resume versions for the given user.
func (s *ResumeService) ListVersions(ctx context.Context, userID uint64) (*VersionListResponse, error) {
	versions, err := s.resumeRepo.ListVersionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list resume versions: %w", err)
	}

	items := make([]VersionListItem, 0, len(versions))
	for _, v := range versions {
		items = append(items, VersionListItem{
			ID:         v.ID,
			VersionNo:  v.VersionNo,
			Name:       v.Name,
			TargetCity: v.TargetCity,
			TargetRole: v.TargetRole,
			CreatedAt:  v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &VersionListResponse{
		Versions: items,
		Total:    len(items),
	}, nil
}

// GetByVersionNo returns a specific resume version by version number.
func (s *ResumeService) GetByVersionNo(ctx context.Context, userID uint64, versionNo int) (*ResumeVersionResponse, error) {
	if versionNo < 1 {
		return nil, errors.New("invalid version number")
	}

	resume, err := s.resumeRepo.GetByUserIDAndVersionNo(ctx, userID, versionNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get resume version: %w", err)
	}
	if resume == nil {
		return nil, errors.New("resume version not found")
	}

	return toResumeVersionResponse(resume), nil
}

// normalizeJSONArray ensures a JSON field is a valid JSON value.
// If the input is empty or not valid JSON, returns an empty JSON array [].
func normalizeJSONArray(raw json.RawMessage) []byte {
	if len(raw) == 0 {
		return []byte("[]")
	}
	if !json.Valid(raw) {
		return []byte("[]")
	}
	return raw
}

// toResumeVersionResponse converts a model.ResumeVersion to a ResumeVersionResponse.
func toResumeVersionResponse(r *model.ResumeVersion) *ResumeVersionResponse {
	return &ResumeVersionResponse{
		ID:             r.ID,
		VersionNo:      r.VersionNo,
		Name:           r.Name,
		Phone:          r.Phone,
		Email:          r.Email,
		Location:       r.Location,
		TargetCity:     r.TargetCity,
		TargetRole:     r.TargetRole,
		Education:      normalizeResponseJSON(r.EducationJSON),
		Experience:     normalizeResponseJSON(r.ExperienceJSON),
		Project:        normalizeResponseJSON(r.ProjectJSON),
		Skills:         normalizeResponseJSON(r.SkillsJSON),
		SelfEvaluation: r.SelfEvaluation,
		RawText:        r.RawText,
		CreatedAt:      r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// normalizeResponseJSON ensures response JSON fields are valid.
// Returns empty array for nil/empty values.
func normalizeResponseJSON(data []byte) json.RawMessage {
	if len(data) == 0 {
		return json.RawMessage("[]")
	}
	return json.RawMessage(data)
}
