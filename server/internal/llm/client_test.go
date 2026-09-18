package llm

import (
	"fmt"
	"testing"
)

func TestNoopClient_Provider(t *testing.T) {
	client := NewNoopClient(DefaultConfig())
	if got := client.Provider(); got != ProviderNoop {
		t.Errorf("Provider() = %q, want %q", got, ProviderNoop)
	}
}

func TestNoopClient_Chat_Disabled(t *testing.T) {
	cfg := DefaultConfig() // Enable=false by default
	client := NewNoopClient(cfg)

	resp, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error when LLM is disabled, got nil")
	}
	if resp != nil {
		t.Fatal("expected nil response when LLM is disabled")
	}

	llmErr, ok := err.(*LLMError)
	if !ok {
		t.Fatalf("expected *LLMError, got %T", err)
	}
	if llmErr.Code != ErrLLMDisabled {
		t.Errorf("error code = %q, want %q", llmErr.Code, ErrLLMDisabled)
	}
}

func TestNoopClient_Chat_Enabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://example.com/v1"
	cfg.APIKey = "test-key-for-noop"
	cfg.Model = "test-model"
	client := NewNoopClient(cfg)

	resp, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "system prompt"},
			{Role: RoleUser, Content: "user message"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Provider != ProviderNoop {
		t.Errorf("Provider = %q, want %q", resp.Provider, ProviderNoop)
	}
	if resp.RawContent == "" {
		t.Error("RawContent should not be empty")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "stop")
	}
	if resp.RawOutputLength != len(resp.RawContent) {
		t.Errorf("RawOutputLength = %d, want %d", resp.RawOutputLength, len(resp.RawContent))
	}
	if resp.Model != "test-model" {
		t.Errorf("Model = %q, want %q", resp.Model, "test-model")
	}
}

func TestNoopClient_Chat_RequestModelOverride(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://example.com/v1"
	cfg.APIKey = "test-key"
	cfg.Model = "default-model"
	client := NewNoopClient(cfg)

	resp, err := client.Chat(t.Context(), ChatRequest{
		Model:    "override-model",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Model != "override-model" {
		t.Errorf("Model = %q, want %q", resp.Model, "override-model")
	}
}

func TestLLMError_Error(t *testing.T) {
	err := NewLLMError(ErrLLMTimeout, "request timed out after 30s")
	want := "LLM_TIMEOUT: request timed out after 30s"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestLLMError_Error_WithCause(t *testing.T) {
	cause := NewLLMError(ErrLLMHTTPError, "connection refused")
	err := WrapLLMError(ErrLLMTimeout, "request timed out", cause)
	got := err.Error()
	if got == "" {
		t.Error("Error() should not be empty")
	}
	// Should contain both codes and messages
	if !contains(got, "LLM_TIMEOUT") {
		t.Errorf("Error() should contain LLM_TIMEOUT, got %q", got)
	}
}

func TestLLMError_Unwrap(t *testing.T) {
	inner := NewLLMError(ErrLLMHTTPError, "inner")
	outer := WrapLLMError(ErrLLMTimeout, "outer", inner)
	if unwrapped := outer.Unwrap(); unwrapped != inner {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, inner)
	}
}

func TestLLMError_Unwrap_Nil(t *testing.T) {
	err := NewLLMError(ErrLLMDisabled, "disabled")
	if unwrapped := err.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestIsErrorCode(t *testing.T) {
	err := NewLLMError(ErrLLMTimeout, "timeout")
	if !IsErrorCode(err, ErrLLMTimeout) {
		t.Error("IsErrorCode should return true for matching code")
	}
	if IsErrorCode(err, ErrLLMDisabled) {
		t.Error("IsErrorCode should return false for non-matching code")
	}
	if IsErrorCode(nil, ErrLLMTimeout) {
		t.Error("IsErrorCode should return false for nil error")
	}
	if IsErrorCode(fmt.Errorf("other"), ErrLLMTimeout) {
		t.Error("IsErrorCode should return false for non-LLMError")
	}
}

func TestIsDirectMockFallback(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected bool
	}{
		{ErrLLMDisabled, true},
		{ErrLLMConfigMissing, true},
		{ErrLLMAPIKeyMissing, true},
		{ErrLLMAuthFailed, true},
		{ErrLLMTimeout, true},
		{ErrLLMHTTPError, true},
		{ErrLLMRateLimit, true},
		{ErrLLMEmptyResponse, true},
		{ErrLLMOutputTruncated, false},
		{ErrUnknown, true},
	}
	for _, tt := range tests {
		err := NewLLMError(tt.code, "test")
		got := IsDirectMockFallback(err)
		if got != tt.expected {
			t.Errorf("IsDirectMockFallback(%s) = %v, want %v", tt.code, got, tt.expected)
		}
	}
	// nil error
	if IsDirectMockFallback(nil) {
		t.Error("IsDirectMockFallback(nil) should be false")
	}
	// non-LLMError goes to Mock
	if !IsDirectMockFallback(fmt.Errorf("other")) {
		t.Error("IsDirectMockFallback(non-LLMError) should be true")
	}
}

func TestChatRequest_Messages(t *testing.T) {
	req := ChatRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "system"},
			{Role: RoleUser, Content: "user"},
		},
		Temperature: 0.5,
		MaxTokens:   4096,
		Metadata:    map[string]string{"prompt_name": "match_analysis"},
	}
	if len(req.Messages) != 2 {
		t.Errorf("len(Messages) = %d, want 2", len(req.Messages))
	}
	if req.Messages[0].Role != RoleSystem {
		t.Errorf("Messages[0].Role = %q, want %q", req.Messages[0].Role, RoleSystem)
	}
	if req.Metadata["prompt_name"] != "match_analysis" {
		t.Errorf("Metadata[prompt_name] = %q, want %q", req.Metadata["prompt_name"], "match_analysis")
	}
}

func TestChatResponse_RawOutputLength(t *testing.T) {
	resp := &ChatResponse{
		Provider:        ProviderHunyuan,
		RawContent:      `{"key": "value"}`,
		RawOutputLength: 16,
	}
	if resp.RawOutputLength != len(resp.RawContent) {
		t.Errorf("RawOutputLength = %d, want %d", resp.RawOutputLength, len(resp.RawContent))
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "(empty)"},
		{"short", "****"},
		{"12345678", "****5678"},
		{"sk-abcdefghijklmnop", "***************mnop"},
	}
	for _, tt := range tests {
		got := maskAPIKey(tt.input)
		if got != tt.expected {
			t.Errorf("maskAPIKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
