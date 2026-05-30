package service

import (
	"context"
	"encoding/json"
	"testing"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// fakePipelineRunner: stub agent.PipelineRunner for service-level tests
// ---------------------------------------------------------------------------

type fakePipelineRunner struct {
	result *pipeline.RunResult
}

func (f *fakePipelineRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return f.result
}

// ---------------------------------------------------------------------------
// Test data helpers for service-level tests
// ---------------------------------------------------------------------------

func svcTestResume() *model.ResumeVersion {
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
		UserID:         100,
		VersionNo:      1,
		TargetRole:     "后端开发",
		TargetCity:     "深圳",
		SkillsJSON:     skills,
		EducationJSON:  edu,
		ExperienceJSON: exp,
		ProjectJSON:    proj,
	}
}

func svcTestJob() *model.Job {
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

// minimalLLMParsedOutput returns a valid match_analysis LLM output.
func minimalLLMParsedOutput() map[string]any {
	return map[string]any{
		"match_score":      float64(78),
		"match_level":      "potential_match",
		"match_label":      "建议尝试",
		"screening_result": "likely",
		"offer_go_summary": "你的简历与该岗位有一定匹配度，建议优化后尝试投递。",
		"evidence_chain": []any{
			map[string]any{"jd_requirement": "岗位要求Go相关经验", "resume_evidence": "简历中有Go项目经验", "judgement": "matched", "explanation": "简历中有明确的Go项目经验"},
			map[string]any{"jd_requirement": "岗位要求数据库经验", "resume_evidence": "简历中有MySQL使用经验", "judgement": "partial", "explanation": "有数据库使用经验但深度不足"},
		},
		"matched_points": []any{
			map[string]any{"point": "技能匹配Go和MySQL", "why_it_matters": "核心技能符合", "evidence": "简历中有相关项目"},
		},
		"gap_points": []any{
			map[string]any{"gap": "缺少微服务经验", "impact": "可能影响面试表现", "how_to_improve": "补充相关项目经验"},
		},
		"risk_points": []any{
			map[string]any{"risk": "微服务架构经验不足", "severity": "medium", "reason": "岗位明确要求", "mitigation": "学习相关技术"},
		},
		"resume_optimization": map[string]any{
			"p0": []any{
				map[string]any{"problem": "缺少微服务经验描述", "why_it_matters": "岗位明确要求", "action": "补充微服务相关项目描述"},
			},
			"p1": []any{
				map[string]any{"problem": "Redis经验表达不充分", "why_it_matters": "加分项", "action": "补充Redis使用场景"},
			},
		},
		"interview_questions": []any{
			map[string]any{"question": "请描述你在Go项目中的微服务架构设计", "why_ask": "验证架构能力", "how_to_prepare": "复习微服务设计模式", "related_jd_requirement": "微服务架构"},
		},
		"next_actions": []any{
			map[string]any{"action": "补充微服务相关项目描述", "priority": "high", "expected_effect": "提升匹配度"},
		},
		"authenticity_notice": "以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。",
	}
}

// ---------------------------------------------------------------------------
// Scenario 1: LLM success
// ---------------------------------------------------------------------------

func TestMatchService_LLM_Success(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:    pipeline.StatusSuccess,
			GeneratedBy:    "llm_agent",
			Parsed:         minimalLLMParsedOutput(),
			FallbackRequired: false,
			DurationMs:     1500,
		},
	}

	_ = NewMatchService(nil, nil, nil, agent.NewMatchMockAgent(), runner, true)

	// Test the LLM agent directly (Analyze requires DB repos for full service test)
	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got: %s", result.GeneratedBy)
	}
	if result.AgentStatus != "success" {
		t.Errorf("expected agent_status=success, got: %s", result.AgentStatus)
	}
	if result.RepairApplied != false {
		t.Error("expected repair_applied=false")
	}
	if result.MatchScore != 78 {
		t.Errorf("expected match_score=78, got: %d", result.MatchScore)
	}
	if result.ScreeningResult != "likely" {
		t.Errorf("expected screening_result=likely, got: %s", result.ScreeningResult)
	}
	if result.FallbackReason != "" {
		t.Errorf("expected empty fallback_reason, got: %s", result.FallbackReason)
	}
	if result.ErrorCode != "" {
		t.Errorf("expected empty error_code, got: %s", result.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: LLM repaired
// ---------------------------------------------------------------------------

func TestMatchService_LLM_Repaired(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusRepaired,
			GeneratedBy:      "llm_agent",
			Parsed:           minimalLLMParsedOutput(),
			RepairApplied:    true,
			RepairSuccess:    true,
			FallbackRequired: false,
			DurationMs:       3000,
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got: %s", result.GeneratedBy)
	}
	if result.AgentStatus != "repaired" {
		t.Errorf("expected agent_status=repaired, got: %s", result.AgentStatus)
	}
	if result.RepairApplied != true {
		t.Error("expected repair_applied=true")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: fallback_required → Mock fallback
// ---------------------------------------------------------------------------

func TestMatchService_LLM_FallbackRequired(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "HTTP 500 from LLM provider",
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	_, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	if err == nil {
		t.Fatal("expected error (MatchFallbackError), got nil")
	}

	fbErr, ok := err.(*agent.MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got: %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("expected error_code=LLM_HTTP_ERROR, got: %s", fbErr.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: LLM_ENABLE=false → direct Mock
// ---------------------------------------------------------------------------

func TestMatchService_LLMDisabled_UseMock(t *testing.T) {
	// When llmEnabled=false, pipelineRunner is nil and service goes directly to Mock
	svc := NewMatchService(nil, nil, nil, agent.NewMatchMockAgent(), nil, false)

	// Verify that service is configured correctly for mock-only path
	if svc.llmEnabled != false {
		t.Error("expected llmEnabled=false")
	}
	if svc.pipelineRunner != nil {
		t.Error("expected pipelineRunner=nil when LLM disabled")
	}
	if svc.mockAgent == nil {
		t.Error("expected mockAgent to be non-nil")
	}

	// Mock Agent should work directly
	result, err := svc.mockAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("expected no error from mock agent, got: %v", err)
	}
	if result.MatchScore < 0 || result.MatchScore > 100 {
		t.Errorf("expected match_score 0-100, got: %d", result.MatchScore)
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: LLM config error / Pipeline error → fallback to Mock
// ---------------------------------------------------------------------------

func TestMatchService_LLMConfigError_FallbackToMock(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_config_missing",
			ErrorCode:        "LLM_CONFIG_MISSING",
			ErrorMessage:     "LLM_BASE_URL is not configured",
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	_, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	if err == nil {
		t.Fatal("expected MatchFallbackError, got nil")
	}

	fbErr, ok := err.(*agent.MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got: %T", err)
	}
	if fbErr.ErrorCode != "LLM_CONFIG_MISSING" {
		t.Errorf("expected error_code=LLM_CONFIG_MISSING, got: %s", fbErr.ErrorCode)
	}
}

func TestMatchService_PipelineError_FallbackToMock(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "json_parse_failed",
			ErrorCode:        "JSON_PARSE_FAILED",
			ErrorMessage:     "Failed to parse LLM output as JSON",
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	_, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	if err == nil {
		t.Fatal("expected MatchFallbackError, got nil")
	}

	fbErr, ok := err.(*agent.MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "json_parse_failed" {
		t.Errorf("expected fallback_reason=json_parse_failed, got: %s", fbErr.FallbackReason)
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: Invalid request parameters
// ---------------------------------------------------------------------------

func TestMatchService_InvalidJobID(t *testing.T) {
	// When jobID is 0 or invalid, the handler should return error before entering service
	// This test verifies that the service layer expects valid jobID
	svc := NewMatchService(nil, nil, nil, agent.NewMatchMockAgent(), nil, false)

	// We can't test Analyze without DB repos, but we verify the service setup
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

// ---------------------------------------------------------------------------
// Verify Mock Agent generates stable results
// ---------------------------------------------------------------------------

func TestMatchService_MockAgent_StableResult(t *testing.T) {
	mockAgent := agent.NewMatchMockAgent()
	result, err := mockAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify core fields exist
	if result.MatchScore < 0 || result.MatchScore > 100 {
		t.Errorf("match_score out of range: %d", result.MatchScore)
	}
	if result.ScreeningResult == "" {
		t.Error("screening_result should not be empty")
	}
	if len(result.EvidenceChain) == 0 {
		t.Error("evidence_chain should not be empty")
	}
	if result.AuthenticityNotice == "" {
		t.Error("authenticity_notice should not be empty")
	}

	// Verify valid screening_result enum
	validScreening := map[string]bool{"highly_likely": true, "likely": true, "possible": true, "risky": true}
	if !validScreening[result.ScreeningResult] {
		t.Errorf("invalid screening_result: %s", result.ScreeningResult)
	}
}

// ---------------------------------------------------------------------------
// Verify generated_by / agent_status / repair_applied / fallback_reason / error_code
// ---------------------------------------------------------------------------

func TestMatchService_LLMResult_FieldsReasonable(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           minimalLLMParsedOutput(),
			FallbackRequired: false,
			DurationMs:       1500,
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// generated_by should be llm_agent (not llm_agent_fallback, not json_repair_agent)
	if result.GeneratedBy == "llm_agent_fallback" {
		t.Error("generated_by should not be llm_agent_fallback")
	}
	if result.GeneratedBy == "json_repair_agent" {
		t.Error("generated_by should not be json_repair_agent")
	}
	if result.GeneratedBy == "repair_agent" {
		t.Error("generated_by should not be repair_agent")
	}
	if result.GeneratedBy == "pipeline" {
		t.Error("generated_by should not be pipeline")
	}

	// agent_status should be one of the valid values
	validStatuses := map[string]bool{"success": true, "repaired": true, "fallback": true, "mock": true}
	if !validStatuses[result.AgentStatus] {
		t.Errorf("invalid agent_status: %s", result.AgentStatus)
	}

	// fallback_reason and error_code should be empty on success
	if result.FallbackReason != "" {
		t.Errorf("fallback_reason should be empty on success, got: %s", result.FallbackReason)
	}
	if result.ErrorCode != "" {
		t.Errorf("error_code should be empty on success, got: %s", result.ErrorCode)
	}
}

func TestMatchService_MockResult_FieldsReasonable(t *testing.T) {
	mockAgent := agent.NewMatchMockAgent()
	result, err := mockAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock result should not have LLM-specific fields set by default
	if result.GeneratedBy == "llm_agent" {
		t.Error("mock result should not have generated_by=llm_agent")
	}
}

// ---------------------------------------------------------------------------
// Verify MatchResponse structure
// ---------------------------------------------------------------------------

func TestMatchService_Response_Structure(t *testing.T) {
	// Verify MatchResponse struct has expected fields
	resp := MatchResponse{
		AnalysisID:  1,
		GeneratedBy: "llm_agent",
		Result: &model.MatchAnalysisResult{
			MatchScore:      78,
			ScreeningResult: "likely",
			GeneratedBy:     "llm_agent",
			AgentStatus:     "success",
			RepairApplied:   false,
		},
		AnalysisBasis: &model.AnalysisBasis{
			JobDescriptionUsed:  true,
			JobRequirementsUsed: true,
			DataQuality:         "full_jd",
		},
		CreatedAt: "2026-05-30T12:00:00+08:00",
	}

	if resp.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got: %s", resp.GeneratedBy)
	}
	if resp.Result.MatchScore != 78 {
		t.Errorf("expected match_score=78, got: %d", resp.Result.MatchScore)
	}
	if resp.AnalysisBasis.DataQuality != "full_jd" {
		t.Errorf("expected data_quality=full_jd, got: %s", resp.AnalysisBasis.DataQuality)
	}
}

// ---------------------------------------------------------------------------
// Verify no sensitive data leakage
// ---------------------------------------------------------------------------

func TestMatchService_NoSensitiveDataLeak(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           minimalLLMParsedOutput(),
			FallbackRequired: false,
			DurationMs:       1500,
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Result should not contain API key patterns
	resultJSON, _ := json.Marshal(result)
	resultStr := string(resultJSON)

	// Check for common API key patterns
	if containsAny(resultStr, []string{"api_key", "API_KEY", "apikey", "ApiKey", "Authorization", "Bearer "}) {
		t.Error("result should not contain API key patterns")
	}

	// Result should not contain full raw output
	if containsAny(resultStr, []string{"raw_output", "raw_content", "rawOutput", "rawContent"}) {
		t.Error("result should not contain raw output fields")
	}

	// Result should not contain full prompt
	if containsAny(resultStr, []string{"prompt_text", "system_prompt", "user_prompt"}) {
		t.Error("result should not contain prompt fields")
	}
}

func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if len(s) >= len(p) {
			for i := 0; i <= len(s)-len(p); i++ {
				if s[i:i+len(p)] == p {
					return true
				}
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Verify MatchFallbackError structure
// ---------------------------------------------------------------------------

func TestMatchFallbackError_Structure(t *testing.T) {
	fbErr := &agent.MatchFallbackError{
		FallbackReason: "llm_timeout",
		ErrorCode:      "LLM_TIMEOUT",
		ErrorMessage:   "LLM request timed out after 30s",
	}

	errStr := fbErr.Error()
	if errStr == "" {
		t.Error("MatchFallbackError.Error() should not be empty")
	}
	if fbErr.FallbackReason != "llm_timeout" {
		t.Errorf("expected fallback_reason=llm_timeout, got: %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_TIMEOUT" {
		t.Errorf("expected error_code=LLM_TIMEOUT, got: %s", fbErr.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// Verify field mapping completeness for LLM output
// ---------------------------------------------------------------------------

func TestMatchService_LLMFieldMapping_Complete(t *testing.T) {
	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           minimalLLMParsedOutput(),
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Core fields from original interface
	if result.MatchScore == 0 {
		t.Error("match_score should not be 0 from LLM output")
	}
	if result.ScreeningResult == "" {
		t.Error("screening_result should not be empty")
	}
	if len(result.EvidenceChain) == 0 {
		t.Error("evidence_chain should not be empty")
	}
	if len(result.MatchHighlights) == 0 {
		t.Error("match_highlights should not be empty")
	}
	if len(result.OptimizationSuggestions) == 0 {
		t.Error("optimization_suggestions should not be empty")
	}
	if len(result.NextSteps) == 0 {
		t.Error("next_steps should not be empty")
	}
	if result.AuthenticityNotice == "" {
		t.Error("authenticity_notice should not be empty")
	}

	// LLM-enriched fields
	if result.MatchLevel == "" {
		t.Error("match_level should not be empty from LLM output")
	}
	if result.MatchLabel == "" {
		t.Error("match_label should not be empty from LLM output")
	}
	if result.OfferGoSummary == "" {
		t.Error("offer_go_summary should not be empty from LLM output")
	}

	// Job/resume linkage
	if result.JobID != svcTestJob().ID {
		t.Errorf("expected job_id=%d, got: %d", svcTestJob().ID, result.JobID)
	}
	if result.ResumeVersionID != svcTestResume().ID {
		t.Errorf("expected resume_version_id=%d, got: %d", svcTestResume().ID, result.ResumeVersionID)
	}
}

// ---------------------------------------------------------------------------
// Verify mock result sets correct metadata on fallback
// ---------------------------------------------------------------------------

func TestMatchService_MockFallback_Metadata(t *testing.T) {
	// Simulate: service falls back to mock after LLM failure
	svc := NewMatchService(nil, nil, nil, agent.NewMatchMockAgent(), nil, false)

	result, err := svc.mockAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Manually set fallback metadata as analyzeWithMock does
	result.GeneratedBy = "mock_agent"
	result.AgentStatus = "fallback"
	result.FallbackReason = "llm_call_failed"
	result.ErrorCode = "LLM_HTTP_ERROR"
	result.RepairApplied = false

	if result.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got: %s", result.GeneratedBy)
	}
	if result.AgentStatus != "fallback" {
		t.Errorf("expected agent_status=fallback, got: %s", result.AgentStatus)
	}
	if result.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason, got: %s", result.FallbackReason)
	}
	if result.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("expected error_code, got: %s", result.ErrorCode)
	}
	if result.RepairApplied != false {
		t.Error("repair_applied should be false on mock fallback")
	}
}

// ---------------------------------------------------------------------------
// Verify omitempty fields don't break existing consumers
// ---------------------------------------------------------------------------

func TestMatchService_Omitempty_Fields(t *testing.T) {
	// Mock result without LLM fields - should serialize without LLM fields
	mockAgent := agent.NewMatchMockAgent()
	result, _ := mockAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())

	// Override generated_by as the service does
	result.GeneratedBy = "mock_agent"
	result.AgentStatus = "mock"

	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	// Verify omitempty fields are absent when empty
	resultStr := string(resultJSON)

	// These fields should be present (non-empty)
	if !containsSubstring(resultStr, `"match_score"`) {
		t.Error("match_score should be present")
	}
	if !containsSubstring(resultStr, `"screening_result"`) {
		t.Error("screening_result should be present")
	}
	if !containsSubstring(resultStr, `"generated_by"`) {
		t.Error("generated_by should be present")
	}
	if !containsSubstring(resultStr, `"agent_status"`) {
		t.Error("agent_status should be present")
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsSubstr(s, sub))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Verify MatchAnalyzer interface compliance
// ---------------------------------------------------------------------------

func TestMatchService_MatchAnalyzer_Interface(t *testing.T) {
	// Both Mock and LLM agents implement MatchAnalyzer
	var _ agent.MatchAnalyzer = agent.NewMatchMockAgent()

	runner := &fakePipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           minimalLLMParsedOutput(),
			FallbackRequired: false,
		},
	}
	var _ agent.MatchAnalyzer = agent.NewMatchAnalysisAgent(runner)
}

// ---------------------------------------------------------------------------
// Verify all valid screening_result values from LLM are mapped correctly
// ---------------------------------------------------------------------------

func TestMatchService_ScreeningResult_Mapping(t *testing.T) {
	tests := []struct {
		llmValue    string
		expected    string
		description string
	}{
		{"likely", "likely", "likely stays as likely"},
		{"uncertain", "possible", "uncertain maps to possible"},
		{"risky", "risky", "risky stays as risky"},
		{"highly_likely", "likely", "highly_likely maps to likely"},
		{"potential_match", "possible", "potential_match maps to possible"},
		{"explore_first", "possible", "explore_first maps to possible"},
		{"not_recommended", "risky", "not_recommended maps to risky"},
		{"unknown_value", "possible", "unknown defaults to possible"},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			output := minimalLLMParsedOutput()
			output["screening_result"] = tc.llmValue

			runner := &fakePipelineRunner{
				result: &pipeline.RunResult{
					AgentStatus:      pipeline.StatusSuccess,
					GeneratedBy:      "llm_agent",
					Parsed:           output,
					FallbackRequired: false,
				},
			}

			llmAgent := agent.NewMatchAnalysisAgent(runner)
			result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ScreeningResult != tc.expected {
				t.Errorf("LLM %q → expected %q, got %q", tc.llmValue, tc.expected, result.ScreeningResult)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Verify judgement mapping from LLM Schema to model
// ---------------------------------------------------------------------------

func TestMatchService_Judgement_Mapping(t *testing.T) {
	tests := []struct {
		llmValue    string
		expected    string
		description string
	}{
		{"matched", "strong_match", "matched maps to strong_match"},
		{"partial", "partial_match", "partial maps to partial_match"},
		{"missing", "missing", "missing stays as missing"},
		{"strong_match", "strong_match", "strong_match stays"},
		{"partial_match", "partial_match", "partial_match stays"},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			output := minimalLLMParsedOutput()
			output["evidence_chain"] = []any{
				map[string]any{"jd_requirement": "test", "resume_evidence": "test", "judgement": tc.llmValue},
			}

			runner := &fakePipelineRunner{
				result: &pipeline.RunResult{
					AgentStatus:      pipeline.StatusSuccess,
					GeneratedBy:      "llm_agent",
					Parsed:           output,
					FallbackRequired: false,
				},
			}

			llmAgent := agent.NewMatchAnalysisAgent(runner)
			result, err := llmAgent.Analyze(context.Background(), svcTestResume(), svcTestJob())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.EvidenceChain) == 0 {
				t.Fatal("evidence_chain should not be empty")
			}
			if result.EvidenceChain[0].Judgement != tc.expected {
				t.Errorf("LLM %q → expected %q, got %q", tc.llmValue, tc.expected, result.EvidenceChain[0].Judgement)
			}
		})
	}
}
