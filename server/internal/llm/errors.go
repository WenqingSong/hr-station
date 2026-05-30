package llm

import (
	"fmt"
	"strings"
)

// ErrorCode categorises failures that occur during LLM configuration
// or invocation. The fallback decision layer uses these codes to
// determine whether to attempt JSON repair or fall back to Mock.
type ErrorCode string

const (
	// ErrLLMDisabled is returned when LLM is not enabled (LLM_ENABLE=false).
	ErrLLMDisabled ErrorCode = "LLM_DISABLED"
	// ErrLLMConfigMissing is returned when required LLM configuration
	// fields are empty.
	ErrLLMConfigMissing ErrorCode = "LLM_CONFIG_MISSING"
	// ErrLLMAPIKeyMissing is returned when the API key is empty.
	ErrLLMAPIKeyMissing ErrorCode = "LLM_API_KEY_MISSING"
	// ErrLLMTimeout is returned when the LLM request exceeds the
	// configured timeout.
	ErrLLMTimeout ErrorCode = "LLM_TIMEOUT"
	// ErrLLMHTTPError is returned for HTTP 5xx responses or network
	// errors.
	ErrLLMHTTPError ErrorCode = "LLM_HTTP_ERROR"
	// ErrLLMRateLimit is returned when the provider responds with
	// HTTP 429.
	ErrLLMRateLimit ErrorCode = "LLM_RATE_LIMIT"
	// ErrLLMEmptyResponse is returned when the provider returns an
	// empty content string.
	ErrLLMEmptyResponse ErrorCode = "LLM_EMPTY_RESPONSE"
	// ErrLLMOutputTruncated is returned when finish_reason is "length",
	// indicating the output was cut off.
	ErrLLMOutputTruncated ErrorCode = "LLM_OUTPUT_TRUNCATED"
	// ErrUnknown is returned for unclassifiable errors.
	ErrUnknown ErrorCode = "UNKNOWN_ERROR"
)

// LLMError is the structured error type returned by LLM client
// operations. It carries an ErrorCode for programmatic decision-making
// and an optional Cause for error-chain inspection.
//
// Invariant: LLMError must never contain an API key, a full Prompt,
// a full resume, or a full JD. The Message field should be a concise
// summary suitable for logging.
type LLMError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error implements the error interface.
func (e *LLMError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements the errors.Unwrap interface for error-chain
// inspection.
func (e *LLMError) Unwrap() error {
	return e.Cause
}

// NewLLMError creates a new LLMError with the given code and message.
func NewLLMError(code ErrorCode, message string) *LLMError {
	return &LLMError{Code: code, Message: message}
}

// WrapLLMError creates a new LLMError wrapping an existing error.
func WrapLLMError(code ErrorCode, message string, cause error) *LLMError {
	return &LLMError{Code: code, Message: message, Cause: cause}
}

// IsErrorCode checks whether an error is an LLMError with the
// specified ErrorCode. This is the recommended way for pipeline
// consumers to classify errors.
func IsErrorCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	llmErr, ok := err.(*LLMError)
	if !ok {
		return false
	}
	return llmErr.Code == code
}

// IsDirectMockFallback returns true for error codes that should
// trigger an immediate Mock fallback without attempting JSON repair.
func IsDirectMockFallback(err error) bool {
	if err == nil {
		return false
	}
	llmErr, ok := err.(*LLMError)
	if !ok {
		return true // unknown errors go directly to Mock
	}
	switch llmErr.Code {
	case ErrLLMDisabled, ErrLLMConfigMissing, ErrLLMAPIKeyMissing,
		ErrLLMTimeout, ErrLLMHTTPError, ErrLLMRateLimit,
		ErrLLMEmptyResponse, ErrUnknown:
		return true
	default:
		return false
	}
}

// maskAPIKey replaces all but the last 4 characters of an API key
// with asterisks. If the key is shorter than 8 characters, the
// entire key is masked.
func maskAPIKey(key string) string {
	if key == "" {
		return "(empty)"
	}
	if len(key) < 8 {
		return "****"
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}
