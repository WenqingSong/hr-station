package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ─── Hy3Client ───────────────────────────────────────────────────────

// Hy3Client implements Client for Tencent Cloud LKEAP hy3-preview.
type Hy3Client struct {
	config   Config
	httpDoer *http.Client
}

// NewHy3Client creates a Hy3Client. httpClient may be nil (uses default).
func NewHy3Client(config Config, httpClient *http.Client) (*Hy3Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		}
	}
	return &Hy3Client{config: config, httpDoer: httpClient}, nil
}

// Provider returns ProviderHy3.
func (c *Hy3Client) Provider() Provider { return ProviderHy3 }

// Chat sends a chat completion request to LKEAP hy3-preview.
func (c *Hy3Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, &LLMError{Code: ErrLLMConfigMissing, Message: "messages is empty"}
	}

	// Merge default config
	model := req.Model
	if model == "" { model = c.config.Model }
	temp := req.Temperature
	if temp == 0 { temp = c.config.Temperature }
	maxTok := req.MaxTokens
	if maxTok <= 0 { maxTok = c.config.MaxOutputTokens }

	// Build extra_body
	extraBody := make(map[string]interface{})
	for k, v := range req.ExtraBody {
		extraBody[k] = v
	}
	if req.ReasoningEffort != "" {
		if extraBody["chat_template_kwargs"] == nil {
			extraBody["chat_template_kwargs"] = map[string]interface{}{"reasoning_effort": req.ReasoningEffort}
		}
	}

	hy3Req := hy3ChatCompletionRequest{
		Model:     model,
		Messages:  req.Messages,
		MaxTokens: maxTok,
	}
	if temp > 0 {
		hy3Req.Temperature = &temp
	}
	if req.TopP > 0 {
		hy3Req.TopP = &req.TopP
	}
	if len(extraBody) > 0 {
		hy3Req.ExtraBody = extraBody
	}

	body, err := json.Marshal(hy3Req)
	if err != nil {
		return nil, fmt.Errorf("hy3 marshal: %w", err)
	}

	timeout := time.Duration(c.config.TimeoutSeconds) * time.Second
	if req.TimeoutSec > 0 {
		timeout = time.Duration(req.TimeoutSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Normalize URL
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")
	if !strings.HasSuffix(baseURL, "/plan/v3/chat/completions") && !strings.Contains(baseURL, "/chat/completions") {
		baseURL = baseURL + "/v1/chat/completions"
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, &LLMError{Code: ErrLLMHTTPError, Message: err.Error()}
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	log.Printf("[LLM-Hy3] model=%s temp=%.2f max_tokens=%d reason_effort=%s body_bytes=%d",
		model, temp, maxTok, req.ReasoningEffort, len(body))

	resp, err := c.httpDoer.Do(httpReq)
	if err != nil {
		dur := time.Since(start).Milliseconds()
		if ctx.Err() != nil {
			return nil, &LLMError{Code: ErrLLMTimeout, Message: fmt.Sprintf("hy3 timeout after %dms", dur), Cause: ctx.Err()}
		}
		return nil, &LLMError{Code: ErrLLMHTTPError, Message: err.Error(), Cause: err}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
	durationMs := time.Since(start).Milliseconds()

	if resp.StatusCode != 200 {
		summary := string(respBody)
		if len(summary) > 300 { summary = summary[:300] }
		code := errorFromHTTPStatus(resp.StatusCode)
		log.Printf("[LLM-Hy3] http %d: %s", resp.StatusCode, summary)
		return nil, &LLMError{Code: code, Message: fmt.Sprintf("hy3 http %d", resp.StatusCode), Cause: fmt.Errorf("%s", summary)}
	}

	var hy3Resp hy3ChatCompletionResponse
	if err := json.Unmarshal(respBody, &hy3Resp); err != nil {
		preview := string(respBody)
		if len(preview) > 200 { preview = preview[:200] }
		return nil, &LLMError{Code: ErrLLMEmptyResponse, Message: "hy3 parse failed", Cause: fmt.Errorf("%s", preview)}
	}

	if len(hy3Resp.Choices) == 0 || hy3Resp.Choices[0].Message.Content == "" {
		return nil, &LLMError{Code: ErrLLMEmptyResponse, Message: "hy3 empty response"}
	}

	return &ChatResponse{
		Provider:        ProviderHy3,
		Model:           hy3Resp.Model,
		RawContent:      hy3Resp.Choices[0].Message.Content,
		FinishReason:    hy3Resp.Choices[0].FinishReason,
		Usage:           Usage{PromptTokens: hy3Resp.Usage.PromptTokens, CompletionTokens: hy3Resp.Usage.CompletionTokens, TotalTokens: hy3Resp.Usage.TotalTokens},
		RequestID:       hy3Resp.ID,
		DurationMs:      durationMs,
		RawOutputLength: len(hy3Resp.Choices[0].Message.Content),
	}, nil
}

func errorFromHTTPStatus(code int) ErrorCode {
	switch {
	case code == 429: return ErrLLMRateLimit
	case code == 401 || code == 403: return ErrLLMAuthFailed
	case code >= 500: return ErrLLMHTTPError
	default: return ErrLLMHTTPError
	}
}
