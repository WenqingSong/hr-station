package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
)

// ---------------------------------------------------------------------------
// fakeRunner: stub pipeline.Runner for unit tests
// ---------------------------------------------------------------------------

type fakeRunner struct {
	result *pipeline.RunResult
	err    error // not used by current Runner API, but reserved
}

func (f *fakeRunner) Run(_ context.Context, _ pipeline.RunRequest) *pipeline.RunResult {
	return f.result
}

// ---------------------------------------------------------------------------
// Test data helpers
// ---------------------------------------------------------------------------

func testResume() *model.ResumeVersion {
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

func testJob() *model.Job {
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

// minimalLLMOutput returns a map[string]any that conforms to
// match_analysis_schema.json, suitable for Pipeline Parsed field.
func minimalLLMOutput() map[string]any {
	return map[string]any{
		"match_score":      float64(78),
		"match_level":      "potential_match",
		"match_label":      "建议尝试",
		"screening_result": "likely",
		"offer_go_summary": "你的简历与该岗位有一定匹配度，建议优化后尝试投递。",
		"evidence_chain": []any{
			map[string]any{"jd_requirement": "岗位要求Go相关经验", "resume_evidence": "简历中有Go项目经验", "judgement": "matched", "explanation": "简历中有明确的Go项目经验"},
			map[string]any{"jd_requirement": "岗位要求数据库经验", "resume_evidence": "简历中有MySQL使用经验", "judgement": "partial", "explanation": "有数据库使用经验但深度不足"},
			map[string]any{"jd_requirement": "岗位要求微服务经验", "resume_evidence": "简历中暂未看到直接证据", "judgement": "missing", "explanation": "简历中没有微服务相关描述"},
		},
		"matched_points": []any{
			map[string]any{"point": "Go项目经验匹配", "why_it_matters": "岗位核心要求Go语言", "evidence": "简历中有Go项目经验"},
			map[string]any{"point": "数据库使用经验", "why_it_matters": "后端开发需要数据库能力", "evidence": "简历中有MySQL经验"},
		},
		"gap_points": []any{
			map[string]any{"gap": "缺少微服务经验", "impact": "可能影响技术面试表现", "how_to_improve": "建议学习微服务框架并完成实践项目"},
		},
		"risk_points": []any{
			map[string]any{"risk": "缺少微服务经验", "severity": "medium", "reason": "岗位可能涉及微服务架构", "mitigation": "通过在线课程补充微服务知识"},
		},
		"resume_optimization": map[string]any{
			"p0": []any{
				map[string]any{"problem": "项目描述缺少量化数据", "why_it_matters": "量化数据能让初筛更有说服力", "action": "回顾项目经历，补充可量化的成果"},
			},
			"p1": []any{
				map[string]any{"problem": "技能栏缺少微服务关键词", "why_it_matters": "岗位可能通过关键词筛选", "action": "学习并在技能栏中添加微服务相关技能"},
			},
			"p2": []any{},
		},
		"interview_questions": []any{
			map[string]any{"question": "请介绍一下你做过的Go项目", "why_ask": "考察Go语言实际使用能力", "how_to_prepare": "准备项目的技术架构和关键决策说明"},
		},
		"next_actions": []any{
			map[string]any{"action": "优化项目描述，补充量化成果", "priority": "p0", "expected_effect": "提升初筛通过率"},
			map[string]any{"action": "学习微服务基础知识", "priority": "p1", "expected_effect": "增加技术面试表现"},
		},
		"authenticity_notice": "以上建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。",
		"analysis_basis": map[string]any{
			"job_description_used":  true,
			"job_requirements_used": true,
			"data_quality":          "full_jd",
		},
	}
}

// ---------------------------------------------------------------------------
// Tests: MatchAnalysisAgent.Analyze
// ---------------------------------------------------------------------------

func TestMatchAnalysisAgent_Analyze_Success(t *testing.T) {
	parsed := minimalLLMOutput()
	agent := NewMatchAnalysisAgent(&fakeRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusSuccess,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: false,
		},
	})

	result, err := agent.Analyze(context.Background(), testResume(), testJob())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify basic fields
	if result.MatchScore != 78 {
		t.Errorf("expected match_score=78, got %d", result.MatchScore)
	}
	if result.MatchLevel != "potential_match" {
		t.Errorf("expected match_level=potential_match, got %s", result.MatchLevel)
	}
	if result.MatchLabel != "建议尝试" {
		t.Errorf("expected match_label=建议尝试, got %s", result.MatchLabel)
	}
	if result.ScreeningResult != "likely" {
		t.Errorf("expected screening_result=likely, got %s", result.ScreeningResult)
	}
	if result.OfferGoSummary == "" {
		t.Error("expected non-empty offer_go_summary")
	}
	if result.GeneratedBy != "llm_agent" {
		t.Errorf("expected generated_by=llm_agent, got %s", result.GeneratedBy)
	}
	if result.AgentStatus != "success" {
		t.Errorf("expected agent_status=success, got %s", result.AgentStatus)
	}
	if result.RepairApplied {
		t.Error("expected repair_applied=false")
	}
}

func TestMatchAnalysisAgent_Analyze_Repaired(t *testing.T) {
	parsed := minimalLLMOutput()
	agent := NewMatchAnalysisAgent(&fakeRunner{
		result: &pipeline.RunResult{
			AgentStatus:  pipeline.StatusRepaired,
			GeneratedBy:  "llm_agent",
			Parsed:       parsed,
			RepairApplied: true,
		},
	})

	result, err := agent.Analyze(context.Background(), testResume(), testJob())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.AgentStatus != "repaired" {
		t.Errorf("expected agent_status=repaired, got %s", result.AgentStatus)
	}
	if !result.RepairApplied {
		t.Error("expected repair_applied=true")
	}
}

func TestMatchAnalysisAgent_Analyze_FallbackRequired(t *testing.T) {
	agent := NewMatchAnalysisAgent(&fakeRunner{
		result: &pipeline.RunResult{
			AgentStatus:      pipeline.StatusFallbackRequired,
			GeneratedBy:      "mock_agent",
			FallbackRequired: true,
			FallbackReason:   "llm_call_failed",
			ErrorCode:        "LLM_HTTP_ERROR",
			ErrorMessage:     "HTTP 500",
		},
	})

	_, err := agent.Analyze(context.Background(), testResume(), testJob())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	fbErr, ok := err.(*MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got %T", err)
	}
	if fbErr.FallbackReason != "llm_call_failed" {
		t.Errorf("expected fallback_reason=llm_call_failed, got %s", fbErr.FallbackReason)
	}
	if fbErr.ErrorCode != "LLM_HTTP_ERROR" {
		t.Errorf("expected error_code=LLM_HTTP_ERROR, got %s", fbErr.ErrorCode)
	}
}

func TestMatchAnalysisAgent_Analyze_FieldMappingFallback(t *testing.T) {
	// Return a parsed output that's missing critical fields — mapping should
	// still succeed (with defaults), so let's test with empty parsed map
	// which will still produce a valid result (with defaults).
	// Instead, test with nil Parsed to trigger mapping fallback.
	agent := NewMatchAnalysisAgent(&fakeRunner{
		result: &pipeline.RunResult{
			AgentStatus: pipeline.StatusSuccess,
			GeneratedBy: "llm_agent",
			Parsed:      nil, // nil parsed should trigger mapping error → fallback
		},
	})

	_, err := agent.Analyze(context.Background(), testResume(), testJob())
	if err == nil {
		t.Fatal("expected error for nil parsed, got nil")
	}

	fbErr, ok := err.(*MatchFallbackError)
	if !ok {
		t.Fatalf("expected *MatchFallbackError, got %T", err)
	}
	if fbErr.ErrorCode != "FIELD_MAPPING_FAILED" {
		t.Errorf("expected error_code=FIELD_MAPPING_FAILED, got %s", fbErr.ErrorCode)
	}
}

// ---------------------------------------------------------------------------
// Tests: Field mapping functions
// ---------------------------------------------------------------------------

func TestMapScreeningResult(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"likely", "likely"},
		{"highly_likely", "likely"},
		{"uncertain", "possible"},
		{"potential_match", "possible"},
		{"explore_first", "possible"},
		{"risky", "risky"},
		{"not_recommended", "risky"},
		{"possible", "possible"},     // existing enum preserved
		{"unknown_value", "possible"}, // default
	}
	for _, tt := range tests {
		got := mapScreeningResult(tt.input)
		if got != tt.expected {
			t.Errorf("mapScreeningResult(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapJudgement(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"matched", "strong_match"},
		{"strong_match", "strong_match"},
		{"partial", "partial_match"},
		{"partial_match", "partial_match"},
		{"missing", "missing"},
		{"unknown", "missing"},
	}
	for _, tt := range tests {
		got := mapJudgement(tt.input)
		if got != tt.expected {
			t.Errorf("mapJudgement(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMapEvidenceChain(t *testing.T) {
	parsed := map[string]any{
		"evidence_chain": []any{
			map[string]any{"jd_requirement": "Go经验", "resume_evidence": "有Go项目", "judgement": "matched"},
			map[string]any{"jd_requirement": "微服务", "resume_evidence": "无", "judgement": "missing"},
		},
	}
	chain := mapEvidenceChain(parsed)
	if len(chain) != 2 {
		t.Fatalf("expected 2 evidence items, got %d", len(chain))
	}
	if chain[0].Judgement != "strong_match" {
		t.Errorf("evidence[0] judgement: expected strong_match, got %s", chain[0].Judgement)
	}
	if chain[0].Requirement != "Go经验" {
		t.Errorf("evidence[0] requirement: expected Go经验, got %s", chain[0].Requirement)
	}
	if chain[1].Judgement != "missing" {
		t.Errorf("evidence[1] judgement: expected missing, got %s", chain[1].Judgement)
	}
}

func TestMapEvidenceChain_Empty(t *testing.T) {
	parsed := map[string]any{}
	chain := mapEvidenceChain(parsed)
	if len(chain) != 1 {
		t.Fatalf("expected 1 default evidence item, got %d", len(chain))
	}
	if chain[0].Judgement != "missing" {
		t.Errorf("default evidence judgement: expected missing, got %s", chain[0].Judgement)
	}
}

func TestMapMatchedPoints(t *testing.T) {
	parsed := map[string]any{
		"matched_points": []any{
			map[string]any{"point": "Go技能匹配"},
			map[string]any{"point": "数据库经验"},
		},
	}
	highlights := mapMatchedPoints(parsed)
	if len(highlights) != 2 {
		t.Fatalf("expected 2 highlights, got %d", len(highlights))
	}
	if highlights[0] != "Go技能匹配" {
		t.Errorf("highlight[0]: expected Go技能匹配, got %s", highlights[0])
	}
}

func TestMapMatchedPoints_Empty(t *testing.T) {
	parsed := map[string]any{}
	highlights := mapMatchedPoints(parsed)
	if len(highlights) != 1 {
		t.Fatalf("expected 1 default highlight, got %d", len(highlights))
	}
}

func TestMapGapPoints(t *testing.T) {
	parsed := map[string]any{
		"gap_points": []any{
			map[string]any{"gap": "缺少微服务经验"},
			map[string]any{"gap": "缺少量化数据"},
		},
	}
	gaps := mapGapPoints(parsed)
	if len(gaps) != 2 {
		t.Fatalf("expected 2 gaps, got %d", len(gaps))
	}
}

func TestMapRiskPoints_Structured(t *testing.T) {
	parsed := map[string]any{
		"risk_points": []any{
			map[string]any{"risk": "缺少微服务经验", "severity": "medium"},
		},
	}
	risks := mapRiskPoints(parsed)
	if len(risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(risks))
	}
	if risks[0] != "缺少微服务经验" {
		t.Errorf("risk[0]: expected 缺少微服务经验, got %s", risks[0])
	}
}

func TestMapRiskPoints_PlainStrings(t *testing.T) {
	parsed := map[string]any{
		"risk_points": []any{
			"城市不一致",
			"经验不足",
		},
	}
	risks := mapRiskPoints(parsed)
	if len(risks) != 2 {
		t.Fatalf("expected 2 risks, got %d", len(risks))
	}
}

func TestMapOptimizationSuggestions(t *testing.T) {
	parsed := map[string]any{
		"resume_optimization": map[string]any{
			"p0": []any{
				map[string]any{"problem": "缺少量化数据", "why_it_matters": "提升说服力", "action": "补充数据"},
			},
			"p1": []any{
				map[string]any{"problem": "缺少关键词", "why_it_matters": "通过筛选", "action": "添加关键词"},
			},
		},
	}
	suggestions := mapOptimizationSuggestions(parsed)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].Priority != "P0" {
		t.Errorf("suggestion[0] priority: expected P0, got %s", suggestions[0].Priority)
	}
	if suggestions[0].Title != "缺少量化数据" {
		t.Errorf("suggestion[0] title: expected 缺少量化数据, got %s", suggestions[0].Title)
	}
	if suggestions[1].Priority != "P1" {
		t.Errorf("suggestion[1] priority: expected P1, got %s", suggestions[1].Priority)
	}
}

func TestMapInterviewQuestions(t *testing.T) {
	parsed := map[string]any{
		"interview_questions": []any{
			map[string]any{"question": "介绍一下Go项目"},
			map[string]any{"question": "微服务了解多少"},
		},
	}
	questions := mapInterviewQuestions(parsed)
	if len(questions) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(questions))
	}
	if questions[0] != "介绍一下Go项目" {
		t.Errorf("question[0]: expected 介绍一下Go项目, got %s", questions[0])
	}
}

func TestMapNextActions(t *testing.T) {
	parsed := map[string]any{
		"next_actions": []any{
			map[string]any{"action": "优化简历", "priority": "p0"},
			map[string]any{"action": "学习微服务", "priority": "p1"},
		},
	}
	steps := mapNextActions(parsed)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0] != "优化简历" {
		t.Errorf("step[0]: expected 优化简历, got %s", steps[0])
	}
}

func TestMapNextActions_PlainStrings(t *testing.T) {
	parsed := map[string]any{
		"next_actions": []any{
			"优化简历后添加新版本",
			"再次分析匹配度",
		},
	}
	steps := mapNextActions(parsed)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
}

func TestMapNextActions_Empty(t *testing.T) {
	parsed := map[string]any{}
	steps := mapNextActions(parsed)
	if len(steps) != 1 {
		t.Fatalf("expected 1 default step, got %d", len(steps))
	}
	if steps[0] != "优化简历后添加新版本" {
		t.Errorf("default step: expected 优化简历后添加新版本, got %s", steps[0])
	}
}

// ---------------------------------------------------------------------------
// Tests: Type-safe helpers
// ---------------------------------------------------------------------------

func TestGetString(t *testing.T) {
	m := map[string]any{
		"str":   "hello",
		"num":   float64(42),
		"other": true,
	}
	if got := getString(m, "str", "def"); got != "hello" {
		t.Errorf("getString(str) = %q, want %q", got, "hello")
	}
	if got := getString(m, "num", "def"); got != "42" {
		t.Errorf("getString(num) = %q, want %q", got, "42")
	}
	if got := getString(m, "missing", "def"); got != "def" {
		t.Errorf("getString(missing) = %q, want %q", got, "def")
	}
	if got := getString(m, "other", "def"); got != "def" {
		t.Errorf("getString(other) = %q, want %q", got, "def")
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]any{
		"int":   float64(42),
		"str":   "100",
		"bad":   "notanumber",
		"other": true,
	}
	if got := getInt(m, "int", 0); got != 42 {
		t.Errorf("getInt(int) = %d, want 42", got)
	}
	if got := getInt(m, "str", 0); got != 100 {
		t.Errorf("getInt(str) = %d, want 100", got)
	}
	if got := getInt(m, "bad", -1); got != -1 {
		t.Errorf("getInt(bad) = %d, want -1", got)
	}
	if got := getInt(m, "missing", -1); got != -1 {
		t.Errorf("getInt(missing) = %d, want -1", got)
	}
}

func TestGetSlice(t *testing.T) {
	m := map[string]any{
		"slice": []any{1, 2, 3},
		"other": "not a slice",
	}
	slice, ok := getSlice(m, "slice")
	if !ok || len(slice) != 3 {
		t.Errorf("getSlice(slice) = %v, ok=%v, want len=3, ok=true", slice, ok)
	}
	_, ok = getSlice(m, "other")
	if ok {
		t.Error("getSlice(other) should return ok=false")
	}
	_, ok = getSlice(m, "missing")
	if ok {
		t.Error("getSlice(missing) should return ok=false")
	}
}

// ---------------------------------------------------------------------------
// Tests: MatchFallbackError
// ---------------------------------------------------------------------------

func TestMatchFallbackError_Error(t *testing.T) {
	err := &MatchFallbackError{
		FallbackReason: "llm_call_failed",
		ErrorCode:      "LLM_HTTP_ERROR",
		ErrorMessage:   "HTTP 500",
	}
	expected := "match fallback required: reason=llm_call_failed code=LLM_HTTP_ERROR msg=HTTP 500"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

// ---------------------------------------------------------------------------
// Tests: MatchAnalyzer interface compliance
// ---------------------------------------------------------------------------

func TestMatchMockAgent_ImplementsMatchAnalyzer(t *testing.T) {
	var _ MatchAnalyzer = (*MatchMockAgent)(nil)
}

func TestMatchAnalysisAgent_ImplementsMatchAnalyzer(t *testing.T) {
	var _ MatchAnalyzer = (*MatchAnalysisAgent)(nil)
}

// ---------------------------------------------------------------------------
// Tests: buildPromptValues
// ---------------------------------------------------------------------------

func TestBuildPromptValues(t *testing.T) {
	agent := NewMatchAnalysisAgent(&fakeRunner{})
	values, err := agent.buildPromptValues(testResume(), testJob())
	if err != nil {
		t.Fatalf("buildPromptValues() error: %v", err)
	}

	// Check all three keys exist
	for _, key := range []string{"RESUME_CONTEXT_JSON", "JOB_CONTEXT_JSON", "BUSINESS_CONTEXT"} {
		if _, ok := values[key]; !ok {
			t.Errorf("missing prompt value key: %s", key)
		}
	}

	// Verify RESUME_CONTEXT_JSON is valid JSON
	var resumeCtx map[string]any
	if err := json.Unmarshal([]byte(values["RESUME_CONTEXT_JSON"]), &resumeCtx); err != nil {
		t.Errorf("RESUME_CONTEXT_JSON is not valid JSON: %v", err)
	}

	// Verify JOB_CONTEXT_JSON is valid JSON
	var jobCtx map[string]any
	if err := json.Unmarshal([]byte(values["JOB_CONTEXT_JSON"]), &jobCtx); err != nil {
		t.Errorf("JOB_CONTEXT_JSON is not valid JSON: %v", err)
	}

	// Verify BUSINESS_CONTEXT is non-empty
	if values["BUSINESS_CONTEXT"] == "" {
		t.Error("BUSINESS_CONTEXT should not be empty")
	}
}

// ---------------------------------------------------------------------------
// Tests: mapParsedToMatchResult
// ---------------------------------------------------------------------------

func TestMapParsedToMatchResult_FullOutput(t *testing.T) {
	parsed := minimalLLMOutput()
	result, err := mapParsedToMatchResult(parsed, testResume(), testJob())
	if err != nil {
		t.Fatalf("mapParsedToMatchResult() error: %v", err)
	}

	// Check job metadata fields
	if result.JobID != 10 {
		t.Errorf("JobID: expected 10, got %d", result.JobID)
	}
	if result.JobTitle != "Go后端开发工程师" {
		t.Errorf("JobTitle: expected Go后端开发工程师, got %s", result.JobTitle)
	}

	// Check score clamping
	if result.MatchScore != 78 {
		t.Errorf("MatchScore: expected 78, got %d", result.MatchScore)
	}

	// Check LLM-enriched fields
	if result.MatchLevel != "potential_match" {
		t.Errorf("MatchLevel: expected potential_match, got %s", result.MatchLevel)
	}
	if result.OfferGoSummary == "" {
		t.Error("OfferGoSummary should not be empty")
	}
	if len(result.GapPoints) != 1 {
		t.Errorf("GapPoints: expected 1, got %d", len(result.GapPoints))
	}
	if len(result.InterviewQuestions) != 1 {
		t.Errorf("InterviewQuestions: expected 1, got %d", len(result.InterviewQuestions))
	}

	// Check evidence chain mapping
	if len(result.EvidenceChain) != 3 {
		t.Fatalf("EvidenceChain: expected 3, got %d", len(result.EvidenceChain))
	}
	// "matched" → "strong_match"
	if result.EvidenceChain[0].Judgement != "strong_match" {
		t.Errorf("EvidenceChain[0] Judgement: expected strong_match, got %s", result.EvidenceChain[0].Judgement)
	}
	// "partial" → "partial_match"
	if result.EvidenceChain[1].Judgement != "partial_match" {
		t.Errorf("EvidenceChain[1] Judgement: expected partial_match, got %s", result.EvidenceChain[1].Judgement)
	}

	// Check optimization suggestions
	if len(result.OptimizationSuggestions) != 2 {
		t.Fatalf("OptimizationSuggestions: expected 2, got %d", len(result.OptimizationSuggestions))
	}
	if result.OptimizationSuggestions[0].Priority != "P0" {
		t.Errorf("OptimizationSuggestions[0] Priority: expected P0, got %s", result.OptimizationSuggestions[0].Priority)
	}
}

func TestMapParsedToMatchResult_ScoreClamping(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected int
	}{
		{"negative", -10, 0},
		{"zero", 0, 0},
		{"normal", 75, 75},
		{"hundred", 100, 100},
		{"over", 150, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := minimalLLMOutput()
			parsed["match_score"] = tt.score
			result, err := mapParsedToMatchResult(parsed, testResume(), testJob())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.MatchScore != tt.expected {
				t.Errorf("match_score %v: expected %d, got %d", tt.score, tt.expected, result.MatchScore)
			}
		})
	}
}

func TestMapParsedToMatchResult_NilParsed(t *testing.T) {
	_, err := mapParsedToMatchResult(nil, testResume(), testJob())
	if err == nil {
		t.Fatal("expected error for nil parsed")
	}
}

// ---------------------------------------------------------------------------
// Tests: safeParseJSONArray
// ---------------------------------------------------------------------------

func TestSafeParseJSONArray(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int // expected length
	}{
		{"empty", []byte{}, 0},
		{"null", []byte("null"), 0},
		{"invalid", []byte("not json"), 0},
		{"valid", []byte(`["a","b"]`), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeParseJSONArray(tt.input)
			if len(got) != tt.want {
				t.Errorf("safeParseJSONArray(%s) len = %d, want %d", tt.name, len(got), tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: stringsToUpper
// ---------------------------------------------------------------------------

func TestStringsToUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"p0", "P0"},
		{"p1", "P1"},
		{"p2", "P2"},
		{"P0", "P0"}, // already upper
		{"other", "other"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stringsToUpper(tt.input)
		if got != tt.expected {
			t.Errorf("stringsToUpper(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration-style test: MatchService with LLM fallback to Mock
// ---------------------------------------------------------------------------

func TestMatchAnalysisAgent_FallbackFlow(t *testing.T) {
	// Simulate Pipeline returning fallback_required, then using Mock
	fbErr := &MatchFallbackError{
		FallbackReason: "llm_call_failed",
		ErrorCode:      "LLM_HTTP_ERROR",
		ErrorMessage:   "HTTP 500",
	}

	// Verify the error implements error interface
	if fbErr.Error() == "" {
		t.Error("MatchFallbackError.Error() should not be empty")
	}

	// Verify fallback error can be type-asserted
	var err error = fbErr
	if _, ok := err.(*MatchFallbackError); !ok {
		t.Error("should be able to type-assert to *MatchFallbackError")
	}
}

// ---------------------------------------------------------------------------
// Benchmark
// ---------------------------------------------------------------------------

func BenchmarkMapParsedToMatchResult(b *testing.B) {
	parsed := minimalLLMOutput()
	resume := testResume()
	job := testJob()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mapParsedToMatchResult(parsed, resume, job)
	}
}

func BenchmarkBuildPromptValues(b *testing.B) {
	agent := NewMatchAnalysisAgent(&fakeRunner{})
	resume := testResume()
	job := testJob()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.buildPromptValues(resume, job)
	}
}

// ---------------------------------------------------------------------------
// Example
// ---------------------------------------------------------------------------

func ExampleMatchFallbackError_Error() {
	err := &MatchFallbackError{
		FallbackReason: "schema_validation_failed",
		ErrorCode:      "SCHEMA_VALIDATION_FAILED",
		ErrorMessage:   "missing required field: match_score",
	}
	fmt.Println(err.Error())
	// Output: match fallback required: reason=schema_validation_failed code=SCHEMA_VALIDATION_FAILED msg=missing required field: match_score
}
