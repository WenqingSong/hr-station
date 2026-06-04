package service

import (
	"context"
	"encoding/json"
	"fmt"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// CreateResumeRequest represents a request to create a new resume version.
// All fields are optional.
// V1 fields: name, phone, email, location, target_city, target_role,
//            education, experience, project, skills, self_evaluation, raw_text.
// V2 fields (K8-5-22-B): schema_version, resume_name, target_direction, description,
//            profile, job_intention, educations, internships, projects, skills (object),
//            certificates, awards, portfolio_links, extra_info.
type CreateResumeRequest struct {
	// V1 fields
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

	// V2 fields (K8-5-22-B)
	SchemaVersion   int             `json:"schema_version"`   // 1=V1 (default), 2=V2
	ResumeName      string          `json:"resume_name"`      // V2 简历名称
	TargetDirection string          `json:"target_direction"` // V2 目标方向
	Description     string          `json:"description"`      // V2 描述
	Profile         json.RawMessage `json:"profile"`          // V2 基础信息
	JobIntention    json.RawMessage `json:"job_intention"`    // V2 求职意向
	Educations      json.RawMessage `json:"educations"`       // V2 教育经历 (extended)
	Internships     json.RawMessage `json:"internships"`      // V2 实习经历
	Certificates    json.RawMessage `json:"certificates"`     // V2 证书与语言
	Awards          json.RawMessage `json:"awards"`           // V2 获奖与竞赛
	PortfolioLinks  json.RawMessage `json:"portfolio_links"`  // V2 作品链接
	ExtraInfo       json.RawMessage `json:"extra_info"`       // V2 补充信息
}

// ResumeVersionResponse represents a resume version in API responses.
type ResumeVersionResponse struct {
	// V1 fields
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

	// V2 fields (K8-5-22-B)
	SchemaVersion   int             `json:"schema_version"`
	TargetDirection string          `json:"target_direction"`
	JobType         string          `json:"job_type"`
	ExpectedCities  string          `json:"expected_cities"`
	School          string          `json:"school"`
	Degree          string          `json:"degree"`
	Major           string          `json:"major"`
	GraduationYear  string          `json:"graduation_year"`
	Profile         json.RawMessage `json:"profile"`
	JobIntention    json.RawMessage `json:"job_intention"`
	Internships     json.RawMessage `json:"internships"`
	Certificates    json.RawMessage `json:"certificates"`
	Awards          json.RawMessage `json:"awards"`
	PortfolioLinks  json.RawMessage `json:"portfolio_links"`
	ExtraInfo       json.RawMessage `json:"extra_info"`
}

// VersionListItem represents a simplified version entry in the version list.
type VersionListItem struct {
	ID         uint64 `json:"id"`
	VersionNo  int    `json:"version_no"`
	Name       string `json:"name"`
	TargetCity string `json:"target_city"`
	TargetRole string `json:"target_role"`
	CreatedAt  string `json:"created_at"`

	// V2 fields (K8-5-22-B)
	SchemaVersion   int    `json:"schema_version"`
	TargetDirection string `json:"target_direction"`
	ExpectedCities  string `json:"expected_cities"`
	School          string `json:"school"`
	Degree          string `json:"degree"`
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
// V1: schema_version missing or 1 → uses V1 request fields.
// V2: schema_version=2 → uses V2 module fields with V1 summary backfill.
func (s *ResumeService) CreateVersion(ctx context.Context, userID uint64, req CreateResumeRequest) (*ResumeVersionResponse, error) {
	// Get next version number
	versionNo, err := s.resumeRepo.GetNextVersionNo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next version number: %w", err)
	}

	// Normalize V1 JSON fields: default to empty arrays if not provided
	educationJSON := normalizeJSONArray(req.Education)
	experienceJSON := normalizeJSONArray(req.Experience)
	projectJSON := normalizeJSONArray(req.Project)
	skillsJSON := normalizeJSONArray(req.Skills)

	isV2 := req.SchemaVersion >= 2
	schemaVer := 1
	if isV2 {
		schemaVer = 2
	}

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
		SchemaVersion:  schemaVer,
	}

	if isV2 {
		// V2 write: use V2 module fields, backfill V1 summary fields
		s.backfillV1FromV2(resume, &req)
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

// backfillV1FromV2 populates V1 summary fields from V2 request data
// so that old list pages don't break when viewing V2 resumes.
func (s *ResumeService) backfillV1FromV2(resume *model.ResumeVersion, req *CreateResumeRequest) {
	// Use V2 resume name if provided
	if req.ResumeName != "" {
		resume.Name = req.ResumeName
	}

	// Backfill target_role from target_direction or job_intention
	if req.TargetDirection != "" {
		resume.TargetRole = req.TargetDirection
	} else if len(req.JobIntention) > 0 {
		if roles := extractStringArrayFromJSON(req.JobIntention, "target_roles"); len(roles) > 0 {
			resume.TargetRole = roles[0]
		}
	}

	// Backfill target_city from expected_cities or job_intention
	if req.TargetCity == "" {
		if cities := extractStringArrayFromJSON(req.JobIntention, "expected_cities"); len(cities) > 0 {
			resume.TargetCity = cities[0]
		}
	}

	// Backfill location from profile.current_city
	if req.Location == "" && len(req.Profile) > 0 {
		if city := extractStringFromJSON(req.Profile, "current_city"); city != "" {
			resume.Location = city
		}
	}

	// Backfill self_evaluation from extra_info
	if req.SelfEvaluation == "" && len(req.ExtraInfo) > 0 {
		if eval := extractStringFromJSON(req.ExtraInfo, "self_evaluation"); eval != "" {
			resume.SelfEvaluation = eval
		}
	}

	// Backfill phone/email from profile
	if req.Phone == "" && len(req.Profile) > 0 {
		if p := extractStringFromJSON(req.Profile, "phone"); p != "" {
			resume.Phone = p
		}
	}
	if req.Email == "" && len(req.Profile) > 0 {
		if e := extractStringFromJSON(req.Profile, "email"); e != "" {
			resume.Email = e
		}
	}

	// V2 JSON fields → model
	resume.ProfileJSON = []byte(req.Profile)
	resume.JobIntentionJSON = []byte(req.JobIntention)
	resume.InternshipJSON = []byte(req.Internships)
	resume.CertificatesJSON = []byte(req.Certificates)
	resume.AwardsJSON = []byte(req.Awards)
	resume.PortfolioJSON = []byte(req.PortfolioLinks)
	resume.ExtraJSON = []byte(req.ExtraInfo)

	// V2 also writes to V1 JSON fields: educations→education_json, internships→experience_json,
	// projects→project_json, skills→skills_json (extended format)
	if len(req.Educations) > 0 {
		resume.EducationJSON = []byte(req.Educations)
	}
	if len(req.Internships) > 0 {
		resume.ExperienceJSON = []byte(req.Internships)
	}
	// Note: req.Project and req.Skills (V1 fields) already handled above;
	// V2 request may send "projects" and "skills" (same field names, extended format).
	if len(req.Project) > 0 {
		resume.ProjectJSON = []byte(req.Project)
	}
	if len(req.Skills) > 0 {
		resume.SkillsJSON = []byte(req.Skills)
	}

	// V2 summary structured fields
	resume.TargetDirection = req.TargetDirection

	// Extract summary from job_intention JSON
	if len(req.JobIntention) > 0 {
		resume.JobType = extractStringFromJSON(req.JobIntention, "job_type")
		if cities := extractStringArrayFromJSON(req.JobIntention, "expected_cities"); len(cities) > 0 {
			first := cities[0]
			for _, c := range cities[1:] {
				first += "," + c
			}
			resume.ExpectedCities = first
		}
	}

	// Extract summary from profile JSON
	if len(req.Profile) > 0 {
		resume.School = extractStringFromJSON(req.Profile, "school")
		resume.Degree = extractStringFromJSON(req.Profile, "degree")
		resume.Major = extractStringFromJSON(req.Profile, "major")
		resume.GraduationYear = extractStringFromJSON(req.Profile, "graduation_year")
	}
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
			ID:              v.ID,
			VersionNo:       v.VersionNo,
			Name:            v.Name,
			TargetCity:      v.TargetCity,
			TargetRole:      v.TargetRole,
			CreatedAt:       v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SchemaVersion:   v.SchemaVersion,
			TargetDirection: v.TargetDirection,
			ExpectedCities:  v.ExpectedCities,
			School:          v.School,
			Degree:          v.Degree,
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
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid version number", 400)
	}

	resume, err := s.resumeRepo.GetByUserIDAndVersionNo(ctx, userID, versionNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get resume version: %w", err)
	}
	if resume == nil {
		return nil, apperror.ErrResumeNotFound
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

		// V2 fields (K8-5-22-B)
		SchemaVersion:   r.SchemaVersion,
		TargetDirection: r.TargetDirection,
		JobType:         r.JobType,
		ExpectedCities:  r.ExpectedCities,
		School:          r.School,
		Degree:          r.Degree,
		Major:           r.Major,
		GraduationYear:  r.GraduationYear,
		Profile:         normalizeResponseJSON(r.ProfileJSON),
		JobIntention:    normalizeResponseJSON(r.JobIntentionJSON),
		Internships:     normalizeResponseJSON(r.InternshipJSON),
		Certificates:    normalizeResponseJSON(r.CertificatesJSON),
		Awards:          normalizeResponseJSON(r.AwardsJSON),
		PortfolioLinks:  normalizeResponseJSON(r.PortfolioJSON),
		ExtraInfo:       normalizeResponseJSON(r.ExtraJSON),
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

// extractStringFromJSON extracts a string value by key from a JSON object.
// Returns empty string if the key is not found or the JSON is invalid.
func extractStringFromJSON(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// extractStringArrayFromJSON extracts a string array by key from a JSON object.
// Returns nil if the key is not found or the JSON is invalid.
func extractStringArrayFromJSON(raw json.RawMessage, key string) []string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
