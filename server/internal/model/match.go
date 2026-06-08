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

	// K9-2-Regression-B: Deep fields
	FitPoints             []string           `json:"fit_points,omitempty"`
	ResumeEvidence        []string           `json:"resume_evidence,omitempty"`
	JobRequirements       []string           `json:"job_requirements,omitempty"`
	InterviewTalkingPoints []string          `json:"interview_talking_points,omitempty"`
	DeliveryStrategy      string             `json:"delivery_strategy,omitempty"`
	ImprovementSuggestions []MatchSuggestion `json:"improvement_suggestions,omitempty"`
	Summary               string             `json:"summary,omitempty"`

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

// MatchResumeFeatureCard represents a runtime feature card for a resume.
type MatchResumeFeatureCard struct {
	ResumeID            uint64   `json:"resume_id"`
	PrimaryDirection    string   `json:"primary_direction"`
	SecondaryDirections []string `json:"secondary_directions,omitempty"`
	StrongCapabilities  []string `json:"strong_capabilities"`
	WeakCapabilities    []string `json:"weak_capabilities,omitempty"`
	EvidenceSummary     []string `json:"evidence_summary"`
	ProjectSignals      []string `json:"project_signals,omitempty"`
	InternshipSignals   []string `json:"internship_signals,omitempty"`
	NegativePreferences []string `json:"negative_preferences,omitempty"`

	IsTechnical        bool `json:"is_technical"`
	IsAlgorithm        bool `json:"is_algorithm"`
	IsBackend          bool `json:"is_backend"`
	IsProduct          bool `json:"is_product"`
	IsPublicService    bool `json:"is_public_service"`
	IsBusinessSolution bool `json:"is_business_solution"`
	IsExploration      bool `json:"is_exploration"`
}

// MatchJobFeatureCard represents a runtime feature card for a job.
type MatchJobFeatureCard struct {
	JobID              uint64   `json:"job_id"`
	Title              string   `json:"title"`
	Direction          string   `json:"direction"`
	Category           string   `json:"category"`
	Department         string   `json:"department"`
	JobFamily          string   `json:"job_family,omitempty"`
	CoreRequirements   []string `json:"core_requirements"`
	HardRequirements   []string `json:"hard_requirements,omitempty"`
	PreferredRequirements []string `json:"preferred_requirements,omitempty"`
	KeywordRequirements []string `json:"keyword_requirements,omitempty"`
	RiskRequirements   []string `json:"risk_requirements,omitempty"`

	IsTechnical         bool `json:"is_technical"`
	IsAlgorithm         bool `json:"is_algorithm"`
	IsComputerVision    bool `json:"is_computer_vision"`
	IsBackend           bool `json:"is_backend"`
	IsProduct           bool `json:"is_product"`
	IsUserResearch      bool `json:"is_user_research"`
	IsProjectManagement bool `json:"is_project_management"`
	IsBusinessSolution  bool `json:"is_business_solution"`
	IsPublicServiceLike bool `json:"is_public_service_like"`
}

// MatchHardGuardResult represents the hard guard decision.
type MatchHardGuardResult struct {
	Result     string   `json:"result"` // pass / downgrade / weak_related / mismatch
	MaxScore   int      `json:"max_score"`
	Reasons    []string `json:"reasons"`
	RiskPoints []string `json:"risk_points,omitempty"`
}

// ApplyReadinessObj represents the structured apply readiness.
type ApplyReadinessObj struct {
	Level  string `json:"level"`  // ready / prepare_first / cautious_try / not_recommended
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

// DeliveryStrategyObj represents a structured delivery strategy.
type DeliveryStrategyObj struct {
	Recommendation string   `json:"recommendation"` // 是否建议优先投递
	Preparations   []string `json:"preparations"`   // 投递前需要补什么
	Highlights     []string `json:"highlights"`     // 面试时应该突出什么
	Risks          []string `json:"risks"`          // 不建议盲投的原因
}

// AnalysisBasis describes what data was used to generate the match analysis.
type AnalysisBasis struct {
	JobDescriptionUsed  bool   `json:"job_description_used"`
	JobRequirementsUsed bool   `json:"job_requirements_used"`
	DataQuality         string `json:"data_quality"`
}
