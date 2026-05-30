package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"offer-catcher/server/internal/agent/repair"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/llmjson"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// Runner implements the unified LLM Agent Pipeline:
//
//	Prompt Render → LLM Call → JSON extract/parse → Schema validate
//	→ if failed: JSONRepairAgent repair once
//	→ repair output parse + validate
//	→ if still failed: return fallback_required
//
// Runner does NOT:
//   - Call specific Mock Agents
//   - Modify business Agent logic
//   - Implement retry loops or recursive repair
type Runner struct {
	llmClient    llm.Client
	promptLoader *prompt.Loader
	schemaLoader *schema.Loader
	validator    *schema.Validator
	repairAgent  *repair.Agent
}

// NewRunner creates a new Pipeline Runner.
func NewRunner(
	llmClient llm.Client,
	promptLoader *prompt.Loader,
	schemaLoader *schema.Loader,
	validator *schema.Validator,
	repairAgent *repair.Agent,
) *Runner {
	return &Runner{
		llmClient:    llmClient,
		promptLoader: promptLoader,
		schemaLoader: schemaLoader,
		validator:    validator,
		repairAgent:  repairAgent,
	}
}

// Run executes the unified LLM Agent Pipeline for the given request.
//
// The pipeline always follows this order:
//  1. Validate RunRequest
//  2. Load and render Prompt
//  3. Load target Schema
//  4. Call llm.Client
//  5. JSON extract + parse
//  6. Schema validate
//  7. If parse/validate failed, call JSONRepairAgent repair once
//  8. Repair output parse + validate
//  9. Return RunResult with appropriate AgentStatus
func (r *Runner) Run(ctx context.Context, req RunRequest) *RunResult {
	start := time.Now()

	// Step 1: Validate request.
	if err := r.validateRequest(req); err != nil {
		return fallbackResult(FallbackRequestInvalid, ErrPipelineRequestInvalid, err.Error(), start)
	}

	// Step 2: Load and render Prompt.
	renderedPrompt, err := r.promptLoader.Render(req.PromptName, req.PromptValues)
	if err != nil {
		return fallbackResult(FallbackPromptLoadFailed, ErrPromptLoadFailed,
			fmt.Sprintf("failed to load/render prompt: %s", safeErr(err)), start)
	}

	// Step 3: Load target Schema.
	schemaDoc, err := r.schemaLoader.Load(req.SchemaName)
	if err != nil {
		return fallbackResult(FallbackSchemaLoadFailed, ErrSchemaLoadFailed,
			fmt.Sprintf("failed to load schema: %s", safeErr(err)), start)
	}

	// Step 4: Call llm.Client.
	chatResp, fbResult := r.callLLM(ctx, req, renderedPrompt, start)
	if fbResult != nil {
		return fbResult
	}

	// Build a base result from the LLM response.
	base := &RunResult{
		RawContent:      chatResp.RawContent,
		RawOutputLength: chatResp.RawOutputLength,
		LLMProvider:     string(chatResp.Provider),
		LLMModel:        chatResp.Model,
		FinishReason:    chatResp.FinishReason,
		DurationMs:      chatResp.DurationMs,
	}

	// Step 5: JSON extract + parse.
	parseResult := llmjson.ExtractAndParse(chatResp.RawContent)
	if !parseResult.Success {
		// Parse failed — attempt repair once.
		return r.attemptRepair(parseResult.RawOutput, "json_parse_failed",
			parseResult.ErrorMessage, schemaDoc, base, start)
	}

	// Step 6: Schema validate.
	valResult, valErr := r.validator.Validate(schemaDoc, parseResult.Parsed)
	if valErr != nil {
		// Validator compilation error — direct fallback.
		return fillFallback(base, FallbackSchemaValidationFailed, ErrSchemaValidationFailed,
			fmt.Sprintf("schema validation error: %s", safeErr(valErr)), start)
	}

	if !valResult.Success {
		// Schema validate failed — attempt repair once.
		validationSummary := valResult.ErrorMessage
		return r.attemptRepair(chatResp.RawContent, "schema_validation_failed",
			validationSummary, schemaDoc, base, start)
	}

	// Steps 9a: First-try success.
	base.AgentStatus = StatusSuccess
	base.GeneratedBy = "llm_agent"
	base.Parsed = parseResult.Parsed
	base.JSONText = parseResult.ExtractedJSON
	base.RepairApplied = false
	base.RepairSuccess = false
	base.FallbackRequired = false
	return base
}

// ---------------------------------------------------------------------------
// Step 1: Request validation
// ---------------------------------------------------------------------------

func (r *Runner) validateRequest(req RunRequest) error {
	if req.AgentName == "" {
		return fmt.Errorf("agent_name is required")
	}
	if req.PromptName == "" {
		return fmt.Errorf("prompt_name is required")
	}
	if req.SchemaName == "" {
		return fmt.Errorf("schema_name is required")
	}
	if !validTargetOutputTypes[req.TargetOutputType] {
		return fmt.Errorf("invalid target_output_type: %q (allowed: match_analysis, recommendation, resume_optimization)", req.TargetOutputType)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Step 4: LLM call
// ---------------------------------------------------------------------------

func (r *Runner) callLLM(ctx context.Context, req RunRequest, renderedPrompt string, start time.Time) (*llm.ChatResponse, *RunResult) {
	chatReq := llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "你是 Offer Catcher AI Agent，只输出严格 JSON，不要输出 Markdown、解释性前后缀或代码块。",
			},
			{
				Role:    llm.RoleUser,
				Content: renderedPrompt,
			},
		},
		Temperature: 0.2,
		MaxTokens:   8192,
		Metadata: map[string]string{
			"agent_name":         req.AgentName,
			"prompt_name":        string(req.PromptName),
			"schema_name":        string(req.SchemaName),
			"target_output_type": string(req.TargetOutputType),
		},
	}

	chatResp, err := r.llmClient.Chat(ctx, chatReq)
	if err != nil {
		return nil, r.handleLLMError(err, start)
	}

	if chatResp.RawContent == "" {
		return nil, &RunResult{
			AgentStatus:      StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   FallbackLLMCallFailed,
			ErrorCode:        ErrLLMEmptyResponse,
			ErrorMessage:     "LLM returned empty content",
			LLMProvider:      string(chatResp.Provider),
			LLMModel:         chatResp.Model,
			DurationMs:       msSince(start),
		}
	}

	return chatResp, nil
}

// handleLLMError maps LLM errors to fallback RunResults.
func (r *Runner) handleLLMError(err error, start time.Time) *RunResult {
	code := ErrLLMHTTPError
	reason := FallbackLLMCallFailed

	if llmErr, ok := err.(*llm.LLMError); ok {
		code = mapLLMErrorCodeToPipeline(string(llmErr.Code))
		reason = mapLLMErrorCodeToFallbackReason(string(llmErr.Code))
	}

	return &RunResult{
		AgentStatus:      StatusFallbackRequired,
		GeneratedBy:      "mock_agent",
		FallbackRequired: true,
		FallbackReason:   reason,
		ErrorCode:        code,
		ErrorMessage:     safeErr(err),
		DurationMs:       msSince(start),
	}
}

// ---------------------------------------------------------------------------
// Steps 7-8: Repair attempt
// ---------------------------------------------------------------------------

// attemptRepair calls JSONRepairAgent once and processes the result.
func (r *Runner) attemptRepair(
	brokenOutput string,
	failureType string,
	validationErrors string,
	schemaDoc *schema.Document,
	base *RunResult,
	start time.Time,
) *RunResult {
	repairReq := repair.RepairRequest{
		TargetOutputType: targetOutputTypeFromSchema(schemaDoc.Name),
		TargetSchemaJSON: schemaDoc.Content,
		BrokenOutput:     brokenOutput,
		ValidationErrors: fmt.Sprintf("%s: %s", failureType, validationErrors),
		RepairContext:    fmt.Sprintf("schema_name=%s", schemaDoc.Name),
	}

	repairResult := r.repairAgent.Repair(context.Background(), repairReq)

	return r.processRepairResult(repairResult, base)
}

// processRepairResult converts a repair.RepairResult into a pipeline.RunResult.
func (r *Runner) processRepairResult(rr *repair.RepairResult, base *RunResult) *RunResult {
	base.RepairApplied = true

	if rr.Success {
		base.AgentStatus = StatusRepaired
		base.GeneratedBy = "llm_agent"
		base.RepairSuccess = true
		base.FallbackRequired = false
		base.Parsed = rr.Parsed
		base.JSONText = rr.RepairedJSON
		return base
	}

	// Repair failed.
	base.AgentStatus = StatusFallbackRequired
	base.GeneratedBy = "mock_agent"
	base.RepairSuccess = false
	base.FallbackRequired = true
	base.ErrorCode = mapRepairErrorCode(rr.ErrorCode)
	base.FallbackReason = mapRepairErrorCodeToFallbackReason(rr.ErrorCode)
	base.ErrorMessage = rr.ErrorMessage
	return base
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func targetOutputTypeFromSchema(name schema.SchemaName) repair.TargetOutputType {
	switch name {
	case schema.SchemaMatchAnalysis:
		return repair.TargetMatchAnalysis
	case schema.SchemaRecommendation:
		return repair.TargetRecommendation
	case schema.SchemaResumeOptimization:
		return repair.TargetResumeOptimization
	default:
		return repair.TargetMatchAnalysis
	}
}

func mapRepairErrorCode(code string) string {
	switch code {
	case repair.ErrRepairParseFailed:
		return ErrJSONRepairParseFailed
	case repair.ErrRepairSchemaFailed:
		return ErrJSONRepairSchemaFailed
	case repair.ErrRepairLLMFailed:
		return ErrJSONRepairLLMFailed
	case repair.ErrRepairPromptLoadFailed, repair.ErrRepairSchemaLoadFailed, repair.ErrRepairTargetInvalid:
		return ErrJSONRepairFailed
	default:
		return ErrJSONRepairFailed
	}
}

func mapRepairErrorCodeToFallbackReason(code string) string {
	switch code {
	case repair.ErrRepairParseFailed:
		return FallbackJSONRepairParseFailed
	case repair.ErrRepairSchemaFailed:
		return FallbackJSONRepairSchemaFailed
	case repair.ErrRepairLLMFailed:
		return FallbackJSONRepairLLMFailed
	default:
		return FallbackJSONRepairFailed
	}
}

func fallbackResult(reason, code, msg string, start time.Time) *RunResult {
	return &RunResult{
		AgentStatus:      StatusFallbackRequired,
		GeneratedBy:      "mock_agent",
		FallbackRequired: true,
		FallbackReason:   reason,
		ErrorCode:        code,
		ErrorMessage:     msg,
		DurationMs:       msSince(start),
	}
}

func fillFallback(base *RunResult, reason, code, msg string, start time.Time) *RunResult {
	base.AgentStatus = StatusFallbackRequired
	base.GeneratedBy = "mock_agent"
	base.FallbackRequired = true
	base.FallbackReason = reason
	base.ErrorCode = code
	base.ErrorMessage = msg
	return base
}

func msSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

// safeErr returns a concise error message that does not leak API keys,
// full Prompts, or full model outputs.
func safeErr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	msg = redactSensitive(msg)
	return msg
}

// redactSensitive replaces common API key patterns with [REDACTED].
func redactSensitive(s string) string {
	patterns := []string{"key=", "token=", "sk-", "api_key=", "apikey="}
	for _, p := range patterns {
		idx := strings.Index(strings.ToLower(s), p)
		if idx != -1 {
			end := idx + len(p)
			for end < len(s) && s[end] != ' ' && s[end] != '"' && s[end] != '\'' {
				end++
			}
			s = s[:idx] + p + "[REDACTED]" + s[end:]
		}
	}
	return s
}
