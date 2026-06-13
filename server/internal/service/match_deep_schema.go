// ─── H19-B: Deep Match Report Schema ─────────────────────────────────────
// 8-layer Agent architecture output schema

package service

import "offer-catcher/server/internal/model"

// ─── DeepMatchReport ─────────────────────────────────────────────────────

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
