package model

import (
	"encoding/json"
	"time"
)

// ─── SEP Semantic Profile Constants ────────────────────────────────────────

const (
	SemanticTargetResume = "resume"
	SemanticTargetJob    = "job"

	SemanticProfileStatusPending    = "pending"
	SemanticProfileStatusProcessing = "processing"
	SemanticProfileStatusVerified   = "verified"
	SemanticProfileStatusFailed     = "failed"
	SemanticProfileStatusStale      = "stale"

	SemanticProfileVersionV1 = "sep_v1"
	SemanticSchemaVersionV1  = "sep_schema_v1"
	SemanticPromptVersionV1  = "sep_prompt_v1"
	SemanticPromptProfileResumeV1 = "resume_profiler_v1"
	SemanticPromptProfileJobV1    = "job_profiler_v1"
	SemanticPromptProfileGenericV1 = "profiler_generic_v1"
)

// ─── Tag Category Constants ───────────────────────────────────────────────

const (
	SemanticTagRoleDirection         = "role_direction"
	SemanticTagTechnicalSkill        = "technical_skill"
	SemanticTagEngineeringCapability = "engineering_capability"
	SemanticTagDomainContext         = "domain_context"
	SemanticTagExperienceSignal      = "experience_signal"
	SemanticTagRiskSignal            = "risk_signal"
	SemanticTagPreferenceSignal      = "preference_signal"
)

// ─── SemanticProfile ──────────────────────────────────────────────────────

// SemanticProfile represents a single SEP semantic profile for either a resume or a job.
type SemanticProfile struct {
	ID              int64     `json:"id"`
	TargetType      string    `json:"target_type"`
	TargetID        int64     `json:"target_id"`
	ProfileVersion  string    `json:"profile_version"`
	PromptVersion   string    `json:"prompt_version"`
	SchemaVersion   string    `json:"schema_version"`
	ContentHash     string    `json:"content_hash"`
	ModelProvider   string    `json:"model_provider"`
	ModelName       string    `json:"model_name"`
	ReasoningEffort string    `json:"reasoning_effort"`
	SummarySentence string    `json:"summary_sentence,omitempty"`
	StrengthsJSON   string    `json:"strengths_json,omitempty"`
	RisksJSON       string    `json:"risks_json,omitempty"`
	FitDirectionsJSON string  `json:"fit_directions_json,omitempty"`
	ProfileJSON     string    `json:"profile_json,omitempty"`
	Status          string    `json:"status"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ─── SemanticProfileTag ───────────────────────────────────────────────────

// ─── SemanticProfile LLM Output Types ─────────────────────────────────────

// SemanticProfileLLMOutput is the unified output schema for resume and job profilers.
type SemanticProfileLLMOutput struct {
	SummarySentence string                    `json:"summary_sentence"`
	Tags            []SemanticProfileLLMTag   `json:"tags"`
	Strengths       []string                  `json:"strengths"`
	Risks           json.RawMessage           `json:"risks"`
	FitDirections   []string                  `json:"fit_directions"`
	ProfileNotes    string                    `json:"profile_notes,omitempty"`
}

// SemanticProfileLLMTag represents a single tag in the LLM output.
type SemanticProfileLLMTag struct {
	TagCode       string  `json:"tag_code"`
	TagLabel      string  `json:"tag_label"`
	Category      string  `json:"category"`
	Confidence    float64 `json:"confidence"`
	EvidenceText  string  `json:"evidence_text"`
	Explanation   string  `json:"explanation"`
	SourceSection string  `json:"source_section"`
	IsUserVisible bool    `json:"is_user_visible"`
	Weight        float64 `json:"weight"`
}

// SemanticProfileLLMRisk represents a risk signal in the LLM output.
type SemanticProfileLLMRisk struct {
	Label        string  `json:"label"`
	Severity     string  `json:"severity"`
	EvidenceText string  `json:"evidence_text"`
	Explanation  string  `json:"explanation"`
	Confidence   float64 `json:"confidence"`
}

// Defensive raw code blacklist — these must never appear as tag_label or display text.
var semanticRawCodeBlacklist = map[string]bool{
	"research": true, "algorithm": true, "backend": true, "frontend": true,
	"redis": true, "golang": true, "python": true, "java": true, "cpp": true,
	"mysql": true, "linux": true, "docker_k8s": true, "pytorch_tf": true,
	"degree_bachelor": true, "degree_master": true,
	"internship_exp": true, "project_exp": true, "paper": true,
	"skill_0": true, "skill_1": true, "skill_2": true, "skill_3": true,
}

// IsBlacklistedRawCode returns true if the given string is a known raw code that must not appear as user-visible text.
func IsBlacklistedRawCode(s string) bool { return semanticRawCodeBlacklist[s] }

// SanitizeLLMOutput checks and cleans a SemanticProfileLLMOutput against the raw code blacklist.
func (o *SemanticProfileLLMOutput) SanitizeLLMOutput() {
	if IsBlacklistedRawCode(o.SummarySentence) {
		o.SummarySentence = ""
	}
	for i := range o.Strengths {
		if IsBlacklistedRawCode(o.Strengths[i]) { o.Strengths[i] = "" }
	}
	for i := range o.FitDirections {
		if IsBlacklistedRawCode(o.FitDirections[i]) { o.FitDirections[i] = "" }
	}
	for i := range o.Tags {
		if IsBlacklistedRawCode(o.Tags[i].TagLabel) {
			o.Tags[i].TagLabel = o.Tags[i].TagCode
		}
		o.Tags[i].IsUserVisible = true
		if o.Tags[i].Weight <= 0 { o.Tags[i].Weight = o.Tags[i].Confidence }
	}
}

// ToPersistenceJSON converts the output to JSON strings for DB storage.
func (o *SemanticProfileLLMOutput) ToPersistenceJSON(profile *SemanticProfile) {
	if b, err := json.Marshal(o.Strengths); err == nil { profile.StrengthsJSON = string(b) }
	if len(o.Risks) > 0 { profile.RisksJSON = string(o.Risks) }
	if b, err := json.Marshal(o.FitDirections); err == nil { profile.FitDirectionsJSON = string(b) }
	if b, err := json.Marshal(o); err == nil { profile.ProfileJSON = string(b) }
	profile.SummarySentence = o.SummarySentence
}

// SemanticProfileTag represents a single tag within a semantic profile, with evidence and confidence.
type SemanticProfileTag struct {
	ID            int64     `json:"id"`
	ProfileID     int64     `json:"profile_id"`
	TargetType    string    `json:"target_type"`
	TargetID      int64     `json:"target_id"`
	TagCode       string    `json:"tag_code"`
	TagLabel      string    `json:"tag_label"`
	Category      string    `json:"category"`
	Confidence    float64   `json:"confidence"`
	EvidenceText  string    `json:"evidence_text,omitempty"`
	Explanation   string    `json:"explanation,omitempty"`
	SourceSection string    `json:"source_section,omitempty"`
	IsUserVisible bool      `json:"is_user_visible"`
	Weight        float64   `json:"weight"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
