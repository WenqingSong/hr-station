// ─── K13-Fix-C4: Deep Compare — 3-Agent Architecture ──────────────────
// Agent 1: CompareContext (rule-based, no LLM)
// Agent 2: CompareDecision (LLM quality)
// Agent 3: CompareVerifier (LLM quality)

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

// ─── Agent 1: CompareContext ──────────────────────────────────────────

type CompareContextInput struct {
	FromResumeID uint64
	ToResumeID   uint64
	FromProfile  *model.ResumeTagProfile
	ToProfile    *model.ResumeTagProfile
}

type CompareContext struct {
	AgentName      string               `json:"agent_name"`
	FromResume     *CompareResumeSummary `json:"from_resume"`
	ToResume       *CompareResumeSummary `json:"to_resume"`
	DiffSummary    *CompareDiffSummary   `json:"diff_summary"`
	ContextQuality *CompareContextQuality `json:"context_quality"`
	GeneratedAt    string               `json:"generated_at"`
}

type CompareResumeSummary struct {
	ResumeID        uint64   `json:"resume_id"`
	DominantProfile string   `json:"dominant_profile"`
	ProjectCount    int      `json:"project_count"`
	InternshipCount int      `json:"internship_count"`
	CoreTags        []string `json:"core_tags"`
	EvidenceRefs    []string `json:"evidence_refs"`
}

type CompareDiffSummary struct {
	AddedTags        []string `json:"added_tags"`
	StrengthenedTags []string `json:"strengthened_tags"`
	RemovedTags      []string `json:"removed_tags"`
	NewProjectEvidence []string `json:"new_project_evidence"`
	NewInternshipEvidence []string `json:"new_internship_evidence"`
	RiskChanges      []string `json:"risk_changes"`
}

type CompareContextQuality struct {
	FromHasEvidence bool `json:"from_has_evidence"`
	ToHasEvidence   bool `json:"to_has_evidence"`
	HasRealDiff     bool `json:"has_real_diff"`
}

func BuildCompareContext(_ context.Context, input *CompareContextInput) (*CompareContext, error) {
	fp, tp := input.FromProfile, input.ToProfile
	if fp == nil || tp == nil { return nil, fmt.Errorf("compare context: profiles required") }

	fromSum := &CompareResumeSummary{
		ResumeID: fp.ResumeID, DominantProfile: fp.DominantProfile,
		ProjectCount: fp.Evidence.ProjectCount, InternshipCount: fp.Evidence.InternshipCount,
		CoreTags: collectTagNames(fp.PrimaryTags, 5), EvidenceRefs: collectEvidenceIDRefs(fp),
	}
	toSum := &CompareResumeSummary{
		ResumeID: tp.ResumeID, DominantProfile: tp.DominantProfile,
		ProjectCount: tp.Evidence.ProjectCount, InternshipCount: tp.Evidence.InternshipCount,
		CoreTags: collectTagNames(tp.PrimaryTags, 5), EvidenceRefs: collectEvidenceIDRefs(tp),
	}

	diff := buildCompareDiff(fp, tp, fromSum, toSum)
	hasDiff := len(diff.AddedTags) > 0 || len(diff.StrengthenedTags) > 0 || len(diff.RemovedTags) > 0 || toSum.ProjectCount != fromSum.ProjectCount || toSum.InternshipCount != fromSum.InternshipCount

	return &CompareContext{
		AgentName:  "compare_context_agent",
		FromResume: fromSum, ToResume: toSum, DiffSummary: diff,
		ContextQuality: &CompareContextQuality{
			FromHasEvidence: len(fp.Evidence.ProjectEvidence) > 0, ToHasEvidence: len(tp.Evidence.ProjectEvidence) > 0, HasRealDiff: hasDiff,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func buildCompareDiff(fp, tp *model.ResumeTagProfile, fs, ts *CompareResumeSummary) *CompareDiffSummary {
	fromTagSet := allTagSet(fp)
	toTagSet := allTagSet(tp)
	diff := &CompareDiffSummary{}
	for t := range toTagSet {
		if !fromTagSet[t] { diff.AddedTags = append(diff.AddedTags, t) }
	}
	for _, e := range tp.Evidence.ProjectEvidence {
		diff.NewProjectEvidence = append(diff.NewProjectEvidence, truncRunes(e.Text, 60))
	}
	for _, e := range tp.Evidence.InternshipEvidence {
		diff.NewInternshipEvidence = append(diff.NewInternshipEvidence, truncRunes(e.Text, 60))
	}
	for _, t := range tp.RiskTags {
		found := false
		for _, ft := range fp.RiskTags { if ft.Tag == t.Tag { found = true; break } }
		if !found { diff.RiskChanges = append(diff.RiskChanges, t.Tag) }
	}
	return diff
}

func allTagSet(rp *model.ResumeTagProfile) map[string]bool {
	s := make(map[string]bool)
	for _, t := range rp.PrimaryTags { s[t.Tag] = true }
	for _, t := range rp.SkillTags { s[t.Tag] = true }
	for _, t := range rp.ProjectTags { s[t.Tag] = true }
	for _, t := range rp.InternshipTags { s[t.Tag] = true }
	return s
}

func truncRunes(s string, max int) string {
	rs := []rune(s)
	if len(rs) <= max { return s }
	return string(rs[:max]) + "…"
}

func collectEvidenceIDRefs(rp *model.ResumeTagProfile) []string {
	var refs []string
	for _, e := range rp.Evidence.ProjectEvidence { refs = append(refs, e.ID) }
	for _, e := range rp.Evidence.InternshipEvidence { refs = append(refs, e.ID) }
	return refs
}

// ─── Agent 2: CompareDecision ────────────────────────────────────────

type CompareDecision struct {
	AgentName          string   `json:"agent_name"`
	Score              int      `json:"score"`
	CompareLevel       string   `json:"compare_level"`
	WinnerResumeID     uint64   `json:"winner_resume_id"`
	OneSentenceSummary string   `json:"one_sentence_summary"`
	KeyDifferences     []string `json:"key_differences"`
	FromStrengths      []string `json:"from_strengths"`
	ToStrengths        []string `json:"to_strengths"`
	FromRisks          []string `json:"from_risks"`
	ToRisks            []string `json:"to_risks"`
	MergeSuggestions   []string `json:"merge_suggestions"`
	DecisionNotes      *CompareDecisionNotes `json:"decision_notes"`
}

type CompareDecisionNotes struct {
	EvidenceBased  bool `json:"evidence_based"`
	WinnerIDValid  bool `json:"winner_id_valid"`
}

// ─── Agent 3: CompareVerifier ─────────────────────────────────────────

type CompareFinalOutput struct {
	GeneratedBy        string   `json:"generated_by"`
	AgentStatus        string   `json:"agent_status"`
	Fallback           bool     `json:"fallback"`
	FallbackReason     string   `json:"fallback_reason"`
	RepairApplied      bool     `json:"repair_applied"`
	GenerationTier     string   `json:"generation_tier"`
	Score              int      `json:"score"`
	CompareLevel       string   `json:"compare_level"`
	WinnerResumeID     uint64   `json:"winner_resume_id"`
	OneSentenceSummary string   `json:"one_sentence_summary"`
	KeyDifferences     []string `json:"key_differences"`
	FromStrengths      []string `json:"from_strengths"`
	ToStrengths        []string `json:"to_strengths"`
	FromRisks          []string `json:"from_risks"`
	ToRisks            []string `json:"to_risks"`
	MergeSuggestions   []string `json:"merge_suggestions"`
	AgentTrace         *CompareAgentTrace `json:"agent_trace,omitempty"`
}

type CompareAgentTrace struct {
	Agent1Type       string `json:"agent1_type"`; Agent1DurationMs int64 `json:"agent1_duration_ms"`
	Agent2Type       string `json:"agent2_type"`; Agent2DurationMs int64 `json:"agent2_duration_ms"`
	Agent3Type       string `json:"agent3_type"`; Agent3DurationMs int64 `json:"agent3_duration_ms"`
}

// ─── Orchestrator ─────────────────────────────────────────────────────

func RunDeepCompare3Agents(
	ctx context.Context,
	fromProfile, toProfile *model.ResumeTagProfile,
	fromID, toID uint64,
	runner2, runner3 PipelineRunner,
) (*CompareFinalOutput, CompareAgentTrace, error) {
	trace := CompareAgentTrace{}

	// Agent 1: Rule-based diff
	t0 := time.Now()
	a1Out, err := BuildCompareContext(ctx, &CompareContextInput{FromResumeID: fromID, ToResumeID: toID, FromProfile: fromProfile, ToProfile: toProfile})
	if err != nil { return nil, trace, fmt.Errorf("agent1 context failed: %w", err) }
	trace.Agent1DurationMs = time.Since(t0).Milliseconds(); trace.Agent1Type = "rule"
	log.Printf("[DeepCompare] agent1_context from_proj=%d to_proj=%d from_intern=%d to_intern=%d added=%d removed=%d duration_ms=%d",
		a1Out.FromResume.ProjectCount, a1Out.ToResume.ProjectCount, a1Out.FromResume.InternshipCount, a1Out.ToResume.InternshipCount, len(a1Out.DiffSummary.AddedTags), len(a1Out.DiffSummary.RemovedTags), trace.Agent1DurationMs)

	// Agent 2: LLM decision
	t1 := time.Now()
	ctxJSON, _ := json.Marshal(a1Out)
	a2Req := pipeline.RunRequest{
		AgentName: "CompareDecisionAgent", PromptName: prompt.PromptCompareDecision, SchemaName: schema.SchemaCompareDecision, TargetOutputType: pipeline.TargetCompareDecision,
		PromptValues: map[string]string{"COMPARE_CONTEXT_JSON": string(ctxJSON)},
	}
	a2Result := runner2.Run(ctx, a2Req)
	if a2Result.FallbackRequired { return nil, trace, fmt.Errorf("agent2 decision failed: %s", a2Result.FallbackReason) }
	trace.Agent2DurationMs = time.Since(t1).Milliseconds(); trace.Agent2Type = "llm_quality"

	a2Out := &CompareDecision{}
	a2JSON, _ := json.Marshal(a2Result.Parsed); json.Unmarshal(a2JSON, a2Out)
	log.Printf("[DeepCompare] agent2_decision score=%d winner=%d body_bytes=%d duration_ms=%d", a2Out.Score, a2Out.WinnerResumeID, len(a2JSON), trace.Agent2DurationMs)

	// Agent 3: LLM verifier
	t2 := time.Now()
	a3In := map[string]interface{}{"decision": a2Out, "context": a1Out, "from_id": fromID, "to_id": toID}
	a3JSON, _ := json.Marshal(a3In)
	a3Req := pipeline.RunRequest{
		AgentName: "CompareVerifierAgent", PromptName: prompt.PromptCompareVerifier, SchemaName: schema.SchemaCompareVerifier, TargetOutputType: pipeline.TargetCompareVerifier,
		PromptValues: map[string]string{"VERIFIER_INPUT_JSON": string(a3JSON)},
	}
	a3Result := runner3.Run(ctx, a3Req)
	trace.Agent3DurationMs = time.Since(t2).Milliseconds(); trace.Agent3Type = "llm_quality"

	final := &CompareFinalOutput{}
	if a3Result.FallbackRequired {
		log.Printf("[DeepCompare] agent3_verifier_llm_failed reason=%s fallback_to_local_guard=true", a3Result.FallbackReason)
		final = localNormalizeCompare(a2Out, fromID, toID)
		final.AgentStatus = "verified_by_local_guard"; final.RepairApplied = true
	} else {
		a3Parsed, _ := json.Marshal(a3Result.Parsed); json.Unmarshal(a3Parsed, final)
		log.Printf("[DeepCompare] agent3_verifier verified=%v body_bytes=%d duration_ms=%d", final.AgentStatus != "fallback", len(a3JSON), trace.Agent3DurationMs)
	}

	enforceCompareConstraints(final, fromID, toID)
	final.AgentTrace = &trace
	return final, trace, nil
}

func localNormalizeCompare(d *CompareDecision, fromID, toID uint64) *CompareFinalOutput {
	w := d.WinnerResumeID
	if w != fromID && w != toID { w = toID }
	return &CompareFinalOutput{
		GeneratedBy: "llm_agent", AgentStatus: "repaired", Fallback: false, RepairApplied: true, GenerationTier: "quality",
		Score: d.Score, CompareLevel: d.CompareLevel, WinnerResumeID: w, OneSentenceSummary: d.OneSentenceSummary,
		KeyDifferences: d.KeyDifferences, FromStrengths: d.FromStrengths, ToStrengths: d.ToStrengths,
		FromRisks: d.FromRisks, ToRisks: d.ToRisks, MergeSuggestions: d.MergeSuggestions,
	}
}

func enforceCompareConstraints(f *CompareFinalOutput, fromID, toID uint64) {
	if f.GeneratedBy == "" { f.GeneratedBy = "llm_agent" }
	if f.AgentStatus == "" { f.AgentStatus = "verified" }
	if f.GenerationTier == "" { f.GenerationTier = "quality" }
	if f.Score < 0 || f.Score > 100 { f.Score = 70 }
	if f.WinnerResumeID != fromID && f.WinnerResumeID != toID { f.WinnerResumeID = toID }
	if f.KeyDifferences == nil { f.KeyDifferences = []string{} }
	if f.FromStrengths == nil { f.FromStrengths = []string{} }
	if f.ToStrengths == nil { f.ToStrengths = []string{} }
	if f.FromRisks == nil { f.FromRisks = []string{} }
	if f.ToRisks == nil { f.ToRisks = []string{} }
	if f.MergeSuggestions == nil { f.MergeSuggestions = []string{} }
}
