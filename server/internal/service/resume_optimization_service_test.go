package service

import (
	"context"
	"strings"
	"testing"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
)

func TestResumeOptimizationService_Optimize_NormalRequest(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	if err.Error() != "resume_text is required" {
		t.Errorf("expected 'resume_text is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_NoTargetJob(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, mockAgent)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for missing target_job")
	}
	if err.Error() != "target_job is required" {
		t.Errorf("expected 'target_job is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_NoTitleOrJDText(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, mockAgent)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 开发",
		TargetJob:  &model.ResumeOptimizationTargetJob{},
	}

	_, err := svc.Optimize(context.Background(), 1, req)
	if err == nil {
		t.Fatal("expected error for missing title and jd_text")
	}
	if err.Error() != "target_job.title or target_job.jd_text is required" {
		t.Errorf("expected 'target_job.title or target_job.jd_text is required' error, got: %v", err)
	}
}

func TestResumeOptimizationService_Optimize_ShortResumeInsufficientInfo(t *testing.T) {
	mockAgent := agent.NewResumeOptimizationMockAgent()
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

	req := ResumeOptimizationRequest{
		ResumeText: "熟悉 Go 和 MySQL，做过项目。",
		TargetJob: &model.ResumeOptimizationTargetJob{
			Title:  "Go 后端开发",
			JDText: "Go 语言 后端开发",
		},
		UserIntention: &model.ResumeOptimizationUserIntention{
			TargetRoles:      []string{"Go 后端开发"},
			ExpectedLocations: []string{"深圳"},
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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
	svc := NewResumeOptimizationService(nil, mockAgent)

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
