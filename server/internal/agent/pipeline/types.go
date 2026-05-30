package pipeline

import (
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// TargetOutputType identifies the business output type for the pipeline run.
type TargetOutputType string

const (
	// TargetMatchAnalysis indicates a match analysis output.
	TargetMatchAnalysis TargetOutputType = "match_analysis"
	// TargetRecommendation indicates a recommendation output.
	TargetRecommendation TargetOutputType = "recommendation"
	// TargetResumeOptimization indicates a resume optimization output.
	TargetResumeOptimization TargetOutputType = "resume_optimization"
)

// validTargetOutputTypes contains all allowed TargetOutputType values.
var validTargetOutputTypes = map[TargetOutputType]bool{
	TargetMatchAnalysis:      true,
	TargetRecommendation:     true,
	TargetResumeOptimization: true,
}

// AgentStatus describes the outcome of a pipeline run.
type AgentStatus string

const (
	// StatusSuccess means the LLM output passed parse + validate on the first try.
	StatusSuccess AgentStatus = "success"
	// StatusRepaired means the initial output failed but JSONRepairAgent fixed it.
	StatusRepaired AgentStatus = "repaired"
	// StatusFallbackRequired means repair failed and the caller should fall back to Mock.
	StatusFallbackRequired AgentStatus = "fallback_required"
)

// RunRequest is the input to a Pipeline.Run call.
type RunRequest struct {
	// AgentName is a diagnostic label for the calling business agent.
	AgentName string
	// PromptName identifies which Prompt template to load.
	PromptName prompt.PromptName
	// SchemaName identifies which Schema to validate against.
	SchemaName schema.SchemaName
	// TargetOutputType identifies the business output type.
	TargetOutputType TargetOutputType
	// PromptValues contains placeholder key-value pairs for Prompt rendering.
	PromptValues map[string]string
	// Metadata carries optional diagnostic key-value pairs.
	// Must not contain API keys.
	Metadata map[string]string
}

// RunResult is the output of a Pipeline.Run call.
type RunResult struct {
	// AgentStatus is one of success / repaired / fallback_required.
	AgentStatus AgentStatus `json:"agent_status"`
	// GeneratedBy is "llm_agent" on success/repaired, or "mock_agent" on
	// fallback_required.
	GeneratedBy string `json:"generated_by"`
	// Parsed is the validated JSON object when AgentStatus is success or repaired.
	Parsed map[string]any `json:"parsed,omitempty"`
	// JSONText is the extracted JSON string when AgentStatus is success or repaired.
	JSONText string `json:"json_text,omitempty"`
	// RawContent is the LLM raw output. Kept for internal diagnostics only;
	// do not log or expose full content.
	RawContent string `json:"-"`
	// RawOutputLength is the byte length of RawContent.
	RawOutputLength int `json:"raw_output_length"`
	// RepairApplied is true when JSONRepairAgent was invoked.
	RepairApplied bool `json:"repair_applied"`
	// RepairSuccess is true when JSONRepairAgent produced a valid result.
	RepairSuccess bool `json:"repair_success"`
	// FallbackRequired is true when the caller should invoke a Mock Agent.
	FallbackRequired bool `json:"fallback_required"`
	// FallbackReason describes why fallback is required (e.g. "llm_call_failed",
	// "json_parse_failed").
	FallbackReason string `json:"fallback_reason,omitempty"`
	// ErrorCode is the pipeline-level error code when the result is not success.
	ErrorCode string `json:"error_code,omitempty"`
	// ErrorMessage is a concise summary when the result is not success.
	ErrorMessage string `json:"error_message,omitempty"`
	// LLMProvider identifies the backend that was used (e.g. "hunyuan").
	LLMProvider string `json:"llm_provider,omitempty"`
	// LLMModel is the actual model name used.
	LLMModel string `json:"llm_model,omitempty"`
	// FinishReason is the provider's stop reason (e.g. "stop", "length").
	FinishReason string `json:"finish_reason,omitempty"`
	// DurationMs is the wall-clock time of the initial LLM call in milliseconds.
	DurationMs int64 `json:"duration_ms"`
	// ValidationErrors lists Schema validation failures when the output
	// passes JSON parse but fails Schema validate.
	ValidationErrors []schema.ValidationError `json:"validation_errors,omitempty"`
}

// Pipeline error codes.
const (
	ErrPipelineRequestInvalid    = "PIPELINE_REQUEST_INVALID"
	ErrPromptLoadFailed          = "PROMPT_LOAD_FAILED"
	ErrPromptRenderFailed        = "PROMPT_RENDER_FAILED"
	ErrSchemaLoadFailed          = "SCHEMA_LOAD_FAILED"
	ErrLLMDisabled               = "LLM_DISABLED"
	ErrLLMConfigMissing          = "LLM_CONFIG_MISSING"
	ErrLLMAPIKeyMissing          = "LLM_API_KEY_MISSING"
	ErrLLMTimeout                = "LLM_TIMEOUT"
	ErrLLMHTTPError              = "LLM_HTTP_ERROR"
	ErrLLMRateLimit              = "LLM_RATE_LIMIT"
	ErrLLMEmptyResponse          = "LLM_EMPTY_RESPONSE"
	ErrLLMOutputTruncated        = "LLM_OUTPUT_TRUNCATED"
	ErrJSONParseFailed           = "JSON_PARSE_FAILED"
	ErrJSONTopLevelNotObject     = "JSON_TOP_LEVEL_NOT_OBJECT"
	ErrSchemaValidationFailed    = "SCHEMA_VALIDATION_FAILED"
	ErrJSONRepairFailed          = "JSON_REPAIR_FAILED"
	ErrJSONRepairParseFailed     = "JSON_REPAIR_PARSE_FAILED"
	ErrJSONRepairSchemaFailed    = "JSON_REPAIR_SCHEMA_FAILED"
	ErrJSONRepairLLMFailed       = "JSON_REPAIR_LLM_FAILED"
	ErrTargetOutputTypeInvalid   = "TARGET_OUTPUT_TYPE_INVALID"
	ErrUnknown                   = "UNKNOWN_ERROR"
)

// Fallback reason values.
const (
	FallbackPromptLoadFailed       = "prompt_load_failed"
	FallbackSchemaLoadFailed       = "schema_load_failed"
	FallbackLLMCallFailed          = "llm_call_failed"
	FallbackLLMDisabled            = "llm_disabled"
	FallbackLLMConfigMissing       = "llm_config_missing"
	FallbackLLMAPIKeyMissing       = "llm_api_key_missing"
	FallbackJSONParseFailed        = "json_parse_failed"
	FallbackSchemaValidationFailed = "schema_validation_failed"
	FallbackJSONRepairFailed       = "json_repair_failed"
	FallbackJSONRepairParseFailed  = "json_repair_parse_failed"
	FallbackJSONRepairSchemaFailed = "json_repair_schema_failed"
	FallbackJSONRepairLLMFailed    = "json_repair_llm_failed"
	FallbackTargetOutputTypeInvalid = "target_output_type_invalid"
	FallbackRequestInvalid         = "request_invalid"
	FallbackUnknownError           = "unknown_error"
)

// mapLLMErrorCodeToPipeline maps an llm.ErrorCode to a pipeline error code.
func mapLLMErrorCodeToPipeline(code string) string {
	switch code {
	case "LLM_DISABLED":
		return ErrLLMDisabled
	case "LLM_CONFIG_MISSING":
		return ErrLLMConfigMissing
	case "LLM_API_KEY_MISSING":
		return ErrLLMAPIKeyMissing
	case "LLM_TIMEOUT":
		return ErrLLMTimeout
	case "LLM_HTTP_ERROR":
		return ErrLLMHTTPError
	case "LLM_RATE_LIMIT":
		return ErrLLMRateLimit
	case "LLM_EMPTY_RESPONSE":
		return ErrLLMEmptyResponse
	case "LLM_OUTPUT_TRUNCATED":
		return ErrLLMOutputTruncated
	default:
		return ErrUnknown
	}
}

// mapLLMErrorCodeToFallbackReason maps an llm.ErrorCode to a fallback reason.
func mapLLMErrorCodeToFallbackReason(code string) string {
	switch code {
	case "LLM_DISABLED":
		return FallbackLLMDisabled
	case "LLM_CONFIG_MISSING":
		return FallbackLLMConfigMissing
	case "LLM_API_KEY_MISSING":
		return FallbackLLMAPIKeyMissing
	case "LLM_TIMEOUT", "LLM_HTTP_ERROR", "LLM_RATE_LIMIT":
		return FallbackLLMCallFailed
	case "LLM_EMPTY_RESPONSE":
		return FallbackLLMCallFailed
	case "LLM_OUTPUT_TRUNCATED":
		return FallbackLLMCallFailed
	default:
		return FallbackUnknownError
	}
}
