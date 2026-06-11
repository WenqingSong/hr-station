// ─── K13-Fix-C2: Deep Match — 3-Agent Architecture ─────────────────────
// Agent 1: MatchContext (rule-based, no LLM)
// Agent 2: MatchDecision (LLM quality)
// Agent 3: MatchVerifier (LLM quality)

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// ─── Agent 1: MatchContext ───────────────────────────────────────────────

type MatchContextInput struct {
	ResumeID      uint64
	JobID         uint64
	ResumeProfile *model.ResumeTagProfile
	JobProfile    *model.JobTagProfile
	ResumeName    string
	JobTitle      string
}

type MatchContext struct {
	AgentName          string                `json:"agent_name"`
	ResumeID           uint64                `json:"resume_id"`
	JobID              uint64                `json:"job_id"`
	ResumeProfileReady bool                  `json:"resume_profile_ready"`
	JobProfileReady    bool                  `json:"job_profile_ready"`
	BaselineScore      int                   `json:"baseline_score"`
	MatchedTags        []string              `json:"matched_tags"`
	MissingTags        []string              `json:"missing_tags"`
	StrongEvidence     []MatchEvidenceSnippet `json:"strong_evidence"`
	WeakEvidence       []MatchEvidenceSnippet `json:"weak_evidence"`
	RiskEvidence       []MatchEvidenceSnippet `json:"risk_evidence"`
	ResumeSummary      *ResumeContextSummary `json:"resume_summary"`
	JobSummary         string                `json:"job_summary"`
	JobRequirements    []string              `json:"job_requirements"`
	ProfileCheck       *ProfileRealityCheck  `json:"profile_check"`
	GeneratedAt        string                `json:"generated_at"`
}

type MatchEvidenceSnippet struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Text     string `json:"text"`
	Strength string `json:"strength"` // high/medium/low
}

func BuildMatchContext(_ context.Context, input *MatchContextInput) (*MatchContext, error) {
	if input.ResumeProfile == nil {
		return nil, fmt.Errorf("match context: resume profile is nil")
	}
	if input.JobProfile == nil {
		return nil, fmt.Errorf("match context: job profile is nil")
	}

	matched, missing, tagScore := computeMatchTags(input.ResumeProfile, input.JobProfile)
	baselineScore := minInt(95, 50+int(tagScore*0.6))
	if baselineScore < 30 {
		baselineScore = 30
	}

	// Build evidence snippets
	strong := buildMatchEvidenceList(input.ResumeProfile, matched, "high")
	weak := buildMatchEvidenceList(input.ResumeProfile, missing, "low")
	risk := buildMatchRiskEvidence(input.JobProfile, missing)

	resumeSummary := &ResumeContextSummary{
		DominantProfile: input.ResumeProfile.DominantProfile,
		CoreTags:        collectTagNames(input.ResumeProfile.PrimaryTags, 6),
		EvidenceRefs:    collectEvidenceRefs(input.ResumeProfile, matched),
		ProjectCount:    input.ResumeProfile.Evidence.ProjectCount,
		InternshipCount: input.ResumeProfile.Evidence.InternshipCount,
		SkillCount:      input.ResumeProfile.Evidence.SkillCount,
	}

	jobReqs := collectTagNames(input.JobProfile.RequirementTags, 6)
	jobSummary := fmt.Sprintf("%s - %s %s", input.JobTitle, input.JobProfile.Direction, input.JobProfile.Category)

	reqTags := make([]string, 0)
	for _, t := range input.JobProfile.RequirementTags {
		reqTags = append(reqTags, t.Tag)
	}

	ctxOut := &MatchContext{
		AgentName:          "match_context_agent",
		ResumeID:           input.ResumeID,
		JobID:              input.JobID,
		ResumeProfileReady: true,
		JobProfileReady:    true,
		BaselineScore:      baselineScore,
		MatchedTags:        matched,
		MissingTags:        missing,
		StrongEvidence:     strong,
		WeakEvidence:       weak,
		RiskEvidence:       risk,
		ResumeSummary:      resumeSummary,
		JobSummary:         jobSummary,
		JobRequirements:    jobReqs,
		ProfileCheck: &ProfileRealityCheck{
			ProjectCountOK:    input.ResumeProfile.Evidence.ProjectCount > 0,
			InternshipCountOK: input.ResumeProfile.Evidence.InternshipCount > 0,
			ProfileConsistent: true,
			Note:              fmt.Sprintf("p=%d i=%d", input.ResumeProfile.Evidence.ProjectCount, input.ResumeProfile.Evidence.InternshipCount),
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
	return ctxOut, nil
}

// ─── Agent 2: MatchDecision ──────────────────────────────────────────────

type MatchDecision struct {
	Score               int                     `json:"score"`
	Level               string                  `json:"level"`
	Summary             string                  `json:"summary"`
	MatchedPoints       []MatchDecisionPoint    `json:"matched_points"`
	GapPoints           []MatchDecisionPoint    `json:"gap_points"`
	RiskPoints          []MatchDecisionRisk     `json:"risk_points"`
	ImprovementSuggestions []MatchImprovement  `json:"improvement_suggestions"`
	Confidence          string                  `json:"confidence"`
}

type MatchDecisionPoint struct {
	Title    string `json:"title"`
	Evidence string `json:"evidence"`
	Reason   string `json:"reason,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

type MatchDecisionRisk struct {
	Title      string `json:"title"`
	Evidence   string `json:"evidence"`
	Risk       string `json:"risk"`
	Mitigation string `json:"mitigation"`
}

type MatchImprovement struct {
	Title    string `json:"title"`
	Action   string `json:"action"`
	Priority string `json:"priority"`
}

// ─── Agent 3: MatchVerifier ──────────────────────────────────────────────

type MatchFinalOutput struct {
	GeneratedBy            string                 `json:"generated_by"`
	AgentStatus            string                 `json:"agent_status"`
	Fallback               bool                   `json:"fallback"`
	FallbackReason         string                 `json:"fallback_reason"`
	RepairApplied          bool                   `json:"repair_applied"`
	Score                  int                    `json:"score"`
	Level                  string                 `json:"level"`
	Summary                string                 `json:"summary"`
	MatchedPoints          []MatchDecisionPoint   `json:"matched_points"`
	GapPoints              []MatchDecisionPoint   `json:"gap_points"`
	RiskPoints             []MatchDecisionRisk    `json:"risk_points"`
	ImprovementSuggestions []MatchImprovement     `json:"improvement_suggestions"`
	Confidence             string                 `json:"confidence"`
	AgentTrace             *MatchAgentTrace       `json:"agent_trace,omitempty"`
}

type MatchAgentTrace struct {
	Agent1Type      string `json:"agent1_type"`
	Agent1DurationMs int64 `json:"agent1_duration_ms"`
	Agent2Type      string `json:"agent2_type"`
	Agent2DurationMs int64 `json:"agent2_duration_ms"`
	Agent3Type      string `json:"agent3_type"`
	Agent3DurationMs int64 `json:"agent3_duration_ms"`
}

// ─── PipelineRunner interface (reuses agent.PipelineRunner) ────────────

// ─── Deep Match 3-Agent Orchestrator ─────────────────────────────────────

func RunDeepMatch3Agents(
	ctx context.Context,
	resumeProfile *model.ResumeTagProfile,
	jobProfile *model.JobTagProfile,
	resumeName string,
	jobTitle string,
	runner2 PipelineRunner,
	runner3 PipelineRunner,
) (*MatchFinalOutput, MatchAgentTrace, error) {
	trace := MatchAgentTrace{}

	// ── Agent 1: Rule-based context building ──────────────────────────
	t0 := time.Now()
	a1Input := &MatchContextInput{
		ResumeID:      resumeProfile.ResumeID,
		JobID:         jobProfile.JobID,
		ResumeProfile: resumeProfile,
		JobProfile:    jobProfile,
		ResumeName:    resumeName,
		JobTitle:      jobTitle,
	}
	a1Output, err := BuildMatchContext(ctx, a1Input)
	if err != nil {
		return nil, trace, fmt.Errorf("agent1 context failed: %w", err)
	}
	trace.Agent1DurationMs = time.Since(t0).Milliseconds()
	trace.Agent1Type = "rule"
	log.Printf("[DeepMatch] stage=agent1_context resume=%d job=%d matched=%d missing=%d baseline=%d duration_ms=%d",
		resumeProfile.ResumeID, jobProfile.JobID, len(a1Output.MatchedTags), len(a1Output.MissingTags), a1Output.BaselineScore, trace.Agent1DurationMs)

	// ── Agent 2: LLM decision generation ──────────────────────────────
	t1 := time.Now()
	ctxJSON, _ := json.Marshal(a1Output)
	a2Req := pipeline.RunRequest{
		AgentName:        "MatchDecisionAgent",
		PromptName:       prompt.PromptMatchDecision,
		SchemaName:       schema.SchemaMatchDecision,
		TargetOutputType: pipeline.TargetMatchDecision,
		PromptValues: map[string]string{
			"MATCH_CONTEXT_JSON": string(ctxJSON),
		},
	}
	a2Result := runner2.Run(ctx, a2Req)
	if a2Result.FallbackRequired {
		return nil, trace, fmt.Errorf("agent2 decision failed: %s", a2Result.FallbackReason)
	}
	trace.Agent2DurationMs = time.Since(t1).Milliseconds()
	trace.Agent2Type = "llm_quality"

	a2Output := &MatchDecision{}
	a2JSON, _ := json.Marshal(a2Result.Parsed)
	json.Unmarshal(a2JSON, a2Output)
	log.Printf("[DeepMatch] stage=agent2_decision score=%d level=%s body_bytes=%d duration_ms=%d",
		a2Output.Score, a2Output.Level, len(a2JSON), trace.Agent2DurationMs)

	// ── Agent 3: LLM verifier ─────────────────────────────────────────
	t2 := time.Now()
	a3Input := map[string]interface{}{
		"decision": a2Output,
		"context":  a1Output,
	}
	a3JSON, _ := json.Marshal(a3Input)
	a3Req := pipeline.RunRequest{
		AgentName:        "MatchVerifierAgent",
		PromptName:       prompt.PromptMatchVerifier,
		SchemaName:       schema.SchemaMatchVerifier,
		TargetOutputType: pipeline.TargetMatchVerifier,
		PromptValues: map[string]string{
			"VERIFIER_INPUT_JSON": string(a3JSON),
		},
	}
	a3Result := runner3.Run(ctx, a3Req)
	trace.Agent3DurationMs = time.Since(t2).Milliseconds()
	trace.Agent3Type = "llm_quality"

	finalOutput := &MatchFinalOutput{}

	if a3Result.FallbackRequired {
		// Agent 3 LLM failed — deterministic fallback
		log.Printf("[DeepMatch] agent3_verifier_llm_failed reason=%s fallback_to_local_guard=true", a3Result.FallbackReason)
		finalOutput = normalizeMatchDecision(a2Output, a1Output)
		finalOutput.AgentStatus = "verified_by_local_guard"
		finalOutput.RepairApplied = true
	} else {
		a3Parsed, _ := json.Marshal(a3Result.Parsed)
		json.Unmarshal(a3Parsed, finalOutput)
		log.Printf("[DeepMatch] stage=agent3_verifier verified=%v body_bytes=%d duration_ms=%d",
			finalOutput.AgentStatus != "fallback", len(a3JSON), trace.Agent3DurationMs)
	}

	// Enforce constraints
	enforceMatchConstraints(finalOutput, a1Output)
	finalOutput.AgentTrace = &trace

	return finalOutput, trace, nil
}

// ─── Local normalization (deterministic, no LLM) ─────────────────────────

func normalizeMatchDecision(d *MatchDecision, ctx *MatchContext) *MatchFinalOutput {
	score := d.Score
	if score < 0 || score > 100 {
		score = ctx.BaselineScore
	}

	return &MatchFinalOutput{
		GeneratedBy:            "llm_agent",
		AgentStatus:            "repaired",
		Fallback:               false,
		FallbackReason:         "agent3_local_normalization",
		RepairApplied:          true,
		Score:                  score,
		Level:                  normalizeMatchLevel(score),
		Summary:                d.Summary,
		MatchedPoints:          d.MatchedPoints,
		GapPoints:              d.GapPoints,
		RiskPoints:             d.RiskPoints,
		ImprovementSuggestions: d.ImprovementSuggestions,
		Confidence:             d.Confidence,
	}
}

func enforceMatchConstraints(f *MatchFinalOutput, ctx *MatchContext) {
	if f.GeneratedBy == "" {
		f.GeneratedBy = "llm_agent"
	}
	if f.AgentStatus == "" {
		f.AgentStatus = "verified"
	}
	if f.Score < 0 || f.Score > 100 {
		f.Score = ctx.BaselineScore
	}
	if f.Level == "" {
		f.Level = normalizeMatchLevel(f.Score)
	}
	if f.MatchedPoints == nil {
		f.MatchedPoints = []MatchDecisionPoint{}
	}
	if f.GapPoints == nil {
		f.GapPoints = []MatchDecisionPoint{}
	}
	if f.RiskPoints == nil {
		f.RiskPoints = []MatchDecisionRisk{}
	}
	if f.ImprovementSuggestions == nil {
		f.ImprovementSuggestions = []MatchImprovement{}
	}
}

func normalizeMatchLevel(score int) string {
	if score >= 80 { return "high" }
	if score >= 60 { return "medium" }
	return "low"
}

// minInt is defined in recommendation_mock.go

// ─── Helpers ─────────────────────────────────────────────────────────────

func computeMatchTags(rp *model.ResumeTagProfile, jp *model.JobTagProfile) (matched, missing []string, score float64) {
	resumeSet := make(map[string]bool)
	for _, t := range rp.PrimaryTags { resumeSet[t.Tag] = true }
	for _, t := range rp.SkillTags { resumeSet[t.Tag] = true }
	for _, t := range rp.ProjectTags { resumeSet[t.Tag] = true }
	for _, t := range rp.InternshipTags { resumeSet[t.Tag] = true }

	jobSet := make(map[string]bool)
	for _, t := range jp.PrimaryTags { jobSet[t.Tag] = true }
	for _, t := range jp.SkillTags { jobSet[t.Tag] = true }
	for _, t := range jp.RequirementTags { jobSet[t.Tag] = true }

	score = 0
	for jt := range jobSet {
		if resumeSet[jt] {
			matched = append(matched, jt)
			score += 1.0
		} else {
			missing = append(missing, jt)
		}
	}
	return
}

func buildMatchEvidenceList(rp *model.ResumeTagProfile, tags []string, strength string) []MatchEvidenceSnippet {
	var result []MatchEvidenceSnippet
	for _, t := range tags {
		if len(result) >= 5 { break }
		result = append(result, MatchEvidenceSnippet{
			ID:       t, Type: "tag_match", Text: t, Strength: strength,
		})
	}
	// Add evidence from resume
	for _, e := range rp.Evidence.ProjectEvidence {
		if len(result) >= 8 { break }
		runes := []rune(e.Text)
		txt := e.Text
		if len(runes) > 60 { txt = string(runes[:60]) + "…" }
		result = append(result, MatchEvidenceSnippet{ID: e.ID, Type: e.Type, Text: txt, Strength: strength})
	}
	return result
}

func buildMatchRiskEvidence(jp *model.JobTagProfile, missing []string) []MatchEvidenceSnippet {
	var result []MatchEvidenceSnippet
	for _, t := range missing {
		if len(result) >= 5 { break }
		result = append(result, MatchEvidenceSnippet{ID: "risk_"+t, Type: "missing_tag", Text: t, Strength: "low"})
	}
	return result
}

func collectTagNames(tags []model.TagItem, limit int) []string {
	var result []string
	for _, t := range tags {
		if len(result) >= limit { break }
		result = append(result, t.Tag)
	}
	return result
}

func collectEvidenceRefs(rp *model.ResumeTagProfile, matched []string) []string {
	refs := matched
	for _, e := range rp.Evidence.ProjectEvidence {
		if len(refs) >= 10 { break }
		refs = append(refs, e.ID)
	}
	return refs
}
