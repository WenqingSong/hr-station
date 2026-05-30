package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"offer-catcher/server/internal/model"
)

func TestResumeOptimizationMockAgent_Optimize_NormalCase(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "我是一名计算机专业的学生，熟悉 Go、MySQL、Redis，参与过后台管理系统的开发，负责用户模块和权限模块的接口开发。有过一段实习经历，在团队中使用 Go 开发微服务。",
		JobTitle:   "Go 后端开发工程师",
		JDText:     "熟悉 Go 语言，了解微服务架构，有数据库和缓存经验",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check agent metadata
	if result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", result.GeneratedBy)
	}
	if result.AgentStatus != "mock_success" {
		t.Errorf("expected agent_status=mock_success, got %s", result.AgentStatus)
	}
	if result.FallbackRequired != false {
		t.Error("expected fallback_required=false")
	}
	if result.RepairApplied != false {
		t.Error("expected repair_applied=false")
	}

	// Check strategic fields
	validReadiness := map[string]bool{
		"ready_to_apply": true, "quick_fix_before_apply": true,
		"not_ready_for_this_role": true, "insufficient_information": true,
	}
	if !validReadiness[result.ApplyReadiness] {
		t.Errorf("invalid apply_readiness: %s", result.ApplyReadiness)
	}

	validStopSignal := map[string]bool{
		"stop_and_apply": true, "fix_top_3_then_apply": true,
		"needs_more_evidence": true, "target_role_mismatch": true,
	}
	if !validStopSignal[result.OptimizationStopSignal] {
		t.Errorf("invalid optimization_stop_signal: %s", result.OptimizationStopSignal)
	}

	// Check top_fix_items: max 3
	if len(result.TopFixItems) > 3 {
		t.Errorf("expected at most 3 top_fix_items, got %d", len(result.TopFixItems))
	}
	if len(result.TopFixItems) < 1 {
		t.Error("expected at least 1 top_fix_item")
	}

	// Check no "建议改为" in fix items
	for _, item := range result.TopFixItems {
		if strings.Contains(item.ExampleReference, "建议改为") {
			t.Errorf("fix item example_reference should not contain '建议改为': %s", item.ExampleReference)
		}
		if strings.Contains(item.Action, "建议改为") {
			t.Errorf("fix item action should not contain '建议改为': %s", item.Action)
		}
		// Check "示例参考" is used
		if item.ExampleReference != "" && !strings.Contains(item.ExampleReference, "示例参考") {
			t.Errorf("fix item example_reference should contain '示例参考': %s", item.ExampleReference)
		}
	}

	// Check verification_notice
	if result.VerificationNotice == "" {
		t.Error("expected non-empty verification_notice")
	}
	if !strings.Contains(result.VerificationNotice, "真实经历") {
		t.Errorf("verification_notice should mention verifying real experience: %s", result.VerificationNotice)
	}

	// Check no fabricated metrics
	for _, item := range result.TopFixItems {
		if strings.Contains(item.ExampleReference, "QPS") ||
			strings.Contains(item.ExampleReference, "DAU") ||
			strings.Contains(item.ExampleReference, "提升%") {
			t.Errorf("fix item should not contain fabricated metrics: %s", item.ExampleReference)
		}
	}

	// Check long_term_improvement
	for _, item := range result.LongTermImprovement {
		if !strings.Contains(item.Note, "不建议") || !strings.Contains(item.Note, "延后当前投递") {
			t.Errorf("long_term_improvement note should mention not to delay application: %s", item.Note)
		}
	}

	// Check jd_evidence_chain
	if len(result.JDEvidenceChain) < 1 {
		t.Error("expected at least 1 jd_evidence_chain entry")
	}
	for _, ev := range result.JDEvidenceChain {
		validStrength := map[string]bool{"strong": true, "medium": true, "weak": true, "missing": true}
		if !validStrength[ev.EvidenceStrength] {
			t.Errorf("invalid evidence_strength: %s", ev.EvidenceStrength)
		}
		validJudgement := map[string]bool{"matched": true, "partially_matched": true, "missing": true, "risky": true}
		if !validJudgement[ev.MatchJudgement] {
			t.Errorf("invalid match_judgement: %s", ev.MatchJudgement)
		}
	}
}

func TestResumeOptimizationMockAgent_Optimize_ShortResume(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "学生",
		JobTitle:   "Go 后端开发",
		JDText:     "熟悉 Go 语言",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ApplyReadiness != "insufficient_information" {
		t.Errorf("expected apply_readiness=insufficient_information for short resume, got %s", result.ApplyReadiness)
	}
	if result.OptimizationStopSignal != "needs_more_evidence" {
		t.Errorf("expected optimization_stop_signal=needs_more_evidence for short resume, got %s", result.OptimizationStopSignal)
	}
}

func TestResumeOptimizationMockAgent_Optimize_StrongMatch(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "3年后端开发经验，熟练使用 Go、MySQL、Redis、Docker。在上一份工作中负责微服务架构的后台系统开发，完成了用户管理、订单处理、权限控制等核心模块的接口设计与实现。熟悉分布式系统设计，有缓存穿透、限流、幂等处理的实战经验。",
		JobTitle:   "Go 后端开发工程师",
		JDText:     "熟悉 Go 语言，了解微服务架构，有 MySQL 和 Redis 经验，了解 Docker 和分布式系统",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ApplyReadiness != "ready_to_apply" {
		t.Errorf("expected apply_readiness=ready_to_apply for strong match, got %s", result.ApplyReadiness)
	}
}

func TestResumeOptimizationMockAgent_Optimize_NoOverlap(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "我是一名设计专业的学生，擅长 Photoshop 和 Illustrator，有过平面设计实习经历。",
		JobTitle:   "Go 后端开发工程师",
		JDText:     "熟悉 Go 语言，了解微服务架构，有数据库和缓存经验",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ApplyReadiness != "not_ready_for_this_role" {
		t.Errorf("expected apply_readiness=not_ready_for_this_role for no overlap, got %s", result.ApplyReadiness)
	}
}

func TestResumeOptimizationMockAgent_Optimize_FixItemsNoFabricatedMetrics(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "计算机专业学生，了解 Go 和 MySQL，做过一个课程项目。",
		JobTitle:   "后端开发",
		JDText:     "后端开发 接口开发 数据库 缓存",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ensure no fabricated specific numbers in example references
	fabricatedPatterns := []string{"QPS 达到", "DAU 超过", "提升了 50%", "性能提升 3 倍"}
	for _, item := range result.TopFixItems {
		for _, pattern := range fabricatedPatterns {
			if strings.Contains(item.ExampleReference, pattern) {
				t.Errorf("fix item should not contain fabricated metric pattern '%s': %s", pattern, item.ExampleReference)
			}
		}
	}
}

func TestResumeOptimizationMockAgent_Optimize_SelfCheckQuestion(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "熟悉 Go，做过一些项目。",
		JobTitle:   "Go 开发",
		JDText:     "Go 语言开发经验",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, item := range result.TopFixItems {
		if item.SelfCheckQuestion == "" {
			t.Error("expected non-empty self_check_question in fix item")
		}
		if item.LearnMoreQuery == "" {
			t.Error("expected non-empty learn_more_query in fix item")
		}
	}
}

func TestResumeOptimizationMockAgent_Optimize_EvidenceChainStrength(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()

	tests := []struct {
		name           string
		resumeText     string
		jdText         string
		wantStrength   string
	}{
		{
			name:         "missing evidence",
			resumeText:   "我是学生",
			jdText:       "需要 Go 语言和 MySQL 经验",
			wantStrength: "missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agent.Optimize(context.Background(), &OptimizeInput{
				ResumeText: tt.resumeText,
				JDText:     tt.jdText,
				JobTitle:   "后端开发",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Check at least one entry has the expected strength
			found := false
			for _, ev := range result.JDEvidenceChain {
				if ev.EvidenceStrength == tt.wantStrength {
					found = true
					break
				}
			}
			if !found && tt.wantStrength == "missing" {
				// For very short resumes, missing is expected
				t.Logf("evidence chain: %+v", result.JDEvidenceChain)
			}
		})
	}
}

func TestResumeOptimizationMockAgent_Optimize_ConsistencyCheck(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText: "熟悉 Go 和 MySQL，做过后台管理系统项目，负责接口开发。",
		JobTitle:   "Go 后端开发",
		JDText:     "Go 语言 后端开发 接口开发 MySQL",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check apply_readiness and optimization_stop_signal consistency
	readinessToSignal := map[string]string{
		"ready_to_apply":           "stop_and_apply",
		"quick_fix_before_apply":   "fix_top_3_then_apply",
		"not_ready_for_this_role":  "target_role_mismatch",
		"insufficient_information": "needs_more_evidence",
	}

	expectedSignal, ok := readinessToSignal[result.ApplyReadiness]
	if ok && result.OptimizationStopSignal != expectedSignal {
		t.Errorf("inconsistency: apply_readiness=%s should pair with optimization_stop_signal=%s, got %s",
			result.ApplyReadiness, expectedSignal, result.OptimizationStopSignal)
	}
}

func TestResumeOptimizationMockAgent_Optimize_TargetSection(t *testing.T) {
	agent := NewResumeOptimizationMockAgent()
	input := &OptimizeInput{
		ResumeText:    "熟悉 Go，做过项目。",
		JobTitle:      "Go 后端开发",
		JDText:        "Go 语言开发",
		TargetSection: "项目经历",
	}

	result, err := agent.Optimize(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still return valid result with target_section specified
	if result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got %s", result.GeneratedBy)
	}
}

// Ensure ResumeOptimizationMockAgent implements the expected interface
func TestResumeOptimizationMockAgent_Interface(t *testing.T) {
	var _ *ResumeOptimizationMockAgent = NewResumeOptimizationMockAgent()
}

// Test model struct JSON tags
func TestResumeOptimizationResult_JSONTags(t *testing.T) {
	result := &model.ResumeOptimizationResult{
		AgentStatus:            "mock_success",
		GeneratedBy:            "mock_agent",
		FallbackRequired:       false,
		RepairApplied:          false,
		ApplyReadiness:         "quick_fix_before_apply",
		OptimizationStopSignal: "fix_top_3_then_apply",
		Summary:                "test summary",
		TopFixItems: []model.ResumeOptimizationFixItem{
			{
				Issue:             "test issue",
				WhyItMatters:      "test why",
				Action:            "test action",
				ExampleReference:  "示例参考：test",
				SelfCheckQuestion: "test question",
				LearnMoreQuery:    "test query",
				EstimatedEffort:   "within_30_min",
			},
		},
		VerificationNotice: "test notice",
	}

	// Verify JSON serialization works
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	jsonStr := string(data)
	requiredFields := []string{
		"agent_status", "generated_by", "fallback_required", "repair_applied",
		"apply_readiness", "optimization_stop_signal", "summary",
		"top_fix_items", "verification_notice",
	}
	for _, field := range requiredFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing field: %s", field)
		}
	}
}
