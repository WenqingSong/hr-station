package agent

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// Fixed test resume loader (for Smoke and future 4-3.8 validation)
// ---------------------------------------------------------------------------

// testResumeFromFile loads a resume from server/testdata/resumes/ and converts
// it to a model.ResumeVersion suitable for Agent consumption.
func testResumeFromFile(t *testing.T, filename string) *model.ResumeVersion {
	t.Helper()
	data, err := os.ReadFile("../../testdata/resumes/" + filename)
	if err != nil {
		t.Fatalf("failed to read test resume %s: %v", filename, err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse test resume %s: %v", filename, err)
	}

	// Extract fields matching model.ResumeVersion
	skillsJSON, _ := json.Marshal(raw["skills"])
	educationJSON, _ := json.Marshal(raw["education"])
	experienceJSON, _ := json.Marshal(raw["experience"])
	projectJSON, _ := json.Marshal(raw["project"])

	selfEval := ""
	if v, ok := raw["self_evaluation"].(string); ok {
		selfEval = v
	}
	targetRole := ""
	if v, ok := raw["target_role"].(string); ok {
		targetRole = v
	}
	targetCity := ""
	if v, ok := raw["target_city"].(string); ok {
		targetCity = v
	}

	id := uint64(0)
	if v, ok := raw["id"].(float64); ok {
		id = uint64(v)
	}
	versionNo := 1
	if v, ok := raw["version_no"].(float64); ok {
		versionNo = int(v)
	}

	return &model.ResumeVersion{
		ID:             id,
		VersionNo:      versionNo,
		TargetRole:     targetRole,
		TargetCity:     targetCity,
		SkillsJSON:     skillsJSON,
		EducationJSON:  educationJSON,
		ExperienceJSON: experienceJSON,
		ProjectJSON:    projectJSON,
		SelfEvaluation: selfEval,
	}
}

// fixedBackendJD returns a fixed JD for smoke testing.
func fixedBackendJD() *model.Job {
	skillKeywords, _ := json.Marshal([]string{"Go", "MySQL", "Redis", "HTTP API", "Docker"})
	return &model.Job{
		ID:                9001,
		Title:             "后台开发实习生",
		Department:        "技术部",
		City:              "深圳",
		Category:          "后端开发",
		Description:       "参与后台系统开发与维护，负责接口设计和实现。",
		Requirements:      "1. 熟悉 Go 或 Java 至少一种后端语言；2. 熟悉 HTTP API 开发；3. 熟悉 MySQL，了解索引、事务；4. 了解 Redis、缓存、登录鉴权优先；5. 有项目开发、问题排查、接口联调经验；6. 具备良好代码习惯和学习能力。",
		DataQuality:       "full_jd",
		SkillKeywordsJSON: skillKeywords,
		ExternalSource:    "test_fixture",
	}
}

// ---------------------------------------------------------------------------
// Smoke Test: MatchAnalysisAgent returns decision fields (4-3.7-fix.4)
// ---------------------------------------------------------------------------

// TestMatchAnalysisAgentReturnsDecisionFieldsSmoke verifies that the
// MatchMockAgent (local deterministic path) returns all critical decision
// fields: apply_readiness, optimization_stop_signal, top_fix_items,
// and evidence/match fields.
//
// This test uses the Mock Agent path, NOT a real LLM.
// Real LLM + fallback behavior is validated in 4-3.8.
func TestMatchAnalysisAgentReturnsDecisionFieldsSmoke(t *testing.T) {
	// Load the "strong Go backend" test resume
	resume := testResumeFromFile(t, "resume_01_go_backend_strong.json")
	job := fixedBackendJD()

	// Use Mock Agent (local deterministic path)
	mockAgent := NewMatchMockAgent()
	result, err := mockAgent.Analyze(context.Background(), resume, job)
	if err != nil {
		t.Fatalf("Mock Agent Analyze() failed: %v", err)
	}

	// 1. Agent call does not panic (reaching here proves it)
	// 2. Result is non-nil (already checked above)

	// 3. Result content is non-empty
	if result.MatchScore == 0 && result.ScreeningResult == "" {
		t.Error("expected non-empty match result content")
	}

	// 4. apply_readiness is non-empty
	if result.ApplyReadiness == "" {
		t.Error("expected non-empty apply_readiness, got empty string")
	}
	t.Logf("apply_readiness = %q", result.ApplyReadiness)

	// 5. optimization_stop_signal is non-empty
	if result.OptimizationStopSignal == "" {
		t.Error("expected non-empty optimization_stop_signal, got empty string")
	}
	t.Logf("optimization_stop_signal = %q", result.OptimizationStopSignal)

	// 6. top_fix_items field exists (may be empty slice but not nil)
	if result.TopFixItems == nil {
		t.Error("expected top_fix_items to be non-nil (can be empty slice), got nil")
	}
	t.Logf("top_fix_items count = %d", len(result.TopFixItems))

	// 7. At least one evidence/match/summary field is non-empty
	hasEvidence := len(result.EvidenceChain) > 0
	hasHighlights := len(result.MatchHighlights) > 0
	hasSummary := result.OfferGoSummary != ""
	if !hasEvidence && !hasHighlights && !hasSummary {
		t.Error("expected at least one of evidence_chain, match_highlights, or offer_go_summary to be non-empty")
	}
	t.Logf("evidence_chain count = %d, match_highlights count = %d, offer_go_summary empty = %v",
		len(result.EvidenceChain), len(result.MatchHighlights), result.OfferGoSummary == "")

	// 8. generated_by / fallback identifier not corrupted
	// Mock Agent does not set GeneratedBy (it's empty by default for Mock)
	// This is expected behavior: Mock results don't have generated_by set.
	// The important check is that the field isn't set to an invalid value.
	if result.GeneratedBy != "" && result.GeneratedBy != "mock_agent" && result.GeneratedBy != "llm_agent" {
		t.Errorf("generated_by has unexpected value: %q", result.GeneratedBy)
	}

	t.Logf("=== Smoke Test Summary ===")
	t.Logf("match_score = %d", result.MatchScore)
	t.Logf("screening_result = %s", result.ScreeningResult)
	t.Logf("apply_readiness = %s", result.ApplyReadiness)
	t.Logf("optimization_stop_signal = %s", result.OptimizationStopSignal)
	t.Logf("top_fix_items = %d items", len(result.TopFixItems))
	t.Logf("evidence_chain = %d items", len(result.EvidenceChain))
}
