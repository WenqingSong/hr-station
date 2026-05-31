package agent

import (
	"context"
	"testing"

	"offer-catcher/server/internal/agent/pipeline"
)

// ---------------------------------------------------------------------------
// fakeResumeOptPipelineRunner for testing
// ---------------------------------------------------------------------------

type fakeResumeOptPipelineRunner struct {
	result *pipeline.RunResult
}

func (r *fakeResumeOptPipelineRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return r.result
}

// ---------------------------------------------------------------------------
// ResumeOptimizationAgent tests
// ---------------------------------------------------------------------------

func TestResumeOptimizationAgent_LLMSuccess(t *testing.T) {
	parsedOutput := map[string]any{
		"optimization_summary": "当前简历与目标岗位有一定匹配基础，建议先完成 2 个快速优化项后投递。",
		"apply_readiness":     "quick_fix_before_apply",
		"optimization_stop_signal": "fix_top_3_then_apply",
		"top_fix_items": []any{
			map[string]any{
				"issue":              "项目职责描述不够清晰",
				"why_it_matters":     "当前表达难以证明个人贡献",
				"action":             "补充负责的模块和技术栈",
				"example_reference":  "示例参考：基于 Go 完成接口开发",
				"self_check_question": "请确认你是否确实负责过这些模块",
				"learn_more_query":   "Go 后端简历项目经历写法",
				"estimated_effort":   "within_30_min",
			},
		},
		"priority_suggestions": []any{
			map[string]any{
				"suggestion": "后续可补充压测或性能优化经验",
				"note":       "这属于长期提升方向，不建议延后投递",
			},
		},
		"keyword_alignment": []any{
			map[string]any{
				"jd_keyword":       "Go 语言",
				"resume_keyword":   "Go",
				"alignment_status": "partial_match",
				"gap_analysis":     "有相关技术栈，但表达不够具体",
				"screening_risk":   "responsibility_unclear",
			},
		},
		"authenticity_notice": "以下建议仅用于表达优化，请基于真实经历修改。",
	}

	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsedOutput,
			FallbackRequired: false,
			RepairApplied:    false,
			DurationMs:       500,
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText:   "熟悉 Go、MySQL，做过后台系统开发",
		JobTitle:     "Go 后端开发工程师",
		JDText:       "Go 微服务 MySQL Redis",
		TargetSection: "project",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got %s", result.GeneratedBy)
	}
	if result.ApplyReadiness != "quick_fix_before_apply" {
		t.Errorf("expected apply_readiness=quick_fix_before_apply, got %s", result.ApplyReadiness)
	}
	if result.OptimizationStopSignal != "fix_top_3_then_apply" {
		t.Errorf("expected optimization_stop_signal=fix_top_3_then_apply, got %s", result.OptimizationStopSignal)
	}
	if len(result.TopFixItems) == 0 {
		t.Error("expected non-empty top_fix_items")
	}
	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if result.VerificationNotice == "" {
		t.Error("expected non-empty verification_notice")
	}
}

func TestResumeOptimizationAgent_FallbackRequired(t *testing.T) {
	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_CALL_FAILED",
			ErrorMessage:     "timeout",
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText: "熟悉 Go 开发",
		JobTitle:   "Go 后端开发",
		JDText:     "Go 微服务",
	}

	_, err := agent.Optimize(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when fallback is required")
	}

	fbErr, ok := err.(*ResumeOptFallbackError)
	if !ok {
		t.Fatalf("expected *ResumeOptFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_CALL_FAILED" {
		t.Errorf("expected error_code=LLM_CALL_FAILED, got %s", fbErr.ErrorCode)
	}
}

func TestResumeOptimizationAgent_RepairedSuccess(t *testing.T) {
	parsedOutput := map[string]any{
		"optimization_summary":     "建议优化后投递。",
		"apply_readiness":         "quick_fix_before_apply",
		"optimization_stop_signal": "fix_top_3_then_apply",
		"top_fix_items":           []any{},
	}

	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusRepaired,
			GeneratedBy:      "llm_agent",
			Parsed:           parsedOutput,
			FallbackRequired: false,
			RepairApplied:    true,
			RepairSuccess:    true,
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText: "Go 开发经验",
		JobTitle:   "后端开发",
		JDText:     "Go 语言后端",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got %s", result.GeneratedBy)
	}
	if !result.RepairApplied {
		t.Error("expected repair_applied=true")
	}
	if result.ApplyReadiness != "quick_fix_before_apply" {
		t.Errorf("expected apply_readiness=quick_fix_before_apply, got %s", result.ApplyReadiness)
	}
}

func TestResumeOptimizationAgent_MappingFailure(t *testing.T) {
	// Parsed is nil — should trigger mapping failure → fallback
	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           nil,
			FallbackRequired: false,
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText: "Go 开发经验",
		JobTitle:   "后端开发",
		JDText:     "Go 语言",
	}

	_, err := agent.Optimize(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for nil parsed output")
	}

	fbErr, ok := err.(*ResumeOptFallbackError)
	if !ok {
		t.Fatalf("expected *ResumeOptFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "field_mapping_failed" {
		t.Errorf("expected fallback_reason=field_mapping_failed, got %s", fbErr.FallbackReason)
	}
}

func TestResumeOptimizationAgent_PreservesCriticalFields(t *testing.T) {
	parsedOutput := map[string]any{
		"apply_readiness":         "ready_to_apply",
		"optimization_stop_signal": "stop_and_apply",
		"top_fix_items": []any{
			map[string]any{
				"issue":            "优化建议1",
				"why_it_matters":   "原因1",
				"action":           "操作1",
				"example_reference": "示例1",
				"estimated_effort": "within_30_min",
			},
		},
	}

	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsedOutput,
			FallbackRequired: false,
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText: "丰富经验",
		JobTitle:   "Go 后端",
		JDText:     "Go 微服务",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ApplyReadiness != "ready_to_apply" {
		t.Errorf("expected apply_readiness=ready_to_apply, got %s", result.ApplyReadiness)
	}
	if result.OptimizationStopSignal != "stop_and_apply" {
		t.Errorf("expected optimization_stop_signal=stop_and_apply, got %s", result.OptimizationStopSignal)
	}
	if len(result.TopFixItems) != 1 {
		t.Errorf("expected 1 top_fix_item, got %d", len(result.TopFixItems))
	}
	if result.TopFixItems[0].Issue != "优化建议1" {
		t.Errorf("expected issue=优化建议1, got %s", result.TopFixItems[0].Issue)
	}
}

func TestResumeOptFallbackError_Error(t *testing.T) {
	err := &ResumeOptFallbackError{
		FallbackReason: "llm_call_failed",
		ErrorCode:      "LLM_CALL_FAILED",
		ErrorMessage:   "timeout after 30s",
	}

	msg := err.Error()
	if msg != "resume_optimization fallback required: reason=llm_call_failed code=LLM_CALL_FAILED msg=timeout after 30s" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestResumeOptimizationAgent_EnsureInterface(t *testing.T) {
	// Compile-time check that ResumeOptimizationAgent implements ResumeOptimizer
	var _ ResumeOptimizer = (*ResumeOptimizationAgent)(nil)
	// Compile-time check that ResumeOptimizationMockAgent implements ResumeOptimizer
	var _ ResumeOptimizer = (*ResumeOptimizationMockAgent)(nil)
}

func TestResumeOptimizationAgent_DefaultFieldsWhenMissing(t *testing.T) {
	// Minimal parsed output — missing most fields
	parsedOutput := map[string]any{}

	runner := &fakeResumeOptPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsedOutput,
			FallbackRequired: false,
		},
	}

	agent := NewResumeOptimizationAgent(runner)
	input := &OptimizeInput{
		ResumeText: "Go 开发经验",
		JobTitle:   "Go 后端",
		JDText:     "Go 语言",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have defaults for critical fields
	if result.ApplyReadiness != "insufficient_information" {
		t.Errorf("expected default apply_readiness=insufficient_information, got %s", result.ApplyReadiness)
	}
	if result.OptimizationStopSignal != "needs_more_evidence" {
		t.Errorf("expected default optimization_stop_signal=needs_more_evidence, got %s", result.OptimizationStopSignal)
	}
	if result.TopFixItems == nil {
		t.Error("expected non-nil top_fix_items (can be empty slice)")
	}
	if result.JDEvidenceChain == nil {
		t.Error("expected non-nil jd_evidence_chain")
	}
	if result.LongTermImprovement == nil {
		t.Error("expected non-nil long_term_improvement")
	}
	if result.VerificationNotice == "" {
		t.Error("expected non-empty verification_notice")
	}
}
