package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// MatchAnalyzer defines the interface for match analysis generation.
// Both MatchMockAgent and MatchAnalysisAgent implement this interface.
type MatchAnalyzer interface {
	// Analyze generates a match analysis for the given resume and job.
	Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error)
}

// Ensure MatchMockAgent implements MatchAnalyzer.
var _ MatchAnalyzer = (*MatchMockAgent)(nil)

// PipelineRunner defines the interface for the LLM Pipeline Runner.
// This abstraction allows testing with fake runners.
type PipelineRunner interface {
	Run(ctx context.Context, req pipeline.RunRequest) *pipeline.RunResult
}

// MatchAnalysisAgent uses the LLM Pipeline to generate match analysis results.
// On Pipeline success/repaired, it maps the LLM output to model.MatchAnalysisResult.
// On Pipeline fallback_required, it returns an error so the caller can fall back to Mock.
type MatchAnalysisAgent struct {
	runner PipelineRunner
}

// NewMatchAnalysisAgent creates a new MatchAnalysisAgent with the given Pipeline Runner.
func NewMatchAnalysisAgent(runner PipelineRunner) *MatchAnalysisAgent {
	return &MatchAnalysisAgent{runner: runner}
}

// Ensure MatchAnalysisAgent implements MatchAnalyzer.
var _ MatchAnalyzer = (*MatchAnalysisAgent)(nil)

// matchAnalysisBusinessContext is the business context injected into the Prompt.
const matchAnalysisBusinessContext = `你正在进行学生求职场景下的岗位匹配分析。
请结合 HR 初筛视角和技术面试官视角。
必须基于输入简历和岗位信息判断。
不要编造经历。
可以给出优化方向和可替换表达，但必须提醒用户基于真实经历修改。
输出必须是严格 JSON。`

// Analyze calls the LLM Pipeline and maps the result to MatchAnalysisResult.
// If the Pipeline returns fallback_required or the mapping fails, it returns an error.
func (a *MatchAnalysisAgent) Analyze(ctx context.Context, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	// Step 1: Assemble PromptValues
	promptValues, err := a.buildPromptValues(resume, job)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt values: %w", err)
	}

	// Step 2: Call Pipeline
	runReq := pipeline.RunRequest{
		AgentName:        "MatchAnalysisAgent",
		PromptName:       prompt.PromptMatchAnalysis,
		SchemaName:       schema.SchemaMatchAnalysis,
		TargetOutputType: pipeline.TargetMatchAnalysis,
		PromptValues:     promptValues,
		Metadata: map[string]string{
			"job_id":    strconv.FormatUint(job.ID, 10),
			"resume_id": strconv.FormatUint(resume.ID, 10),
		},
	}

	runResult := a.runner.Run(ctx, runReq)

	// Step 3: Log pipeline result (safe, no API key or full content)
	log.Printf("[MatchAnalysisAgent] pipeline result: status=%s generated_by=%s repair_applied=%v fallback=%v reason=%s error_code=%s duration_ms=%d raw_output_length=%d",
		runResult.AgentStatus, runResult.GeneratedBy, runResult.RepairApplied,
		runResult.FallbackRequired, runResult.FallbackReason, runResult.ErrorCode,
		runResult.DurationMs, runResult.RawOutputLength)

	// Step 4: Check if fallback is required
	if runResult.FallbackRequired {
		return nil, &MatchFallbackError{
			FallbackReason: runResult.FallbackReason,
			ErrorCode:      runResult.ErrorCode,
			ErrorMessage:   runResult.ErrorMessage,
		}
	}

	// Step 5: Map Parsed to MatchAnalysisResult
	result, err := mapParsedToMatchResult(runResult.Parsed, resume, job)
	if err != nil {
		// Mapping failure should also trigger fallback
		return nil, &MatchFallbackError{
			FallbackReason: "field_mapping_failed",
			ErrorCode:      "FIELD_MAPPING_FAILED",
			ErrorMessage:   fmt.Sprintf("failed to map LLM output: %s", err),
		}
	}

	// Step 6: Set LLM-specific metadata on the result
	result.GeneratedBy = runResult.GeneratedBy
	result.AgentStatus = string(runResult.AgentStatus)
	result.RepairApplied = runResult.RepairApplied
	result.FallbackReason = ""
	result.ErrorCode = ""

	return result, nil
}

// MatchFallbackError indicates the Pipeline returned fallback_required or mapping failed.
// The caller should fall back to the Mock Agent.
type MatchFallbackError struct {
	FallbackReason string
	ErrorCode      string
	ErrorMessage   string
}

func (e *MatchFallbackError) Error() string {
	return fmt.Sprintf("match fallback required: reason=%s code=%s msg=%s",
		e.FallbackReason, e.ErrorCode, e.ErrorMessage)
}

// buildPromptValues assembles the three placeholders for match_analysis_v1.md.
func (a *MatchAnalysisAgent) buildPromptValues(resume *model.ResumeVersion, job *model.Job) (map[string]string, error) {
	// RESUME_CONTEXT_JSON: serialize resume (safe subset, no API key)
	resumeJSON, err := json.Marshal(map[string]any{
		"version_no":       resume.VersionNo,
		"target_role":      resume.TargetRole,
		"target_city":      resume.TargetCity,
		"skills":           safeParseJSONArray(resume.SkillsJSON),
		"education":        safeParseJSONArray(resume.EducationJSON),
		"experience":       safeParseJSONArray(resume.ExperienceJSON),
		"project":          safeParseJSONArray(resume.ProjectJSON),
		"self_evaluation":  resume.SelfEvaluation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resume context: %w", err)
	}

	// JOB_CONTEXT_JSON: serialize job
	jobJSON, err := json.Marshal(map[string]any{
		"id":            job.ID,
		"title":         job.Title,
		"department":    job.Department,
		"city":          job.City,
		"category":      job.Category,
		"description":   job.Description,
		"requirements":  job.Requirements,
		"data_quality":  job.DataQuality,
		"skill_keywords": safeParseJSONArray(job.SkillKeywordsJSON),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job context: %w", err)
	}

	return map[string]string{
		"RESUME_CONTEXT_JSON": string(resumeJSON),
		"JOB_CONTEXT_JSON":    string(jobJSON),
		"BUSINESS_CONTEXT":    matchAnalysisBusinessContext,
	}, nil
}

// safeParseJSONArray parses a JSON byte array, returning empty slice on error.
func safeParseJSONArray(data []byte) []any {
	if len(data) == 0 {
		return []any{}
	}
	var result []any
	if json.Unmarshal(data, &result) != nil {
		return []any{}
	}
	return result
}

// ---------------------------------------------------------------------------
// Field mapping: RunResult.Parsed (map[string]any) → model.MatchAnalysisResult
// ---------------------------------------------------------------------------

// mapParsedToMatchResult maps the Pipeline's parsed JSON output to a MatchAnalysisResult.
// This function handles the structural differences between the LLM Schema output
// and the existing model.MatchAnalysisResult.
func mapParsedToMatchResult(parsed map[string]any, resume *model.ResumeVersion, job *model.Job) (*model.MatchAnalysisResult, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed output is nil")
	}

	result := &model.MatchAnalysisResult{
		JobID:           job.ID,
		JobTitle:        job.Title,
		ResumeVersionID: resume.ID,
		ResumeVersionNo: resume.VersionNo,
		ExternalSource:  job.ExternalSource,
		ExternalID:      job.ExternalID,
		DataQuality:     job.DataQuality,
		SourceURL:       job.SourceURL,
		IsFeatured:      job.IsFeatured,
	}

	// match_score (required in Schema)
	result.MatchScore = getInt(parsed, "match_score", 0)
	if result.MatchScore < 0 {
		result.MatchScore = 0
	}
	if result.MatchScore > 100 {
		result.MatchScore = 100
	}

	// match_level (LLM-only field, safe to add)
	result.MatchLevel = getString(parsed, "match_level", "")

	// match_label (LLM-only field, safe to add)
	result.MatchLabel = getString(parsed, "match_label", "")

	// screening_result: map LLM values to existing enum
	screeningResult := getString(parsed, "screening_result", "possible")
	result.ScreeningResult = mapScreeningResult(screeningResult)

	// offer_go_summary (LLM-only field)
	result.OfferGoSummary = getString(parsed, "offer_go_summary", "")

	// evidence_chain: map Schema structure to model.MatchEvidence
	result.EvidenceChain = mapEvidenceChain(parsed)

	// matched_points → match_highlights
	result.MatchHighlights = mapMatchedPoints(parsed)

	// gap_points (LLM-only field)
	result.GapPoints = mapGapPoints(parsed)

	// risk_points: extract risk strings from Schema structure
	result.RiskPoints = mapRiskPoints(parsed)

	// resume_optimization → optimization_suggestions
	result.OptimizationSuggestions = mapOptimizationSuggestions(parsed)

	// interview_questions (LLM-only field)
	result.InterviewQuestions = mapInterviewQuestions(parsed)

	// next_actions → next_steps
	result.NextSteps = mapNextActions(parsed)

	// authenticity_notice
	result.AuthenticityNotice = getString(parsed, "authenticity_notice",
		"以下建议仅用于表达优化，请只补充真实经历，不要虚构项目或能力。")

	// apply_readiness (4-3.7.C.1)
	result.ApplyReadiness = getString(parsed, "apply_readiness", "insufficient_information")

	// optimization_stop_signal (4-3.7.C.1)
	result.OptimizationStopSignal = getString(parsed, "optimization_stop_signal", "needs_more_evidence")

	// top_fix_items (4-3.7.C.1)
	result.TopFixItems = mapTopFixItems(parsed)

	return result, nil
}

// mapScreeningResult maps LLM screening_result values to existing enum values.
// LLM Schema uses: likely / uncertain / risky
// Mock uses: highly_likely / likely / possible / risky
func mapScreeningResult(val string) string {
	switch val {
	case "highly_likely", "likely":
		return "likely"
	case "potential_match", "uncertain":
		return "possible"
	case "explore_first":
		return "possible"
	case "not_recommended", "risky":
		return "risky"
	default:
		// If it already matches existing enum, keep it
		for _, v := range []string{"highly_likely", "likely", "possible", "risky"} {
			if val == v {
				return v
			}
		}
		return "possible"
	}
}

// mapEvidenceChain maps Schema evidence_chain items to model.MatchEvidence.
// Schema: {jd_requirement, resume_evidence, judgement, explanation}
// Model: {requirement, resume_point, judgement}
func mapEvidenceChain(parsed map[string]any) []model.MatchEvidence {
	items, ok := getSlice(parsed, "evidence_chain")
	if !ok || len(items) == 0 {
		return []model.MatchEvidence{{Requirement: "无证据链数据", ResumePoint: "-", Judgement: "missing"}}
	}

	var chain []model.MatchEvidence
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		judgement := getString(m, "judgement", "missing")
		// Map Schema judgement values to existing enum
		mappedJudgement := mapJudgement(judgement)
		chain = append(chain, model.MatchEvidence{
			Requirement: getString(m, "jd_requirement", ""),
			ResumePoint: getString(m, "resume_evidence", ""),
			Judgement:   mappedJudgement,
		})
	}
	return chain
}

// mapJudgement maps LLM Schema judgement values to existing enum.
// Schema: matched / partial / missing
// Model: strong_match / partial_match / missing
func mapJudgement(val string) string {
	switch val {
	case "matched", "strong_match":
		return "strong_match"
	case "partial", "partial_match":
		return "partial_match"
	case "missing":
		return "missing"
	default:
		return "missing"
	}
}

// mapMatchedPoints extracts highlight strings from matched_points array.
// Schema: [{point, why_it_matters, evidence}]
func mapMatchedPoints(parsed map[string]any) []string {
	items, ok := getSlice(parsed, "matched_points")
	if !ok || len(items) == 0 {
		return []string{"岗位方向与求职目标部分相关"}
	}

	var highlights []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if point := getString(m, "point", ""); point != "" {
			highlights = append(highlights, point)
		}
	}
	if len(highlights) == 0 {
		return []string{"岗位方向与求职目标部分相关"}
	}
	return highlights
}

// mapGapPoints extracts gap strings from gap_points array.
// Schema: [{gap, impact, how_to_improve}]
func mapGapPoints(parsed map[string]any) []string {
	items, ok := getSlice(parsed, "gap_points")
	if !ok || len(items) == 0 {
		return []string{}
	}

	var gaps []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if gap := getString(m, "gap", ""); gap != "" {
			gaps = append(gaps, gap)
		}
	}
	return gaps
}

// mapRiskPoints extracts risk strings from risk_points array.
// Schema: [{risk, severity, reason, mitigation}]
func mapRiskPoints(parsed map[string]any) []string {
	items, ok := getSlice(parsed, "risk_points")
	if !ok || len(items) == 0 {
		return []string{}
	}

	var risks []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			// Could be plain string
			if s, ok := item.(string); ok {
				risks = append(risks, s)
			}
			continue
		}
		if risk := getString(m, "risk", ""); risk != "" {
			risks = append(risks, risk)
		}
	}
	return risks
}

// mapOptimizationSuggestions maps resume_optimization from Schema to model.MatchSuggestion.
// Schema: {p0: [{problem, why_it_matters, rewrite_example, action}], p1: [...], p2: [...]}
// Model: [{priority, title, reason, action}]
func mapOptimizationSuggestions(parsed map[string]any) []model.MatchSuggestion {
	opt, ok := parsed["resume_optimization"].(map[string]any)
	if !ok {
		return []model.MatchSuggestion{}
	}

	var suggestions []model.MatchSuggestion

	for _, priority := range []string{"p0", "p1", "p2"} {
		items, ok := getSlice(opt, priority)
		if !ok {
			continue
		}
		for _, item := range items {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			suggestions = append(suggestions, model.MatchSuggestion{
				Priority: stringsToUpper(priority),
				Title:    getString(m, "problem", ""),
				Reason:   getString(m, "why_it_matters", ""),
				Action:   getString(m, "action", ""),
			})
		}
	}

	return suggestions
}

// stringsToUpper converts "p0" → "P0", "p1" → "P1", etc.
func stringsToUpper(s string) string {
	if len(s) == 2 && s[0] == 'p' && s[1] >= '0' && s[1] <= '9' {
		return "P" + string(s[1])
	}
	return s
}

// mapInterviewQuestions extracts question strings from interview_questions array.
// Schema: [{question, why_ask, how_to_prepare, related_jd_requirement}]
func mapInterviewQuestions(parsed map[string]any) []string {
	items, ok := getSlice(parsed, "interview_questions")
	if !ok || len(items) == 0 {
		return []string{}
	}

	var questions []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if q := getString(m, "question", ""); q != "" {
			questions = append(questions, q)
		}
	}
	return questions
}

// mapNextActions extracts action strings from next_actions array.
// Schema: [{action, priority, expected_effect}]
func mapNextActions(parsed map[string]any) []string {
	items, ok := getSlice(parsed, "next_actions")
	if !ok || len(items) == 0 {
		return []string{"优化简历后添加新版本"}
	}

	var steps []string
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			// Could be plain string (from Mock)
			if s, ok := item.(string); ok {
				steps = append(steps, s)
			}
			continue
		}
		if action := getString(m, "action", ""); action != "" {
			steps = append(steps, action)
		}
	}
	if len(steps) == 0 {
		return []string{"优化简历后添加新版本"}
	}
	return steps
}

// mapTopFixItems maps top_fix_items from Schema to []model.TopFixItem (4-3.7.C.1).
// Schema: [{issue, why_it_matters, action, example_reference, self_check_question, estimated_effort}]
func mapTopFixItems(parsed map[string]any) []model.TopFixItem {
	items, ok := getSlice(parsed, "top_fix_items")
	if !ok || len(items) == 0 {
		return []model.TopFixItem{}
	}

	var fixItems []model.TopFixItem
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fixItems = append(fixItems, model.TopFixItem{
			Issue:            getString(m, "issue", ""),
			WhyItMatters:     getString(m, "why_it_matters", ""),
			Action:           getString(m, "action", ""),
			ExampleReference: getString(m, "example_reference", ""),
			SelfCheckQuestion: getString(m, "self_check_question", ""),
			EstimatedEffort:  getString(m, "estimated_effort", ""),
		})
	}
	return fixItems
}

// ---------------------------------------------------------------------------
// Type-safe helpers for map[string]any access
// ---------------------------------------------------------------------------

func getString(m map[string]any, key, defaultVal string) string {
	val, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return defaultVal
	}
}

func getInt(m map[string]any, key string, defaultVal int) int {
	val, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		return defaultVal
	default:
		return defaultVal
	}
}

func getSlice(m map[string]any, key string) ([]any, bool) {
	val, ok := m[key]
	if !ok {
		return nil, false
	}
	slice, ok := val.([]any)
	return slice, ok
}
