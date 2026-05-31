package model

import (
	"encoding/json"
	"testing"
)

func TestAgentResponseMetaJSONMarshal(t *testing.T) {
	meta := AgentResponseMeta{
		AgentName:      AgentNameMatchAnalysis,
		GeneratedBy:    GeneratedByLLMAgent,
		AgentStatus:    AgentStatusSuccess,
		FallbackReason: FallbackReasonNone,
		ErrorCode:      ErrorCodeNone,
		RepairApplied:  false,
		SchemaValid:    true,
		RequestID:      "req-20260531-abc123",
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal AgentResponseMeta: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal AgentResponseMeta: %v", err)
	}

	// Verify all fields are present
	expectedFields := []string{
		"agent_name", "generated_by", "agent_status",
		"fallback_reason", "error_code", "repair_applied",
		"schema_valid", "request_id",
	}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing field %q in marshalled AgentResponseMeta", field)
		}
	}

	// Verify values
	if result["agent_name"] != "match_analysis" {
		t.Errorf("agent_name = %v, want match_analysis", result["agent_name"])
	}
	if result["generated_by"] != "llm_agent" {
		t.Errorf("generated_by = %v, want llm_agent", result["generated_by"])
	}
	if result["agent_status"] != "success" {
		t.Errorf("agent_status = %v, want success", result["agent_status"])
	}
	if result["fallback_reason"] != "none" {
		t.Errorf("fallback_reason = %v, want none", result["fallback_reason"])
	}
	if result["error_code"] != "none" {
		t.Errorf("error_code = %v, want none", result["error_code"])
	}
	if result["repair_applied"] != false {
		t.Errorf("repair_applied = %v, want false", result["repair_applied"])
	}
	if result["schema_valid"] != true {
		t.Errorf("schema_valid = %v, want true", result["schema_valid"])
	}
	if result["request_id"] != "req-20260531-abc123" {
		t.Errorf("request_id = %v, want req-20260531-abc123", result["request_id"])
	}
}

func TestAgentResponseMetaNoOmitEmpty(t *testing.T) {
	// Zero-value meta should still have all fields present in JSON
	meta := AgentResponseMeta{}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal zero-value AgentResponseMeta: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal zero-value AgentResponseMeta: %v", err)
	}

	// Even with zero values, all fields must be present
	expectedFields := []string{
		"agent_name", "generated_by", "agent_status",
		"fallback_reason", "error_code", "repair_applied",
		"schema_valid", "request_id",
	}
	for _, field := range expectedFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Missing field %q in zero-value AgentResponseMeta (omitempty leak)", field)
		}
	}
}

func TestNewAgentResponseMeta(t *testing.T) {
	meta := NewAgentResponseMeta(AgentNameMatchAnalysis)

	if meta.AgentName != AgentNameMatchAnalysis {
		t.Errorf("AgentName = %q, want %q", meta.AgentName, AgentNameMatchAnalysis)
	}
	if meta.GeneratedBy != GeneratedByMockAgent {
		t.Errorf("GeneratedBy = %q, want %q", meta.GeneratedBy, GeneratedByMockAgent)
	}
	if meta.AgentStatus != AgentStatusMockSuccess {
		t.Errorf("AgentStatus = %q, want %q", meta.AgentStatus, AgentStatusMockSuccess)
	}
	if meta.FallbackReason != FallbackReasonNone {
		t.Errorf("FallbackReason = %q, want %q", meta.FallbackReason, FallbackReasonNone)
	}
	if meta.ErrorCode != ErrorCodeNone {
		t.Errorf("ErrorCode = %q, want %q", meta.ErrorCode, ErrorCodeNone)
	}
	if meta.RepairApplied != false {
		t.Errorf("RepairApplied = %v, want false", meta.RepairApplied)
	}
	if meta.SchemaValid != true {
		t.Errorf("SchemaValid = %v, want true", meta.SchemaValid)
	}
	if meta.RequestID != "" {
		t.Errorf("RequestID = %q, want empty string", meta.RequestID)
	}
}

func TestDeriveMetaFromResult(t *testing.T) {
	tests := []struct {
		name           string
		agentName      string
		generatedBy    string
		agentStatus    string
		fallbackReason string
		errorCode      string
		repairApplied  bool
		wantStatus     string
		wantGenerated  string
		wantReason     string
		wantErrCode    string
	}{
		{
			name:           "LLM success",
			agentName:      AgentNameMatchAnalysis,
			generatedBy:    GeneratedByLLMAgent,
			agentStatus:    AgentStatusSuccess,
			fallbackReason: "",
			errorCode:      "",
			repairApplied:  false,
			wantStatus:     AgentStatusSuccess,
			wantGenerated:  GeneratedByLLMAgent,
			wantReason:     FallbackReasonNone,
			wantErrCode:    ErrorCodeNone,
		},
		{
			name:           "LLM repaired",
			agentName:      AgentNameMatchAnalysis,
			generatedBy:    GeneratedByLLMAgent,
			agentStatus:    AgentStatusRepaired,
			fallbackReason: "",
			errorCode:      "",
			repairApplied:  true,
			wantStatus:     AgentStatusRepaired,
			wantGenerated:  GeneratedByLLMAgent,
			wantReason:     FallbackReasonNone,
			wantErrCode:    ErrorCodeNone,
		},
		{
			name:           "Fallback",
			agentName:      AgentNameRecommendation,
			generatedBy:    GeneratedByMockAgent,
			agentStatus:    AgentStatusFallback,
			fallbackReason: "json_repair_failed",
			errorCode:      "JSON_REPAIR_FAILED",
			repairApplied:  false,
			wantStatus:     AgentStatusFallback,
			wantGenerated:  GeneratedByMockAgent,
			wantReason:     "json_repair_failed",
			wantErrCode:    "JSON_REPAIR_FAILED",
		},
		{
			name:           "Mock direct with old mock status",
			agentName:      AgentNameMatchAnalysis,
			generatedBy:    GeneratedByMockAgent,
			agentStatus:    "mock",
			fallbackReason: "",
			errorCode:      "",
			repairApplied:  false,
			wantStatus:     AgentStatusMockSuccess,
			wantGenerated:  GeneratedByMockAgent,
			wantReason:     FallbackReasonNone,
			wantErrCode:    ErrorCodeNone,
		},
		{
			name:           "Mock direct with new mock_success status",
			agentName:      AgentNameResumeOptimization,
			generatedBy:    GeneratedByMockAgent,
			agentStatus:    AgentStatusMockSuccess,
			fallbackReason: "",
			errorCode:      "",
			repairApplied:  false,
			wantStatus:     AgentStatusMockSuccess,
			wantGenerated:  GeneratedByMockAgent,
			wantReason:     FallbackReasonNone,
			wantErrCode:    ErrorCodeNone,
		},
		{
			name:           "Empty values default to mock_success",
			agentName:      AgentNameRecommendation,
			generatedBy:    "",
			agentStatus:    "",
			fallbackReason: "",
			errorCode:      "",
			repairApplied:  false,
			wantStatus:     AgentStatusMockSuccess,
			wantGenerated:  GeneratedByMockAgent,
			wantReason:     FallbackReasonNone,
			wantErrCode:    ErrorCodeNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := DeriveMetaFromResult(tt.agentName, tt.generatedBy, tt.agentStatus, tt.fallbackReason, tt.errorCode, tt.repairApplied)

			if meta.AgentName != tt.agentName {
				t.Errorf("AgentName = %q, want %q", meta.AgentName, tt.agentName)
			}
			if meta.AgentStatus != tt.wantStatus {
				t.Errorf("AgentStatus = %q, want %q", meta.AgentStatus, tt.wantStatus)
			}
			if meta.GeneratedBy != tt.wantGenerated {
				t.Errorf("GeneratedBy = %q, want %q", meta.GeneratedBy, tt.wantGenerated)
			}
			if meta.FallbackReason != tt.wantReason {
				t.Errorf("FallbackReason = %q, want %q", meta.FallbackReason, tt.wantReason)
			}
			if meta.ErrorCode != tt.wantErrCode {
				t.Errorf("ErrorCode = %q, want %q", meta.ErrorCode, tt.wantErrCode)
			}
			if meta.SchemaValid != true {
				t.Errorf("SchemaValid = %v, want true", meta.SchemaValid)
			}
		})
	}
}

func TestWithFallback(t *testing.T) {
	meta := NewAgentResponseMeta(AgentNameMatchAnalysis)
	meta = meta.WithFallback("json_repair_failed", "JSON_REPAIR_FAILED")

	if meta.AgentStatus != AgentStatusFallback {
		t.Errorf("AgentStatus = %q, want %q", meta.AgentStatus, AgentStatusFallback)
	}
	if meta.FallbackReason != "json_repair_failed" {
		t.Errorf("FallbackReason = %q, want json_repair_failed", meta.FallbackReason)
	}
	if meta.ErrorCode != "JSON_REPAIR_FAILED" {
		t.Errorf("ErrorCode = %q, want JSON_REPAIR_FAILED", meta.ErrorCode)
	}
	if meta.GeneratedBy != GeneratedByMockAgent {
		t.Errorf("GeneratedBy = %q, want %q", meta.GeneratedBy, GeneratedByMockAgent)
	}
}

func TestWithRepaired(t *testing.T) {
	meta := NewAgentResponseMeta(AgentNameMatchAnalysis)
	meta = meta.WithRepaired()

	if meta.AgentStatus != AgentStatusRepaired {
		t.Errorf("AgentStatus = %q, want %q", meta.AgentStatus, AgentStatusRepaired)
	}
	if meta.RepairApplied != true {
		t.Errorf("RepairApplied = %v, want true", meta.RepairApplied)
	}
	if meta.GeneratedBy != GeneratedByLLMAgent {
		t.Errorf("GeneratedBy = %q, want %q", meta.GeneratedBy, GeneratedByLLMAgent)
	}
}

func TestAgentResponseMetaInEnvelope(t *testing.T) {
	// Test that meta can be embedded in a response envelope structure
	type TestEnvelope struct {
		Result interface{}       `json:"result"`
		Meta   AgentResponseMeta `json:"meta"`
	}

	envelope := TestEnvelope{
		Result: map[string]interface{}{
			"match_score": 82,
		},
		Meta: AgentResponseMeta{
			AgentName:      AgentNameMatchAnalysis,
			GeneratedBy:    GeneratedByLLMAgent,
			AgentStatus:    AgentStatusSuccess,
			FallbackReason: FallbackReasonNone,
			ErrorCode:      ErrorCodeNone,
			RepairApplied:  false,
			SchemaValid:    true,
			RequestID:      "",
		},
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal envelope: %v", err)
	}

	// Verify result exists
	if _, ok := result["result"]; !ok {
		t.Error("Missing 'result' in envelope")
	}
	// Verify meta exists
	if _, ok := result["meta"]; !ok {
		t.Error("Missing 'meta' in envelope")
	}

	// Verify meta fields inside envelope
	metaMap, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("meta is not a map")
	}
	if metaMap["agent_name"] != "match_analysis" {
		t.Errorf("meta.agent_name = %v, want match_analysis", metaMap["agent_name"])
	}
}
