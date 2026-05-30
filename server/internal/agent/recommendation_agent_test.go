package agent

import (
	"context"
	"encoding/json"
	"testing"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// fakePipelineRunner for testing (same pattern as match_agent_test.go)
// ---------------------------------------------------------------------------

type fakeRecPipelineRunner struct {
	result *pipeline.RunResult
}

func (r *fakeRecPipelineRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return r.result
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func recTestResume() *model.ResumeVersion {
	skillsJSON, _ := json.Marshal([]string{"Go", "MySQL", "Docker", "Redis"})
	eduJSON, _ := json.Marshal([]any{
		map[string]any{"school": "Test University", "degree": "本科", "major": "计算机科学"},
	})
	expJSON, _ := json.Marshal([]any{})
	projJSON, _ := json.Marshal([]any{})

	return &model.ResumeVersion{
		ID:           1,
		UserID:       100,
		VersionNo:    2,
		TargetRole:   "Go 后端开发工程师",
		TargetCity:   "深圳",
		SkillsJSON:   skillsJSON,
		EducationJSON: eduJSON,
		ExperienceJSON: expJSON,
		ProjectJSON:  projJSON,
		SelfEvaluation: "熟悉 Go 和微服务",
	}
}

func recTestJobs() []*model.Job {
	externalID1 := "ext-001"
	externalID2 := "ext-002"
	return []*model.Job{
		{
			ID:             1,
			Title:          "Go 后端开发工程师",
			Department:      "技术部",
			City:           "深圳",
			Category:       "技术",
			JobType:        "full_time",
			Description:    "负责 Go 后端开发",
			Requirements:   "3年 Go 经验",
			ExternalSource: "local_mock",
			ExternalID:     &externalID1,
			DataQuality:    "full_jd",
			IsFeatured:     true,
		},
		{
			ID:             2,
			Title:          "Java 后端开发工程师",
			Department:      "平台部",
			City:           "北京",
			Category:       "技术",
			JobType:        "full_time",
			Description:    "负责 Java 后端开发",
			Requirements:   "3年 Java 经验",
			ExternalSource: "tencent_join",
			ExternalID:     &externalID2,
			DataQuality:    "full_jd",
			IsFeatured:     false,
		},
	}
}

// minimalRecParsedOutput returns a valid recommendation LLM output for testing.
func minimalRecParsedOutput(jobID uint64) map[string]any {
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
				"source_url":      "",
				"recommend_level": "priority_apply",
				"recommend_label": "强烈推荐",
				"match_score":     float64(85),
				"offer_go_summary": "你的 Go 和 MySQL 经验非常匹配",
				"recommend_reason": "目标岗位与简历高度一致，技能匹配度高",
				"evidence": []any{
					map[string]any{
						"job_requirement":  "3年 Go 经验",
						"resume_evidence":  "技能列表包含 Go",
						"judgement":        "matched",
						"explanation":      "技能完全匹配",
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
				"apply_readiness":           "quick_fix_before_apply",
				"optimization_stop_signal":  "fix_top_3_then_apply",
				"top_fix_items": []any{
					map[string]any{
						"issue":             "项目描述缺少量化数据",
						"why_it_matters":    "量化数据能让初筛更有说服力",
						"action":            "回顾项目经历，补充可量化的成果",
						"example_reference": "示例参考：独立完成 12 个 RESTful API 接口开发",
						"self_check_question": "请确认所有数字均为真实数据",
						"estimated_effort":  "within_30_min",
					},
				},
			},
		},
		"overall_risks": []any{
			map[string]any{
				"risk":       "部分岗位需要 Java 经验",
				"severity":   "low",
				"suggestion": "可以学习 Java 基础",
			},
		},
		"resume_focus": []any{
			map[string]any{
				"focus":           "强化 Go 项目经验",
				"why_it_matters":  "岗位核心要求",
				"suggested_action": "补充项目细节",
			},
		},
		"authenticity_notice": "请基于真实经历优化简历，不要虚构项目或能力。",
		"analysis_basis": map[string]any{
			"candidate_count":     float64(2),
			"data_quality_filter": "full_jd",
		},
		"generated_by": "llm_agent",
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRecommendationAgent_LLM_Success(t *testing.T) {
	parsed := minimalRecParsedOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:     pipeline.StatusSuccess,
			GeneratedBy:     "llm_agent",
			Parsed:          parsed,
			JSONText:        string(parsedJSON),
			FallbackRequired: false,
			RepairApplied:   false,
			DurationMs:      1200,
		},
	}

	agent := NewRecommendationAgent(runner)
	results, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(results))
	}
	if results[0].JobID != 1 {
		t.Errorf("expected job_id=1, got %d", results[0].JobID)
	}
	if results[0].Priority != "strong_recommend" {
		t.Errorf("expected priority=strong_recommend, got %s", results[0].Priority)
	}
	if results[0].RecommendationScore != 85 {
		t.Errorf("expected recommendation_score=85, got %d", results[0].RecommendationScore)
	}
	if results[0].RecommendLabel != "强烈推荐" {
		t.Errorf("expected recommend_label=强烈推荐, got %s", results[0].RecommendLabel)
	}
	if results[0].OfferGoSummary != "你的 Go 和 MySQL 经验非常匹配" {
		t.Errorf("expected offer_go_summary, got %s", results[0].OfferGoSummary)
	}
	if len(results[0].MatchHighlights) == 0 {
		t.Error("expected match_highlights")
	}
	if len(results[0].RiskPoints) == 0 {
		t.Error("expected risk_points")
	}
	if len(results[0].QuerySuggestions) == 0 {
		t.Error("expected query_suggestions")
	}
}

func TestRecommendationAgent_LLM_Repaired(t *testing.T) {
	parsed := minimalRecParsedOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:     pipeline.StatusRepaired,
			GeneratedBy:     "llm_agent",
			Parsed:          parsed,
			JSONText:        string(parsedJSON),
			FallbackRequired: false,
			RepairApplied:   true,
			DurationMs:      2500,
		},
	}

	agent := NewRecommendationAgent(runner)
	results, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(results))
	}
	if results[0].JobID != 1 {
		t.Errorf("expected job_id=1, got %d", results[0].JobID)
	}
}

func TestRecommendationAgent_FallbackRequired(t *testing.T) {
	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "LLM service unavailable",
		},
	}

	agent := NewRecommendationAgent(runner)
	_, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err == nil {
		t.Fatal("expected fallback error, got nil")
	}
	fbErr, ok := err.(*RecFallbackError)
	if !ok {
		t.Fatalf("expected RecFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("expected error_code=LLM_HTTP_ERROR, got %s", fbErr.ErrorCode)
	}
}

func TestRecommendationAgent_InvalidJobID_Filtered(t *testing.T) {
	// LLM returns a job_id that doesn't exist in candidates
	parsed := minimalRecParsedOutput(999)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:     pipeline.StatusSuccess,
			GeneratedBy:     "llm_agent",
			Parsed:          parsed,
			JSONText:        string(parsedJSON),
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	_, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err == nil {
		t.Fatal("expected error when all job_ids are invalid, got nil")
	}
}

func TestRecommendationAgent_EmptyRecommendations(t *testing.T) {
	parsed := map[string]any{
		"strategy_summary": "无匹配岗位",
		"recommendations": []any{},
		"overall_risks": []any{},
		"resume_focus": []any{},
		"authenticity_notice": "暂无推荐",
		"analysis_basis": map[string]any{},
		"generated_by": "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:     pipeline.StatusSuccess,
			GeneratedBy:     "llm_agent",
			Parsed:          parsed,
			JSONText:        string(parsedJSON),
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	_, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err == nil {
		t.Fatal("expected error for empty recommendations, got nil")
	}
}

func TestRecommendationAgent_ParsedNil(t *testing.T) {
	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:     pipeline.StatusSuccess,
			GeneratedBy:     "llm_agent",
			Parsed:          nil,
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	_, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err == nil {
		t.Fatal("expected error for nil parsed, got nil")
	}
}

func TestRecommendationAgent_MixedValidInvalidJobIDs(t *testing.T) {
	// One valid job_id (1) and one invalid (999)
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
				"external_source": "local_mock",
				"external_id":     "ext-001",
				"data_quality":    "full_jd",
				"recommend_level": "priority_apply",
				"recommend_label": "强烈推荐",
				"match_score":     float64(85),
				"offer_go_summary": "匹配度高",
				"recommend_reason": "技能匹配",
				"evidence":        []any{},
				"risk_points":     []any{},
				"next_action":     map[string]any{"action": "投递"},
				"apply_readiness":           "ready_to_apply",
				"optimization_stop_signal":  "stop_and_apply",
				"top_fix_items":             []any{},
			},
			map[string]any{
				"job_id":          float64(999),
				"title":           "虚构岗位",
				"department":      "虚构部门",
				"city":            "虚构城市",
				"category":        "技术",
				"job_type":        "full_time",
				"external_source": "local_mock",
				"external_id":     "fake-id",
				"data_quality":    "full_jd",
				"recommend_level": "explore_first",
				"recommend_label": "探索",
				"match_score":     float64(50),
				"offer_go_summary": "",
				"recommend_reason": "虚构理由",
				"evidence":        []any{},
				"risk_points":     []any{},
				"next_action":     map[string]any{"action": "忽略"},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "请基于真实经历",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	results, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 recommendation (invalid job_id filtered), got %d", len(results))
	}
	if results[0].JobID != 1 {
		t.Errorf("expected job_id=1, got %d", results[0].JobID)
	}
}

func TestRecommendationAgent_PromptValuesContainKeys(t *testing.T) {
	agent := NewRecommendationAgent(&fakeRecPipelineRunner{})
	values, err := agent.buildPromptValues(recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	requiredKeys := []string{"RESUME_CONTEXT_JSON", "JOB_CANDIDATES_JSON", "BUSINESS_CONTEXT"}
	for _, key := range requiredKeys {
		if _, ok := values[key]; !ok {
			t.Errorf("expected prompt value key %s", key)
		}
	}

	// Verify JOB_CANDIDATES_JSON contains job_id
	if values["JOB_CANDIDATES_JSON"] == "" {
		t.Error("JOB_CANDIDATES_JSON should not be empty")
	}

	// Verify no API key leakage
	for _, v := range values {
		if containsAPIKeyPattern(v) {
			t.Error("prompt values should not contain API key patterns")
		}
	}
}

func TestRecommendationAgent_NoSensitiveDataInResults(t *testing.T) {
	parsed := minimalRecParsedOutput(1)
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			RawContent:       "some raw output with sk-abc123 apikey value",
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	results, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify no API key in serialized results
	resultJSON, _ := json.Marshal(results)
	if containsAPIKeyPattern(string(resultJSON)) {
		t.Error("serialized results should not contain API key patterns")
	}
}

func TestRecommendationAgent_RecommendLevelMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"priority_apply", "strong_recommend"},
		{"recommended", "recommend"},
		{"explore_first", "try"},
		{"not_recommended", "try"},
		{"strong_recommend", "strong_recommend"},
		{"recommend", "recommend"},
		{"try", "try"},
		{"unknown_value", "try"},
	}

	for _, tt := range tests {
		got := mapRecommendLevel(tt.input)
		if got != tt.expected {
			t.Errorf("mapRecommendLevel(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRecommendationAgent_FieldsFromCandidate(t *testing.T) {
	// Verify that job fields come from candidate, not LLM output
	parsed := map[string]any{
		"strategy_summary": "推荐",
		"recommendations": []any{
			map[string]any{
				"job_id":          float64(1),
				"title":           "LLM 编造的标题", // LLM tries to change title
				"department":      "LLM 编造的部门",
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
				"apply_readiness":           "quick_fix_before_apply",
				"optimization_stop_signal":  "fix_top_3_then_apply",
				"top_fix_items":             []any{},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "提醒",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	results, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Title/department/city should come from candidate, not LLM
	if results[0].Title != "Go 后端开发工程师" {
		t.Errorf("title should come from candidate (Go 后端开发工程师), got %s", results[0].Title)
	}
	if results[0].Department != "技术部" {
		t.Errorf("department should come from candidate (技术部), got %s", results[0].Department)
	}
	if results[0].City != "深圳" {
		t.Errorf("city should come from candidate (深圳), got %s", results[0].City)
	}
}

func TestRecommendationAgent_MissingJobID_Skipped(t *testing.T) {
	// Recommendation without job_id should be skipped
	parsed := map[string]any{
		"strategy_summary": "推荐",
		"recommendations": []any{
			map[string]any{
				"title":           "无 job_id 岗位",
				"department":      "部门",
				"recommend_level": "explore_first",
				"match_score":     float64(50),
				"recommend_reason": "推荐",
				"evidence":        []any{},
				"risk_points":     []any{},
			},
		},
		"overall_risks":       []any{},
		"resume_focus":        []any{},
		"authenticity_notice": "提醒",
		"analysis_basis":      map[string]any{},
		"generated_by":        "llm_agent",
	}
	parsedJSON, _ := json.Marshal(parsed)

	runner := &fakeRecPipelineRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusSuccess,
			GeneratedBy:      "llm_agent",
			Parsed:           parsed,
			JSONText:         string(parsedJSON),
			FallbackRequired: false,
		},
	}

	agent := NewRecommendationAgent(runner)
	_, err := agent.Recommend(context.Background(), recTestResume(), recTestJobs())
	if err == nil {
		t.Fatal("expected error (all filtered), got nil")
	}
}

// containsAPIKeyPattern checks for common API key patterns.
func containsAPIKeyPattern(s string) bool {
	patterns := []string{"sk-", "api.key", "apikey", "secret_key", "bearer "}
	for _, p := range patterns {
		if len(s) > 0 && contains(s, p) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
