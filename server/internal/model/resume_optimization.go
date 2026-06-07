package model

import "time"

// ResumeOptimizationSnapshot represents a saved resume optimization snapshot.
// Maps to the `resume_optimization_results` table in MySQL.
type ResumeOptimizationSnapshot struct {
	ID                     uint64
	UserID                 uint64
	ResumeID               uint64
	JobID                  uint64 // optional, 0 if no job reference
	OptimizationMode       string
	GenerationTier         string
	GeneratedBy            string
	Status                 string
	ProfileType            string
	ProfileDominantProfile string
	TargetDirection        string
	ResultJSON             []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// ResumeOptimizationTargetJob represents the target job information in the request.
type ResumeOptimizationTargetJob struct {
	JobID   string `json:"job_id,omitempty"`
	Title   string `json:"title,omitempty"`
	Location string `json:"location,omitempty"`
	JDText  string `json:"jd_text,omitempty"`
}

// ResumeOptimizationUserIntention represents the user's job search intention.
type ResumeOptimizationUserIntention struct {
	TargetRoles      []string `json:"target_roles,omitempty"`
	ExpectedLocations []string `json:"expected_locations,omitempty"`
	AcceptRelocation  bool     `json:"accept_relocation,omitempty"`
}

// ResumeOptimizationFixItem represents a quick fix item before applying.
type ResumeOptimizationFixItem struct {
	Issue            string `json:"issue"`
	WhyItMatters     string `json:"why_it_matters"`
	Action           string `json:"action"`
	ExampleReference string `json:"example_reference"`
	SelfCheckQuestion string `json:"self_check_question"`
	LearnMoreQuery   string `json:"learn_more_query"`
	EstimatedEffort  string `json:"estimated_effort"`
	ResumeModule     string `json:"resume_module,omitempty"` // K8-5-22-I-R
}

// ResumeOptimizationLongTermItem represents a long-term improvement direction.
type ResumeOptimizationLongTermItem struct {
	Direction string `json:"direction"`
	Note      string `json:"note"`
}

// ResumeOptimizationJDEvidenceItem represents a single JD evidence chain entry.
type ResumeOptimizationJDEvidenceItem struct {
	JDRequirement    string `json:"jd_requirement"`
	ResumeEvidence   string `json:"resume_evidence"`
	EvidenceStrength string `json:"evidence_strength"`
	MatchJudgement   string `json:"match_judgement"`
	GapAnalysis      string `json:"gap_analysis,omitempty"`
	ScreeningRisk    string `json:"screening_risk,omitempty"`
}

// ResumeOptimizationResult represents the structured resume optimization output.
// This is the core response model that incorporates 4-3.5 strategic boundaries.
type ResumeOptimizationResult struct {
	// Agent metadata
	AgentStatus       string `json:"agent_status"`
	GeneratedBy       string `json:"generated_by"`
	FallbackRequired  bool   `json:"fallback_required"`
	RepairApplied     bool   `json:"repair_applied"`

	// Core strategic fields (4-3.5.C)
	ApplyReadiness         string `json:"apply_readiness"`
	OptimizationStopSignal string `json:"optimization_stop_signal"`

	// Summary
	Summary string `json:"summary"`

	// Quick fix items (max 3, 4-3.5.C)
	TopFixItems []ResumeOptimizationFixItem `json:"top_fix_items"`

	// Long-term improvement (4-3.5.C)
	LongTermImprovement []ResumeOptimizationLongTermItem `json:"long_term_improvement"`

	// JD evidence chain (4-3.5.D)
	JDEvidenceChain []ResumeOptimizationJDEvidenceItem `json:"jd_evidence_chain"`

	// Verification notice (4-3.5.B)
	VerificationNotice string `json:"verification_notice"`

	// Optional fallback metadata
	FallbackReason string `json:"fallback_reason,omitempty"`
	ErrorCode      string `json:"error_code,omitempty"`
}

// ─── K9-3-C: New optimization result types ──────────────────────────────────

type OptimizationFixItem struct {
	Section string `json:"section"`
	Issue   string `json:"issue"`
	Action  string `json:"action"`
}

type OptimizationRewriteExample struct {
	Section string `json:"section"`
	Before  string `json:"before"`
	After   string `json:"after"`
	Reason  string `json:"reason"`
}

type OptimizationLongTermItem struct {
	Direction string `json:"direction"`
	Note      string `json:"note"`
}

// OptimizationResultV2 is the new K9-3 result format.
type OptimizationResultV2 struct {
	OptimizationID          uint64                      `json:"optimization_id"`
	ResumeID                uint64                      `json:"resume_id"`
	JobID                   uint64                      `json:"job_id"`
	OptimizationMode        string                      `json:"optimization_mode"`
	TargetDirection         string                      `json:"target_direction,omitempty"`
	GenerationTier          string                      `json:"generation_tier"`
	GeneratedBy             string                      `json:"generated_by"`
	Status                  string                      `json:"status"`
	ProfileType             string                      `json:"profile_type"`
	ProfileDominantProfile  string                      `json:"profile_dominant_profile"`
	OptimizationScore       int                         `json:"optimization_score"`
	OptimizationLevel       string                      `json:"optimization_level"`
	OneSentenceSummary      string                      `json:"one_sentence_summary"`
	QuickFixItems           []OptimizationFixItem       `json:"quick_fix_items"`
	StructureFixItems       []OptimizationFixItem       `json:"structure_fix_items"`
	ContentFixItems         []OptimizationFixItem       `json:"content_fix_items"`
	KeywordFixItems         []OptimizationFixItem       `json:"keyword_fix_items"`
	EvidenceFixItems        []OptimizationFixItem       `json:"evidence_fix_items"`
	RiskPoints              []string                    `json:"risk_points"`
	RewriteExamples         []OptimizationRewriteExample `json:"rewrite_examples"`
	DeliveryFocus           string                      `json:"delivery_focus"`
	LongTermImprovements    []OptimizationLongTermItem  `json:"long_term_improvements"`
}

// OptimizationLatestResult is the DTO for GET /api/resumes/{id}/optimization/latest.
type OptimizationLatestResult struct {
	HasResult             bool        `json:"has_result"`
	OptimizationID        uint64      `json:"optimization_id,omitempty"`
	ResumeID              uint64      `json:"resume_id,omitempty"`
	JobID                 uint64      `json:"job_id,omitempty"`
	OptimizationMode      string      `json:"optimization_mode,omitempty"`
	TargetDirection       string      `json:"target_direction,omitempty"`
	GenerationTier        string      `json:"generation_tier,omitempty"`
	GeneratedBy           string      `json:"generated_by,omitempty"`
	Status                string      `json:"status,omitempty"`
	ProfileType           string      `json:"profile_type,omitempty"`
	ProfileDominantProfile string      `json:"profile_dominant_profile,omitempty"`
	CreatedAt             string      `json:"created_at,omitempty"`
	Result                interface{} `json:"result,omitempty"`
}
