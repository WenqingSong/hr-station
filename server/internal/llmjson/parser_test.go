package llmjson

import (
	"strings"
	"testing"
)

func TestParseObject_PureJSON(t *testing.T) {
	input := `{"generated_by":"llm_agent","score":85}`
	got, err := ParseObject(input)
	if err != nil {
		t.Fatalf("ParseObject error: %v", err)
	}
	if got["generated_by"] != "llm_agent" {
		t.Errorf("generated_by = %v, want llm_agent", got["generated_by"])
	}
}

func TestParseObject_MarkdownFence(t *testing.T) {
	input := "```json\n{\"generated_by\":\"llm_agent\"}\n```"
	got, err := ParseObject(input)
	if err != nil {
		t.Fatalf("ParseObject error: %v", err)
	}
	if got["generated_by"] != "llm_agent" {
		t.Errorf("generated_by = %v, want llm_agent", got["generated_by"])
	}
}

func TestParseObject_SurroundingText(t *testing.T) {
	input := "以下是分析结果：\n{\"match_score\": 78}\n请查收。"
	got, err := ParseObject(input)
	if err != nil {
		t.Fatalf("ParseObject error: %v", err)
	}
	if got["match_score"] != float64(78) {
		t.Errorf("match_score = %v, want 78", got["match_score"])
	}
}

func TestParseObject_InvalidJSON(t *testing.T) {
	input := `{invalid json}`
	_, err := ParseObject(input)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), string(ErrorJSONParseFailed)) {
		t.Errorf("error %q should contain %q", err.Error(), ErrorJSONParseFailed)
	}
}

func TestParseObject_TopLevelArray(t *testing.T) {
	input := `[{"id": 1}, {"id": 2}]`
	_, err := ParseObject(input)
	if err == nil {
		t.Fatal("expected error for top-level array, got nil")
	}
	if !strings.Contains(err.Error(), string(ErrorJSONTopLevelNotObject)) {
		t.Errorf("error %q should contain %q", err.Error(), ErrorJSONTopLevelNotObject)
	}
}

func TestParseObject_EmptyString(t *testing.T) {
	_, err := ParseObject("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
}

func TestExtractAndParse_Success(t *testing.T) {
	input := `{"generated_by":"llm_agent","match_score":85}`
	result := ExtractAndParse(input)
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.ErrorMessage)
	}
	if result.Parsed["generated_by"] != "llm_agent" {
		t.Errorf("generated_by = %v, want llm_agent", result.Parsed["generated_by"])
	}
	if result.ExtractedJSON != input {
		t.Errorf("ExtractedJSON = %q, want %q", result.ExtractedJSON, input)
	}
	if result.RawOutputLength != len(input) {
		t.Errorf("RawOutputLength = %d, want %d", result.RawOutputLength, len(input))
	}
}

func TestExtractAndParse_Empty(t *testing.T) {
	result := ExtractAndParse("")
	if result.Success {
		t.Fatal("expected failure for empty input")
	}
	if result.ErrorCode != ErrorJSONEmpty {
		t.Errorf("ErrorCode = %q, want %q", result.ErrorCode, ErrorJSONEmpty)
	}
}

func TestExtractAndParse_NoJSONObject(t *testing.T) {
	result := ExtractAndParse("Just plain text here.")
	if result.Success {
		t.Fatal("expected failure for plain text")
	}
	if result.ErrorCode != ErrorJSONObjectNotFound {
		t.Errorf("ErrorCode = %q, want %q", result.ErrorCode, ErrorJSONObjectNotFound)
	}
}

func TestExtractAndParse_InvalidJSON(t *testing.T) {
	result := ExtractAndParse(`{broken json}`)
	if result.Success {
		t.Fatal("expected failure for invalid JSON")
	}
	if result.ErrorCode != ErrorJSONParseFailed {
		t.Errorf("ErrorCode = %q, want %q", result.ErrorCode, ErrorJSONParseFailed)
	}
}

func TestExtractAndParse_TopLevelArray(t *testing.T) {
	result := ExtractAndParse(`[1, 2, 3]`)
	if result.Success {
		t.Fatal("expected failure for top-level array")
	}
	if result.ErrorCode != ErrorJSONTopLevelNotObject {
		t.Errorf("ErrorCode = %q, want %q", result.ErrorCode, ErrorJSONTopLevelNotObject)
	}
}

func TestExtractAndParse_MarkdownFence(t *testing.T) {
	input := "```json\n{\"score\": 42}\n```"
	result := ExtractAndParse(input)
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.ErrorMessage)
	}
	if result.Parsed["score"] != float64(42) {
		t.Errorf("score = %v, want 42", result.Parsed["score"])
	}
}
