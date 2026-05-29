package model

import "time"

// RecommendationResult represents a saved recommendation snapshot.
// Maps to the `recommendation_results` table in MySQL.
type RecommendationResult struct {
	ID              uint64
	UserID          uint64
	ResumeVersionID uint64
	ResultJSON      []byte
	CreatedAt       time.Time
}

// RecommendedJob represents a single recommended job in the response.
type RecommendedJob struct {
	JobID                uint64   `json:"job_id"`
	Title                string   `json:"title"`
	Department           string   `json:"department"`
	City                 string   `json:"city"`
	Category             string   `json:"category"`
	JobType              string   `json:"job_type"`
	Priority             string   `json:"priority"`
	RecommendationScore  int      `json:"recommendation_score"`
	RecommendationReason string   `json:"recommendation_reason"`
	MatchHighlights      []string `json:"match_highlights"`
	RiskPoints           []string `json:"risk_points"`
	SuggestedAction      string   `json:"suggested_action"`
	ExternalSource       string   `json:"external_source"`
	ExternalID           *string  `json:"external_id,omitempty"`
	DataQuality          string   `json:"data_quality"`
	SourceURL            *string  `json:"source_url,omitempty"`
	IsFeatured           bool     `json:"is_featured"`
}
