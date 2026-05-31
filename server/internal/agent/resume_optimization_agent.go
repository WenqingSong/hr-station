package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// ResumeOptimizer defines the interface for resume optimization generation.
// Both ResumeOptimizationMockAgent and ResumeOptimizationAgent implement this interface.
type ResumeOptimizer interface {
	// Optimize generates a resume optimization for the given input.
	Optimize(ctx context.Context, input *OptimizeInput) (*model.ResumeOptimizationResult, error)
}

// Ensure ResumeOptimizationMockAgent implements ResumeOptimizer.
var _ ResumeOptimizer = (*ResumeOptimizationMockAgent)(nil)

// ResumeOptimizationAgent uses the LLM Pipeline to generate resume optimization results.
// On Pipeline success/repaired, it maps the LLM output to model.ResumeOptimizationResult.
// On Pipeline fallback_required, it returns an error so the caller can fall back to Mock.
type ResumeOptimizationAgent struct {
	runner PipelineRunner
}

// NewResumeOptimizationAgent creates a new ResumeOptimizationAgent with the given Pipeline Runner.
func NewResumeOptimizationAgent(runner PipelineRunner) *ResumeOptimizationAgent {
	return &ResumeOptimizationAgent{runner: runner}
}

// Ensure ResumeOptimizationAgent implements ResumeOptimizer.
var _ ResumeOptimizer = (*ResumeOptimizationAgent)(nil)

// resumeOptimizationBusinessContext is the business context injected into the Prompt.
const resumeOptimizationBusinessContext = `你正在进行学生求职场景下的简历优化建议。
必须基于输入的简历内容和目标岗位 JD 判断。
不要编造经历或指标。
可以给出优化方向和可替换表达，但必须提醒用户基于真实经历修改。
输出必须是严格 JSON。`

// Optimize calls the LLM Pipeline and maps the result to ResumeOptimizationResult.
// If the Pipeline returns fallback_required or the mapping fails, it returns an error.
func (a *ResumeOptimizationAgent) Optimize(ctx context.Context, input *OptimizeInput) (*model.ResumeOptimizationResult, error) {
	// Step 1: Assemble PromptValues
	promptValues, err := a.buildPromptValues(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt values: %w", err)
	}

	// Step 2: Call Pipeline
	runReq := pipeline.RunRequest{
		AgentName:        "ResumeOptimizationAgent",
		PromptName:       prompt.PromptResumeOptimization,
		SchemaName:       schema.SchemaResumeOptimization,
		TargetOutputType: pipeline.TargetResumeOptimization,
		PromptValues:     promptValues,
		Metadata: map[string]string{
			"job_title": input.JobTitle,
		},
	}

	runResult := a.runner.Run(ctx, runReq)

	// Step 3: Log pipeline result (safe, no API key or full content)
	log.Printf("[ResumeOptimizationAgent] pipeline result: status=%s generated_by=%s repair_applied=%v fallback=%v reason=%s error_code=%s duration_ms=%d raw_output_length=%d",
		runResult.AgentStatus, runResult.GeneratedBy, runResult.RepairApplied,
		runResult.FallbackRequired, runResult.FallbackReason, runResult.ErrorCode,
		runResult.DurationMs, runResult.RawOutputLength)

	// Step 4: Check if fallback is required
	if runResult.FallbackRequired {
		return nil, &ResumeOptFallbackError{
			FallbackReason: runResult.FallbackReason,
			ErrorCode:      runResult.ErrorCode,
			ErrorMessage:   runResult.ErrorMessage,
		}
	}

	// Step 5: Map Parsed to ResumeOptimizationResult
	result, err := mapParsedToResumeOptResult(runResult.Parsed, input)
	if err != nil {
		// Mapping failure should also trigger fallback
		return nil, &ResumeOptFallbackError{
			FallbackReason: "field_mapping_failed",
			ErrorCode:      "FIELD_MAPPING_FAILED",
			ErrorMessage:   fmt.Sprintf("failed to map LLM output: %s", err),
		}
	}

	// Step 6: Set LLM-specific metadata on the result
	result.GeneratedBy = runResult.GeneratedBy
	result.AgentStatus = string(runResult.AgentStatus)
	result.RepairApplied = runResult.RepairApplied
	result.FallbackReason = ""
	result.ErrorCode = ""

	return result, nil
}

// ResumeOptFallbackError indicates the Pipeline returned fallback_required or mapping failed.
// The caller should fall back to the Mock Agent.
type ResumeOptFallbackError struct {
	FallbackReason string
	ErrorCode      string
	ErrorMessage   string
}

func (e *ResumeOptFallbackError) Error() string {
	return fmt.Sprintf("resume_optimization fallback required: reason=%s code=%s msg=%s",
		e.FallbackReason, e.ErrorCode, e.ErrorMessage)
}

// buildPromptValues assembles the placeholders for resume_optimization_v1.md.
func (a *ResumeOptimizationAgent) buildPromptValues(input *OptimizeInput) (map[string]string, error) {
	// RESUME_CONTEXT_JSON: serialize resume text and context
	resumeContext := map[string]any{
		"resume_text":     input.ResumeText,
		"target_section":  input.TargetSection,
		"target_roles":    input.TargetRoles,
		"expected_locations": input.ExpectedLocations,
		"accept_relocation":  input.AcceptRelocation,
	}
	resumeJSON, err := json.Marshal(resumeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resume context: %w", err)
	}

	// TARGET_JOB_CONTEXT_JSON: serialize target job
	jobContext := map[string]any{
		"job_title": input.JobTitle,
		"jd_text":   input.JDText,
	}
	jobJSON, err := json.Marshal(jobContext)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target job context: %w", err)
	}

	return map[string]string{
		"RESUME_CONTEXT_JSON":      string(resumeJSON),
		"TARGET_JOB_CONTEXT_JSON":  string(jobJSON),
		"MATCH_ANALYSIS_JSON":      "{}", // No pre-existing match analysis in basic flow
		"BUSINESS_CONTEXT":         resumeOptimizationBusinessContext,
	}, nil
}

// ---------------------------------------------------------------------------
// Field mapping: RunResult.Parsed (map[string]any) → model.ResumeOptimizationResult
// ---------------------------------------------------------------------------

// mapParsedToResumeOptResult maps the Pipeline's parsed JSON output to a ResumeOptimizationResult.
func mapParsedToResumeOptResult(parsed map[string]any, input *OptimizeInput) (*model.ResumeOptimizationResult, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed output is nil")
	}

	result := &model.ResumeOptimizationResult{}

	// optimization_summary
	result.Summary = getString(parsed, "optimization_summary", "")

	// offer_go_summary (LLM-only field, stored in summary if present)
	offerGoSummary := getString(parsed, "offer_go_summary", "")
	if result.Summary == "" && offerGoSummary != "" {
		result.Summary = offerGoSummary
	}

	// apply_readiness (4-3.7.C.1)
	result.ApplyReadiness = getString(parsed, "apply_readiness", "insufficient_information")

	// optimization_stop_signal (4-3.7.C.1)
	result.OptimizationStopSignal = getString(parsed, "optimization_stop_signal", "needs_more_evidence")

	// top_fix_items (4-3.7.C.1)
	result.TopFixItems = mapResumeOptTopFixItems(parsed)

	// screening_focus (LLM-only field)
	// Not in model, skip

	// priority_suggestions → long_term_improvement
	result.LongTermImprovement = mapResumeOptLongTermItems(parsed)

	// keyword_alignment → jd_evidence_chain
	result.JDEvidenceChain = mapResumeOptKeywordAlignment(parsed)

	// project_optimization (LLM-only enrichment, not directly mapped)

	// skill_section_optimization (LLM-only enrichment, not directly mapped)

	// summary_optimization (LLM-only enrichment, not directly mapped)

	// missing_evidence_plan (LLM-only enrichment, not directly mapped)

	// rewrite_examples (LLM-only enrichment, not directly mapped)

	// next_actions
	// Not directly in model, skip

	// authenticity_notice → verification_notice
	result.VerificationNotice = getString(parsed, "authenticity_notice",
		"以下建议仅作为简历表达和投递策略参考，AI 可能理解错误。请基于你的真实经历核实后使用，不要加入没有实际发生过的经历、指标或职责。")

	// If summary is still empty, generate a default
	if result.Summary == "" {
		result.Summary = generateOptimizationSummary(result.ApplyReadiness, len(result.TopFixItems), input.JobTitle)
	}

	// Ensure non-empty critical fields
	if result.ApplyReadiness == "" {
		result.ApplyReadiness = "insufficient_information"
	}
	if result.OptimizationStopSignal == "" {
		result.OptimizationStopSignal = "needs_more_evidence"
	}
	if result.TopFixItems == nil {
		result.TopFixItems = []model.ResumeOptimizationFixItem{}
	}
	if result.JDEvidenceChain == nil {
		result.JDEvidenceChain = []model.ResumeOptimizationJDEvidenceItem{}
	}
	if result.LongTermImprovement == nil {
		result.LongTermImprovement = []model.ResumeOptimizationLongTermItem{}
	}

	return result, nil
}

// mapResumeOptTopFixItems maps top_fix_items from Schema to []model.ResumeOptimizationFixItem.
// Schema: [{issue, why_it_matters, action, example_reference, self_check_question, learn_more_query, estimated_effort}]
func mapResumeOptTopFixItems(parsed map[string]any) []model.ResumeOptimizationFixItem {
	items, ok := getSlice(parsed, "top_fix_items")
	if !ok || len(items) == 0 {
		return []model.ResumeOptimizationFixItem{}
	}

	var fixItems []model.ResumeOptimizationFixItem
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fixItems = append(fixItems, model.ResumeOptimizationFixItem{
			Issue:             getString(m, "issue", ""),
			WhyItMatters:      getString(m, "why_it_matters", ""),
			Action:            getString(m, "action", ""),
			ExampleReference:  getString(m, "example_reference", ""),
			SelfCheckQuestion: getString(m, "self_check_question", ""),
			LearnMoreQuery:    getString(m, "learn_more_query", ""),
			EstimatedEffort:   getString(m, "estimated_effort", ""),
		})
	}
	return fixItems
}

// mapResumeOptLongTermItems maps priority_suggestions from Schema to []model.ResumeOptimizationLongTermItem.
// Schema: [{suggestion, note, priority}]
func mapResumeOptLongTermItems(parsed map[string]any) []model.ResumeOptimizationLongTermItem {
	items, ok := getSlice(parsed, "priority_suggestions")
	if !ok || len(items) == 0 {
		return []model.ResumeOptimizationLongTermItem{}
	}

	var longTermItems []model.ResumeOptimizationLongTermItem
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		direction := getString(m, "suggestion", "")
		note := getString(m, "note", "")
		if direction == "" {
			continue
		}
		if note == "" {
			note = "这属于长期提升方向，不建议因为这一点延后当前投递。"
		}
		longTermItems = append(longTermItems, model.ResumeOptimizationLongTermItem{
			Direction: direction,
			Note:      note,
		})
	}
	return longTermItems
}

// mapResumeOptKeywordAlignment maps keyword_alignment from Schema to []model.ResumeOptimizationJDEvidenceItem.
// Schema: [{jd_keyword, resume_keyword, alignment_status, gap_analysis, screening_risk}]
func mapResumeOptKeywordAlignment(parsed map[string]any) []model.ResumeOptimizationJDEvidenceItem {
	items, ok := getSlice(parsed, "keyword_alignment")
	if !ok || len(items) == 0 {
		return []model.ResumeOptimizationJDEvidenceItem{}
	}

	var evidenceItems []model.ResumeOptimizationJDEvidenceItem
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		evidenceItems = append(evidenceItems, model.ResumeOptimizationJDEvidenceItem{
			JDRequirement:    getString(m, "jd_keyword", ""),
			ResumeEvidence:   getString(m, "resume_keyword", ""),
			EvidenceStrength: mapAlignmentStatusToStrength(getString(m, "alignment_status", "")),
			MatchJudgement:   mapAlignmentStatusToJudgement(getString(m, "alignment_status", "")),
			GapAnalysis:      getString(m, "gap_analysis", ""),
			ScreeningRisk:    getString(m, "screening_risk", ""),
		})
	}
	return evidenceItems
}

// mapAlignmentStatusToStrength maps LLM alignment_status to evidence_strength.
func mapAlignmentStatusToStrength(status string) string {
	switch status {
	case "matched", "strong_match":
		return "strong"
	case "partial_match", "partial":
		return "medium"
	case "missing", "no_match":
		return "missing"
	default:
		return "weak"
	}
}

// mapAlignmentStatusToJudgement maps LLM alignment_status to match_judgement.
func mapAlignmentStatusToJudgement(status string) string {
	switch status {
	case "matched", "strong_match":
		return "partially_matched"
	case "partial_match", "partial":
		return "partially_matched"
	case "missing", "no_match":
		return "missing"
	default:
		return "partially_matched"
	}
}
