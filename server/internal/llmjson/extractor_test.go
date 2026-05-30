package llmjson

import (
	"strings"
	"testing"
)

func TestExtractObject_PureJSON(t *testing.T) {
	input := `{"generated_by":"llm_agent"}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestExtractObject_MarkdownJSONFence(t *testing.T) {
	input := "```json\n{\"generated_by\":\"llm_agent\"}\n```"
	want := `{"generated_by":"llm_agent"}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractObject_CodeFenceNoLanguage(t *testing.T) {
	input := "```\n{\"generated_by\":\"llm_agent\"}\n```"
	want := `{"generated_by":"llm_agent"}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractObject_SurroundingText(t *testing.T) {
	input := "以下是结果：\n{\"generated_by\":\"llm_agent\"}\n请查收。"
	want := `{"generated_by":"llm_agent"}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractObject_MultiLineJSON(t *testing.T) {
	input := "{\n  \"generated_by\": \"llm_agent\"\n}"
	want := "{\n  \"generated_by\": \"llm_agent\"\n}"
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractObject_EmptyString(t *testing.T) {
	_, err := ExtractObject("")
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if !strings.Contains(err.Error(), string(ErrorJSONEmpty)) {
		t.Errorf("error %q should contain %q", err.Error(), ErrorJSONEmpty)
	}
}

func TestExtractObject_NoJSONObject(t *testing.T) {
	_, err := ExtractObject("This is just plain text without any JSON.")
	if err == nil {
		t.Fatal("expected error for no JSON object, got nil")
	}
	if !strings.Contains(err.Error(), string(ErrorJSONObjectNotFound)) {
		t.Errorf("error %q should contain %q", err.Error(), ErrorJSONObjectNotFound)
	}
}

func TestExtractObject_BOMPrefixed(t *testing.T) {
	bom := "\xEF\xBB\xBF"
	input := bom + `{"generated_by":"llm_agent"}`
	want := `{"generated_by":"llm_agent"}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExtractObject_NestedJSON(t *testing.T) {
	input := `{"outer": {"inner": "value"}, "list": [1, 2, 3]}`
	got, err := ExtractObject(input)
	if err != nil {
		t.Fatalf("ExtractObject error: %v", err)
	}
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}
