package model

import "time"

// VersionCompareResult represents a saved version comparison snapshot.
// Maps to the `version_compare_results` table in MySQL.
type VersionCompareResult struct {
	ID            uint64
	UserID        uint64
	FromVersionID uint64
	ToVersionID   uint64
	ResultJSON    []byte
	CreatedAt     time.Time
}

// FieldChange represents a change in a specific resume field between versions.
type FieldChange struct {
	Field      string `json:"field"`
	Before     string `json:"before"`
	After      string `json:"after"`
	ChangeType string `json:"change_type"`
	Impact     string `json:"impact"`
}

// RiskChange represents a change in risk level between versions.
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

// VersionCompareAnalysis represents the structured version comparison output.
type VersionCompareAnalysis struct {
	FromVersionNo        int                    `json:"from_version_no"`
	ToVersionNo          int                    `json:"to_version_no"`
	ScoreBefore          int                    `json:"score_before"`
	ScoreAfter           int                    `json:"score_after"`
	ScoreChange          int                    `json:"score_change"`
	FieldChanges         []FieldChange          `json:"field_changes"`
	RiskChanges          []RiskChange           `json:"risk_changes"`
	SuggestionCompletion []SuggestionCompletion `json:"suggestion_completion"`
	KeyImprovements      []string               `json:"key_improvements"`
	RemainingRisks       []string               `json:"remaining_risks"`
	Summary              string                 `json:"summary"`
	AuthenticityNotice   string                 `json:"authenticity_notice"`
}
