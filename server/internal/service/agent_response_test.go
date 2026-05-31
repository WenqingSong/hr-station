package service

import (
	"encoding/json"
	"testing"

	"offer-catcher/server/internal/model"
)

func TestMatchResponseContainsMeta(t *testing.T) {
	resp := &MatchResponse{
		AnalysisID:  12,
		GeneratedBy: "mock_agent",
		Result: &model.MatchAnalysisResult{
			JobID:       1,
			MatchScore:  78,
			GeneratedBy: "mock_agent",
			AgentStatus: model.AgentStatusMockSuccess,
		},
		Meta: model.NewAgentResponseMeta(model.AgentNameMatchAnalysis),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal MatchResponse: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal MatchResponse: %v", err)
	}

	// Verify result exists
	if _, ok := result["result"]; !ok {
		t.Error("Missing 'result' in MatchResponse")
	}
	// Verify meta exists
	if _, ok := result["meta"]; !ok {
		t.Error("Missing 'meta' in MatchResponse")
	}
	// Verify meta fields
	metaMap, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if metaMap["agent_name"] != "match_analysis" {
		t.Errorf("meta.agent_name = %v, want match_analysis", metaMap["agent_name"])
	}
	if metaMap["agent_status"] != "mock_success" {
		t.Errorf("meta.agent_status = %v, want mock_success", metaMap["agent_status"])
	}
	// Verify deprecated generated_by still present
	if _, ok := result["generated_by"]; !ok {
		t.Error("Missing deprecated 'generated_by' field")
	}
}

func TestRecommendationResponseContainsResultAndMeta(t *testing.T) {
	jobs := []*model.RecommendedJob{
		{
			JobID:       1,
			Title:       "Go 后端开发工程师",
			GeneratedBy: "mock_agent",
			AgentStatus: model.AgentStatusMockSuccess,
		},
	}

	resp := &RecommendationResponse{
		ResumeVersionID: 2,
		ResumeVersionNo: 2,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: jobs,
		Result:          &RecommendationBusinessResult{RecommendedJobs: jobs},
		Summary:         "测试摘要",
		CreatedAt:       "2026-05-31T12:00:00+08:00",
		Meta:            model.NewAgentResponseMeta(model.AgentNameRecommendation),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal RecommendationResponse: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal RecommendationResponse: %v", err)
	}

	// Verify result.recommended_jobs exists (new entry)
	resultMap, ok := result["result"].(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	if _, ok := resultMap["recommended_jobs"]; !ok {
		t.Error("Missing 'result.recommended_jobs' in RecommendationResponse")
	}
	// Verify deprecated recommended_jobs still present at data level
	if _, ok := result["recommended_jobs"]; !ok {
		t.Error("Missing deprecated 'recommended_jobs' at data level")
	}
	// Verify meta exists
	if _, ok := result["meta"]; !ok {
		t.Error("Missing 'meta' in RecommendationResponse")
	}
	// Verify meta fields
	metaMap, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if metaMap["agent_name"] != "recommendation" {
		t.Errorf("meta.agent_name = %v, want recommendation", metaMap["agent_name"])
	}
}

func TestResumeOptimizationResponseContainsMeta(t *testing.T) {
	resp := &ResumeOptimizationResponse{
		OptimizationID: 0,
		GeneratedBy:    "mock_agent",
		Result: &model.ResumeOptimizationResult{
			AgentStatus:    model.AgentStatusMockSuccess,
			GeneratedBy:    "mock_agent",
			ApplyReadiness: "quick_fix_before_apply",
			Summary:        "测试",
		},
		Meta: model.NewAgentResponseMeta(model.AgentNameResumeOptimization),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ResumeOptimizationResponse: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal ResumeOptimizationResponse: %v", err)
	}

	// Verify result exists
	if _, ok := result["result"]; !ok {
		t.Error("Missing 'result' in ResumeOptimizationResponse")
	}
	// Verify meta exists
	if _, ok := result["meta"]; !ok {
		t.Error("Missing 'meta' in ResumeOptimizationResponse")
	}
	// Verify meta fields
	metaMap, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if metaMap["agent_name"] != "resume_optimization" {
		t.Errorf("meta.agent_name = %v, want resume_optimization", metaMap["agent_name"])
	}
}

func TestMatchResponseMetaFieldsComplete(t *testing.T) {
	meta := model.NewAgentResponseMeta(model.AgentNameMatchAnalysis)

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal meta: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal meta: %v", err)
	}

	requiredFields := []string{
		"agent_name", "generated_by", "agent_status",
		"fallback_reason", "error_code", "repair_applied",
		"schema_valid", "request_id",
	}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing meta field: %s", field)
		}
	}
}

func TestRecommendationResponseMetaFieldsComplete(t *testing.T) {
	meta := model.NewAgentResponseMeta(model.AgentNameRecommendation)

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal meta: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal meta: %v", err)
	}

	requiredFields := []string{
		"agent_name", "generated_by", "agent_status",
		"fallback_reason", "error_code", "repair_applied",
		"schema_valid", "request_id",
	}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing meta field: %s", field)
		}
	}
}

func TestResumeOptimizationResponseMetaFieldsComplete(t *testing.T) {
	meta := model.NewAgentResponseMeta(model.AgentNameResumeOptimization)

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal meta: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal meta: %v", err)
	}

	requiredFields := []string{
		"agent_name", "generated_by", "agent_status",
		"fallback_reason", "error_code", "repair_applied",
		"schema_valid", "request_id",
	}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing meta field: %s", field)
		}
	}
}

func TestMockDirectAgentStatusIsMockSuccess(t *testing.T) {
	// Verify that mock direct (no fallback error) results in mock_success status
	meta := model.NewAgentResponseMeta(model.AgentNameMatchAnalysis)
	if meta.AgentStatus != model.AgentStatusMockSuccess {
		t.Errorf("Default agent_status = %q, want %q", meta.AgentStatus, model.AgentStatusMockSuccess)
	}

	meta2 := model.NewAgentResponseMeta(model.AgentNameRecommendation)
	if meta2.AgentStatus != model.AgentStatusMockSuccess {
		t.Errorf("Default agent_status = %q, want %q", meta2.AgentStatus, model.AgentStatusMockSuccess)
	}

	meta3 := model.NewAgentResponseMeta(model.AgentNameResumeOptimization)
	if meta3.AgentStatus != model.AgentStatusMockSuccess {
		t.Errorf("Default agent_status = %q, want %q", meta3.AgentStatus, model.AgentStatusMockSuccess)
	}
}

func TestFallbackMetaStatusIsFallback(t *testing.T) {
	meta := model.NewAgentResponseMeta(model.AgentNameMatchAnalysis)
	meta = meta.WithFallback("json_repair_failed", "JSON_REPAIR_FAILED")
	if meta.AgentStatus != model.AgentStatusFallback {
		t.Errorf("Fallback agent_status = %q, want %q", meta.AgentStatus, model.AgentStatusFallback)
	}
	if meta.FallbackReason != "json_repair_failed" {
		t.Errorf("FallbackReason = %q, want json_repair_failed", meta.FallbackReason)
	}
	if meta.ErrorCode != "JSON_REPAIR_FAILED" {
		t.Errorf("ErrorCode = %q, want JSON_REPAIR_FAILED", meta.ErrorCode)
	}
}

func TestRecommendationOldAndNewRecommendedJobsConsistency(t *testing.T) {
	jobs := []*model.RecommendedJob{
		{JobID: 1, Title: "Job 1"},
		{JobID: 2, Title: "Job 2"},
	}

	resp := &RecommendationResponse{
		ResumeVersionID: 2,
		ResumeVersionNo: 2,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: jobs,
		Result:          &RecommendationBusinessResult{RecommendedJobs: jobs},
		Meta:            model.NewAgentResponseMeta(model.AgentNameRecommendation),
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Both recommended_jobs (deprecated) and result.recommended_jobs should exist
	oldJobs, ok1 := result["recommended_jobs"].([]interface{})
	resultMap, ok2 := result["result"].(map[string]interface{})
	newJobs, ok3 := resultMap["recommended_jobs"].([]interface{})

	if !ok1 {
		t.Error("Deprecated recommended_jobs not found")
	}
	if !ok2 {
		t.Error("result not found")
	}
	if !ok3 {
		t.Error("result.recommended_jobs not found")
	}
	if len(oldJobs) != len(newJobs) {
		t.Errorf("recommended_jobs length = %d, result.recommended_jobs length = %d, should be equal", len(oldJobs), len(newJobs))
	}
}
