package repair

import (
	"context"
	"fmt"
	"strings"

	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/llmjson"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// Agent implements the JSONRepairAgent repair-once call chain.
// It loads the json_repair Prompt, renders it with the repair request
// variables, calls llm.Client once, and then applies JSON extract/parse
// and Schema validate on the repair output.
//
// It does NOT:
//   - Retry or recurse
//   - Perform business analysis
//   - Implement Mock fallback
//   - Modify generated_by
type Agent struct {
	llmClient    llm.Client
	promptLoader *prompt.Loader
	schemaLoader *schema.Loader
	validator    *schema.Validator
}

// NewAgent creates a new JSONRepairAgent.
func NewAgent(
	llmClient llm.Client,
	promptLoader *prompt.Loader,
	schemaLoader *schema.Loader,
	validator *schema.Validator,
) *Agent {
	return &Agent{
		llmClient:    llmClient,
		promptLoader: promptLoader,
		schemaLoader: schemaLoader,
		validator:    validator,
	}
}

// Repair executes a single JSON repair attempt. It:
//  1. Validates the TargetOutputType
//  2. Loads and renders the json_repair Prompt template
//  3. Calls llm.Client once
//  4. Extracts and parses JSON from the repair output
//  5. Validates the parsed object against the target Schema
//
// Repair is always exactly one attempt — no recursion, no retry loop.
func (a *Agent) Repair(ctx context.Context, req RepairRequest) *RepairResult {
	result := &RepairResult{
		RepairAttempted: true,
		RepairSuccess:   false,
	}

	// Step 1: Validate TargetOutputType.
	if !ValidTargetOutputTypes[req.TargetOutputType] {
		result.ErrorCode = ErrRepairTargetInvalid
		result.ErrorMessage = fmt.Sprintf(
			"invalid target output type: %q (allowed: match_analysis, recommendation, resume_optimization)",
			req.TargetOutputType,
		)
		return result
	}

	// Step 2: Load the json_repair Prompt template.
	rendered, err := a.renderPrompt(req)
	if err != nil {
		result.ErrorCode = ErrRepairPromptLoadFailed
		result.ErrorMessage = fmt.Sprintf("failed to load/render repair prompt: %s", err.Error())
		return result
	}

	// Step 3: Call llm.Client exactly once.
	chatResp, err := a.callLLM(ctx, req, rendered)
	if err != nil {
		result.ErrorCode = ErrRepairLLMFailed
		result.ErrorMessage = fmt.Sprintf("LLM call failed: %s", safeErrorMessage(err))
		return result
	}

	result.RawContent = chatResp.RawContent
	result.RawOutputLength = chatResp.RawOutputLength

	if chatResp.RawContent == "" {
		result.ErrorCode = ErrRepairLLMFailed
		result.ErrorMessage = "LLM returned empty content"
		return result
	}

	// Step 4: Extract and parse JSON from the repair output.
	parseResult := llmjson.ExtractAndParse(chatResp.RawContent)
	if !parseResult.Success {
		result.ErrorCode = ErrRepairParseFailed
		result.ErrorMessage = fmt.Sprintf(
			"repair output JSON parse failed: %s", parseResult.ErrorMessage,
		)
		return result
	}

	// Step 5: Validate the parsed object against the target Schema.
	schemaDoc, err := a.schemaLoader.Load(schemaNameFromTarget(req.TargetOutputType))
	if err != nil {
		result.ErrorCode = ErrRepairSchemaLoadFailed
		result.ErrorMessage = fmt.Sprintf("failed to load target schema: %s", err.Error())
		return result
	}

	valResult, err := a.validator.Validate(schemaDoc, parseResult.Parsed)
	if err != nil {
		// Validator compilation error (unexpected).
		result.ErrorCode = ErrRepairSchemaFailed
		result.ErrorMessage = fmt.Sprintf("schema validation error: %s", err.Error())
		return result
	}

	if !valResult.Success {
		result.ErrorCode = ErrRepairSchemaFailed
		result.ErrorMessage = valResult.ErrorMessage
		result.ValidationErrors = valResult.Errors
		return result
	}

	// Success.
	result.Success = true
	result.RepairSuccess = true
	result.Parsed = parseResult.Parsed
	result.RepairedJSON = parseResult.ExtractedJSON

	return result
}

// renderPrompt loads the json_repair Prompt template and replaces the
// five placeholders with the repair request values.
func (a *Agent) renderPrompt(req RepairRequest) (string, error) {
	values := map[string]string{
		"TARGET_OUTPUT_TYPE": string(req.TargetOutputType),
		"TARGET_SCHEMA_JSON": req.TargetSchemaJSON,
		"BROKEN_MODEL_OUTPUT": req.BrokenOutput,
		"VALIDATION_ERRORS":  req.ValidationErrors,
		"REPAIR_CONTEXT":     req.RepairContext,
	}

	rendered, err := a.promptLoader.Render(prompt.PromptJSONRepair, values)
	if err != nil {
		return "", fmt.Errorf("prompt render: %w", err)
	}

	return rendered, nil
}

// callLLM constructs a ChatRequest and calls the LLM once.
func (a *Agent) callLLM(ctx context.Context, req RepairRequest, renderedPrompt string) (*llm.ChatResponse, error) {
	chatReq := llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "你是 JSONRepairAgent，只输出严格 JSON。",
			},
			{
				Role:    llm.RoleUser,
				Content: renderedPrompt,
			},
		},
		Temperature: 0,
		MaxTokens:   8192,
		Metadata: map[string]string{
			"agent_name":        "JSONRepairAgent",
			"prompt_name":       "json_repair_v1",
			"target_output_type": string(req.TargetOutputType),
		},
	}

	return a.llmClient.Chat(ctx, chatReq)
}

// schemaNameFromTarget maps a TargetOutputType to a schema.SchemaName.
func schemaNameFromTarget(t TargetOutputType) schema.SchemaName {
	switch t {
	case TargetMatchAnalysis:
		return schema.SchemaMatchAnalysis
	case TargetRecommendation:
		return schema.SchemaRecommendation
	case TargetResumeOptimization:
		return schema.SchemaResumeOptimization
	case TargetCompare:
		return schema.SchemaVersionCompare
	default:
		return schema.SchemaMatchAnalysis
	}
}

// safeErrorMessage returns an error message that does not leak API keys,
// full Prompts, or full model outputs. It truncates long messages and
// redacts patterns that look like API keys.
func safeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// Truncate very long error messages to avoid leaking full content.
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	// Redact patterns that look like API keys (sk-xxx, key=xxx, token=xxx).
	msg = redactSensitivePatterns(msg)
	return msg
}

// redactSensitivePatterns replaces common API key patterns with [REDACTED].
func redactSensitivePatterns(s string) string {
	patterns := []string{"key=", "token=", "sk-", "api_key=", "apikey="}
	for _, p := range patterns {
		idx := strings.Index(strings.ToLower(s), p)
		if idx != -1 {
			// Replace everything from the pattern to the end or next space.
			end := idx + len(p)
			for end < len(s) && s[end] != ' ' && s[end] != '"' && s[end] != '\'' {
				end++
			}
			s = s[:idx] + p + "[REDACTED]" + s[end:]
		}
	}
	return s
}

// safePreview returns a short preview of a string, suitable for logging
// without exposing full content.
func safePreview(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
