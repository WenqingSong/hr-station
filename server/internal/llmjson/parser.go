package llmjson

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseObject extracts a JSON object from raw LLM output and unmarshals it
// into a map[string]any. It rejects top-level arrays.
func ParseObject(raw string) (map[string]any, error) {
	extracted, err := ExtractObject(raw)
	if err != nil {
		// If extraction itself failed with a parse error, propagate it.
		// If extraction found nothing, return the original error.
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(extracted), &result); err != nil {
		// Check if the top-level value is an array.
		var arr []any
		if json.Unmarshal([]byte(extracted), &arr) == nil {
			return nil, fmt.Errorf("llmjson: %s: top-level JSON value is an array, expected object",
				ErrorJSONTopLevelNotObject)
		}
		return nil, fmt.Errorf("llmjson: %s: %w", ErrorJSONParseFailed, err)
	}

	return result, nil
}

// ExtractAndParse is a convenience function that extracts JSON from raw LLM
// output and returns a structured ParseResult. It combines extraction,
// parsing, and top-level object validation in one call.
func ExtractAndParse(raw string) *ParseResult {
	result := &ParseResult{
		RawOutput:       raw,
		RawOutputLength: len(raw),
	}

	// Check for empty input.
	s := trimBOM(raw)
	s = strings.TrimSpace(s)
	if s == "" {
		result.ErrorCode = ErrorJSONEmpty
		result.ErrorMessage = "raw output is empty"
		return result
	}

	// Extract JSON string.
	extracted, err := ExtractObject(raw)
	if err != nil {
		result.ErrorCode = errorCodeFromErr(err)
		result.ErrorMessage = err.Error()
		// Even if extraction failed, try to get the best candidate
		// so downstream repair can use it.
		if candidate, candErr := ExtractBestCandidate(raw); candErr == nil {
			result.ExtractedJSON = candidate
		}
		return result
	}
	result.ExtractedJSON = extracted

	// Parse.
	var parsed map[string]any
	if err := json.Unmarshal([]byte(extracted), &parsed); err != nil {
		// Check for top-level array.
		var arr []any
		if json.Unmarshal([]byte(extracted), &arr) == nil {
			result.ErrorCode = ErrorJSONTopLevelNotObject
			result.ErrorMessage = "top-level JSON value is an array, expected object"
			return result
		}
		result.ErrorCode = ErrorJSONParseFailed
		result.ErrorMessage = fmt.Sprintf("JSON parse failed: %s", err.Error())
		return result
	}

	result.Parsed = parsed
	result.Success = true
	return result
}

// errorCodeFromErr extracts an ErrorCode from an ExtractObject error.
func errorCodeFromErr(err error) ErrorCode {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case containsStr(msg, string(ErrorJSONEmpty)):
		return ErrorJSONEmpty
	case containsStr(msg, string(ErrorJSONParseFailed)):
		return ErrorJSONParseFailed
	case containsStr(msg, string(ErrorJSONObjectNotFound)):
		return ErrorJSONObjectNotFound
	default:
		return ErrorJSONObjectNotFound
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
