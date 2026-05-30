package llmjson

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ExtractObject takes a raw LLM output string and returns the JSON object
// text extracted from it. It handles Markdown code fences, surrounding
// explanatory text, and BOM/whitespace cleanup.
//
// It does NOT repair malformed JSON — that is the job of JSONRepairAgent.
func ExtractObject(raw string) (string, error) {
	// Trim BOM and surrounding whitespace.
	s := trimBOM(raw)
	s = strings.TrimSpace(s)

	if s == "" {
		return "", fmt.Errorf("llmjson: %s: raw output is empty", ErrorJSONEmpty)
	}

	// Try multiple extraction strategies, returning the first that yields
	// valid JSON. If we find something that looks like JSON but is invalid,
	// we return JSON_PARSE_FAILED so the pipeline knows repair is viable.

	// Strategy 1: try to parse the whole string as JSON.
	if json.Valid([]byte(s)) {
		return s, nil
	}

	// Strategy 2: try extracting from a Markdown code fence.
	if extracted, ok := extractFromCodeFence(s); ok {
		if json.Valid([]byte(extracted)) {
			return extracted, nil
		}
		// Found a code fence with invalid JSON.
		return "", fmt.Errorf("llmjson: %s: extracted content from code fence is not valid JSON", ErrorJSONParseFailed)
	}

	// Strategy 3: find the outermost JSON object.
	if extracted, ok := extractBraceObject(s); ok {
		if json.Valid([]byte(extracted)) {
			return extracted, nil
		}
		// Found a brace-delimited object but it's not valid JSON.
		return "", fmt.Errorf("llmjson: %s: found JSON-like object but failed to parse", ErrorJSONParseFailed)
	}

	return "", fmt.Errorf("llmjson: %s: no JSON object found in raw output (length=%d)",
		ErrorJSONObjectNotFound, len(s))
}

// ExtractBestCandidate returns the most likely JSON object string from raw
// output, even if it is not valid JSON. This is useful for the repair
// pipeline which needs the raw candidate to send to JSONRepairAgent.
// Returns ("", JSON_EMPTY) if the input is empty,
// ("", JSON_OBJECT_NOT_FOUND) if no candidate is found,
// (candidate, nil) if a candidate is found (regardless of validity).
func ExtractBestCandidate(raw string) (string, error) {
	s := trimBOM(raw)
	s = strings.TrimSpace(s)

	if s == "" {
		return "", fmt.Errorf("llmjson: %s: raw output is empty", ErrorJSONEmpty)
	}

	// Strategy 1: whole string.
	if strings.HasPrefix(s, "{") {
		return s, nil
	}

	// Strategy 2: code fence.
	if extracted, ok := extractFromCodeFence(s); ok {
		return extracted, nil
	}

	// Strategy 3: brace object.
	if extracted, ok := extractBraceObject(s); ok {
		return extracted, nil
	}

	return "", fmt.Errorf("llmjson: %s: no JSON object found in raw output (length=%d)",
		ErrorJSONObjectNotFound, len(s))
}

// trimBOM removes a UTF-8 BOM prefix if present.
func trimBOM(s string) string {
	if strings.HasPrefix(s, "\xEF\xBB\xBF") {
		return strings.TrimPrefix(s, "\xEF\xBB\xBF")
	}
	return s
}

// codeFenceRe matches ```json ... ``` or ``` ... ``` code fences.
var codeFenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*\\n?(.*?)\\n?```")

// extractFromCodeFence attempts to extract content from a Markdown code fence.
func extractFromCodeFence(s string) (string, bool) {
	matches := codeFenceRe.FindStringSubmatch(s)
	if len(matches) < 2 {
		return "", false
	}
	content := strings.TrimSpace(matches[1])
	if content == "" {
		return "", false
	}
	return content, true
}

// extractBraceObject finds the outermost { ... } pair in the text.
// It uses brace counting to handle nested objects.
func extractBraceObject(s string) (string, bool) {
	firstBrace := strings.IndexByte(s, '{')
	if firstBrace == -1 {
		return "", false
	}

	depth := 0
	inString := false
	escape := false

	for i := firstBrace; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if escape {
			escape = false
			i += size
			continue
		}
		switch r {
		case '\\':
			if inString {
				escape = true
			}
		case '"':
			inString = !inString
		case '{':
			if !inString {
				depth++
			}
		case '}':
			if !inString {
				depth--
				if depth == 0 {
					return s[firstBrace : i+size], true
				}
			}
		}
		i += size
	}

	return "", false
}
