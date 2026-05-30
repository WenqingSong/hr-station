// llm-smoke is a one-shot development tool that verifies HunyuanClient
// connectivity using the minimal possible prompt. It does NOT connect to
// any business Agent, database, or external service beyond the LLM API.
//
// Usage:
//
//	LLM_ENABLE=true LLM_BASE_URL=... LLM_API_KEY=... LLM_MODEL=... go run ./cmd/llm-smoke
//
// Safety:
//   - Never prints the API key.
//   - Never prints full raw output (at most 120 chars preview).
//   - Never reads resumes, jobs, or user data.
//   - Never accesses the database.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/llmjson"
)

func main() {
	fmt.Println("LLM smoke test")
	fmt.Println("===============")
	fmt.Println()

	// Read configuration from environment variables.
	cfg := configFromEnv()

	// Check if LLM is enabled.
	if !cfg.Enable {
		fmt.Println("LLM smoke skipped:")
		fmt.Println("  LLM_ENABLE is not true")
		os.Exit(1)
	}

	// Validate config before creating client.
	if err := cfg.Validate(); err != nil {
		fmt.Println("LLM smoke skipped or failed:")
		if llmErr, ok := err.(*llm.LLMError); ok {
			fmt.Printf("  error_code=%s\n", llmErr.Code)
		}
		fmt.Printf("  message=%s\n", err.Error())
		os.Exit(1)
	}

	// Check provider.
	if cfg.Provider != llm.ProviderHunyuan {
		fmt.Printf("LLM smoke skipped: provider %q is not supported (expected %q)\n", cfg.Provider, llm.ProviderHunyuan)
		os.Exit(1)
	}

	fmt.Println("Configuration check:")
	fmt.Printf("  provider=%s\n", cfg.Provider)
	fmt.Printf("  model=%s\n", cfg.Model)
	fmt.Printf("  base_url=%s\n", cfg.BaseURL)
	fmt.Printf("  api_key=%s\n", maskKey(cfg.APIKey))
	fmt.Printf("  temperature=%.1f\n", cfg.Temperature)
	fmt.Printf("  max_output_tokens=%d\n", cfg.MaxOutputTokens)
	fmt.Printf("  timeout_seconds=%d\n", cfg.TimeoutSeconds)
	fmt.Println()

	// Create HunyuanClient.
	client, err := llm.NewHunyuanClient(cfg, nil)
	if err != nil {
		fmt.Println("LLM smoke failed: cannot create HunyuanClient")
		fmt.Printf("  error=%s\n", err.Error())
		os.Exit(1)
	}

	// Build minimal prompt.
	req := llm.ChatRequest{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "你是一个连通性测试助手。你只能输出严格 JSON，不要输出 Markdown。"},
			{Role: llm.RoleUser, Content: "请只输出一个 JSON 对象：{\"ok\":true,\"message\":\"pong\"}"},
		},
		Temperature: 0,
		MaxTokens:   256,
		TimeoutSec:  cfg.TimeoutSeconds,
	}

	fmt.Println("Sending smoke test request...")
	fmt.Printf("  messages_count=%d\n", len(req.Messages))
	fmt.Printf("  temperature=%.1f\n", req.Temperature)
	fmt.Printf("  max_tokens=%d\n", req.MaxTokens)
	fmt.Println()

	// Call with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := client.Chat(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Println("LLM smoke test failed:")
		if llmErr, ok := err.(*llm.LLMError); ok {
			fmt.Printf("  error_code=%s\n", llmErr.Code)
		}
		fmt.Printf("  message=%s\n", err.Error())
		fmt.Printf("  elapsed_ms=%d\n", elapsed.Milliseconds())
		os.Exit(1)
	}

	// Print results.
	fmt.Println("LLM smoke test result:")
	fmt.Printf("  provider=%s\n", resp.Provider)
	fmt.Printf("  model=%s\n", resp.Model)
	fmt.Printf("  finish_reason=%s\n", resp.FinishReason)
	fmt.Printf("  raw_output_length=%d\n", resp.RawOutputLength)
	fmt.Printf("  duration_ms=%d\n", resp.DurationMs)
	fmt.Printf("  usage.prompt_tokens=%d\n", resp.Usage.PromptTokens)
	fmt.Printf("  usage.completion_tokens=%d\n", resp.Usage.CompletionTokens)
	fmt.Printf("  usage.total_tokens=%d\n", resp.Usage.TotalTokens)

	// Attempt JSON parse using llmjson.
	jsonParseSuccess := false
	if resp.RawContent != "" {
		result := llmjson.ExtractAndParse(resp.RawContent)
		jsonParseSuccess = result.Success
	}
	fmt.Printf("  json_parse_success=%v\n", jsonParseSuccess)

	// Print a safe preview of raw output (max 120 chars, no newlines).
	preview := safePreview(resp.RawContent, 120)
	fmt.Printf("  raw_output_preview=%s\n", preview)

	fmt.Println()

	// Verify basic pass conditions.
	if resp.RawContent == "" {
		fmt.Println("LLM smoke test WARNING: raw content is empty")
		os.Exit(1)
	}
	if resp.RawOutputLength <= 0 {
		fmt.Println("LLM smoke test WARNING: raw output length is 0")
		os.Exit(1)
	}

	fmt.Println("LLM smoke test PASSED: HunyuanClient connectivity verified")
}

func configFromEnv() llm.Config {
	cfg := llm.DefaultConfig()
	cfg.Enable = envBool("LLM_ENABLE", false)
	cfg.Provider = llm.Provider(envString("LLM_PROVIDER", "hunyuan"))
	cfg.BaseURL = envString("LLM_BASE_URL", "")
	cfg.APIKey = envString("LLM_API_KEY", "")
	cfg.Model = envString("LLM_MODEL", "")
	cfg.Temperature = envFloat("LLM_TEMPERATURE", 0.2)
	cfg.MaxOutputTokens = envInt("LLM_MAX_OUTPUT_TOKENS", 8192)
	cfg.TimeoutSeconds = envInt("LLM_TIMEOUT_SECONDS", 30)
	cfg.RepairEnable = envBool("LLM_REPAIR_ENABLE", true)
	cfg.RepairMaxAttempts = envInt("LLM_REPAIR_MAX_ATTEMPTS", 1)
	cfg.FallbackEnable = envBool("LLM_FALLBACK_ENABLE", true)
	cfg.LogRawOutput = envBool("LLM_LOG_RAW_OUTPUT", false)
	cfg.PromptDir = envString("LLM_PROMPT_DIR", "server/prompts")
	cfg.SchemaDir = envString("LLM_SCHEMA_DIR", "server/prompts/schema")
	return cfg
}

func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

// maskKey returns a masked version of an API key for safe display.
func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) < 8 {
		return "****"
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}

// safePreview returns a safe, single-line preview of s with maxLen characters.
func safePreview(s string, maxLen int) string {
	if s == "" {
		return "(empty)"
	}
	// Replace newlines with spaces.
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Trim extra spaces.
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
