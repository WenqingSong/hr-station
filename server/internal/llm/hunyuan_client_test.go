package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ─── Helper: valid config for tests ───

func testHunyuanConfig() Config {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://api.example.com/v1"
	cfg.APIKey = "fake-test-key-1234"
	cfg.Model = "test-model"
	return cfg
}

// ─── Constructor Tests ───

func TestNewHunyuanClient_ValidConfig(t *testing.T) {
	client, err := NewHunyuanClient(testHunyuanConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewHunyuanClient_MissingAPIKey(t *testing.T) {
	cfg := testHunyuanConfig()
	cfg.APIKey = ""
	_, err := NewHunyuanClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error when APIKey is missing")
	}
}

func TestNewHunyuanClient_MissingBaseURL(t *testing.T) {
	cfg := testHunyuanConfig()
	cfg.BaseURL = ""
	_, err := NewHunyuanClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error when BaseURL is missing")
	}
}

func TestNewHunyuanClient_MissingModel(t *testing.T) {
	cfg := testHunyuanConfig()
	cfg.Model = ""
	_, err := NewHunyuanClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error when Model is missing")
	}
}

func TestNewHunyuanClient_Provider(t *testing.T) {
	client, _ := NewHunyuanClient(testHunyuanConfig(), nil)
	if got := client.Provider(); got != ProviderHunyuan {
		t.Errorf("Provider() = %q, want %q", got, ProviderHunyuan)
	}
}

// ─── Endpoint Normalization Tests ───

func TestNormalizeEndpoint_V1(t *testing.T) {
	got := normalizeEndpoint("https://api.example.com/v1")
	want := "https://api.example.com/v1/chat/completions"
	if got != want {
		t.Errorf("normalizeEndpoint(%q) = %q, want %q", "https://api.example.com/v1", got, want)
	}
}

func TestNormalizeEndpoint_V1TrailingSlash(t *testing.T) {
	got := normalizeEndpoint("https://api.example.com/v1/")
	want := "https://api.example.com/v1/chat/completions"
	if got != want {
		t.Errorf("normalizeEndpoint(%q) = %q, want %q", "https://api.example.com/v1/", got, want)
	}
}

func TestNormalizeEndpoint_ChatCompletions(t *testing.T) {
	got := normalizeEndpoint("https://api.example.com/v1/chat/completions")
	want := "https://api.example.com/v1/chat/completions"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNormalizeEndpoint_NoDoubleSlash(t *testing.T) {
	got := normalizeEndpoint("https://api.example.com/v1/")
	if strings.Contains(got, "//chat") {
		t.Errorf("endpoint should not contain double slash: %q", got)
	}
}

func TestNormalizeEndpoint_BareHost(t *testing.T) {
	got := normalizeEndpoint("https://api.example.com")
	want := "https://api.example.com/v1/chat/completions"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// ─── Chat Success Tests ───

func TestHunyuanClient_Chat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify Authorization header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer fake-test-key-1234" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer fake-test-key-1234")
		}
		// Verify Content-Type.
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		// Read and verify request body.
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		if reqBody["model"] != "test-model" {
			t.Errorf("model = %v, want %q", reqBody["model"], "test-model")
		}
		if msgs, ok := reqBody["messages"].([]interface{}); !ok || len(msgs) != 2 {
			t.Errorf("messages length = %v, want 2", len(msgs))
		}
		if _, ok := reqBody["temperature"]; !ok {
			t.Error("request body should contain temperature")
		}
		if _, ok := reqBody["max_tokens"]; !ok {
			t.Error("request body should contain max_tokens")
		}

		// Return success response.
		resp := chatCompletionResponse{
			ID: "chatcmpl-test-123",
			Choices: []chatChoice{
				{
					Index: 0,
					Message: chatMessage{
						Role:    "assistant",
						Content: `{"match_score": 85, "summary": "good match"}`,
					},
					FinishReason: "stop",
				},
			},
			Usage: chatCompletionUsage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, err := NewHunyuanClient(cfg, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{
			{Role: RoleSystem, Content: "system"},
			{Role: RoleUser, Content: "user"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Provider != ProviderHunyuan {
		t.Errorf("Provider = %q, want %q", resp.Provider, ProviderHunyuan)
	}
	if resp.RawContent != `{"match_score": 85, "summary": "good match"}` {
		t.Errorf("RawContent = %q, unexpected", resp.RawContent)
	}
	if resp.Usage.TotalTokens != 150 {
		t.Errorf("Usage.TotalTokens = %d, want 150", resp.Usage.TotalTokens)
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "stop")
	}
	if resp.RequestID != "chatcmpl-test-123" {
		t.Errorf("RequestID = %q, want %q", resp.RequestID, "chatcmpl-test-123")
	}
	if resp.RawOutputLength != len(resp.RawContent) {
		t.Errorf("RawOutputLength = %d, want %d", resp.RawOutputLength, len(resp.RawContent))
	}
	if resp.DurationMs < 0 {
		t.Errorf("DurationMs = %d, want >= 0", resp.DurationMs)
	}
}

// ─── Error Tests ───

func TestHunyuanClient_Chat_EmptyMessages(t *testing.T) {
	cfg := testHunyuanConfig()
	client, _ := NewHunyuanClient(cfg, nil)

	_, err := client.Chat(t.Context(), ChatRequest{Messages: []Message{}})
	if err == nil {
		t.Fatal("expected error for empty messages")
	}
	if !IsErrorCode(err, ErrLLMEmptyResponse) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMEmptyResponse)
	}
}

func TestHunyuanClient_Chat_RateLimit429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit"}`))
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 429")
	}
	if !IsErrorCode(err, ErrLLMRateLimit) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMRateLimit)
	}
}

func TestHunyuanClient_Chat_ServerError500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !IsErrorCode(err, ErrLLMHTTPError) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMHTTPError)
	}
}

func TestHunyuanClient_Chat_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !IsErrorCode(err, ErrLLMEmptyResponse) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMEmptyResponse)
	}
}

func TestHunyuanClient_Chat_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			ID:      "test-id",
			Choices: []chatChoice{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !IsErrorCode(err, ErrLLMEmptyResponse) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMEmptyResponse)
	}
}

func TestHunyuanClient_Chat_EmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			ID: "test-id",
			Choices: []chatChoice{
				{Index: 0, Message: chatMessage{Role: "assistant", Content: ""}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
	if !IsErrorCode(err, ErrLLMEmptyResponse) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMEmptyResponse)
	}
}

func TestHunyuanClient_Chat_NonJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("this is not json"))
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
	if !IsErrorCode(err, ErrLLMHTTPError) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMHTTPError)
	}
}

func TestHunyuanClient_Chat_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	cfg.TimeoutSeconds = 1
	client, _ := NewHunyuanClient(cfg, server.Client())

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Chat(ctx, ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for context timeout")
	}
	if !IsErrorCode(err, ErrLLMTimeout) {
		llmErr, ok := err.(*LLMError)
		if ok {
			t.Errorf("error code = %q, want %q", llmErr.Code, ErrLLMTimeout)
		} else {
			t.Errorf("expected *LLMError, got %T", err)
		}
	}
}

// ─── Security Tests ───

func TestHunyuanClient_Chat_ErrorNoAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "fake-test-key-1234") {
		t.Errorf("error string should not contain API key, got: %q", errStr)
	}
}

func TestHunyuanClient_Chat_AuthHeaderNotInError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	errStr := err.Error()
	if strings.Contains(errStr, "Bearer") {
		t.Errorf("error string should not contain Authorization header, got: %q", errStr)
	}
}

func TestHunyuanClient_Chat_ConfigIncomplete_NoFullConfig(t *testing.T) {
	cfg := testHunyuanConfig()
	cfg.APIKey = ""
	_, err := NewHunyuanClient(cfg, nil)
	if err == nil {
		t.Fatal("expected error for incomplete config")
	}
	errStr := err.Error()
	// The error should not print the full config (which would contain BaseURL etc.)
	if strings.Contains(errStr, "https://api.example.com") {
		t.Errorf("error should not print full config details, got: %q", errStr)
	}
}

func TestHunyuanClient_Chat_401AuthFailed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if !IsErrorCode(err, ErrLLMAuthFailed) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMAuthFailed)
	}
	// Error message should not contain the API key.
	errStr := err.Error()
	if strings.Contains(errStr, "fake-test-key") {
		t.Errorf("error should not contain API key, got: %q", errStr)
	}
}

// ─── Integration-like Tests ───

func TestHunyuanClient_Chat_FinishReasonLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := chatCompletionResponse{
			ID: "chatcmpl-truncated",
			Choices: []chatChoice{
				{
					Index:        0,
					Message:      chatMessage{Role: "assistant", Content: `{"match_score": 85, "sum`},
					FinishReason: "length",
				},
			},
			Usage: chatCompletionUsage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	resp, err := client.Chat(t.Context(), ChatRequest{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// finish_reason=length is NOT treated as a hard error.
	// The response is returned with FinishReason="length".
	// Pipeline can decide whether to attempt repair.
	if resp.FinishReason != "length" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "length")
	}
}

func TestHunyuanClient_Chat_DefaultHTTPClient(t *testing.T) {
	cfg := testHunyuanConfig()
	cfg.TimeoutSeconds = 5
	client, err := NewHunyuanClient(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.httpClient == nil {
		t.Error("httpClient should be initialized when nil is passed")
	}
}

func TestHunyuanClient_Chat_RequestModelOverride(t *testing.T) {
	var capturedModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		capturedModel, _ = reqBody["model"].(string)

		resp := chatCompletionResponse{
			ID: "test-id",
			Choices: []chatChoice{
				{Index: 0, Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Model:    "override-model",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedModel != "override-model" {
		t.Errorf("model in request = %q, want %q", capturedModel, "override-model")
	}
}

func TestHunyuanClient_Chat_TemperatureZero(t *testing.T) {
	var capturedTemp interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		capturedTemp = reqBody["temperature"]

		resp := chatCompletionResponse{
			ID: "test-id",
			Choices: []chatChoice{
				{Index: 0, Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	cfg.Temperature = 0.5
	client, _ := NewHunyuanClient(cfg, server.Client())

	// When Temperature=0 and MaxTokens>0, the request Temperature should be used
	_, err := client.Chat(t.Context(), ChatRequest{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: 0,
		MaxTokens:   100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedTemp != float64(0) {
		t.Errorf("temperature in request = %v, want 0", capturedTemp)
	}
}

func TestHunyuanClient_Chat_MaxTokensOverride(t *testing.T) {
	var capturedMaxTokens interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)
		capturedMaxTokens = reqBody["max_tokens"]

		resp := chatCompletionResponse{
			ID: "test-id",
			Choices: []chatChoice{
				{Index: 0, Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := testHunyuanConfig()
	cfg.BaseURL = server.URL + "/v1"
	cfg.MaxOutputTokens = 8192
	client, _ := NewHunyuanClient(cfg, server.Client())

	_, err := client.Chat(t.Context(), ChatRequest{
		Messages:  []Message{{Role: RoleUser, Content: "hi"}},
		MaxTokens: 4096,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// json.Unmarshal decodes numbers as float64
	if fmt.Sprintf("%v", capturedMaxTokens) != "4096" {
		t.Errorf("max_tokens in request = %v, want 4096", capturedMaxTokens)
	}
}
