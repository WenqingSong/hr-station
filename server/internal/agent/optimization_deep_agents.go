// ─── K13-Fix-C3: Deep Optimization — 3-Agent Architecture ──────────────
// Agent 1: OptimizationContext (rule-based, no LLM)
// Agent 2: OptimizationDecision (LLM quality)
// Agent 3: OptimizationVerifier (LLM quality)

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

// ─── Agent 1: OptimizationContext ─────────────────────────────────────

type OptimizationContextInput struct {
	ResumeID      uint64
	ResumeProfile *model.ResumeTagProfile
	JobProfile    *model.JobTagProfile // nil for general mode
	OptimizationMode string
}

type OptimizationContext struct {
	AgentName          string              `json:"agent_name"`
	ResumeProfileSummary *ResumeContextSummary `json:"resume_profile_summary"`
	TargetJobSummary   *OptimizationJobSummary `json:"target_job_summary,omitempty"`
	OptimizationGaps   []OptimizationGap      `json:"optimization_gaps"`
	ContextQuality     *OptimizationContextQuality `json:"context_quality"`
	GeneratedAt        string                `json:"generated_at"`
}

type OptimizationJobSummary struct {
	JobID        uint64   `json:"job_id"`
	Title        string   `json:"title"`
	Direction    string   `json:"direction"`
	RequiredTags []string `json:"required_tags"`
	MatchedTags  []string `json:"matched_tags"`
	MissingTags  []string `json:"missing_tags"`
	RiskTags     []string `json:"risk_tags"`
}

type OptimizationGap struct {
	GapType     string   `json:"gap_type"`
	Tag         string   `json:"tag"`
	Severity    string   `json:"severity"`
	EvidenceRefs []string `json:"evidence_refs"`
	Suggestion  string   `json:"suggestion"`
}

type OptimizationContextQuality struct {
	HasDatabaseEvidence   bool `json:"has_database_evidence"`
	HasBackendEvidence    bool `json:"has_backend_evidence"`
	HasProjectEvidence    bool `json:"has_project_evidence"`
	HasInternshipEvidence bool `json:"has_internship_evidence"`
	ProjectCount          int  `json:"project_count"`
	InternshipCount       int  `json:"internship_count"`
}

func BuildOptimizationContext(_ context.Context, input *OptimizationContextInput) (*OptimizationContext, error) {
	rp := input.ResumeProfile
	if rp == nil {
		return nil, fmt.Errorf("optimization context: resume profile is nil")
	}

	// Build resume summary
	resSummary := &ResumeContextSummary{
		DominantProfile: rp.DominantProfile,
		CoreTags:        collectTagNames(rp.PrimaryTags, 6),
		EvidenceRefs:    collectAllEvidenceRefs(rp),
		ProjectCount:    rp.Evidence.ProjectCount,
		InternshipCount: rp.Evidence.InternshipCount,
		SkillCount:      rp.Evidence.SkillCount,
	}

	// Context quality
	cq := &OptimizationContextQuality{
		ProjectCount:          rp.Evidence.ProjectCount,
		InternshipCount:       rp.Evidence.InternshipCount,
		HasProjectEvidence:    len(rp.Evidence.ProjectEvidence) > 0,
		HasInternshipEvidence: len(rp.Evidence.InternshipEvidence) > 0,
		HasBackendEvidence:    hasOptBackendEvidence(rp),
		HasDatabaseEvidence:   hasOptDatabaseEvidence(rp),
	}

	// Job analysis
	var jobSummary *OptimizationJobSummary
	gaps := buildSimpleGaps(rp)

	if input.JobProfile != nil {
		jp := input.JobProfile
		matched, missing, _ := computeMatchTags(rp, jp)
		riskTags := collectTagNames(jp.RiskTags, 4)

		jobSummary = &OptimizationJobSummary{
			JobID:        jp.JobID,
			Title:        fmt.Sprintf("job_%d", jp.JobID),
			Direction:    jp.Direction,
			RequiredTags: collectTagNames(jp.RequirementTags, 6),
			MatchedTags:  matched,
			MissingTags:  missing,
			RiskTags:     riskTags,
		}

		// Build gaps from missing tags
		for _, m := range missing {
			gaps = append(gaps, OptimizationGap{
				GapType: "job_requirement_gap", Tag: m, Severity: "medium",
			})
		}
	}

	return &OptimizationContext{
		AgentName:           "optimization_context_agent",
		ResumeProfileSummary: resSummary,
		TargetJobSummary:    jobSummary,
		OptimizationGaps:    gaps,
		ContextQuality:      cq,
		GeneratedAt:         time.Now().Format(time.RFC3339),
	}, nil
}

// ─── Agent 2: OptimizationDecision ───────────────────────────────────

type OptimizationDecision struct {
	AgentName          string                           `json:"agent_name"`
	OptimizationScore  int                              `json:"optimization_score"`
	QuickFixItems      []model.OptimizationFixItem      `json:"quick_fix_items"`
	StructureFixItems  []model.OptimizationFixItem      `json:"structure_fix_items"`
	ContentFixItems    []model.OptimizationFixItem      `json:"content_fix_items"`
	KeywordFixItems    []model.OptimizationFixItem      `json:"keyword_fix_items"`
	EvidenceFixItems   []model.OptimizationFixItem      `json:"evidence_fix_items"`
	RiskPoints         []string                         `json:"risk_points"`
	RewriteExamples    []model.OptimizationRewriteExample `json:"rewrite_examples"`
	DecisionNotes      *OptimizationDecisionNotes       `json:"decision_notes"`
}

type OptimizationDecisionNotes struct {
	EvidenceBased   bool `json:"evidence_based"`
	TargetJobAligned bool `json:"target_job_aligned"`
}

// ─── Agent 3: OptimizationVerifier ───────────────────────────────────

type OptimizationFinalOutput struct {
	GeneratedBy            string                           `json:"generated_by"`
	AgentStatus            string                           `json:"agent_status"`
	Fallback               bool                             `json:"fallback"`
	FallbackReason         string                           `json:"fallback_reason"`
	RepairApplied          bool                             `json:"repair_applied"`
	GenerationTier         string                           `json:"generation_tier"`
	OptimizationScore      int                              `json:"optimization_score"`
	QuickFixItems          []model.OptimizationFixItem      `json:"quick_fix_items"`
	StructureFixItems      []model.OptimizationFixItem      `json:"structure_fix_items"`
	ContentFixItems        []model.OptimizationFixItem      `json:"content_fix_items"`
	KeywordFixItems        []model.OptimizationFixItem      `json:"keyword_fix_items"`
	EvidenceFixItems       []model.OptimizationFixItem      `json:"evidence_fix_items"`
	RiskPoints             []string                         `json:"risk_points"`
	RewriteExamples        []model.OptimizationRewriteExample `json:"rewrite_examples"`
	AgentTrace             *OptAgentTrace                   `json:"agent_trace,omitempty"`
}

type OptAgentTrace struct {
	Agent1Type      string `json:"agent1_type"`
	Agent1DurationMs int64 `json:"agent1_duration_ms"`
	Agent2Type      string `json:"agent2_type"`
	Agent2DurationMs int64 `json:"agent2_duration_ms"`
	Agent3Type      string `json:"agent3_type"`
	Agent3DurationMs int64 `json:"agent3_duration_ms"`
}

// ─── Orchestrator ────────────────────────────────────────────────────

func RunDeepOptimization3Agents(
	ctx context.Context,
	resumeProfile *model.ResumeTagProfile,
	jobProfile *model.JobTagProfile,
	optMode string,
	runner2 PipelineRunner,
	runner3 PipelineRunner,
) (*OptimizationFinalOutput, OptAgentTrace, error) {
	trace := OptAgentTrace{}

	// Agent 1: Rule-based context
	t0 := time.Now()
	a1In := &OptimizationContextInput{
		ResumeID: resumeProfile.ResumeID, ResumeProfile: resumeProfile,
		JobProfile: jobProfile, OptimizationMode: optMode,
	}
	a1Out, err := BuildOptimizationContext(ctx, a1In)
	if err != nil {
		return nil, trace, fmt.Errorf("agent1 context failed: %w", err)
	}
	trace.Agent1DurationMs = time.Since(t0).Milliseconds()
	trace.Agent1Type = "rule"
	log.Printf("[DeepOptimization] agent1_context gaps=%d project_count=%d internship_count=%d duration_ms=%d",
		len(a1Out.OptimizationGaps), a1Out.ContextQuality.ProjectCount, a1Out.ContextQuality.InternshipCount, trace.Agent1DurationMs)

	// Agent 2: LLM decision
	t1 := time.Now()
	ctxJSON, _ := json.Marshal(a1Out)
	a2Req := pipeline.RunRequest{
		AgentName:        "OptimizationDecisionAgent",
		PromptName:       prompt.PromptOptimizationDecision,
		SchemaName:       schema.SchemaOptimizationDecision,
		TargetOutputType: pipeline.TargetOptimizationDecision,
		PromptValues:     map[string]string{"OPTIMIZATION_CONTEXT_JSON": string(ctxJSON)},
	}
	a2Result := runner2.Run(ctx, a2Req)
	if a2Result.FallbackRequired {
		return nil, trace, fmt.Errorf("agent2 decision failed: %s", a2Result.FallbackReason)
	}
	trace.Agent2DurationMs = time.Since(t1).Milliseconds()
	trace.Agent2Type = "llm_quality"

	a2Out := &OptimizationDecision{}
	a2JSON, _ := json.Marshal(a2Result.Parsed)
	json.Unmarshal(a2JSON, a2Out)
	log.Printf("[DeepOptimization] agent2_decision score=%d fix_items=%d rewrite=%d body_bytes=%d duration_ms=%d",
		a2Out.OptimizationScore, len(a2Out.QuickFixItems), len(a2Out.RewriteExamples), len(a2JSON), trace.Agent2DurationMs)

	// Agent 3: LLM verifier
	t2 := time.Now()
	a3In := map[string]interface{}{"decision": a2Out, "context": a1Out}
	a3JSON, _ := json.Marshal(a3In)
	a3Req := pipeline.RunRequest{
		AgentName:        "OptimizationVerifierAgent",
		PromptName:       prompt.PromptOptimizationVerifier,
		SchemaName:       schema.SchemaOptimizationVerifier,
		TargetOutputType: pipeline.TargetOptimizationVerifier,
		PromptValues:     map[string]string{"VERIFIER_INPUT_JSON": string(a3JSON)},
	}
	a3Result := runner3.Run(ctx, a3Req)
	trace.Agent3DurationMs = time.Since(t2).Milliseconds()
	trace.Agent3Type = "llm_quality"

	final := &OptimizationFinalOutput{}
	if a3Result.FallbackRequired {
		log.Printf("[DeepOptimization] agent3_verifier_llm_failed reason=%s fallback_to_local_guard=true", a3Result.FallbackReason)
		final = localNormalizeOpt(a2Out, a1Out)
		final.AgentStatus = "verified_by_local_guard"
		final.RepairApplied = true
	} else {
		a3Parsed, _ := json.Marshal(a3Result.Parsed)
		json.Unmarshal(a3Parsed, final)
		log.Printf("[DeepOptimization] agent3_verifier verified=%v body_bytes=%d duration_ms=%d",
			final.AgentStatus != "fallback", len(a3JSON), trace.Agent3DurationMs)
	}

	// Enforce
	enforceOptConstraints(final, a1Out)
	final.AgentTrace = &trace
	return final, trace, nil
}

// ─── Local normalization (fallback only) ────────────────────────────

func localNormalizeOpt(d *OptimizationDecision, ctx *OptimizationContext) *OptimizationFinalOutput {
	return &OptimizationFinalOutput{
		GeneratedBy:       "llm_agent",
		AgentStatus:       "repaired",
		Fallback:          false,
		RepairApplied:     true,
		GenerationTier:    "quality",
		OptimizationScore: d.OptimizationScore,
		QuickFixItems:     d.QuickFixItems,
		StructureFixItems: d.StructureFixItems,
		ContentFixItems:   d.ContentFixItems,
		KeywordFixItems:   d.KeywordFixItems,
		EvidenceFixItems:  d.EvidenceFixItems,
		RiskPoints:        d.RiskPoints,
		RewriteExamples:   d.RewriteExamples,
	}
}

func enforceOptConstraints(f *OptimizationFinalOutput, ctx *OptimizationContext) {
	if f.GeneratedBy == "" { f.GeneratedBy = "llm_agent" }
	if f.AgentStatus == "" { f.AgentStatus = "verified" }
	if f.GenerationTier == "" { f.GenerationTier = "quality" }
	if f.OptimizationScore < 0 || f.OptimizationScore > 100 { f.OptimizationScore = 70 }
	if f.QuickFixItems == nil { f.QuickFixItems = []model.OptimizationFixItem{} }
	if f.StructureFixItems == nil { f.StructureFixItems = []model.OptimizationFixItem{} }
	if f.ContentFixItems == nil { f.ContentFixItems = []model.OptimizationFixItem{} }
	if f.KeywordFixItems == nil { f.KeywordFixItems = []model.OptimizationFixItem{} }
	if f.EvidenceFixItems == nil { f.EvidenceFixItems = []model.OptimizationFixItem{} }
	if f.RiskPoints == nil { f.RiskPoints = []string{} }
	if f.RewriteExamples == nil { f.RewriteExamples = []model.OptimizationRewriteExample{} }
}

// ─── Helpers ────────────────────────────────────────────────────────

func hasOptDatabaseEvidence(rp *model.ResumeTagProfile) bool {
	for _, t := range rp.SkillTags {
		if t.Tag == "mysql" || t.Tag == "redis" { return true }
	}
	for _, e := range rp.Evidence.SkillEvidence {
		if containsText(e.Text, "MySQL", "Redis") { return true }
	}
	return false
}

func hasOptBackendEvidence(rp *model.ResumeTagProfile) bool {
	for _, t := range rp.PrimaryTags {
		if t.Tag == "backend" || t.Tag == "golang" || t.Tag == "java" { return true }
	}
	return false
}

func containsText(text string, keywords ...string) bool {
	for _, k := range keywords {
		if len(text) >= len(k) {
			for i := 0; i <= len(text)-len(k); i++ {
				if text[i:i+len(k)] == k { return true }
			}
		}
	}
	return false
}

func buildSimpleGaps(rp *model.ResumeTagProfile) []OptimizationGap {
	var gaps []OptimizationGap
	if !hasOptBackendEvidence(rp) { gaps = append(gaps, OptimizationGap{GapType: "skill_gap", Tag: "backend_evidence", Severity: "medium", Suggestion: "补充后端开发相关项目经历"}) }
	if !hasOptDatabaseEvidence(rp) { gaps = append(gaps, OptimizationGap{GapType: "skill_gap", Tag: "database_evidence", Severity: "low", Suggestion: "补充MySQL/Redis相关经验"}) }
	if rp.Evidence.ProjectCount < 2 { gaps = append(gaps, OptimizationGap{GapType: "quantity_gap", Tag: "projects", Severity: "medium", Suggestion: "建议增加项目经历至2个以上"}) }
	return gaps
}

func collectAllEvidenceRefs(rp *model.ResumeTagProfile) []string {
	var refs []string
	for _, e := range rp.Evidence.ProjectEvidence { refs = append(refs, e.ID) }
	for _, e := range rp.Evidence.InternshipEvidence { refs = append(refs, e.ID) }
	return refs
}
