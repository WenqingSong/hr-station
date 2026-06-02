package repair

import "offer-catcher/server/internal/schema"

// TargetOutputType identifies the business output type being repaired.
type TargetOutputType string

const (
	// TargetMatchAnalysis indicates a match analysis output.
	TargetMatchAnalysis TargetOutputType = "match_analysis"
	// TargetRecommendation indicates a recommendation output.
	TargetRecommendation TargetOutputType = "recommendation"
	// TargetResumeOptimization indicates a resume optimization output.
	TargetResumeOptimization TargetOutputType = "resume_optimization"
	TargetCompare            TargetOutputType = "version_compare"
	TargetResumeOptimizationCore TargetOutputType = "resume_optimization_core"
)

// ValidTargetOutputTypes contains all allowed TargetOutputType values.
var ValidTargetOutputTypes = map[TargetOutputType]bool{
	TargetMatchAnalysis:      true,
	TargetRecommendation:     true,
	TargetResumeOptimization:     true,
	TargetCompare:                true,
	TargetResumeOptimizationCore: true,
}

// RepairRequest is the input to a JSONRepairAgent.Repair call.
type RepairRequest struct {
	// TargetOutputType identifies which business output type is being repaired.
	TargetOutputType TargetOutputType
	// TargetSchemaJSON is the full JSON text of the target business Schema.
	TargetSchemaJSON string
	// BrokenOutput is the raw model output that failed parse or validate.
	BrokenOutput string
	// ValidationErrors is a summary of the parse/validate errors.
	ValidationErrors string
	// RepairContext is optional non-sensitive context (e.g. schema name,
	// allowed enum values). Must not contain API keys, full resumes, or
	// full JDs.
	RepairContext string
	// Metadata carries optional diagnostic key-value pairs. Must not
	// contain API keys.
	Metadata map[string]string
}

// RepairResult is the output of a JSONRepairAgent.Repair call.
type RepairResult struct {
	// Success indicates whether repair produced a valid, Schema-conforming
	// JSON object.
	Success bool `json:"success"`
	// RepairedJSON is the repaired JSON string when Success is true.
	RepairedJSON string `json:"repaired_json,omitempty"`
	// Parsed is the unmarshalled repaired object when Success is true.
	Parsed map[string]any `json:"parsed,omitempty"`
	// RawContent is the raw repair output from the LLM. Kept for internal
	// diagnostics only; do not log or expose full content.
	RawContent string `json:"-"`
	// ErrorCode is set when Success is false.
	ErrorCode string `json:"error_code,omitempty"`
	// ErrorMessage is a concise summary when Success is false.
	ErrorMessage string `json:"error_message,omitempty"`
	// RepairAttempted is always true after a Repair call (repair was
	// attempted exactly once).
	RepairAttempted bool `json:"repair_attempted"`
	// RepairSuccess indicates whether the repair attempt succeeded.
	RepairSuccess bool `json:"repair_success"`
	// RawOutputLength is the byte length of RawContent.
	RawOutputLength int `json:"raw_output_length"`
	// ValidationErrors lists Schema validation failures when the repair
	// output passes JSON parse but fails Schema validate.
	ValidationErrors []schema.ValidationError `json:"validation_errors,omitempty"`
}

// Repair error codes.
const (
	// ErrRepairFailed is a generic repair failure.
	ErrRepairFailed = "JSON_REPAIR_FAILED"
	// ErrRepairParseFailed indicates the repair output could not be
	// parsed as JSON.
	ErrRepairParseFailed = "JSON_REPAIR_PARSE_FAILED"
	// ErrRepairSchemaFailed indicates the repair output parsed as JSON
	// but did not conform to the target Schema.
	ErrRepairSchemaFailed = "JSON_REPAIR_SCHEMA_FAILED"
	// ErrRepairLLMFailed indicates the LLM call itself failed.
	ErrRepairLLMFailed = "JSON_REPAIR_LLM_FAILED"
	// ErrRepairPromptLoadFailed indicates the repair Prompt template
	// could not be loaded.
	ErrRepairPromptLoadFailed = "JSON_REPAIR_PROMPT_LOAD_FAILED"
	// ErrRepairSchemaLoadFailed indicates the target Schema could not
	// be loaded.
	ErrRepairSchemaLoadFailed = "JSON_REPAIR_SCHEMA_LOAD_FAILED"
	// ErrRepairTargetInvalid indicates the TargetOutputType is not one
	// of the allowed values.
	ErrRepairTargetInvalid = "JSON_REPAIR_TARGET_INVALID"
)
