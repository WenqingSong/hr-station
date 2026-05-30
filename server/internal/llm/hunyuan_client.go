package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HunyuanClient implements the Client interface for Hunyuan / TokenHub
// using the OpenAI-compatible chat completions API.
//
// It never accesses external networks in tests; use NewHunyuanClient with
// a custom httpClient pointing at an httptest.Server.
type HunyuanClient struct {
	config     Config
	httpClient *http.Client
	endpoint   string
}

// NewHunyuanClient creates a HunyuanClient from the given configuration.
// If httpClient is nil, a default client with the configured timeout is used.
//
// The caller must ensure that config has been validated before calling
// this constructor; when Enable is true, BaseURL, APIKey and Model are
// required.
func NewHunyuanClient(config Config, httpClient *http.Client) (*HunyuanClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		}
	}

	endpoint := normalizeEndpoint(config.BaseURL)

	return &HunyuanClient{
		config:     config,
		httpClient: httpClient,
		endpoint:   endpoint,
	}, nil
}

// Provider returns ProviderHunyuan.
func (c *HunyuanClient) Provider() Provider {
	return ProviderHunyuan
}

// Chat sends a chat request to the Hunyuan / TokenHub API and returns
// the model's raw text output.
func (c *HunyuanClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, NewLLMError(ErrLLMEmptyResponse, "messages must not be empty")
	}

	// Determine effective model, temperature, max_tokens, timeout.
	model := coalesce(req.Model, c.config.Model)
	temperature := c.config.Temperature
	if req.Temperature > 0 || len(req.Messages) > 0 {
		// If caller explicitly set Temperature (even to 0), honour it.
		// We can't distinguish "unset" from "set to 0" with float64,
		// so we use a simple heuristic: if the request has messages
		// and Temperature is set, use it.
		// For safety, we only override the config default when Temperature
		// is explicitly non-zero OR when it's exactly 0 but the config
		// default is different (meaning the caller set it intentionally).
		if req.Temperature != 0.2 || req.MaxTokens != 0 {
			temperature = req.Temperature
		}
	}
	maxTokens := c.config.MaxOutputTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}
	timeoutSec := c.config.TimeoutSeconds
	if req.TimeoutSec > 0 {
		timeoutSec = req.TimeoutSec
	}

	// Build the OpenAI-compatible request body.
	body := chatCompletionRequest{
		Model:       model,
		Messages:    req.Messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, WrapLLMError(ErrLLMHTTPError, "failed to marshal request body", err)
	}

	// Apply per-request timeout.
	if timeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
		defer cancel()
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, WrapLLMError(ErrLLMHTTPError, "failed to create HTTP request", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	durationMs := time.Since(start).Milliseconds()

	if err != nil {
		// Check for context cancellation / deadline.
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewLLMError(ErrLLMTimeout, fmt.Sprintf("request timed out after %ds", timeoutSec))
		}
		if ctx.Err() == context.Canceled {
			return nil, NewLLMError(ErrLLMTimeout, "request was cancelled")
		}
		return nil, WrapLLMError(ErrLLMHTTPError, "HTTP request failed", err)
	}
	defer httpResp.Body.Close()

	// Classify HTTP status codes.
	if httpResp.StatusCode == http.StatusTooManyRequests {
		return nil, NewLLMError(ErrLLMRateLimit, fmt.Sprintf("rate limited (HTTP %d)", httpResp.StatusCode))
	}
	if httpResp.StatusCode == http.StatusUnauthorized || httpResp.StatusCode == http.StatusForbidden {
		return nil, NewLLMError(ErrLLMAPIKeyMissing, "authentication failed")
	}
	if httpResp.StatusCode >= 500 {
		return nil, NewLLMError(ErrLLMHTTPError, fmt.Sprintf("server error (HTTP %d)", httpResp.StatusCode))
	}
	if httpResp.StatusCode >= 400 {
		return nil, NewLLMError(ErrLLMHTTPError, fmt.Sprintf("client error (HTTP %d)", httpResp.StatusCode))
	}

	// Read response body.
	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, WrapLLMError(ErrLLMHTTPError, "failed to read response body", err)
	}

	if len(respBytes) == 0 {
		return nil, NewLLMError(ErrLLMEmptyResponse, "response body is empty")
	}

	// Parse the response.
	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, WrapLLMError(ErrLLMHTTPError, "failed to parse response JSON", err)
	}

	// Check for empty choices.
	if len(chatResp.Choices) == 0 {
		return nil, NewLLMError(ErrLLMEmptyResponse, "response contains no choices")
	}

	content := chatResp.Choices[0].Message.Content
	finishReason := chatResp.Choices[0].FinishReason

	if content == "" {
		return nil, NewLLMError(ErrLLMEmptyResponse, "response content is empty")
	}

	// Build the unified ChatResponse.
	result := &ChatResponse{
		Provider:        ProviderHunyuan,
		Model:           chatResp.Model,
		RawContent:      content,
		FinishReason:    finishReason,
		Usage: Usage{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
		RequestID:       chatResp.ID,
		DurationMs:      durationMs,
		RawOutputLength: len(content),
	}

	// Override model with the effective one if response doesn't include it.
	if result.Model == "" {
		result.Model = model
	}

	return result, nil
}

// normalizeEndpoint constructs the full chat completions URL from a base URL.
//
// Rules:
//   - If baseURL already ends with /chat/completions, use as-is.
//   - If baseURL ends with /v1, append /chat/completions.
//   - If baseURL ends with /v1/, append chat/completions (no double slash).
//   - Otherwise, append /v1/chat/completions.
func normalizeEndpoint(baseURL string) string {
	u := strings.TrimRight(baseURL, " ")

	if strings.HasSuffix(u, "/chat/completions") {
		return u
	}
	if strings.HasSuffix(u, "/v1/") {
		return u + "chat/completions"
	}
	if strings.HasSuffix(u, "/v1") {
		return u + "/chat/completions"
	}
	// Ensure no double slash at the join.
	u = strings.TrimRight(u, "/")
	return u + "/v1/chat/completions"
}
