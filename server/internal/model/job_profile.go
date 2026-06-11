package model

import "time"

// JobProfile represents a job profile row in job_profiles.
type JobProfile struct {
	ID                    uint64     `json:"id"`
	JobID                 uint64     `json:"job_id"`
	Direction             string     `json:"direction"`
	Category              string     `json:"category"`
	PrimaryTagsJSON       string     `json:"primary_tags_json"`
	SkillTagsJSON         string     `json:"skill_tags_json"`
	RequirementTagsJSON   string     `json:"requirement_tags_json"`
	ResponsibilityTagsJSON string    `json:"responsibility_tags_json"`
	RiskTagsJSON          string     `json:"risk_tags_json"`
	EvidenceJSON          string     `json:"evidence_json"`
	Source                string     `json:"source"`
	Version               string     `json:"version"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
