package model

import "time"

// RecommendationResult represents a saved recommendation snapshot.
type RecommendationResult struct {
	ID         uint64
	UserID     uint64
	ResumeID   uint64
	ResultJSON []byte
	CreatedAt  time.Time
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

	RecommendLabel   string   `json:"recommend_label,omitempty"`
	OfferGoSummary   string   `json:"offer_go_summary,omitempty"`
	QuerySuggestions []string `json:"query_suggestions,omitempty"`

	RecommendationLabel    string   `json:"recommendation_label,omitempty"`
	RecommendationPriority string   `json:"recommendation_priority,omitempty"`

	ApplyReadiness         string       `json:"apply_readiness,omitempty"`
	OptimizationStopSignal string       `json:"optimization_stop_signal,omitempty"`
	TopFixItems            []TopFixItem `json:"top_fix_items,omitempty"`

	MatchEvidence []RecommendationEvidence `json:"match_evidence,omitempty"`

	GeneratedBy    string `json:"generated_by,omitempty"`
	AgentStatus    string `json:"agent_status,omitempty"`
	RepairApplied  bool   `json:"repair_applied,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ErrorCode      string `json:"error_code,omitempty"`
}

// RecommendationEvidence represents match evidence for a single job.
type RecommendationEvidence struct {
	ResumeSignal   string `json:"resume_signal"`
	JobRequirement string `json:"job_requirement"`
	Strength       string `json:"strength"`
	Explanation    string `json:"explanation"`
}

// RecommendationJudgeResult represents the result of the Judge layer.
type RecommendationJudgeResult struct {
	Pass        bool     `json:"pass"`
	Severity    string   `json:"severity"`
	Issues      []string `json:"issues"`
	FixStrategy string   `json:"fix_strategy"`
}

