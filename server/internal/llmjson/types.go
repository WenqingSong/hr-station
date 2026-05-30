package llmjson

// ErrorCode identifies the category of failure when extracting or parsing
// LLM raw output. Pipeline consumers use these codes to decide whether to
// enter JSON repair or fall back to Mock.
type ErrorCode string

const (
	// ErrorJSONEmpty is returned when the raw output is empty or only whitespace.
	ErrorJSONEmpty ErrorCode = "JSON_EMPTY"
	// ErrorJSONObjectNotFound is returned when no JSON object can be located
	// in the raw output.
	ErrorJSONObjectNotFound ErrorCode = "JSON_OBJECT_NOT_FOUND"
	// ErrorJSONParseFailed is returned when the extracted string cannot be
	// parsed as valid JSON.
	ErrorJSONParseFailed ErrorCode = "JSON_PARSE_FAILED"
	// ErrorJSONTopLevelNotObject is returned when the top-level JSON value is
	// not an object (e.g. an array).
	ErrorJSONTopLevelNotObject ErrorCode = "JSON_TOP_LEVEL_NOT_OBJECT"
)

// ParseResult holds the outcome of extracting and parsing JSON from LLM raw
// output. When Success is true, Parsed contains the resulting object.
// When Success is false, ErrorCode and ErrorMessage describe the failure.
type ParseResult struct {
	// RawOutput is the original model output (kept for internal use only;
	// do not log or expose full content).
	RawOutput string
	// ExtractedJSON is the JSON string that was extracted from RawOutput.
	ExtractedJSON string
	// Parsed is the unmarshalled object when parsing succeeds.
	Parsed map[string]any
	// Success indicates whether extraction and parsing both succeeded.
	Success bool
	// ErrorCode categorises the failure when Success is false.
	ErrorCode ErrorCode
	// ErrorMessage is a human-readable description of the failure.
	ErrorMessage string
	// RawOutputLength is the length of the original raw output in bytes.
	RawOutputLength int
}
