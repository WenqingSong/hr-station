// ─── H19-B: Deep Match V3 Schema ──────────────────────────────────────────
// 8-layer Agent architecture output schema with score_alignment + quality_checks

package service

import "offer-catcher/server/internal/model"

// ─── DeepMatchReport (legacy v2, kept for backward compat) ────────────────

type DeepMatchReport struct {
	GenerationTier string `json:"generation_tier"`
	GeneratedBy    string `json:"generated_by"`
	AgentStatus    string `json:"agent_status"`

	MatchScore           int    `json:"match_score"`
	MatchLevel           string `json:"match_level"`
	ApplicationDecision  string `json:"application_decision"`
	DecisionLabel        string `json:"decision_label"`
	OneSentenceConclusion string `json:"one_sentence_conclusion"`

	ScoreBreakdown     ScoreBreakdownResult      `json:"score_breakdown"`
	JobUnderstanding   DeepJobUnderstanding      `json:"job_understanding"`
	RequirementCoverage []RequirementCoverageItem `json:"requirement_coverage"`
	ResumeEvidence     []ResumeEvidenceItem      `json:"resume_evidence"`
	CoreMatchPoints    []DeepCoreMatchPoint      `json:"core_match_points"`
	KeyRisks           []DeepKeyRisk             `json:"key_risks"`

	ResumeRewriteStrategy DeepResumeRewriteStrategy `json:"resume_rewrite_strategy"`
	ApplicationStrategy   DeepApplicationStrategy   `json:"application_strategy"`
	InterviewRiskQuestions []DeepInterviewQuestion    `json:"interview_risk_questions"`
	AgentTrace             []DeepAgentTrace          `json:"agent_trace"`

	ScoreDeltaExplanation *DeepScoreDeltaExplanation `json:"score_delta_explanation,omitempty"`
}

// ─── Deep Job Understanding ──────────────────────────────────────────────

type DeepJobUnderstanding struct {
	JobPositioning       string   `json:"job_positioning"`
	CoreResponsibilities []string `json:"core_responsibilities"`
	MustHaveRequirements []string `json:"must_have_requirements"`
	ImportantRequirements []string `json:"important_requirements"`
	BonusRequirements    []string `json:"bonus_requirements"`
	HiddenExpectations   []string `json:"hidden_expectations"`
}

// ─── Deep Core Match Point ───────────────────────────────────────────────

type DeepCoreMatchPoint struct {
	PointName          string   `json:"point_name"`
	MatchStrength      string   `json:"match_strength"` // strong/partial/transferable/weak
	WhyItMatches       string   `json:"why_it_matches"`
	SupportingEvidence []string `json:"supporting_evidence"`
	JobRequirement     string   `json:"job_requirement"`
}

// ─── Deep Key Risk ───────────────────────────────────────────────────────

type DeepKeyRisk struct {
	RiskName       string `json:"risk_name"`
	RiskLevel      string `json:"risk_level"` // high/medium/low
	WhyItMatters   string `json:"why_it_matters"`
	EvidenceGap    string `json:"evidence_gap"`
	FixSuggestion  string `json:"fix_suggestion"`
	InterviewImpact string `json:"interview_impact"`
}

// ─── Deep Resume Rewrite Strategy ────────────────────────────────────────

type DeepResumeRewriteStrategy struct {
	SkillsReorder         []string `json:"skills_reorder"`
	ProjectRewritePoints  []string `json:"project_rewrite_points"`
	InternshipRewritePoints []string `json:"internship_rewrite_points"`
	KeywordsToAdd         []string `json:"keywords_to_add"`
	ContentToWeaken       []string `json:"content_to_weaken"`
}

// ─── Deep Application Strategy ───────────────────────────────────────────

type DeepApplicationStrategy struct {
	BestTiming          string `json:"best_timing"`
	DeliveryStrategy    string `json:"delivery_strategy"`
	CompetingPositioning string `json:"competing_positioning"`
}

// ─── Deep Interview Question ─────────────────────────────────────────────

type DeepInterviewQuestion struct {
	Question       string `json:"question"`
	WhyAsked       string `json:"why_asked"`
	AnswerDirection string `json:"answer_direction"`
	RelatedRisk    string `json:"related_risk"`
}

// ─── Deep Agent Trace ────────────────────────────────────────────────────

type DeepAgentTrace struct {
	Stage   string `json:"stage"`
	Agent   string `json:"agent"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

// ─── Deep Context ────────────────────────────────────────────────────────

type DeepMatchContext struct {
	ResumeID    int64
	JobID       int64
	ResumeRaw   string
	JobRaw      string
	ResumeSEP   *MatchSEPContext
	JobSEP      *MatchSEPContext
	FastResult  *model.MatchAnalysisResult
}

// ─── Score Delta Explanation ─────────────────────────────────────────────

type DeepScoreDeltaExplanation struct {
	FastScore int    `json:"fast_score"`
	DeepScore int    `json:"deep_score"`
	Delta     int    `json:"delta"`
	Reason    string `json:"reason"`
}

// ─── H19-B: Deep Match V3 Output Types ────────────────────────────────────

// DeepMatchV3Report is the full deep_match_v3 output.
type DeepMatchV3Report struct {
	GenerationTier  string `json:"generation_tier"`
	GeneratedBy     string `json:"generated_by"`
	AgentStatus     string `json:"agent_status"`
	ReportVersion   string `json:"report_version"`

	MatchScore            int    `json:"match_score"`
	MatchLevel            string `json:"match_level"`
	ApplicationDecision   string `json:"application_decision"`
	DecisionLabel         string `json:"decision_label"`
	OneSentenceConclusion string `json:"one_sentence_conclusion"`

	ScoreBreakdown     ScoreBreakdownResultV3    `json:"score_breakdown"`
	ScoreAlignment     ScoreAlignmentResult      `json:"score_alignment"`
	JobUnderstanding   DeepJobUnderstanding      `json:"job_understanding"`
	RequirementCoverage []RequirementCoverageItem `json:"requirement_coverage"`
	ResumeEvidenceMapping []ResumeEvidenceItem    `json:"resume_evidence_mapping"`
	CoreMatchPoints    []DeepCoreMatchPoint      `json:"core_match_points"`
	KeyRisks           []DeepKeyRisk             `json:"key_risks"`

	ResumeRewriteStrategy DeepResumeRewriteStrategy `json:"resume_rewrite_strategy"`
	ApplicationStrategy   DeepApplicationStrategy   `json:"application_strategy"`
	InterviewRiskQuestions []DeepInterviewQuestion    `json:"interview_risk_questions"`
	ActionPlan             []map[string]any          `json:"action_plan"`
	AgentTrace             []DeepAgentTrace          `json:"agent_trace"`

	QualityChecks QualityChecksResult `json:"quality_checks"`
}

// ScoreBreakdownResultV3 extends ScoreBreakdownResult with jd_competitiveness.
type ScoreBreakdownResultV3 struct {
	DirectionFit        int `json:"direction_fit"`
	MustHaveCoverage    int `json:"must_have_coverage"`
	EvidenceStrength    int `json:"evidence_strength"`
	ExperienceRelevance int `json:"experience_relevance"`
	RiskControl         int `json:"risk_control"`
	JDCompetitiveness   int `json:"jd_competitiveness"`
}

// ScoreAlignmentResult compares Fast vs Deep scores.
type ScoreAlignmentResult struct {
	FastScore    int    `json:"fast_score"`
	DeepScore    int    `json:"deep_score"`
	ScoreDelta   int    `json:"score_delta"`
	IsConsistent bool   `json:"is_consistent"`
	DeltaReason  string `json:"delta_reason"`
}

// QualityChecksResult provides field-level completeness audit.
type QualityChecksResult struct {
	HasJDUnderstanding      bool `json:"has_jd_understanding"`
	HasResumeEvidence       bool `json:"has_resume_evidence"`
	HasRequirementCoverage  bool `json:"has_requirement_coverage"`
	HasStructuredRisks      bool `json:"has_structured_risks"`
	HasRewriteStrategy      bool `json:"has_rewrite_strategy"`
	HasInterviewQuestions   bool `json:"has_interview_questions"`
	HasScoreAlignment       bool `json:"has_score_alignment"`
	NoNullArrays            bool `json:"no_null_arrays"`
	NoRawTagCode            bool `json:"no_raw_tag_code"`
}

// DeepMatchContextV3 holds all inputs for the v3 pipeline.
type DeepMatchContextV3 struct {
	ResumeID int64
	JobID    int64

	ResumeRaw string
	JobRaw    string

	ResumeSEPTags []*model.SemanticProfileTag
	JobSEPTags    []*model.SemanticProfileTag

	FastResult          *model.MatchAnalysisResult
	FastScoreBreakdown  ScoreBreakdownResult
	FastHardGate        HardGateResult
	FastRequirementCoverage []RequirementCoverageItem
	FastResumeEvidence  []ResumeEvidenceItem
	FastStructuredRisks []MatchRiskPoint

	Status string
	Reason string
}

// ─── JM-R3: Screening Match V1 Types ──────────────────────────────────────
// Replacement for DeepMatchV3Report. 5-module Hy3 + Judge output.

// ScreeningMatchV1 is the top-level output of the new Deep pipeline.
type ScreeningMatchV1 struct {
	Mode                      string                      `json:"mode"`
	Version                   string                      `json:"version"`
	ScreeningDecision         ScreeningDecision            `json:"screening_decision"`
	ScreeningPassPoints       []ScreeningPassPoint          `json:"screening_pass_points"`
	ScreeningEliminationRisks []ScreeningEliminationRisk     `json:"screening_elimination_risks"`
	ResumeRewriteCommands     []ResumeRewriteCommand         `json:"resume_rewrite_commands"`
	JobStrategy               JobStrategy                   `json:"job_strategy"`
}

type ScreeningDecision struct {
	Score           int      `json:"score"`
	Level           string   `json:"level"`
	Decision        string   `json:"decision"`
	DecisionLabel   string   `json:"decision_label"`
	Summary         string   `json:"summary"`
	Confidence      string   `json:"confidence"`
	DecisionReasons []string `json:"decision_reasons"`
}

type ScreeningPassPoint struct {
	Title          string `json:"title"`
	WhyItHelps     string `json:"why_it_helps"`
	ResumeEvidence string `json:"resume_evidence"`
	ScreeningValue string `json:"screening_value"`
}

type ScreeningEliminationRisk struct {
	RiskTitle            string `json:"risk_title"`
	Severity             string `json:"severity"`
	WhyItBlocksScreening string `json:"why_it_blocks_screening"`
	CurrentResumeProblem string `json:"current_resume_problem"`
	FixDirection         string `json:"fix_direction"`
}

type ResumeRewriteCommand struct {
	TargetSection          string `json:"target_section"`
	Problem                string `json:"problem"`
	RewriteInstruction     string `json:"rewrite_instruction"`
	BeforeExample          string `json:"before_example"`
	AfterExample           string `json:"after_example"`
	WhyItImprovesScreening string `json:"why_it_improves_screening"`
}

type JobStrategy struct {
	Positioning         string   `json:"positioning"`
	RecommendedAction   string   `json:"recommended_action"`
	IfApplyNow          string   `json:"if_apply_now"`
	BetterFitDirections []string `json:"better_fit_directions"`
	AvoidDirections     []string `json:"avoid_directions"`
}

// ScreeningMatchContext holds all inputs for the 5-module Hy3 pipeline.
type ScreeningMatchContext struct {
	ResumeID          int64
	JobID             int64
	ResumeName        string
	ResumeRawText     string
	ResumeSEPProfile  any
	JobTitle          string
	JobDirection      string
	JobCategory       string
	JobCity           string
	JobRawJD          string
	JobSEPProfile     any
	FastResult        *model.MatchAnalysisResult
}
