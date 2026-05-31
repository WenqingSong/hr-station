package agent

import (
	"context"
	"encoding/json"
	"testing"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// Unified Fake Pipeline Runner for integration tests
// ---------------------------------------------------------------------------

// integrationRunner is a configurable fake runner for integration testing.
// It supports three behaviors: success, repaired, and fallback.
type integrationRunner struct {
	result *pipeline.RunResult
}

func (r *integrationRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return r.result
}

// ---------------------------------------------------------------------------
// Shared test data for integration tests
// ---------------------------------------------------------------------------

func intTestResume() *model.ResumeVersion {
	skills, _ := json.Marshal([]string{"Go", "MySQL", "Redis"})
	edu, _ := json.Marshal([]map[string]any{
		{"school": "TestUniversity", "major": "Computer Science", "degree": "本科"},
	})
	exp, _ := json.Marshal([]map[string]any{
		{"company": "TestCo", "role": "Backend Intern", "description": "Developed Go microservices"},
	})
	proj, _ := json.Marshal([]map[string]any{
		{"name": "GoWebApp", "description": "Built a RESTful API with Go and MySQL"},
	})
	return &model.ResumeVersion{
		ID:             1,
		VersionNo:      1,
		TargetRole:     "后端开发",
		TargetCity:     "深圳",
		SkillsJSON:     skills,
		EducationJSON:  edu,
		ExperienceJSON: exp,
		ProjectJSON:    proj,
	}
}

func intTestJob() *model.Job {
	skillKeywords, _ := json.Marshal([]string{"Go", "MySQL", "微服务"})
	return &model.Job{
		ID:                10,
		Title:             "Go后端开发工程师",
		Department:        "技术部",
		City:              "深圳",
		Category:          "后端开发",
		Description:       "负责后端系统开发",
		Requirements:      "熟悉Go语言，了解微服务架构",
		DataQuality:       "full_jd",
		SkillKeywordsJSON: skillKeywords,
		ExternalSource:    "local_mock",
	}
}

func intTestJobs() []*model.Job {
	extID := "ext-001"
	return []*model.Job{
		{
			ID:             10,
			Title:          "Go后端开发工程师",
			Department:     "技术部",
			City:           "深圳",
			Category:       "后端开发",
			JobType:        "full_time",
			Description:    "负责后端系统开发",
			Requirements:   "熟悉Go语言",
			ExternalSource: "local_mock",
			ExternalID:     &extID,
			DataQuality:    "full_jd",
		},
	}
}

// intTestLLMMatchOutput returns a full LLM match analysis output with strategic fields.
func intTestLLMMatchOutput() map[string]any {
	return map[string]any{
		"match_score":      float64(78),
		"match_level":      "potential_match",
		"match_label":      "建议尝试",
		"screening_result": "likely",
		"offer_go_summary": "你的简历与该岗位有一定匹配度，建议优化后尝试投递。",
		"evidence_chain": []any{
			map[string]any{"jd_requirement": "Go经验", "resume_evidence": "有Go项目", "judgement": "matched", "explanation": "简历中有Go项目"},
		},
		"matched_points": []any{
			map[string]any{"point": "Go技能匹配", "why_it_matters": "核心要求", "evidence": "Go项目经验"},
		},
		"gap_points":    []any{},
		"risk_points":   []any{},
		"resume_optimization": map[string]any{"p0": []any{}, "p1": []any{}, "p2": []any{}},
		"interview_questions": []any{},
		"next_actions": []any{
			map[string]any{"action": "优化项目描述", "priority": "high", "expected_effect": "提升匹配度"},
		},
		"authenticity_notice":      "请基于真实经历修改",
		"apply_readiness":         "quick_fix_before_apply",
		"optimization_stop_signal": "fix_top_3_then_apply",
		"top_fix_items": []any{
			map[string]any{
				"issue":              "项目描述缺少量化数据",
				"why_it_matters":     "量化数据提升说服力",
				"action":             "补充量化成果",
				"example_reference":  "示例：完成 12 个 API",
				"self_check_question": "确认数字真实",
				"estimated_effort":   "within_30_min",
			},
		},
	}
}

// intTestLLMRecOutput returns a full LLM recommendation output with strategic fields.
func intTestLLMRecOutput() map[string]any {
	return map[string]any{
		"strategy_summary": "基于简历推荐 Go 后端岗位",
		"recommendations": []any{
			map[string]any{
				"job_id":           float64(10),
				"title":            "Go后端开发工程师",
				"department":       "技术部",
				"city":             "深圳",
				"category":         "后端开发",
				"job_type":         "full_time",
				"external_source":  "local_mock",
				"external_id":      "ext-001",
				"data_quality":     "full_jd",
				"recommend_level":  "priority_apply",
				"recommend_label":  "强烈推荐",
				"match_score":      float64(85),
				"offer_go_summary": "Go 经验匹配",
				"recommend_reason": "技能高度一致",
				"evidence":         []any{},
				"risk_points":      []any{},
				"next_action":      map[string]any{"action": "建议优先投递"},
				"apply_readiness":          "quick_fix_before_apply",
				"optimization_stop_signal":  "fix_top_3_then_apply",
				"top_fix_items": []any{
					map[string]any{
						"issue":              "缺少微服务经验",
						"why_it_matters":     "岗位核心要求",
						"action":             "补充微服务项目",
						"example_reference":  "示例：设计微服务架构",
						"self_check_question": "确认是否做过",
						"estimated_effort":   "within_1_hour",
					},
				},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "请基于真实经历",
		"analysis_basis":      map[string]any{},
	}
}

// intTestLLMResumeOptOutput returns a full LLM resume optimization output with strategic fields.
func intTestLLMResumeOptOutput() map[string]any {
	return map[string]any{
		"optimization_summary":      "建议优化后投递",
		"apply_readiness":          "quick_fix_before_apply",
		"optimization_stop_signal": "fix_top_3_then_apply",
		"top_fix_items": []any{
			map[string]any{
				"issue":              "项目描述不够清晰",
				"why_it_matters":     "难以证明个人贡献",
				"action":             "补充负责模块",
				"example_reference":  "示例：基于 Go 完成接口开发",
				"self_check_question": "确认是否负责过",
				"learn_more_query":   "Go 后端简历写法",
				"estimated_effort":   "within_30_min",
			},
		},
		"priority_suggestions": []any{},
		"keyword_alignment":    []any{},
		"authenticity_notice":  "请基于真实经历修改",
	}
}

// intTestOptimizeInput returns a standard OptimizeInput for testing.
func intTestOptimizeInput() *OptimizeInput {
	return &OptimizeInput{
		ResumeText:   "熟悉 Go、MySQL，做过后台系统开发",
		JobTitle:     "Go 后端开发工程师",
		JDText:       "Go 微服务 MySQL Redis",
		TargetSection: "project",
	}
}

// ===========================================================================
// MatchAnalysisAgent Integration Tests
// ===========================================================================

// --- LLM success ---

func TestIntegration_MatchAnalysis_LLM_Success(t *testing.T) {
	parsed := intTestLLMMatchOutput()
	agent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusSuccess,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: false,
		},
	})

	result, err := agent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// generated_by / agent_status
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "success" {
		t.Errorf("agent_status: got %q, want success", result.AgentStatus)
	}

	// Strategic fields
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty on LLM success")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty on LLM success")
	}
	if result.TopFixItems == nil {
		t.Error("top_fix_items should not be nil on LLM success")
	}

	// Core content fields non-empty
	if len(result.EvidenceChain) == 0 {
		t.Error("evidence_chain should not be empty")
	}
	if len(result.MatchHighlights) == 0 {
		t.Error("match_highlights should not be empty")
	}
	if result.OfferGoSummary == "" {
		t.Error("offer_go_summary should not be empty")
	}
}

// --- LLM error fallback ---

func TestIntegration_MatchAnalysis_LLM_Error_Fallback(t *testing.T) {
	agent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "service unavailable",
		},
	})

	_, err := agent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err == nil {
		t.Fatal("expected MatchFallbackError")
	}

	fbErr, ok := err.(*MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("fallback_reason: got %q, want llm_call_failed", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("error_code: got %q, want LLM_HTTP_ERROR", fbErr.ErrorCode)
	}
}

// --- Malformed JSON → Pipeline returns StatusRepaired (repair success) ---

func TestIntegration_MatchAnalysis_MalformedJSON_Repaired(t *testing.T) {
	// Simulate: initial LLM output was malformed, but repair succeeded
	parsed := intTestLLMMatchOutput()
	agent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusRepaired,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: true,
			RepairSuccess: true,
		},
	})

	result, err := agent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Repair success → still llm_agent
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "repaired" {
		t.Errorf("agent_status: got %q, want repaired", result.AgentStatus)
	}
	if !result.RepairApplied {
		t.Error("repair_applied should be true for repaired result")
	}

	// Strategic fields still preserved after repair
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty after repair")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty after repair")
	}
}

// --- Malformed JSON → repair also failed → fallback required ---

func TestIntegration_MatchAnalysis_MalformedJSON_RepairFailed_Fallback(t *testing.T) {
	// Simulate: malformed JSON, repair failed, fallback required
	agent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "json_repair_failed",
			ErrorCode:        "JSON_REPAIR_FAILED",
			ErrorMessage:     "repair could not fix the JSON",
			RepairApplied:    true,
			RepairSuccess:    false,
		},
	})

	_, err := agent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err == nil {
		t.Fatal("expected MatchFallbackError for repair failure")
	}

	fbErr, ok := err.(*MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "json_repair_failed" {
		t.Errorf("fallback_reason: got %q, want json_repair_failed", fbErr.FallbackReason)
	}
}

// --- Mock direct ---

func TestIntegration_MatchAnalysis_Mock_Direct(t *testing.T) {
	mockAgent := NewMatchMockAgent()
	result, err := mockAgent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock result should have valid score
	if result.MatchScore < 0 || result.MatchScore > 100 {
		t.Errorf("match_score out of range: %d", result.MatchScore)
	}
	// Core fields should be non-empty
	if result.ScreeningResult == "" {
		t.Error("screening_result should not be empty")
	}
	if len(result.EvidenceChain) == 0 {
		t.Error("evidence_chain should not be empty")
	}
}

// --- Mock fallback: verify metadata set by service layer ---

func TestIntegration_MatchAnalysis_Mock_Fallback_Metadata(t *testing.T) {
	mockAgent := NewMatchMockAgent()
	result, err := mockAgent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simulate what MatchService.analyzeWithMock does
	result.GeneratedBy = "mock_agent"
	result.AgentStatus = "fallback"
	result.FallbackReason = "llm_call_failed"
	result.ErrorCode = "LLM_HTTP_ERROR"
	result.RepairApplied = false

	if result.GeneratedBy != "mock_agent" {
		t.Errorf("generated_by: got %q, want mock_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "fallback" {
		t.Errorf("agent_status: got %q, want fallback", result.AgentStatus)
	}
	// Mock fallback should still have strategic fields
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty on mock fallback")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty on mock fallback")
	}
}

// ===========================================================================
// RecommendationAgent Integration Tests
// ===========================================================================

// --- LLM success ---

func TestIntegration_Recommendation_LLM_Success(t *testing.T) {
	parsed := intTestLLMRecOutput()
	agent := NewRecommendationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusSuccess,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: false,
		},
	})

	results, err := agent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 recommendation")
	}

	j := results[0]

	// generated_by / agent_status
	if j.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", j.GeneratedBy)
	}
	if j.AgentStatus != "success" {
		t.Errorf("agent_status: got %q, want success", j.AgentStatus)
	}

	// Strategic fields
	if j.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty on LLM success")
	}
	if j.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty on LLM success")
	}
	if j.TopFixItems == nil {
		t.Error("top_fix_items should not be nil on LLM success")
	}

	// Core content
	if j.RecommendationReason == "" {
		t.Error("recommend_reason should not be empty")
	}
	if j.OfferGoSummary == "" {
		t.Error("offer_go_summary should not be empty")
	}
}

// --- LLM error fallback ---

func TestIntegration_Recommendation_LLM_Error_Fallback(t *testing.T) {
	agent := NewRecommendationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "service unavailable",
		},
	})

	_, err := agent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err == nil {
		t.Fatal("expected RecFallbackError")
	}

	fbErr, ok := err.(*RecFallbackError)
	if !ok {
		t.Fatalf("expected *RecFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("fallback_reason: got %q, want llm_call_failed", fbErr.FallbackReason)
	}
}

// --- Malformed JSON → Pipeline returns StatusRepaired ---

func TestIntegration_Recommendation_MalformedJSON_Repaired(t *testing.T) {
	parsed := intTestLLMRecOutput()
	agent := NewRecommendationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusRepaired,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: true,
			RepairSuccess: true,
		},
	})

	results, err := agent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 recommendation after repair")
	}

	j := results[0]
	if j.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", j.GeneratedBy)
	}
	if j.AgentStatus != "repaired" {
		t.Errorf("agent_status: got %q, want repaired", j.AgentStatus)
	}
	if !j.RepairApplied {
		t.Error("repair_applied should be true for repaired result")
	}

	// Strategic fields preserved after repair
	if j.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty after repair")
	}
	if j.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty after repair")
	}
}

// --- Malformed JSON → repair failed → fallback ---

func TestIntegration_Recommendation_MalformedJSON_RepairFailed_Fallback(t *testing.T) {
	agent := NewRecommendationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "json_repair_schema_failed",
			ErrorCode:        "JSON_REPAIR_SCHEMA_FAILED",
			ErrorMessage:     "repaired JSON failed schema validation",
			RepairApplied:    true,
			RepairSuccess:    false,
		},
	})

	_, err := agent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err == nil {
		t.Fatal("expected RecFallbackError for repair failure")
	}

	fbErr, ok := err.(*RecFallbackError)
	if !ok {
		t.Fatalf("expected *RecFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "json_repair_schema_failed" {
		t.Errorf("fallback_reason: got %q, want json_repair_schema_failed", fbErr.FallbackReason)
	}
}

// --- Mock direct ---

func TestIntegration_Recommendation_Mock_Direct(t *testing.T) {
	mockAgent := NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 mock recommendation")
	}
	for _, j := range results {
		if j.Priority == "" {
			t.Error("priority should not be empty")
		}
		if j.RecommendationReason == "" {
			t.Error("recommend_reason should not be empty")
		}
	}
}

// --- Mock fallback metadata ---

func TestIntegration_Recommendation_Mock_Fallback_Metadata(t *testing.T) {
	mockAgent := NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simulate what RecommendationService.recommendWithMock does
	for i := range results {
		results[i].GeneratedBy = "mock_agent"
		results[i].AgentStatus = "fallback"
		results[i].FallbackReason = "llm_call_failed"
		results[i].ErrorCode = "LLM_HTTP_ERROR"
		results[i].RepairApplied = false
	}

	for _, j := range results {
		if j.GeneratedBy != "mock_agent" {
			t.Errorf("generated_by: got %q, want mock_agent", j.GeneratedBy)
		}
		if j.AgentStatus != "fallback" {
			t.Errorf("agent_status: got %q, want fallback", j.AgentStatus)
		}
	}
}

// ===========================================================================
// ResumeOptimizationAgent Integration Tests
// ===========================================================================

// --- LLM success ---

func TestIntegration_ResumeOptimization_LLM_Success(t *testing.T) {
	parsed := intTestLLMResumeOptOutput()
	agent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusSuccess,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: false,
		},
	})

	result, err := agent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// generated_by / agent_status
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "success" {
		t.Errorf("agent_status: got %q, want success", result.AgentStatus)
	}

	// Strategic fields
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty on LLM success")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty on LLM success")
	}
	if result.TopFixItems == nil {
		t.Error("top_fix_items should not be nil on LLM success")
	}

	// Core content
	if result.Summary == "" {
		t.Error("summary should not be empty")
	}
	if result.VerificationNotice == "" {
		t.Error("verification_notice should not be empty")
	}
}

// --- LLM error fallback ---

func TestIntegration_ResumeOptimization_LLM_Error_Fallback(t *testing.T) {
	agent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "service unavailable",
		},
	})

	_, err := agent.Optimize(context.Background(), intTestOptimizeInput())
	if err == nil {
		t.Fatal("expected ResumeOptFallbackError")
	}

	fbErr, ok := err.(*ResumeOptFallbackError)
	if !ok {
		t.Fatalf("expected *ResumeOptFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("fallback_reason: got %q, want llm_call_failed", fbErr.FallbackReason)
	}
}

// --- Malformed JSON → repair succeeded ---

func TestIntegration_ResumeOptimization_MalformedJSON_Repaired(t *testing.T) {
	parsed := intTestLLMResumeOptOutput()
	agent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusRepaired,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: true,
			RepairSuccess: true,
		},
	})

	result, err := agent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by: got %q, want llm_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "repaired" {
		t.Errorf("agent_status: got %q, want repaired", result.AgentStatus)
	}
	if !result.RepairApplied {
		t.Error("repair_applied should be true for repaired result")
	}

	// Strategic fields preserved
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty after repair")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty after repair")
	}
}

// --- Malformed JSON → repair failed → fallback ---

func TestIntegration_ResumeOptimization_MalformedJSON_RepairFailed_Fallback(t *testing.T) {
	agent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "json_repair_failed",
			ErrorCode:        "JSON_REPAIR_FAILED",
			ErrorMessage:     "repair could not fix the JSON",
			RepairApplied:    true,
			RepairSuccess:    false,
		},
	})

	_, err := agent.Optimize(context.Background(), intTestOptimizeInput())
	if err == nil {
		t.Fatal("expected ResumeOptFallbackError for repair failure")
	}

	fbErr, ok := err.(*ResumeOptFallbackError)
	if !ok {
		t.Fatalf("expected *ResumeOptFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "json_repair_failed" {
		t.Errorf("fallback_reason: got %q, want json_repair_failed", fbErr.FallbackReason)
	}
}

// --- Mock direct ---

func TestIntegration_ResumeOptimization_Mock_Direct(t *testing.T) {
	mockAgent := NewResumeOptimizationMockAgent()
	result, err := mockAgent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Summary == "" {
		t.Error("summary should not be empty from mock")
	}
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty from mock")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty from mock")
	}
	if result.VerificationNotice == "" {
		t.Error("verification_notice should not be empty from mock")
	}
}

// --- Mock fallback: verify metadata ---

func TestIntegration_ResumeOptimization_Mock_Fallback_Metadata(t *testing.T) {
	mockAgent := NewResumeOptimizationMockAgent()
	result, err := mockAgent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simulate what ResumeOptimizationService.optimizeWithMock does
	result.GeneratedBy = "mock_agent"
	result.AgentStatus = "fallback"
	result.FallbackReason = "llm_call_failed"
	result.ErrorCode = "LLM_HTTP_ERROR"
	result.RepairApplied = false

	if result.GeneratedBy != "mock_agent" {
		t.Errorf("generated_by: got %q, want mock_agent", result.GeneratedBy)
	}
	if result.AgentStatus != "fallback" {
		t.Errorf("agent_status: got %q, want fallback", result.AgentStatus)
	}
	// Strategic fields still present on fallback
	if result.ApplyReadiness == "" {
		t.Error("apply_readiness should not be empty on mock fallback")
	}
	if result.OptimizationStopSignal == "" {
		t.Error("optimization_stop_signal should not be empty on mock fallback")
	}
	if result.TopFixItems == nil {
		t.Error("top_fix_items should not be nil on mock fallback")
	}
}

// ===========================================================================
// Cross-Agent generated_by / agent_status Validation
// ===========================================================================

func TestIntegration_GeneratedBy_AgentStatus_Rules(t *testing.T) {
	tests := []struct {
		name         string
		agentStatus  pipeline.AgentStatus
		generatedBy  string
		wantStatus   string // expected model agent_status
		wantGenBy    string // expected model generated_by
	}{
		{
			name:        "llm_success",
			agentStatus: pipeline.StatusSuccess,
			generatedBy: "llm_agent",
			wantStatus:  "success",
			wantGenBy:   "llm_agent",
		},
		{
			name:        "llm_repaired",
			agentStatus: pipeline.StatusRepaired,
			generatedBy: "llm_agent",
			wantStatus:  "repaired",
			wantGenBy:   "llm_agent",
		},
		{
			name:        "llm_fallback",
			agentStatus: pipeline.StatusFallbackRequired,
			generatedBy: "mock_agent",
			wantStatus:  "fallback", // set by service layer on mock fallback
			wantGenBy:   "mock_agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the pipeline-level values are consistent
			if tt.agentStatus == pipeline.StatusSuccess || tt.agentStatus == pipeline.StatusRepaired {
				if tt.generatedBy != "llm_agent" {
					t.Errorf("success/repaired should have generated_by=llm_agent, got %q", tt.generatedBy)
				}
			}
			if tt.agentStatus == pipeline.StatusFallbackRequired {
				if tt.generatedBy != "mock_agent" {
					t.Errorf("fallback_required should have generated_by=mock_agent, got %q", tt.generatedBy)
				}
			}
		})
	}
}

// ===========================================================================
// Strategic Fields Cross-Agent Validation
// ===========================================================================

func TestIntegration_StrategicFields_AllAgents(t *testing.T) {
	// MatchAnalysis
	matchParsed := intTestLLMMatchOutput()
	matchAgent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      matchParsed,
		},
	})
	matchResult, err := matchAgent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("MatchAnalysis error: %v", err)
	}
	if matchResult.ApplyReadiness == "" {
		t.Error("MatchAnalysis: apply_readiness empty on LLM success")
	}
	if matchResult.OptimizationStopSignal == "" {
		t.Error("MatchAnalysis: optimization_stop_signal empty on LLM success")
	}
	if matchResult.TopFixItems == nil {
		t.Error("MatchAnalysis: top_fix_items nil on LLM success")
	}

	// Recommendation
	recParsed := intTestLLMRecOutput()
	recAgent := NewRecommendationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      recParsed,
		},
	})
	recResults, err := recAgent.Recommend(context.Background(), intTestResume(), intTestJobs())
	if err != nil {
		t.Fatalf("Recommendation error: %v", err)
	}
	if len(recResults) == 0 {
		t.Fatal("Recommendation: expected at least 1 result")
	}
	if recResults[0].ApplyReadiness == "" {
		t.Error("Recommendation: apply_readiness empty on LLM success")
	}
	if recResults[0].OptimizationStopSignal == "" {
		t.Error("Recommendation: optimization_stop_signal empty on LLM success")
	}
	if recResults[0].TopFixItems == nil {
		t.Error("Recommendation: top_fix_items nil on LLM success")
	}

	// ResumeOptimization
	resumeParsed := intTestLLMResumeOptOutput()
	resumeAgent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      resumeParsed,
		},
	})
	resumeResult, err := resumeAgent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("ResumeOptimization error: %v", err)
	}
	if resumeResult.ApplyReadiness == "" {
		t.Error("ResumeOptimization: apply_readiness empty on LLM success")
	}
	if resumeResult.OptimizationStopSignal == "" {
		t.Error("ResumeOptimization: optimization_stop_signal empty on LLM success")
	}
	if resumeResult.TopFixItems == nil {
		t.Error("ResumeOptimization: top_fix_items nil on LLM success")
	}
}

// ===========================================================================
// Default Strategic Fields When Missing from LLM Output
// ===========================================================================

func TestIntegration_DefaultStrategicFields_WhenMissing(t *testing.T) {
	// MatchAnalysis: LLM output missing strategic fields → defaults applied
	matchParsed := map[string]any{
		"match_score":      float64(50),
		"screening_result": "possible",
		"evidence_chain":   []any{},
		"matched_points":   []any{},
		"next_actions":     []any{},
	}
	matchAgent := NewMatchAnalysisAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      matchParsed,
		},
	})
	matchResult, err := matchAgent.Analyze(context.Background(), intTestResume(), intTestJob())
	if err != nil {
		t.Fatalf("MatchAnalysis error: %v", err)
	}
	if matchResult.ApplyReadiness != "insufficient_information" {
		t.Errorf("default apply_readiness: got %q, want insufficient_information", matchResult.ApplyReadiness)
	}
	if matchResult.OptimizationStopSignal != "needs_more_evidence" {
		t.Errorf("default optimization_stop_signal: got %q, want needs_more_evidence", matchResult.OptimizationStopSignal)
	}

	// ResumeOptimization: LLM output missing strategic fields → defaults applied
	resumeParsed := map[string]any{}
	resumeAgent := NewResumeOptimizationAgent(&integrationRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      resumeParsed,
		},
	})
	resumeResult, err := resumeAgent.Optimize(context.Background(), intTestOptimizeInput())
	if err != nil {
		t.Fatalf("ResumeOptimization error: %v", err)
	}
	if resumeResult.ApplyReadiness != "insufficient_information" {
		t.Errorf("default apply_readiness: got %q, want insufficient_information", resumeResult.ApplyReadiness)
	}
	if resumeResult.OptimizationStopSignal != "needs_more_evidence" {
		t.Errorf("default optimization_stop_signal: got %q, want needs_more_evidence", resumeResult.OptimizationStopSignal)
	}
	if resumeResult.TopFixItems == nil {
		t.Error("top_fix_items should be non-nil empty slice when missing")
	}
}

// ===========================================================================
// Interface Compliance Cross-Check
// ===========================================================================

func TestIntegration_AllAgents_InterfaceCompliance(t *testing.T) {
	var _ MatchAnalyzer = (*MatchAnalysisAgent)(nil)
	var _ MatchAnalyzer = (*MatchMockAgent)(nil)
	var _ Recommender = (*RecommendationAgent)(nil)
	var _ Recommender = (*RecommendationMockAgent)(nil)
	var _ ResumeOptimizer = (*ResumeOptimizationAgent)(nil)
	var _ ResumeOptimizer = (*ResumeOptimizationMockAgent)(nil)
}
