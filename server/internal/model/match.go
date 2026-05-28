package model

import "time"

// MatchAnalysis represents a saved match analysis snapshot.
// Maps to the `match_analyses` table in MySQL.
type MatchAnalysis struct {
	ID              uint64
	UserID          uint64
	ResumeVersionID uint64
	JobID           uint64
	Score           int
	ResultJSON      []byte
	CreatedAt       time.Time
}

// MatchEvidence represents a single JD-resume evidence chain entry.
type MatchEvidence struct {
	Requirement string `json:"requirement"`
	ResumePoint string `json:"resume_point"`
	Judgement   string `json:"judgement"`
}

// MatchSuggestion represents an optimization suggestion with priority.
type MatchSuggestion struct {
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Reason   string `json:"reason"`
	Action   string `json:"action"`
}

// MatchAnalysisResult represents the structured match analysis output.
type MatchAnalysisResult struct {
	JobID                   uint64            `json:"job_id"`
	JobTitle                string            `json:"job_title"`
	ResumeVersionID         uint64            `json:"resume_version_id"`
	ResumeVersionNo         int               `json:"resume_version_no"`
	MatchScore              int               `json:"match_score"`
	ScreeningResult         string            `json:"screening_result"`
	EvidenceChain           []MatchEvidence   `json:"evidence_chain"`
	MatchHighlights         []string          `json:"match_highlights"`
	RiskPoints              []string          `json:"risk_points"`
	OptimizationSuggestions []MatchSuggestion `json:"optimization_suggestions"`
	NextSteps               []string          `json:"next_steps"`
	AuthenticityNotice      string            `json:"authenticity_notice"`
}
