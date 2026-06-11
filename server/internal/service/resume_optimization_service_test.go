package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// fakeResumeOptSvcPipelineRunner for service-level tests
// ---------------------------------------------------------------------------

type fakeResumeOptSvcPipelineRunner struct {
	result *pipeline.RunResult
}

func (f *fakeResumeOptSvcPipelineRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return f.result
}

// ---------------------------------------------------------------------------
// Existing Mock-only tests (updated for new constructor signature)
// ---------------------------------------------------------------------------

func TestResumeOptimizationService_Optimize_NormalRequest(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go、MySQL、Redis，参与过后台管理系统开发，负责用户模块和权限模块的接口开发。",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发工程师",
			JDText: "熟悉 Go 语言，了解微服务架构，有数据库和缓存经验",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", resp.GeneratedBy)
	}
	if resp.Result == nil {
		t.Fatal("expected non-nil result")
	}
	if resp.Result.ApplyReadiness == "" {
		t.Error("expected non-empty apply_readiness")
	}
	if resp.Result.OptimizationStopSignal == "" {
		t.Error("expected non-empty optimization_stop_signal")
	}
	if resp.Result.VerificationNotice == "" {
		t.Error("expected non-empty verification_notice")
	}
}

func TestResumeOptimizationService_Optimize_EmptyResumeText(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发",
		},
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for empty resume_text")
	}
	var ae *apperror.AppError; if !errors.As(err, &ae) || ae.Message != "resume_text is required" {
		t.Errorf("expected 'resume_text is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_NoTargetJob(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for missing target_job")
	}
	var ae2 *apperror.AppError; if !errors.As(err, &ae2) || ae2.Message != "target_job is required" {
		t.Errorf("expected 'target_job is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_NoTitleOrJDText(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
		TargetJob:  &model.ResumeOptimizationTargetJob{},
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for missing title and jd_text")
	}
	var ae3 *apperror.AppError; if !errors.As(err, &ae3) || ae3.Message != "target_job.title or target_job.jd_text is required" {
		t.Errorf("expected 'target_job.title or target_job.jd_text is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_ShortResumeInsufficientInfo(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "学生",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Result.ApplyReadiness != "insufficient_information" {
		t.Errorf("expected apply_readiness=insufficient_information for short resume, got %s", resp.Result.ApplyReadiness)
	}
}

func TestResumeOptimizationService_Optimize_WithUserIntention(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 和 MySQL，做过项目。",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发",
			JDText: "Go 语言 后端开发",
		},
		UserIntention: &model.ResumeOptimizationUserIntention{
			AcceptRelocation:  true,
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", resp.Result.GeneratedBy)
	}
}

func TestResumeOptimizationService_Optimize_TopFixItemsMax3(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go，做过后台项目，了解 MySQL 和 Redis。",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发工程师",
			JDText: "Go 微服务 MySQL Redis Docker 分布式 缓存 接口开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Result.TopFixItems) > 3 {
		t.Errorf("expected at most 3 top_fix_items, got %d", len(resp.Result.TopFixItems))
	}
}

func TestResumeOptimizationService_Optimize_NoSuggestChangeTo(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "做过一些 Go 项目",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "后端开发",
			JDText: "后端开发 接口开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, item := range resp.Result.TopFixItems {
		if strings.Contains(item.ExampleReference, "建议改为") {
			t.Errorf("example_reference should not contain '建议改为': %s", item.ExampleReference)
		}
		if strings.Contains(item.Action, "建议改为") {
			t.Errorf("action should not contain '建议改为': %s", item.Action)
		}
	}
}

func TestResumeOptimizationService_Optimize_NoFabricatedMetrics(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go，做过项目。",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "后端开发",
			JDText: "Go 后端开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, item := range resp.Result.TopFixItems {
		if strings.Contains(item.ExampleReference, "QPS") {
			t.Errorf("example_reference should not contain fabricated QPS: %s", item.ExampleReference)
		}
	}
}

func TestResumeOptimizationService_Optimize_WithJDTextOnly(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发，做过后台管理系统",
		TargetJob: &model.ResumeOptimizationTargetJob{
			JDText: "熟悉 Go 语言，了解微服务架构，有数据库和缓存经验",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", resp.Result.GeneratedBy)
	}
}

func TestResumeOptimizationService_Optimize_WithTitleOnly(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发工程师",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", resp.Result.GeneratedBy)
	}
}

func TestResumeOptimizationService_Optimize_WhitespaceOnlyResumeText(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "   ",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发",
		},
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for whitespace-only resume_text")
	}
}

// ---------------------------------------------------------------------------
// LLM + fallback tests
// ---------------------------------------------------------------------------

func TestResumeOptimizationService_LLMDisabled_UseMock(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	// llmEnabled=false — should use Mock regardless of pipelineRunner
	runner := &fakeResumeOptSvcPipelineRunner{}
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, runner, runner, false, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发，做过后台系统",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent when llm disabled, got %s", resp.GeneratedBy)
	}
}

func TestResumeOptimizationService_LLMEnabled_PipelineSuccess(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	runner := &fakeResumeOptSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed: map[string]any{
				"optimization_summary":      "建议优化后投递",
				"apply_readiness":          "quick_fix_before_apply",
				"optimization_stop_signal": "fix_top_3_then_apply",
				"top_fix_items":            []any{},
				"authenticity_notice":      "请基于真实经历修改",
			},
			FallbackRequired: false,
		},
	}
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, runner, runner, true, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发，做过项目",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发",
			JDText: "Go 微服务",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent on LLM success, got %s", resp.GeneratedBy)
	}
	if resp.Result.ApplyReadiness != "quick_fix_before_apply" {
		t.Errorf("expected apply_readiness=quick_fix_before_apply, got %s", resp.Result.ApplyReadiness)
	}
	if resp.Result.OptimizationStopSignal != "fix_top_3_then_apply" {
		t.Errorf("expected optimization_stop_signal=fix_top_3_then_apply, got %s", resp.Result.OptimizationStopSignal)
	}
}

func TestResumeOptimizationService_LLMEnabled_PipelineFallbackToMock(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	runner := &fakeResumeOptSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_CALL_FAILED",
			ErrorMessage:     "timeout",
		},
	}
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, runner, runner, true, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发，做过后台系统",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发",
			JDText: "Go 微服务",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should fall back to mock
	if resp.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent on fallback, got %s", resp.GeneratedBy)
	}
	// Should indicate fallback status
	if resp.Result.AgentStatus != "fallback" {
		t.Errorf("expected agent_status=fallback, got %s", resp.Result.AgentStatus)
	}
	if resp.Result.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got %s", resp.Result.FallbackReason)
	}
	// Critical fields should still be present
	if resp.Result.ApplyReadiness == "" {
		t.Error("expected non-empty apply_readiness after fallback")
	}
	if resp.Result.OptimizationStopSignal == "" {
		t.Error("expected non-empty optimization_stop_signal after fallback")
	}
	if resp.Result.TopFixItems == nil {
		t.Error("expected non-nil top_fix_items after fallback")
	}
}

func TestResumeOptimizationService_LLMEnabled_NilRunner_UseMock(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	// pipelineRunner is nil — should fall back to Mock
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, nil, nil, true, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title: "Go 后端开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent when pipelineRunner is nil, got %s", resp.GeneratedBy)
	}
}

func TestResumeOptimizationService_LLMFallback_PreservesTopFixItems(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	runner := &fakeResumeOptSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "schema_validation_failed",
			ErrorCode:        "SCHEMA_VALIDATION_FAILED",
		},
	}
	svc := NewResumeOptimizationService(nil, nil, nil, mockAgent, runner, runner, true, true, true)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go、MySQL，做过后台系统开发，负责接口开发",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发工程师",
			JDText: "Go 微服务 MySQL Redis Docker 接口开发",
		},
	}

	resp, err := svc.Optimize(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Fallback to mock should still have critical fields
	if resp.Result.ApplyReadiness == "" {
		t.Error("expected non-empty apply_readiness after fallback")
	}
	if resp.Result.OptimizationStopSignal == "" {
		t.Error("expected non-empty optimization_stop_signal after fallback")
	}
	if resp.Result.TopFixItems == nil {
		t.Error("expected non-nil top_fix_items after fallback")
	}
	if len(resp.Result.TopFixItems) > 3 {
		t.Errorf("expected at most 3 top_fix_items, got %d", len(resp.Result.TopFixItems))
	}
	if resp.Result.VerificationNotice == "" {
		t.Error("expected non-empty verification_notice after fallback")
	}
}
