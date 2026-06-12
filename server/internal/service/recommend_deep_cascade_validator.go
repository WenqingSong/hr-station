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
	if strings.TrimSpace(out.ResumeImprovementHint) == "" {
		return fmt.Errorf("resume_improvement_hint empty")
	}
	if len(out.DisplayTags) < 3 {
		return fmt.Errorf("display_tags count %d < 3", len(out.DisplayTags))
	}
	if strings.TrimSpace(out.WhyRankedHere) == "" {
		return fmt.Errorf("why_ranked_here empty")
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
		o.WhyRankedHere = sanitizeStr(o.WhyRankedHere)
		o.ApplicationStrategy = sanitizeStr(o.ApplicationStrategy)
		o.ResumeImprovementHint = sanitizeStr(o.ResumeImprovementHint)
		o.DisplayTags = sanitizeStrList(o.DisplayTags)
		o.WhyRecommended = sanitizeStrList(o.WhyRecommended)
		o.ResumeEvidence = sanitizeStrList(o.ResumeEvidence)
		o.RiskPoints = sanitizeStrList(o.RiskPoints)
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
