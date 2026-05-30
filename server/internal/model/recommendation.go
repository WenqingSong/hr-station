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
// Fields added for LLM Agent output (4-3.4) are safe additions that won't break
// existing consumers: RecommendLabel, OfferGoSummary, QuerySuggestions,
// AgentStatus, RepairApplied, FallbackReason, ErrorCode.
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

	// LLM-enriched fields (added in 4-3.4, safe additions with omitempty)
	RecommendLabel   string   `json:"recommend_label,omitempty"`
	OfferGoSummary   string   `json:"offer_go_summary,omitempty"`
	QuerySuggestions []string `json:"query_suggestions,omitempty"`

	// Strategic fields (4-3.7.C.1)
	ApplyReadiness         string       `json:"apply_readiness,omitempty"`
	OptimizationStopSignal string       `json:"optimization_stop_signal,omitempty"`
	TopFixItems            []TopFixItem `json:"top_fix_items,omitempty"`

	// Agent metadata (internal, 4-3.5 will unify API meta)
	GeneratedBy    string `json:"generated_by,omitempty"`
	AgentStatus    string `json:"agent_status,omitempty"`
	RepairApplied  bool   `json:"repair_applied,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ErrorCode      string `json:"error_code,omitempty"`
}
