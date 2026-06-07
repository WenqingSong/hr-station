package model

import "time"

// ResumeCompareResult represents a saved resume comparison snapshot.
// Maps to the `resume_compare_results` table in MySQL.
type ResumeCompareResult struct {
	ID            uint64
	UserID        uint64
	FromResumeID  uint64
	ToResumeID    uint64
	ResultJSON    []byte
	CreatedAt     time.Time
}

// VersionCompareResult is an alias (backward compatibility).
// Deprecated: use ResumeCompareResult.
type VersionCompareResult = ResumeCompareResult

// FieldChange represents a change in a specific resume field between resumes.
type FieldChange struct {
	Field      string `json:"field"`
	Before     string `json:"before"`
	After      string `json:"after"`
	ChangeType string `json:"change_type"`
	Impact     string `json:"impact"`
	ModuleName string `json:"module_name,omitempty"`
}

// RiskChange represents a change in risk level between resumes.
type RiskChange struct {
	Risk   string `json:"risk"`
	Before string `json:"before"`
	After  string `json:"after"`
	Trend  string `json:"trend"`
}

// SuggestionCompletion represents the completion status of a suggestion.
type SuggestionCompletion struct {
	Suggestion string `json:"suggestion"`
	Status     string `json:"status"`
	Evidence   string `json:"evidence"`
}

// VersionCompareAnalysis represents the structured resume comparison output.
type VersionCompareAnalysis struct {
	FromResumeID        uint64                 `json:"from_resume_id"`
	ToResumeID          uint64                 `json:"to_resume_id"`
	ScoreBefore         int                    `json:"score_before"`
	ScoreAfter          int                    `json:"score_after"`
	ScoreChange         int                    `json:"score_change"`
	FieldChanges        []FieldChange          `json:"field_changes"`
	RiskChanges         []RiskChange           `json:"risk_changes"`
	SuggestionCompletion []SuggestionCompletion `json:"suggestion_completion"`
	KeyImprovements     []string               `json:"key_improvements"`
	RemainingRisks      []string               `json:"remaining_risks"`
	Summary             string                 `json:"summary"`
	AuthenticityNotice  string                 `json:"authenticity_notice"`
	OverallChange       string                 `json:"overall_change,omitempty"`
	Recommendation      string                 `json:"recommendation,omitempty"`
	NextActions         []string               `json:"next_actions,omitempty"`
	ImprovedItems       []CompareChangeItem    `json:"improved_items,omitempty"`
	WeakenedItems       []CompareChangeItem    `json:"weakened_items,omitempty"`
	MissingItems        []string               `json:"missing_items,omitempty"`
	GeneratedBy         string                 `json:"generated_by,omitempty"`
	AgentStatus         string                 `json:"agent_status,omitempty"`
}

// CompareChangeItem represents a change in an improved or weakened item.
type CompareChangeItem struct {
	Field  string `json:"field"`
	Change string `json:"change"`
}

// ── K9-4-C: V2 Compare Types ──────────────────────────────────────────────

// CompareRequestV2 is the request for POST /api/resumes/compare.
type CompareRequestV2 struct {
	FromResumeID   uint64 `json:"from_resume_id"`
	ToResumeID     uint64 `json:"to_resume_id"`
	ComparisonMode string `json:"comparison_mode,omitempty"` // general / target_direction / target_job
	GenerationTier string `json:"generation_tier,omitempty"` // fast / quality
	TargetDirection string `json:"target_direction,omitempty"`
	JobID           uint64 `json:"job_id,omitempty"`
}

// CompareResponseV2 is the response from POST /api/resumes/compare.
type CompareResponseV2 struct {
	CompareID       uint64 `json:"compare_id"`
	FromResumeID    uint64 `json:"from_resume_id"`
	ToResumeID      uint64 `json:"to_resume_id"`
	ComparisonMode  string `json:"comparison_mode"`
	TargetDirection string `json:"target_direction,omitempty"`
	JobID           *uint64 `json:"job_id"`
	GenerationTier  string `json:"generation_tier"`
	GeneratedBy     string `json:"generated_by"`
	Status          string `json:"status"`

	FromProfileType            string `json:"from_profile_type,omitempty"`
	FromProfileDominantProfile string `json:"from_profile_dominant_profile,omitempty"`
	ToProfileType              string `json:"to_profile_type,omitempty"`
	ToProfileDominantProfile   string `json:"to_profile_dominant_profile,omitempty"`

	CompareScore     int      `json:"compare_score"`
	CompareLevel     string   `json:"compare_level"`
	WinnerResumeID   uint64   `json:"winner_resume_id"`
	WinnerLabel      string   `json:"winner_label"`
	OneSentenceSummary string `json:"one_sentence_summary"`

	FromResumeSummary string `json:"from_resume_summary,omitempty"`
	ToResumeSummary   string `json:"to_resume_summary,omitempty"`

	KeyDifferences          []string              `json:"key_differences"`
	FromStrengths           []string              `json:"from_strengths"`
	ToStrengths             []string              `json:"to_strengths"`
	FromRisks               []string              `json:"from_risks"`
	ToRisks                 []string              `json:"to_risks"`
	KeepFromItems           []string              `json:"keep_from_items"`
	KeepToItems             []string              `json:"keep_to_items"`
	MergeSuggestions        []string              `json:"merge_suggestions"`
	RewriteExamples         []RewriteCompareExample `json:"rewrite_examples"`
	NextVersionSuggestions  []string              `json:"next_version_suggestions"`
	DeliveryAdvice          string                `json:"delivery_advice"`

	CreatedAt string `json:"created_at"`
}

// RewriteCompareExample is a before/after rewrite example for compare.
type RewriteCompareExample struct {
	Section        string `json:"section"`
	FromVersion    string `json:"from_version"`
	ToVersion      string `json:"to_version"`
	SuggestedNext  string `json:"suggested_next"`
	Reason         string `json:"reason"`
}

// ResumeCompareResultV2 is the DB model with full metadata for persistence.
type ResumeCompareResultV2 struct {
	ID                         uint64
	UserID                     uint64
	FromResumeID               uint64
	ToResumeID                 uint64
	ComparisonMode             string
	GenerationTier             string
	GeneratedBy                string
	Status                     string
	TargetDirection            string
	JobID                      *uint64
	FromProfileType            string
	FromProfileDominantProfile string
	ToProfileType              string
	ToProfileDominantProfile   string
	ResultJSON                 []byte
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

// ── K9-4-D: Latest types ──────────────────────────────────────────────────

// CompareLatestResponseV2 is the DTO for GET /api/resumes/compare/latest.
type CompareLatestResponseV2 struct {
	HasResult                  bool        `json:"has_result"`
	CompareID                  uint64      `json:"compare_id,omitempty"`
	FromResumeID               uint64      `json:"from_resume_id,omitempty"`
	ToResumeID                 uint64      `json:"to_resume_id,omitempty"`
	ComparisonMode             string      `json:"comparison_mode,omitempty"`
	TargetDirection            string      `json:"target_direction,omitempty"`
	JobID                      *uint64     `json:"job_id,omitempty"`
	GenerationTier             string      `json:"generation_tier,omitempty"`
	GeneratedBy                string      `json:"generated_by,omitempty"`
	Status                     string      `json:"status,omitempty"`
	FromProfileType            string      `json:"from_profile_type,omitempty"`
	FromProfileDominantProfile string      `json:"from_profile_dominant_profile,omitempty"`
	ToProfileType              string      `json:"to_profile_type,omitempty"`
	ToProfileDominantProfile   string      `json:"to_profile_dominant_profile,omitempty"`
	CreatedAt                  string      `json:"created_at,omitempty"`
	Result                     interface{} `json:"result,omitempty"`
}
