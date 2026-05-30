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

// Recommender defines the interface for job recommendation generation.
// Both RecommendationMockAgent and RecommendationAgent implement this interface.
type Recommender interface {
	// Recommend generates job recommendations for the given resume and candidate jobs.
	Recommend(ctx context.Context, resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error)
}

// Ensure RecommendationMockAgent implements Recommender.
var _ Recommender = (*RecommendationMockAgent)(nil)

// RecommendationAgent uses the LLM Pipeline to generate job recommendations.
// On Pipeline success/repaired, it maps the LLM output to []model.RecommendedJob.
// On Pipeline fallback_required, it returns an error so the caller can fall back to Mock.
type RecommendationAgent struct {
	runner PipelineRunner
}

// NewRecommendationAgent creates a new RecommendationAgent with the given Pipeline Runner.
func NewRecommendationAgent(runner PipelineRunner) *RecommendationAgent {
	return &RecommendationAgent{runner: runner}
}

// Ensure RecommendationAgent implements Recommender.
var _ Recommender = (*RecommendationAgent)(nil)

// recommendationBusinessContext is the business context injected into the Prompt.
const recommendationBusinessContext = `你正在进行学生求职场景下的岗位推荐。
只能基于输入的 JOB_CANDIDATES_JSON 推荐岗位。
不要编造岗位。
不要编造 job_id、external_id、title、department、location。
推荐理由必须能从候选岗位和简历信息中找到依据。
推荐结果可分为 priority_apply / recommended / explore_first / not_recommended。
输出必须是严格 JSON。`

// Recommend calls the LLM Pipeline and maps the result to []model.RecommendedJob.
// If the Pipeline returns fallback_required or the mapping fails, it returns an error.
func (a *RecommendationAgent) Recommend(ctx context.Context, resume *model.ResumeVersion, jobs []*model.Job) ([]model.RecommendedJob, error) {
	// Step 1: Assemble PromptValues
	promptValues, err := a.buildPromptValues(resume, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt values: %w", err)
	}

	// Step 2: Call Pipeline
	runReq := pipeline.RunRequest{
		AgentName:        "RecommendationAgent",
		PromptName:       prompt.PromptRecommendation,
		SchemaName:       schema.SchemaRecommendation,
		TargetOutputType: pipeline.TargetRecommendation,
		PromptValues:     promptValues,
		Metadata: map[string]string{
			"resume_id":      strconv.FormatUint(resume.ID, 10),
			"candidate_count": strconv.Itoa(len(jobs)),
		},
	}

	runResult := a.runner.Run(ctx, runReq)

	// Step 3: Log pipeline result (safe, no API key or full content)
	log.Printf("[RecommendationAgent] pipeline result: status=%s generated_by=%s repair_applied=%v fallback=%v reason=%s error_code=%s duration_ms=%d raw_output_length=%d",
		runResult.AgentStatus, runResult.GeneratedBy, runResult.RepairApplied,
		runResult.FallbackRequired, runResult.FallbackReason, runResult.ErrorCode,
		runResult.DurationMs, runResult.RawOutputLength)

	// Step 4: Check if fallback is required
	if runResult.FallbackRequired {
		return nil, &RecFallbackError{
			FallbackReason: runResult.FallbackReason,
			ErrorCode:      runResult.ErrorCode,
			ErrorMessage:   runResult.ErrorMessage,
		}
	}

	// Step 5: Build candidate job ID set for validation
	candidateIDs := make(map[uint64]*model.Job, len(jobs))
	for _, j := range jobs {
		candidateIDs[j.ID] = j
	}

	// Step 6: Map Parsed to []model.RecommendedJob
	recommendedJobs, err := mapParsedToRecommendedJobs(runResult.Parsed, candidateIDs, resume)
	if err != nil {
		// Mapping failure should also trigger fallback
		return nil, &RecFallbackError{
			FallbackReason: "field_mapping_failed",
			ErrorCode:      "FIELD_MAPPING_FAILED",
			ErrorMessage:   fmt.Sprintf("failed to map LLM output: %s", err),
		}
	}

	// Step 7: Set LLM-specific metadata on each recommended job
	for i := range recommendedJobs {
		recommendedJobs[i].GeneratedBy = runResult.GeneratedBy
		recommendedJobs[i].AgentStatus = string(runResult.AgentStatus)
		recommendedJobs[i].RepairApplied = runResult.RepairApplied
		recommendedJobs[i].FallbackReason = ""
		recommendedJobs[i].ErrorCode = ""
	}

	return recommendedJobs, nil
}

// RecFallbackError indicates the Pipeline returned fallback_required or mapping failed.
// The caller should fall back to the Mock Agent.
type RecFallbackError struct {
	FallbackReason string
	ErrorCode      string
	ErrorMessage   string
}

func (e *RecFallbackError) Error() string {
	return fmt.Sprintf("recommendation fallback required: reason=%s code=%s msg=%s",
		e.FallbackReason, e.ErrorCode, e.ErrorMessage)
}

// buildPromptValues assembles the three placeholders for recommendation_v1.md.
func (a *RecommendationAgent) buildPromptValues(resume *model.ResumeVersion, jobs []*model.Job) (map[string]string, error) {
	// RESUME_CONTEXT_JSON: serialize resume (safe subset, no API key)
	resumeJSON, err := json.Marshal(map[string]any{
		"version_no":      resume.VersionNo,
		"target_role":     resume.TargetRole,
		"target_city":     resume.TargetCity,
		"skills":          safeParseJSONArray(resume.SkillsJSON),
		"education":       safeParseJSONArray(resume.EducationJSON),
		"experience":      safeParseJSONArray(resume.ExperienceJSON),
		"project":         safeParseJSONArray(resume.ProjectJSON),
		"self_evaluation": resume.SelfEvaluation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resume context: %w", err)
	}

	// JOB_CANDIDATES_JSON: serialize candidate jobs
	// Only include fields the LLM needs; avoid including raw_json or other internal data.
	type candidateJob struct {
		JobID        uint64   `json:"job_id"`
		Title        string   `json:"title"`
		Department   string   `json:"department"`
		City         string   `json:"city"`
		Category     string   `json:"category"`
		JobType      string   `json:"job_type"`
		Description  string   `json:"description,omitempty"`
		Requirements string   `json:"requirements,omitempty"`
		SkillKeywords []string `json:"skill_keywords,omitempty"`
		ExternalID   string   `json:"external_id,omitempty"`
		DataQuality  string   `json:"data_quality"`
	}

	candidates := make([]candidateJob, 0, len(jobs))
	for _, j := range jobs {
		cj := candidateJob{
			JobID:        j.ID,
			Title:        j.Title,
			Department:   j.Department,
			City:         j.City,
			Category:     j.Category,
			JobType:      j.JobType,
			Description:  j.Description,
			Requirements: j.Requirements,
			DataQuality:  j.DataQuality,
		}
		if j.ExternalID != nil {
			cj.ExternalID = *j.ExternalID
		}
		cj.SkillKeywords = parseJSONStringArray(j.SkillKeywordsJSON)
		candidates = append(candidates, cj)
	}

	candidatesJSON, err := json.Marshal(candidates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job candidates: %w", err)
	}

	return map[string]string{
		"RESUME_CONTEXT_JSON":  string(resumeJSON),
		"JOB_CANDIDATES_JSON":  string(candidatesJSON),
		"BUSINESS_CONTEXT":     recommendationBusinessContext,
	}, nil
}

// ---------------------------------------------------------------------------
// Field mapping: RunResult.Parsed (map[string]any) → []model.RecommendedJob
// ---------------------------------------------------------------------------

// mapParsedToRecommendedJobs maps the Pipeline's parsed JSON output to []model.RecommendedJob.
// It validates that each recommended job exists in the candidate set.
func mapParsedToRecommendedJobs(parsed map[string]any, candidateIDs map[uint64]*model.Job, resume *model.ResumeVersion) ([]model.RecommendedJob, error) {
	if parsed == nil {
		return nil, fmt.Errorf("parsed output is nil")
	}

	// Extract recommendations array
	items, ok := getSlice(parsed, "recommendations")
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("recommendations array is empty or missing")
	}

	var results []model.RecommendedJob

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Validate job_id against candidates
		jobID := getUint64(m, "job_id")
		if jobID == 0 {
			// Missing or invalid job_id, skip this recommendation
			continue
		}

		candidate, exists := candidateIDs[jobID]
		if !exists {
			// LLM returned a job_id not in the candidate set — skip (don't fabricate)
			continue
		}

		rec := model.RecommendedJob{
			JobID:       candidate.ID,
			Title:       candidate.Title,
			Department:  candidate.Department,
			City:        candidate.City,
			Category:    candidate.Category,
			JobType:     candidate.JobType,
			Priority:    mapRecommendLevel(getString(m, "recommend_level", "explore_first")),
			RecommendationScore: getInt(m, "match_score", 0),
			RecommendationReason: getString(m, "recommend_reason", ""),
			ExternalSource: candidate.ExternalSource,
			ExternalID:   candidate.ExternalID,
			DataQuality:  candidate.DataQuality,
			SourceURL:    candidate.SourceURL,
			IsFeatured:   candidate.IsFeatured,
		}

		// match_highlights from evidence
		rec.MatchHighlights = mapRecEvidence(m)

		// risk_points from risk_points
		rec.RiskPoints = mapRecRiskPoints(m)

		// suggested_action from next_action
		rec.SuggestedAction = mapRecNextAction(m)

		// LLM-enriched fields
		rec.RecommendLabel = getString(m, "recommend_label", "")
		rec.OfferGoSummary = getString(m, "offer_go_summary", "")
		rec.QuerySuggestions = mapRecQuerySuggestions(m)

		// Strategic fields (4-3.7.C.1)
		rec.ApplyReadiness = getString(m, "apply_readiness", "insufficient_information")
		rec.OptimizationStopSignal = getString(m, "optimization_stop_signal", "needs_more_evidence")
		rec.TopFixItems = mapRecTopFixItems(m)

		// Agent metadata
		rec.AgentStatus = "success" // will be overridden by caller for repaired

		results = append(results, rec)
	}

	// If all recommendations were filtered out (none matched candidates), trigger fallback
	if len(results) == 0 {
		return nil, fmt.Errorf("all LLM recommendations were filtered: no valid job_id in candidate set")
	}

	return results, nil
}

// mapRecommendLevel maps LLM recommend_level values to existing priority enum.
// LLM Schema: priority_apply / recommended / explore_first / not_recommended
// Mock uses: strong_recommend / recommend / try
func mapRecommendLevel(val string) string {
	switch val {
	case "priority_apply":
		return "strong_recommend"
	case "recommended":
		return "recommend"
	case "explore_first":
		return "try"
	case "not_recommended":
		// We still include not_recommended jobs but with a neutral priority
		return "try"
	default:
		// If it already matches existing enum, keep it
		for _, v := range []string{"strong_recommend", "recommend", "try"} {
			if val == v {
				return v
			}
		}
		return "try"
	}
}

// mapRecEvidence extracts highlight strings from evidence array.
// Schema: [{job_requirement, resume_evidence, judgement, explanation}]
func mapRecEvidence(m map[string]any) []string {
	items, ok := getSlice(m, "evidence")
	if !ok || len(items) == 0 {
		return []string{"岗位方向与求职目标部分相关"}
	}

	var highlights []string
	for _, item := range items {
		em, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if explanation := getString(em, "explanation", ""); explanation != "" {
			highlights = append(highlights, explanation)
		} else if evidence := getString(em, "resume_evidence", ""); evidence != "" {
			highlights = append(highlights, evidence)
		}
	}
	if len(highlights) == 0 {
		return []string{"岗位方向与求职目标部分相关"}
	}
	return highlights
}

// mapRecRiskPoints extracts risk strings from risk_points array.
// Schema: [{risk, severity, mitigation}]
func mapRecRiskPoints(m map[string]any) []string {
	items, ok := getSlice(m, "risk_points")
	if !ok || len(items) == 0 {
		return []string{}
	}

	var risks []string
	for _, item := range items {
		rm, ok := item.(map[string]any)
		if !ok {
			if s, ok := item.(string); ok {
				risks = append(risks, s)
			}
			continue
		}
		if risk := getString(rm, "risk", ""); risk != "" {
			risks = append(risks, risk)
		}
	}
	return risks
}

// mapRecNextAction extracts action string from next_action.
// Schema: {action, priority, expected_effect}
func mapRecNextAction(m map[string]any) string {
	na, ok := m["next_action"].(map[string]any)
	if !ok {
		return "建议关注该岗位详情。"
	}
	if action := getString(na, "action", ""); action != "" {
		return action
	}
	return "建议关注该岗位详情。"
}

// mapRecQuerySuggestions extracts query strings from query_suggestions array.
// Schema: [{query, why_search}]
func mapRecQuerySuggestions(m map[string]any) []string {
	items, ok := getSlice(m, "query_suggestions")
	if !ok || len(items) == 0 {
		return []string{}
	}

	var queries []string
	for _, item := range items {
		qm, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if q := getString(qm, "query", ""); q != "" {
			queries = append(queries, q)
		}
	}
	return queries
}

// mapRecTopFixItems maps top_fix_items from Schema to []model.TopFixItem (4-3.7.C.1).
// Schema: [{issue, why_it_matters, action, example_reference, self_check_question, estimated_effort}]
func mapRecTopFixItems(m map[string]any) []model.TopFixItem {
	items, ok := getSlice(m, "top_fix_items")
	if !ok || len(items) == 0 {
		return []model.TopFixItem{}
	}

	var fixItems []model.TopFixItem
	for _, item := range items {
		fm, ok := item.(map[string]any)
		if !ok {
			continue
		}
		fixItems = append(fixItems, model.TopFixItem{
			Issue:            getString(fm, "issue", ""),
			WhyItMatters:     getString(fm, "why_it_matters", ""),
			Action:           getString(fm, "action", ""),
			ExampleReference: getString(fm, "example_reference", ""),
			SelfCheckQuestion: getString(fm, "self_check_question", ""),
			EstimatedEffort:  getString(fm, "estimated_effort", ""),
		})
	}
	return fixItems
}

// getUint64 safely extracts a uint64 value from map[string]any.
func getUint64(m map[string]any, key string) uint64 {
	val, ok := m[key]
	if !ok {
		return 0
	}
	switch v := val.(type) {
	case float64:
		if v >= 0 {
			return uint64(v)
		}
		return 0
	case int:
		if v >= 0 {
			return uint64(v)
		}
		return 0
	case string:
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
		return 0
	default:
		return 0
	}
}
