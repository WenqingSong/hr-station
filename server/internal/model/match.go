package model

import "time"

// MatchAnalysis represents a saved match analysis snapshot.
type MatchAnalysis struct {
	ID                     uint64
	UserID                 uint64
	ResumeID               uint64
	JobID                  uint64
	GenerationTier         string
	GeneratedBy            string
	Status                 string
	ProfileType            string
	ProfileDominantProfile string
	Score                  int
	ResultJSON             []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// MatchEvidence represents a single JD-resume evidence chain entry.
type MatchEvidence struct {
	Requirement  string `json:"requirement"`
	ResumePoint  string `json:"resume_point"`
	Judgement    string `json:"judgement"`
	ResumeModule string `json:"resume_module,omitempty"`
}

// MatchSuggestion represents an optimization suggestion with priority.
type MatchSuggestion struct {
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Reason   string `json:"reason"`
	Action   string `json:"action"`
}

// TopFixItem represents a quick fix item before applying.
type TopFixItem struct {
	Issue             string `json:"issue"`
	WhyItMatters      string `json:"why_it_matters"`
	Action            string `json:"action"`
	ExampleReference  string `json:"example_reference"`
	SelfCheckQuestion string `json:"self_check_question"`
	EstimatedEffort   string `json:"estimated_effort"`
	ResumeModule      string `json:"resume_module,omitempty"`
}

// MatchAnalysisResult represents the structured match analysis output.
type MatchAnalysisResult struct {
	JobID                   uint64            `json:"job_id"`
	JobTitle                string            `json:"job_title"`
	ResumeID                uint64            `json:"resume_id"`
	MatchScore              int               `json:"match_score"`
	ScreeningResult         string            `json:"screening_result"`
	EvidenceChain           []MatchEvidence   `json:"evidence_chain"`
	MatchHighlights         []string          `json:"match_highlights"`
	RiskPoints              []string          `json:"risk_points"`
	OptimizationSuggestions []MatchSuggestion `json:"optimization_suggestions"`
	NextSteps               []string          `json:"next_steps"`
	AuthenticityNotice      string            `json:"authenticity_notice"`
	ExternalSource          string            `json:"external_source"`
	ExternalID              *string           `json:"external_id,omitempty"`
	DataQuality             string            `json:"data_quality"`
	SourceURL               *string           `json:"source_url,omitempty"`
	IsFeatured              bool              `json:"is_featured"`

	MatchLevel         string   `json:"match_level,omitempty"`
	MatchLabel         string   `json:"match_label,omitempty"`
	OfferGoSummary     string   `json:"offer_go_summary,omitempty"`
	GapPoints          []string `json:"gap_points,omitempty"`
	InterviewQuestions []string `json:"interview_questions,omitempty"`

	ApplyReadiness         string       `json:"apply_readiness,omitempty"`
	OptimizationStopSignal string       `json:"optimization_stop_signal,omitempty"`
	TopFixItems            []TopFixItem `json:"top_fix_items,omitempty"`

	ProfileType            string `json:"profile_type,omitempty"`
	ProfileDominantProfile string `json:"profile_dominant_profile,omitempty"`

	GeneratedBy    string `json:"generated_by,omitempty"`
	AgentStatus    string `json:"agent_status,omitempty"`
	RepairApplied  bool   `json:"repair_applied,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
	ErrorCode      string `json:"error_code,omitempty"`
}

// MatchLatestResult is the DTO returned by GET /api/jobs/{id}/match/latest.
type MatchLatestResult struct {
	HasResult             bool        `json:"has_result"`
	AnalysisID            uint64      `json:"analysis_id,omitempty"`
	ResumeID              uint64      `json:"resume_id,omitempty"`
	JobID                 uint64      `json:"job_id,omitempty"`
	GenerationTier        string      `json:"generation_tier,omitempty"`
	GeneratedBy           string      `json:"generated_by,omitempty"`
	Status                string      `json:"status,omitempty"`
	ProfileType           string      `json:"profile_type,omitempty"`
	ProfileDominantProfile string      `json:"profile_dominant_profile,omitempty"`
	CreatedAt             string      `json:"created_at,omitempty"`
	Result                interface{} `json:"result,omitempty"`
}

// AnalysisBasis describes what data was used to generate the match analysis.
type AnalysisBasis struct {
	JobDescriptionUsed  bool   `json:"job_description_used"`
	JobRequirementsUsed bool   `json:"job_requirements_used"`
	DataQuality         string `json:"data_quality"`
}
