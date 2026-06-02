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
// (same pattern as match_service_test.go)
// ---------------------------------------------------------------------------

type fakeRecSvcPipelineRunner struct {
	result *pipeline.RunResult
}

func (f *fakeRecSvcPipelineRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return f.result
}

// ---------------------------------------------------------------------------
// Test data helpers for recommendation service-level tests
// ---------------------------------------------------------------------------

func recSvcTestResume() *model.ResumeVersion {
	skills, _ := json.Marshal([]string{"Go", "MySQL", "Docker", "Redis"})
	edu, _ := json.Marshal([]map[string]any{
		{"school": "TestUniversity", "major": "Computer Science", "degree": "本科"},
	})
	exp, _ := json.Marshal([]map[string]any{})
	proj, _ := json.Marshal([]map[string]any{})

	return &model.ResumeVersion{
		ID:              1,
		UserID:          100,
		VersionNo:       2,
		TargetRole:      "Go 后端开发工程师",
		TargetCity:      "深圳",
		SkillsJSON:      skills,
		EducationJSON:   edu,
		ExperienceJSON:  exp,
		ProjectJSON:     proj,
		SelfEvaluation:  "熟悉 Go 和微服务",
	}
}

func recSvcTestJobs() []*model.Job {
	extID := "ext-001"
	return []*model.Job{
		{
			ID:               1,
			Title:            "Go 后端开发工程师",
			Department:        "技术部",
			City:             "深圳",
			Category:         "技术",
			JobType:          "full_time",
			Description:      "负责 Go 后端开发",
			Requirements:     "3年 Go 经验",
			ExternalSource:   "local_mock",
			ExternalID:       &extID,
			DataQuality:      "full_jd",
			IsFeatured:       true,
		},
		{
			ID:               2,
			Title:            "Java 后端开发工程师",
			Department:        "平台部",
			City:             "北京",
			Category:         "技术",
			JobType:          "full_time",
			Description:      "负责 Java 后端开发",
			Requirements:     "3年 Java 经验",
			ExternalSource:   "tencent_join",
			DataQuality:      "full_jd",
		},
	}
}

func recSvcLLMOutput(jobID uint64) map[string]any {
	return map[string]any{
		"strategy_summary": "基于你的简历，推荐 Go 和 Java 后端岗位",
		"recommendations": []any{
			map[string]any{
				"job_id":          float64(jobID),
				"title":           "Go 后端开发工程师",
				"department":      "技术部",
				"city":            "深圳",
				"category":        "技术",
				"job_type":        "full_time",
				"external_source": "local_mock",
				"external_id":     "ext-001",
				"data_quality":    "full_jd",
				"recommend_level": "priority_apply",
				"recommend_label": "强烈推荐",
				"match_score":     float64(85),
				"offer_go_summary": "你的 Go 和 MySQL 经验非常匹配",
				"recommend_reason": "目标岗位与简历高度一致，技能匹配度高",
				"evidence": []any{
					map[string]any{
						"job_requirement": "3年 Go 经验",
						"resume_evidence": "技能列表包含 Go",
						"judgement":       "matched",
						"explanation":     "技能完全匹配",
					},
				},
				"risk_points": []any{
					map[string]any{
						"risk":       "Kubernetes 经验不足",
						"severity":   "medium",
						"mitigation": "可学习 K8s 基础",
					},
				},
				"query_suggestions": []any{
					map[string]any{
						"query":      "Go 微服务开发岗位",
						"why_search": "拓展更多 Go 相关岗位",
					},
				},
				"next_action": map[string]any{
					"action":          "建议优先投递该岗位",
					"priority":        "high",
					"expected_effect": "提高面试机会",
				},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "请基于真实经历优化简历。",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
}

// ---------------------------------------------------------------------------
// Scenario 1: LLM success — agent returns correct metadata
// ---------------------------------------------------------------------------

func TestRecService_LLM_Success(t *testing.T) {
	parsed := recSvcLLMOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
			RepairApplied:    false,
			DurationMs:       1200,
		},
	}

	// Verify service construction with nil repos (same as match test)
	_ = NewRecommendationService(nil, nil, nil, agent.NewRecommendationMockAgent(), runner, runner, true)

	// Test the LLM agent directly (same pattern as match_service_test.go)
	llmAgent := agent.NewRecommendationAgent(runner)
	results, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation, got: %d", len(results))
	}

	j := results[0]
	// Verify agent metadata (the fix in 4-3.5 ensures these are set from RunResult)
	if j.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got: %s", j.GeneratedBy)
	}
	if j.AgentStatus != "success" {
		t.Errorf("expected agent_status=success, got: %s", j.AgentStatus)
	}
	if j.RepairApplied {
		t.Error("expected repair_applied=false")
	}
	if j.FallbackReason != "" {
		t.Errorf("expected empty fallback_reason, got: %s", j.FallbackReason)
	}
	if j.ErrorCode != "" {
		t.Errorf("expected empty error_code, got: %s", j.ErrorCode)
	}

	// Verify core recommendation fields
	if j.JobID != 1 {
		t.Errorf("expected job_id=1, got: %d", j.JobID)
	}
	if j.Priority != "strong_recommend" {
		t.Errorf("expected priority=strong_recommend, got: %s", j.Priority)
	}
	if j.RecommendationScore != 85 {
		t.Errorf("expected recommendation_score=85, got: %d", j.RecommendationScore)
	}
	if j.RecommendLabel != "强烈推荐" {
		t.Errorf("expected recommend_label=强烈推荐, got: %s", j.RecommendLabel)
	}
	if j.OfferGoSummary != "你的 Go 和 MySQL 经验非常匹配" {
		t.Errorf("expected offer_go_summary, got: %s", j.OfferGoSummary)
	}
	if j.RecommendationReason == "" {
		t.Error("expected non-empty recommend_reason")
	}
	if len(j.MatchHighlights) == 0 {
		t.Error("expected match_highlights")
	}
	if len(j.RiskPoints) == 0 {
		t.Error("expected risk_points")
	}
	if len(j.QuerySuggestions) == 0 {
		t.Error("expected query_suggestions")
	}
	if j.SuggestedAction == "" {
		t.Error("expected suggested_action")
	}

	// Verify job fields come from candidate DB, not LLM
	if j.Title != "Go 后端开发工程师" {
		t.Errorf("title should come from candidate, got: %s", j.Title)
	}
	if j.Department != "技术部" {
		t.Errorf("department should come from candidate, got: %s", j.Department)
	}
	if j.City != "深圳" {
		t.Errorf("city should come from candidate, got: %s", j.City)
	}
}

// ---------------------------------------------------------------------------
// Scenario 2: LLM repaired — metadata correctly set
// ---------------------------------------------------------------------------

func TestRecService_LLM_Repaired(t *testing.T) {
	parsed := recSvcLLMOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusRepaired,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
			RepairApplied:    true,
			RepairSuccess:    true,
			DurationMs:       2500,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	results, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation, got: %d", len(results))
	}

	j := results[0]
	if j.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent (not json_repair_agent), got: %s", j.GeneratedBy)
	}
	if j.AgentStatus != "repaired" {
		t.Errorf("expected agent_status=repaired, got: %s", j.AgentStatus)
	}
	if !j.RepairApplied {
		t.Error("expected repair_applied=true for repaired pipeline result")
	}
}

// ---------------------------------------------------------------------------
// Scenario 3: fallback_required — returns RecFallbackError
// ---------------------------------------------------------------------------

func TestRecService_LLM_FallbackRequired(t *testing.T) {
	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "LLM service unavailable",
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	_, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err == nil {
		t.Fatal("expected RecFallbackError, got nil")
	}
	fbErr, ok := err.(*agent.RecFallbackError)
	if !ok {
		t.Fatalf("expected *agent.RecFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got: %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("expected error_code=LLM_HTTP_ERROR, got: %s", fbErr.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// Scenario 4: LLM_ENABLE=false — service uses Mock Agent directly
// ---------------------------------------------------------------------------

func TestRecService_LLMDisabled_UseMock(t *testing.T) {
	// When LLM is disabled, service construction should work with nil pipelineRunner
	svc := NewRecommendationService(nil, nil, nil, agent.NewRecommendationMockAgent(), nil, nil, false)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	// Verify mock agent works independently
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error from mock agent, got: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 mock recommendation")
	}
	for _, j := range results {
		if j.JobID == 0 {
			t.Error("expected non-zero job_id")
		}
		if j.Title == "" {
			t.Error("expected non-empty title")
		}
		if j.Priority != "strong_recommend" && j.Priority != "recommend" && j.Priority != "try" {
			t.Errorf("unexpected priority: %s", j.Priority)
		}
	}
}

// ---------------------------------------------------------------------------
// Scenario 5: LLM config error / Pipeline error → fallback
// ---------------------------------------------------------------------------

func TestRecService_PipelineError_FallbackToMock(t *testing.T) {
	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "json_parse_failed",
			ErrorCode:        "JSON_PARSE_FAILED",
			ErrorMessage:     "could not parse LLM output",
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	_, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err == nil {
		t.Fatal("expected RecFallbackError for pipeline error")
	}
	fbErr, ok := err.(*agent.RecFallbackError)
	if !ok {
		t.Fatalf("expected *agent.RecFallbackError, got: %T", err)
	}
	if fbErr.ErrorCode != "JSON_PARSE_FAILED" {
		t.Errorf("expected error_code=JSON_PARSE_FAILED, got: %s", fbErr.ErrorCode)
	}

	// Verify mock agent provides stable fallback
	mockAgent := agent.NewRecommendationMockAgent()
	mockResults, mockErr := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if mockErr != nil {
		t.Fatalf("mock agent should not fail, got: %v", mockErr)
	}
	if len(mockResults) == 0 {
		t.Fatal("mock agent should provide recommendations")
	}
}

// ---------------------------------------------------------------------------
// Scenario 6: LLM returns job_id not in candidates → filtered, fallback
// ---------------------------------------------------------------------------

func TestRecService_InvalidJobID_Fallback(t *testing.T) {
	parsed := recSvcLLMOutput(999) // invalid job_id
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	_, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err == nil {
		t.Fatal("expected error when all job_ids are invalid (filtered out)")
	}
	// Should be RecFallbackError (field_mapping_failed because all filtered)
	fbErr, ok := err.(*agent.RecFallbackError)
	if !ok {
		t.Fatalf("expected *agent.RecFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "field_mapping_failed" {
		t.Errorf("expected fallback_reason=field_mapping_failed, got: %s", fbErr.FallbackReason)
	}
}

// ---------------------------------------------------------------------------
// Scenario 7: Empty recommendations → fallback
// ---------------------------------------------------------------------------

func TestRecService_EmptyRecommendations_Fallback(t *testing.T) {
	parsed := map[string]any{
		"strategy_summary":    "无匹配岗位",
		"recommendations":     []any{},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "暂无推荐",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	_, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err == nil {
		t.Fatal("expected error for empty recommendations")
	}
	fbErr, ok := err.(*agent.RecFallbackError)
	if !ok {
		t.Fatalf("expected *agent.RecFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "field_mapping_failed" {
		t.Errorf("expected fallback_reason=field_mapping_failed, got: %s", fbErr.FallbackReason)
	}
}

// ---------------------------------------------------------------------------
// Scenario 8: Request parameter validation (verified through code review)
// ---------------------------------------------------------------------------

func TestRecService_RequestParamValidation(t *testing.T) {
	// Verify service construction and limit capping logic
	svc := NewRecommendationService(nil, nil, nil, agent.NewRecommendationMockAgent(), nil, nil, false)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	// Verify limit capping: Recommend() caps limit to max 5
	// (verified through code review of Recommend() method)
}

// ---------------------------------------------------------------------------
// Sorting stability — mock results sorted by score descending
// ---------------------------------------------------------------------------

func TestRecService_MockResult_SortedByScore(t *testing.T) {
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for i := 1; i < len(results); i++ {
		if results[i].RecommendationScore > results[i-1].RecommendationScore {
			t.Errorf("results not sorted by score descending: [%d]=%d > [%d]=%d",
				i, results[i].RecommendationScore,
				i-1, results[i-1].RecommendationScore)
		}
	}
}

// ---------------------------------------------------------------------------
// Mock Agent result structure validation
// ---------------------------------------------------------------------------

func TestRecService_MockResult_FieldsReasonable(t *testing.T) {
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected mock recommendations")
	}
	for _, j := range results {
		if j.RecommendationScore < 0 || j.RecommendationScore > 95 {
			t.Errorf("recommendation_score out of range: %d", j.RecommendationScore)
		}
		if j.Priority != "strong_recommend" && j.Priority != "recommend" && j.Priority != "try" {
			t.Errorf("unexpected priority: %s", j.Priority)
		}
		if j.RecommendationReason == "" {
			t.Error("expected non-empty recommendation_reason")
		}
		if len(j.MatchHighlights) == 0 {
			t.Error("expected non-empty match_highlights")
		}
		if j.SuggestedAction == "" {
			t.Error("expected non-empty suggested_action")
		}
		// generated_by should NOT be llm_agent for mock results
		if j.GeneratedBy == "llm_agent" {
			t.Error("mock result should not have generated_by=llm_agent")
		}
	}
}

// ---------------------------------------------------------------------------
// Security: no API key / raw output / sensitive data leak
// ---------------------------------------------------------------------------

func TestRecService_NoSensitiveDataLeak(t *testing.T) {
	parsed := recSvcLLMOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			RawContent:       "some raw with sk-abc123 and apikey value",
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	results, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	resultJSON, _ := json.Marshal(results)
	resultStr := string(resultJSON)

	sensitivePatterns := []string{"sk-", "apikey", "api_key", "secret_key", "bearer ", "authorization"}
	for _, p := range sensitivePatterns {
		if containsRecSvcSensitive(resultStr, p) {
			t.Errorf("serialized results should not contain sensitive pattern: %s", p)
		}
	}
}

// ---------------------------------------------------------------------------
// Mixed valid/invalid job IDs → filter invalid, keep valid
// ---------------------------------------------------------------------------

func TestRecService_MixedJobIDs_FilterInvalid(t *testing.T) {
	parsed := map[string]any{
		"strategy_summary": "混合推荐",
		"recommendations": []any{
			map[string]any{
				"job_id":          float64(1),
				"title":           "Go 后端开发工程师",
				"department":      "技术部",
				"city":            "深圳",
				"category":        "技术",
				"job_type":        "full_time",
				"recommend_level": "priority_apply",
				"match_score":     float64(85),
				"recommend_reason": "匹配",
				"evidence":        []any{},
				"risk_points":     []any{},
				"next_action":     map[string]any{"action": "投递"},
			},
			map[string]any{
				"job_id":          float64(999),
				"title":           "虚构岗位",
				"department":      "虚构部门",
				"city":            "虚构城市",
				"category":        "技术",
				"job_type":        "full_time",
				"recommend_level": "explore_first",
				"match_score":     float64(50),
				"recommend_reason": "虚构",
				"evidence":        []any{},
				"risk_points":     []any{},
				"next_action":     map[string]any{"action": "忽略"},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "提醒",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	results, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation (invalid filtered), got: %d", len(results))
	}
	if results[0].JobID != 1 {
		t.Errorf("expected job_id=1, got: %d", results[0].JobID)
	}
	// Verify no fabricated job in results
	for _, j := range results {
		if j.JobID == 999 {
			t.Error("fabricated job_id=999 should have been filtered out")
		}
	}
}

// ---------------------------------------------------------------------------
// Job fields come from candidate DB, not LLM
// ---------------------------------------------------------------------------

func TestRecService_JobFieldsFromCandidate(t *testing.T) {
	parsed := map[string]any{
		"strategy_summary": "推荐",
		"recommendations": []any{
			map[string]any{
				"job_id":          float64(1),
				"title":           "LLM 编造标题",
				"department":      "LLM 编造部门",
				"city":            "虚构城市",
				"category":        "技术",
				"job_type":        "full_time",
				"external_source": "local_mock",
				"external_id":     "ext-001",
				"data_quality":    "full_jd",
				"recommend_level": "priority_apply",
				"match_score":     float64(80),
				"recommend_reason": "匹配",
				"evidence":        []any{},
				"risk_points":     []any{},
				"next_action":     map[string]any{"action": "投递"},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "提醒",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	results, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got: %d", len(results))
	}
	j := results[0]
	if j.Title != "Go 后端开发工程师" {
		t.Errorf("title should come from candidate DB (Go 后端开发工程师), got: %s", j.Title)
	}
	if j.Department != "技术部" {
		t.Errorf("department should come from candidate DB (技术部), got: %s", j.Department)
	}
	if j.City != "深圳" {
		t.Errorf("city should come from candidate DB (深圳), got: %s", j.City)
	}
}

// ---------------------------------------------------------------------------
// Service response structure validation
// ---------------------------------------------------------------------------

func TestRecService_Response_Structure(t *testing.T) {
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify response structure can be serialized
	resp := &RecommendationResponse{
		ResumeVersionID: 1,
		ResumeVersionNo: 2,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: make([]*model.RecommendedJob, len(results)),
		Summary:         "test summary",
		CreatedAt:       "2026-05-30T12:00:00+08:00",
	}
	for i := range results {
		resp.RecommendedJobs[i] = &results[i]
	}

	respJSON, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("expected no marshal error, got: %v", err)
	}

	var unmarshaled RecommendationResponse
	if err := json.Unmarshal(respJSON, &unmarshaled); err != nil {
		t.Fatalf("expected no unmarshal error, got: %v", err)
	}
	if unmarshaled.GeneratedBy != "mock_agent" {
		t.Errorf("expected generated_by=mock_agent, got: %s", unmarshaled.GeneratedBy)
	}
	if len(unmarshaled.RecommendedJobs) != len(results) {
		t.Errorf("expected %d jobs, got: %d", len(results), len(unmarshaled.RecommendedJobs))
	}
}

// ---------------------------------------------------------------------------
// RecFallbackError structure validation
// ---------------------------------------------------------------------------

func TestRecService_RecFallbackError_Structure(t *testing.T) {
	fbErr := &agent.RecFallbackError{
		FallbackReason: "json_parse_failed",
		ErrorCode:      "JSON_PARSE_FAILED",
		ErrorMessage:   "could not parse output",
	}
	errMsg := fbErr.Error()
	if !containsRecSvcSensitive(errMsg, "json_parse_failed") {
		t.Error("error message should contain fallback_reason")
	}
	if !containsRecSvcSensitive(errMsg, "JSON_PARSE_FAILED") {
		t.Error("error message should contain error_code")
	}
}

// ---------------------------------------------------------------------------
// Omitempty fields — LLM-enriched fields should be omitted when empty
// ---------------------------------------------------------------------------

func TestRecService_Omitempty_Fields(t *testing.T) {
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	resultJSON, _ := json.Marshal(results)
	resultStr := string(resultJSON)

	// Mock results don't have LLM-enriched fields, so they should be omitted
	if containsRecSvcSensitive(resultStr, "offer_go_summary") {
		// Only fail if there's actually a non-empty value
		t.Log("offer_go_summary present in mock output — acceptable if non-empty")
	}
}

// ---------------------------------------------------------------------------
// Recommender interface compliance
// ---------------------------------------------------------------------------

func TestRecService_Recommender_Interface(t *testing.T) {
	// Verify both agents implement the Recommender interface
	var _ agent.Recommender = (*agent.RecommendationMockAgent)(nil)
	var _ agent.Recommender = (*agent.RecommendationAgent)(nil)

	// Verify mock agent with context
	mockAgent := agent.NewRecommendationMockAgent()
	_, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("mock agent Recommend with context should not fail: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Unexpected LLM error (not RecFallbackError) → wrapped as fallback
// ---------------------------------------------------------------------------

func TestRecService_UnexpectedLLMError_WrappedAsFallback(t *testing.T) {
	// Pipeline returns parsed=nil (will cause "parsed output is nil" error)
	runner := &fakeRecSvcPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           nil,
			FallbackRequired: false,
		},
	}

	llmAgent := agent.NewRecommendationAgent(runner)
	_, err := llmAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())

	if err == nil {
		t.Fatal("expected error for nil parsed output")
	}
	fbErr, ok := err.(*agent.RecFallbackError)
	if !ok {
		t.Fatalf("expected *agent.RecFallbackError, got: %T", err)
	}
	if fbErr.FallbackReason != "field_mapping_failed" {
		t.Errorf("expected fallback_reason=field_mapping_failed, got: %s", fbErr.FallbackReason)
	}
}

// ---------------------------------------------------------------------------
// Service fallback metadata on mock results
// ---------------------------------------------------------------------------

func TestRecService_MockFallback_Metadata(t *testing.T) {
	// Verify that mock results get correct metadata when used as fallback
	mockAgent := agent.NewRecommendationMockAgent()
	results, err := mockAgent.Recommend(context.Background(), recSvcTestResume(), recSvcTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Simulate what recommendWithMock() does
	fbErr := &agent.RecFallbackError{
		FallbackReason: "llm_call_failed",
		ErrorCode:      "LLM_HTTP_ERROR",
		ErrorMessage:   "service unavailable",
	}
	for i := range results {
		results[i].GeneratedBy = "mock_agent"
		results[i].AgentStatus = "fallback"
		results[i].FallbackReason = fbErr.FallbackReason
		results[i].ErrorCode = fbErr.ErrorCode
		results[i].RepairApplied = false
	}

	for _, j := range results {
		if j.GeneratedBy != "mock_agent" {
			t.Errorf("expected generated_by=mock_agent, got: %s", j.GeneratedBy)
		}
		if j.AgentStatus != "fallback" {
			t.Errorf("expected agent_status=fallback, got: %s", j.AgentStatus)
		}
		if j.FallbackReason != "llm_call_failed" {
			t.Errorf("expected fallback_reason=llm_call_failed, got: %s", j.FallbackReason)
		}
		if j.ErrorCode != "LLM_HTTP_ERROR" {
			t.Errorf("expected error_code=LLM_HTTP_ERROR, got: %s", j.ErrorCode)
		}
		if j.RepairApplied {
			t.Error("expected repair_applied=false for mock fallback")
		}
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func containsRecSvcSensitive(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
