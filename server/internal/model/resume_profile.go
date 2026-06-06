package model

import "time"

// Profile status constants
const (
	ProfileTypeFast    = "fast"
	ProfileTypeQuality = "quality"

	ProfileStatusPending = "pending"
	ProfileStatusReady   = "ready"
	ProfileStatusFailed  = "failed"
	ProfileStatusStale   = "stale"

	ProfileVersionV1 = "v1"
)

// ResumeProfile represents a candidate profile row in resume_profiles.
type ResumeProfile struct {
	ID                uint64     `json:"id"`
	ResumeID          uint64     `json:"resume_id"`
	UserID            uint64     `json:"user_id"`
	ProfileType       string     `json:"profile_type"`
	ProfileVersion    string     `json:"profile_version"`
	Status            string     `json:"status"`
	DominantProfile   string     `json:"dominant_profile"`
	PrimaryTagsJSON   string     `json:"primary_tags_json"`
	CandidateTagsJSON string     `json:"candidate_tags_json"`
	RiskTagsJSON      string     `json:"risk_tags_json"`
	EvidenceJSON      string     `json:"evidence_json"`
	ProfileSummary    string     `json:"profile_summary"`
	Source            string     `json:"source"`
	Generator         string     `json:"generator"`
	ErrorMessage      string     `json:"error_message"`
	ResumeUpdatedAt   *time.Time `json:"resume_updated_at,omitempty"`
	GeneratedAt       *time.Time `json:"generated_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ResumeProfileStatusResponse is the API response for GET /api/resumes/{id}/profile/status.
type ResumeProfileStatusResponse struct {
	ResumeID        uint64       `json:"resume_id"`
	FastStatus      string       `json:"fast_status"`
	QualityStatus   string       `json:"quality_status"`
	DominantProfile string       `json:"dominant_profile"`
	PrimaryTags     []ProfileTag `json:"primary_tags"`
	UpdatedAt       string       `json:"updated_at"`
}

// ProfileTag represents a single tag with confidence and evidence.
type ProfileTag struct {
	Tag        string `json:"tag"`
	Group      string `json:"group,omitempty"`
	Confidence int    `json:"confidence"`
	Evidence   string `json:"evidence,omitempty"`
}
