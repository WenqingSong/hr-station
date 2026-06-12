// ─── Deep Cascade Validators ──────────────────────────────────────────────

package service

import (
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// ─── Layer 1 ──────────────────────────────────────────────────────────────

func validateLayer1(out *DeepCandidateDiscoveryOutput, jobSEP map[int64]*gateJobSep) error {
	if len(out.Candidates) < 5 {
		return fmt.Errorf("candidates count %d < 5", len(out.Candidates))
	}
	if len(out.Candidates) > 15 {
		return fmt.Errorf("candidates count %d > 15", len(out.Candidates))
	}
	seen := map[int64]bool{}
	for i, c := range out.Candidates {
		if _, ok := jobSEP[c.JobID]; !ok {
			return fmt.Errorf("candidates[%d]: job_id=%d not in SEP pool", i, c.JobID)
		}
		if seen[c.JobID] {
			return fmt.Errorf("candidates[%d]: duplicate job_id=%d", i, c.JobID)
		}
		seen[c.JobID] = true
		if strings.TrimSpace(c.CandidateReason) == "" {
			return fmt.Errorf("candidates[%d]: candidate_reason empty", i)
		}
		if len(c.MatchedSignals) < 2 {
			return fmt.Errorf("candidates[%d]: matched_signals < 2", i)
		}
		if c.InitialScore < 0 || c.InitialScore > 100 {
			return fmt.Errorf("candidates[%d]: initial_score %d out of range", i, c.InitialScore)
		}
	}
	return nil
}

// ─── Layer 2 ──────────────────────────────────────────────────────────────

func validateLayer2(out *DeepSemanticRerankOutput, l1 *DeepCandidateDiscoveryOutput) error {
	if len(out.TopJobs) != 5 {
		return fmt.Errorf("top_jobs count %d != 5", len(out.TopJobs))
	}
	l1Set := map[int64]bool{}
	for _, c := range l1.Candidates { l1Set[c.JobID] = true }

	seen := map[int64]bool{}
	ranks := map[int]bool{}
	for i, tj := range out.TopJobs {
		if !l1Set[tj.JobID] {
			return fmt.Errorf("top_jobs[%d]: job_id=%d not in L1 candidates", i, tj.JobID)
		}
		if seen[tj.JobID] {
			return fmt.Errorf("top_jobs[%d]: duplicate job_id=%d", i, tj.JobID)
		}
		seen[tj.JobID] = true
		if tj.Rank < 1 || tj.Rank > 5 {
			return fmt.Errorf("top_jobs[%d]: rank %d out of range", i, tj.Rank)
		}
		if ranks[tj.Rank] {
			return fmt.Errorf("top_jobs[%d]: duplicate rank %d", i, tj.Rank)
		}
		ranks[tj.Rank] = true
		if tj.Score < 0 || tj.Score > 100 {
			return fmt.Errorf("top_jobs[%d]: score %d out of range", i, tj.Score)
		}
		if !isValidFitLevel(tj.FitLevel) {
			return fmt.Errorf("top_jobs[%d]: invalid fit_level %q", i, tj.FitLevel)
		}
		if strings.TrimSpace(tj.WhyThisRank) == "" {
			return fmt.Errorf("top_jobs[%d]: why_this_rank empty", i)
		}
	}
	// Validate excluded jobs
	for i, ej := range out.ExcludedJobs {
		if !l1Set[ej.JobID] {
			return fmt.Errorf("excluded_jobs[%d]: job_id=%d not in L1 candidates", i, ej.JobID)
		}
	}
	return nil
}

// ─── Layer 3 ──────────────────────────────────────────────────────────────

func validateLayer3(out *DeepPerJobSynthesisOutput) error {
	if out.FinalScore < 0 || out.FinalScore > 100 {
		return fmt.Errorf("final_score %d out of range", out.FinalScore)
	}
	if !isValidFitLevel(out.FitLevel) {
		return fmt.Errorf("invalid fit_level %q", out.FitLevel)
	}
	if strings.TrimSpace(out.OneSentenceConclusion) == "" {
		return fmt.Errorf("one_sentence_conclusion empty")
	}
	if strings.TrimSpace(out.JobSpecificReason) == "" {
		return fmt.Errorf("job_specific_reason empty")
	}
	if len(out.WhyRecommended) < 3 {
		return fmt.Errorf("why_recommended count %d < 3", len(out.WhyRecommended))
	}
	if len(out.ResumeEvidence) < 2 {
		return fmt.Errorf("resume_evidence count %d < 2", len(out.ResumeEvidence))
	}
	if len(out.JobRequirementMapping) < 2 {
		return fmt.Errorf("job_requirement_mapping count %d < 2", len(out.JobRequirementMapping))
	}
	if len(out.RiskPoints) < 1 {
		return fmt.Errorf("risk_points count < 1")
	}
	if strings.TrimSpace(out.ApplicationStrategy) == "" {
		return fmt.Errorf("application_strategy empty")
	}
	if strings.TrimSpace(out.SuggestedAction) == "" {
		return fmt.Errorf("suggested_action empty")
	}
	if strings.TrimSpace(out.ResumeImprovementHint) == "" {
		return fmt.Errorf("resume_improvement_hint empty")
	}
	if len(out.DisplayTags) < 3 {
		return fmt.Errorf("display_tags count %d < 3", len(out.DisplayTags))
	}
	if strings.TrimSpace(out.WhyRankedHere) == "" {
		return fmt.Errorf("why_ranked_here empty")
	}
	if strings.TrimSpace(out.CandidateCoreStrength) == "" {
		return fmt.Errorf("candidate_core_strength empty")
	}

	// ── L2: duplicate detection ──
	if sameText(out.OneSentenceConclusion, out.JobSpecificReason) {
		return fmt.Errorf("one_sentence_conclusion and job_specific_reason are identical: %q", truncateStr(out.OneSentenceConclusion, 60))
	}
	if sameText(out.ApplicationStrategy, out.SuggestedAction) {
		return fmt.Errorf("application_strategy and suggested_action are identical: %q", truncateStr(out.ApplicationStrategy, 60))
	}
	if sameText(out.ApplicationStrategy, out.ResumeImprovementHint) {
		return fmt.Errorf("application_strategy and resume_improvement_hint are identical: %q", truncateStr(out.ApplicationStrategy, 60))
	}
	if sameText(out.SuggestedAction, out.ResumeImprovementHint) {
		return fmt.Errorf("suggested_action and resume_improvement_hint are identical: %q", truncateStr(out.SuggestedAction, 60))
	}
	if strListsEqual(out.RiskPoints, out.JobSpecificRisks) && len(out.JobSpecificRisks) > 0 {
		return fmt.Errorf("risk_points and job_specific_risks are identical")
	}

	// ── L2: boilerplate/generic text detection ──
	boilerplateStrs := []string{
		out.OneSentenceConclusion, out.JobSpecificReason, out.ApplicationStrategy, out.SuggestedAction, out.ResumeImprovementHint, out.CandidateCoreStrength, out.FitConfidenceReason,
	}
	for _, s := range boilerplateStrs {
		if isBoilerplateText(s) {
			return fmt.Errorf("boilerplate/generic text detected: %q", truncateStr(s, 60))
		}
	}

	// ── L2: raw code / tag in display_tags ──
	for _, t := range out.DisplayTags {
		if model.IsBlacklistedRawCode(t) {
			return fmt.Errorf("display_tags contains raw code: %q", t)
		}
		if isBoilerplateText(t) {
			return fmt.Errorf("display_tags contains boilerplate: %q", t)
		}
	}

	return nil
}

// ─── Layer 4 ──────────────────────────────────────────────────────────────

func validateLayer4(out *DeepCalibrationOutput, l3Results []*DeepPerJobSynthesisOutput) error {
	if len(out.FinalOrder) != len(l3Results) {
		return fmt.Errorf("final_order count %d != l3 count %d", len(out.FinalOrder), len(l3Results))
	}
	l3Set := map[int64]bool{}
	for _, l3 := range l3Results { l3Set[l3.JobID] = true }

	ranks := map[int]bool{}
	for i, cj := range out.FinalOrder {
		if !l3Set[cj.JobID] {
			return fmt.Errorf("final_order[%d]: job_id=%d not in L3 results", i, cj.JobID)
		}
		if cj.FinalRank < 1 || cj.FinalRank > len(l3Results) {
			return fmt.Errorf("final_order[%d]: rank %d out of range", i, cj.FinalRank)
		}
		if ranks[cj.FinalRank] {
			return fmt.Errorf("final_order[%d]: duplicate rank %d", i, cj.FinalRank)
		}
		ranks[cj.FinalRank] = true
		if cj.CalibratedScore < 0 || cj.CalibratedScore > 100 {
			return fmt.Errorf("final_order[%d]: calibrated_score %d out of range", i, cj.CalibratedScore)
		}
	}
	if !isValidQualityStatus(out.QualityStatus) {
		return fmt.Errorf("invalid quality_status %q", out.QualityStatus)
	}

	// ── L2: global_summary must be meaningful ──
	if strings.TrimSpace(out.GlobalSummary) == "" {
		return fmt.Errorf("global_summary empty")
	}
	if isBoilerplateText(out.GlobalSummary) {
		return fmt.Errorf("global_summary is boilerplate: %q", truncateStr(out.GlobalSummary, 60))
	}
	if len([]rune(out.GlobalSummary)) > 250 {
		return fmt.Errorf("global_summary too long (%d chars, max 250)", len([]rune(out.GlobalSummary)))
	}

	return nil
}

// ─── Sanity ───────────────────────────────────────────────────────────────

func isValidFitLevel(f string) bool {
	switch f {
	case "优先投递", "重点关注", "可以尝试", "谨慎尝试", "谨慎投递", "可以考虑", "可以备选", "建议投递", "值得尝试", "中等匹配", "可投递", "一般匹配":
		return true
	}
	return false
}

func isValidQualityStatus(s string) bool {
	switch s {
	case "verified", "repaired", "failed":
		return true
	}
	return false
}

// ─── Raw Code Sanity (reuse from explainer) ──────────────────────────────

func sanitizeCascadeOutputs(outputs []*DeepPerJobSynthesisOutput) {
	for _, o := range outputs {
		o.OneSentenceConclusion = sanitizeStr(o.OneSentenceConclusion)
		o.JobSpecificReason = sanitizeStr(o.JobSpecificReason)
		o.WhyRankedHere = sanitizeStr(o.WhyRankedHere)
		o.ApplicationStrategy = sanitizeStr(o.ApplicationStrategy)
		o.SuggestedAction = sanitizeStr(o.SuggestedAction)
		o.ResumeImprovementHint = sanitizeStr(o.ResumeImprovementHint)
		o.CandidateCoreStrength = sanitizeStr(o.CandidateCoreStrength)
		o.FitConfidenceReason = sanitizeStr(o.FitConfidenceReason)
		o.DisplayTags = sanitizeStrList(o.DisplayTags)
		o.WhyRecommended = sanitizeStrList(o.WhyRecommended)
		o.ResumeEvidence = sanitizeStrList(o.ResumeEvidence)
		o.RiskPoints = sanitizeStrList(o.RiskPoints)
		o.JobSpecificRisks = sanitizeStrList(o.JobSpecificRisks)
	}
}

func sanitizeStr(s string) string {
	if model.IsBlacklistedRawCode(s) { return "" }
	return s
}

func sanitizeStrList(items []string) []string {
	var r []string
	for _, item := range items {
		if model.IsBlacklistedRawCode(item) { continue }
		if item != "" { r = append(r, item) }
	}
	return r
}

// ─── L2: Validation Helpers ──────────────────────────────────────────────────

func sameText(a, b string) bool {
	na := strings.TrimSpace(strings.ReplaceAll(a, " ", ""))
	nb := strings.TrimSpace(strings.ReplaceAll(b, " ", ""))
	return na == nb
}

func strListsEqual(a, b []string) bool {
	if len(a) != len(b) || len(a) == 0 { return false }
	for i := range a {
		if !sameText(a[i], b[i]) { return false }
	}
	return true
}

func isBoilerplateText(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" { return false }
	boilerplates := []string{
		"存在一定关联", "建议进一步确认", "具备一定基础",
		"可考虑投递相关岗位", "适合投递相关岗位",
		"该岗位与当前简历存在一定关联", "具备岗位所需能力",
		"需要提升综合能力", "建议进一步补充",
		"标签匹配", "缺失标签", "建议进一步优化",
	}
	for _, b := range boilerplates {
		if strings.Contains(s, b) { return true }
	}
	return false
}
