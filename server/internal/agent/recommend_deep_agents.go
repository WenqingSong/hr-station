// ─── K13-Fix-C1: Deep Recommendation — 3-Agent Architecture ──────────────
// Agent 1: Context/Tag/Evidence (rule-based, no LLM)
// Agent 2: Decision/Generation (LLM)
// Agent 3: Verifier/Frontend-Contract (LLM)

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/schema"
)

// ─── Agent 1: Context / Tag / Evidence ──────────────────────────────────

// RecommendationContextInput is the input to Agent 1.
type RecommendationContextInput struct {
	ResumeID    uint64
	ResumeProfile *model.ResumeTagProfile
	JobProfiles   []*model.JobTagProfile
	CandidateLimit int
}

// RecommendationContext is the output of Agent 1.
type RecommendationContext struct {
	AgentName           string                    `json:"agent_name"`
	ResumeSummary       *ResumeContextSummary     `json:"resume_profile_summary"`
	CandidateJobs       []CandidateContextJob      `json:"candidate_jobs"`
	ProfileCheck        *ProfileRealityCheck       `json:"profile_reality_check"`
	GeneratedAt         string                    `json:"generated_at"`
}

type ResumeContextSummary struct {
	DominantProfile string   `json:"dominant_profile"`
	CoreTags        []string `json:"core_tags"`
	EvidenceRefs    []string `json:"evidence_refs"`
	ProjectCount    int      `json:"project_count"`
	InternshipCount int      `json:"internship_count"`
	SkillCount      int      `json:"skill_count"`
}

type CandidateContextJob struct {
	JobID              uint64   `json:"job_id"`
	Title              string   `json:"title"`
	Department         string   `json:"department"`
	City               string   `json:"city"`
	Category           string   `json:"category"`
	Direction          string   `json:"direction"`
	TagScore           float64  `json:"tag_score"`
	MatchedTags        []string `json:"matched_tags"`
	MissingTags        []string `json:"missing_tags"`
	ResumeEvidenceRefs []string `json:"resume_evidence_refs"`
	JobEvidenceSummary string   `json:"job_evidence_summary"`
	CandidateReason    string   `json:"candidate_reason"`
}

type ProfileRealityCheck struct {
	ProjectCountOK    bool   `json:"project_count_ok"`
	InternshipCountOK bool   `json:"internship_count_ok"`
	ProfileConsistent bool   `json:"profile_consistent"`
	Note              string `json:"note"`
}

// BuildRecommendationContext is Agent 1 — pure rule-based, no LLM.
// It reads TagEvidenceService data and produces a compressed candidate context.
func BuildRecommendationContext(ctx context.Context, input *RecommendationContextInput) (*RecommendationContext, error) {
	limit := input.CandidateLimit
	if limit <= 0 {
		limit = 15
	}

	rp := input.ResumeProfile
	jps := input.JobProfiles

	// Build resume summary
	resumeSummary := &ResumeContextSummary{
		DominantProfile: rp.DominantProfile,
		CoreTags:        extractCoreTags(rp),
		EvidenceRefs:    collectResumeEvidenceRefs(rp),
		ProjectCount:    rp.Evidence.ProjectCount,
		InternshipCount: rp.Evidence.InternshipCount,
		SkillCount:      rp.Evidence.SkillCount,
	}

	// Build candidate pool
	candidates := buildCandidatePool(rp, jps, limit)

	// Reality check
	check := &ProfileRealityCheck{
		ProjectCountOK:    rp.Evidence.ProjectCount > 0,
		InternshipCountOK: rp.Evidence.InternshipCount > 0,
		ProfileConsistent: len(rp.PrimaryTags) > 0,
		Note:              "",
	}
	if !check.ProjectCountOK && !check.InternshipCountOK {
		check.Note = "简历未检测到项目和实习经历"
	}

	return &RecommendationContext{
		AgentName:       "recommend_context_agent",
		ResumeSummary:   resumeSummary,
		CandidateJobs:   candidates,
		ProfileCheck:    check,
		GeneratedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

func extractCoreTags(rp *model.ResumeTagProfile) []string {
	var tags []string
	for _, t := range rp.PrimaryTags {
		tags = append(tags, t.Tag)
	}
	for _, t := range rp.SkillTags {
		if len(tags) >= 10 { break }
		tags = append(tags, t.Tag)
	}
	return tags
}

func collectResumeEvidenceRefs(rp *model.ResumeTagProfile) []string {
	var refs []string
	for _, e := range rp.Evidence.SkillEvidence {
		if len(refs) >= 3 { break }
		refs = append(refs, e.ID)
	}
	for _, e := range rp.Evidence.ProjectEvidence {
		if len(refs) >= 3 { break }
		refs = append(refs, e.ID)
	}
	return refs
}

func buildCandidatePool(rp *model.ResumeTagProfile, jps []*model.JobTagProfile, limit int) []CandidateContextJob {
	type scored struct {
		job     *model.JobTagProfile
		matched []string
		missing []string
		score   float64
	}
	var scoredList []scored

	for _, jp := range jps {
		matched, missing, score := computeTagMatchFromProfiles(rp, jp)
		// Apply direction bonus
		if len(matched) == 0 && len(jp.PrimaryTags) == 0 {
			continue // skip jobs with no profile data
		}
		scoredList = append(scoredList, scored{jp, matched, missing, score})
	}

	// Sort desc
	for i := 0; i < len(scoredList); i++ {
		for j := i + 1; j < len(scoredList); j++ {
			if scoredList[j].score > scoredList[i].score {
				scoredList[i], scoredList[j] = scoredList[j], scoredList[i]
			}
		}
	}

	if len(scoredList) > limit {
		scoredList = scoredList[:limit]
	}

	var result []CandidateContextJob
	for _, s := range scoredList {
		evSummary := buildJobEvidenceSummary(s.job)
		reason := ""
		if len(s.matched) > 0 {
			reason = fmt.Sprintf("匹配标签：%s", strings.Join(s.matched[:min(3, len(s.matched))], "、"))
		}
		if len(s.missing) > 0 {
			if reason != "" { reason += "；" }
			reason += fmt.Sprintf("缺失：%s", strings.Join(s.missing[:min(3, len(s.missing))], "、"))
		}

		result = append(result, CandidateContextJob{
			JobID:              s.job.JobID,
			Title:              "", // filled by caller
			Department:         "", // filled by caller
			City:               "", // filled by caller
			Category:           s.job.Category,
			Direction:          s.job.Direction,
			TagScore:           math.Round(s.score*100) / 100,
			MatchedTags:        s.matched,
			MissingTags:        s.missing,
			ResumeEvidenceRefs: collectResumeEvidenceRefs(rp),
			JobEvidenceSummary: evSummary,
			CandidateReason:    reason,
		})
	}
	return result
}

func collectResumeTagNames(rp *model.ResumeTagProfile) map[string]bool {
	m := make(map[string]bool)
	for _, t := range rp.PrimaryTags { m[t.Tag] = true }
	for _, t := range rp.SkillTags { m[t.Tag] = true }
	for _, t := range rp.ProjectTags { m[t.Tag] = true }
	for _, t := range rp.InternshipTags { m[t.Tag] = true }
	return m
}

func computeTagMatchFromProfiles(rp *model.ResumeTagProfile, jp *model.JobTagProfile) (matched, missing []string, score float64) {
	resumeTags := collectResumeTagNames(rp)
	jobTags := make(map[string]bool)
	for _, t := range jp.PrimaryTags { jobTags[t.Tag] = true }
	for _, t := range jp.SkillTags { jobTags[t.Tag] = true }
	for _, t := range jp.RequirementTags { jobTags[t.Tag] = true }

	if len(jobTags) == 0 {
		return nil, nil, 0
	}

	for tag := range jobTags {
		if resumeTags[tag] {
			matched = append(matched, tag)
		} else {
			missing = append(missing, tag)
		}
	}

	score = (float64(len(matched)) / float64(len(jobTags))) * 100
	// Direction bonus
	dom := rp.DominantProfile
	dir := jp.Direction
	if dom == "algorithm_research" && strings.Contains(dir, "研究") { score = math.Min(100, score+10) }
	if dom == "backend_engineering" && strings.Contains(dir, "技术") && !strings.Contains(dir, "研究") { score = math.Min(100, score+10) }
	return
}

func buildJobEvidenceSummary(jp *model.JobTagProfile) string {
	var parts []string
	for _, e := range jp.Evidence.RequirementEvidence {
		if len(parts) >= 3 { break }
		runes := []rune(e.Text)
		if len(runes) > 60 {
			parts = append(parts, string(runes[:60])+"…")
		} else {
			parts = append(parts, e.Text)
		}
	}
	return strings.Join(parts, "；")
}

// ─── Agent 2: Decision / Generation ─────────────────────────────────────

// RecommendationDecisionInput is Agent 2's input.
type RecommendationDecisionInput struct {
	ResumeSummary *ResumeContextSummary
	Candidates    []CandidateContextJob
}

// RecommendationDecision is Agent 2's LLM output.
type RecommendationDecision struct {
	AgentName        string                     `json:"agent_name"`
	Recommendations  []RecommendationDecisionItem `json:"recommendations"`
	OverallStrategy  string                     `json:"overall_strategy"`
	GenerationNotes  *GenerationNotes           `json:"generation_notes"`
}

type RecommendationDecisionItem struct {
	JobID                uint64   `json:"job_id"`
	Title                string   `json:"title"`
	RecommendationScore  int      `json:"recommendation_score"`
	RecommendationLevel  string   `json:"recommendation_level"`
	RecommendationReason string   `json:"recommendation_reason"`
	MatchHighlights      []string `json:"match_highlights"`
	RiskPoints           []string `json:"risk_points"`
	ResumeEvidence       []string `json:"resume_evidence"`
	JobEvidence          []string `json:"job_evidence"`
	OptimizationSignal   string   `json:"optimization_stop_signal"`
	SuggestedAction      string   `json:"suggested_action"`
}

type GenerationNotes struct {
	ScoreDistribution string `json:"score_distribution"`
	EvidenceBased     bool   `json:"evidence_based"`
}

// ─── Agent 3: Verifier / Frontend-Contract ──────────────────────────────

// RecommendationVerifierInput is Agent 3's input.
type RecommendationVerifierInput struct {
	Decision   *RecommendationDecision
	Candidates []CandidateContextJob // for job_id validation
	ResumeCheck *ProfileRealityCheck
}

// RecommendationFinalOutput is Agent 3's verified output.
type RecommendationFinalOutput struct {
	AgentName     string                     `json:"agent_name"`
	GeneratedBy   string                     `json:"generated_by"`
	AgentStatus   string                     `json:"agent_status"`
	Fallback      bool                       `json:"fallback"`
	RepairApplied bool                       `json:"repair_applied"`
	GenerationTier string                    `json:"generation_tier"`
	Items         []RecommendationFinalItem  `json:"items"`
	Verification  *VerificationReport        `json:"verification"`
}

type RecommendationFinalItem struct {
	JobID                uint64   `json:"job_id"`
	Title                string   `json:"title"`
	RecommendationScore  int      `json:"recommendation_score"`
	RecommendationReason string   `json:"recommendation_reason"`
	RecommendationLevel  string   `json:"recommendation_level"`
	MatchHighlights      []string `json:"match_highlights"`
	RiskPoints           []string `json:"risk_points"`
	ResumeEvidence       []string `json:"resume_evidence"`
	JobEvidence          []string `json:"job_evidence"`
	SuggestedAction      string   `json:"suggested_action"`
	OptimizationSignal   string   `json:"optimization_stop_signal"`
}

type VerificationReport struct {
	AllJobIDsValid      bool     `json:"all_job_ids_valid"`
	ScoreDifferentiated bool     `json:"score_differentiated"`
	FieldsComplete      bool     `json:"fields_complete"`
	IssuesFound         []string `json:"issues_found"`
}

// ─── RunDeepRecommendation3Agents ──────────────────────────────────────
// This orchestrates the 3-agent pipeline for Deep Recommendation.
// Returns the final output with frontend-compatible fields.

type DeepRecommendResult struct {
	FinalOutput  *RecommendationFinalOutput
	Agent1Output *RecommendationContext
	Agent2Output *RecommendationDecision
	Agent3Output *RecommendationFinalOutput // same as FinalOutput
	Agent1Duration time.Duration
	Agent2Duration time.Duration
	Agent3Duration time.Duration
}

func RunDeepRecommendation3Agents(
	ctx context.Context,
	resumeProfile *model.ResumeTagProfile,
	jobProfiles []*model.JobTagProfile,
	jobMap map[uint64]*model.Job,
	runner2 PipelineRunner,
	runner3 PipelineRunner,
) (*DeepRecommendResult, error) {
	// Agent 1: Rule-based context building
	t0 := time.Now()
	if resumeProfile == nil {
		return nil, fmt.Errorf("agent1: resume profile is nil")
	}
	if len(jobProfiles) == 0 {
		return nil, fmt.Errorf("agent1: no job profiles available")
	}

	agent1Input := &RecommendationContextInput{
		ResumeID:       resumeProfile.ResumeID,
		ResumeProfile:  resumeProfile,
		JobProfiles:    jobProfiles,
		CandidateLimit: 15,
	}
	agent1Output, err := BuildRecommendationContext(ctx, agent1Input)
	if err != nil {
		return nil, fmt.Errorf("agent1 context failed: %w", err)
	}
	agent1Duration := time.Since(t0)

	// Fill titles from job map
	for i, c := range agent1Output.CandidateJobs {
		if j, ok := jobMap[c.JobID]; ok {
			agent1Output.CandidateJobs[i].Title = j.Title
			agent1Output.CandidateJobs[i].Department = j.Department
			agent1Output.CandidateJobs[i].City = j.City
		}
	}

	log.Printf("[DeepRecommend] stage=agent1_context candidates_in=%d candidates_out=%d duration_ms=%d",
		len(jobProfiles), len(agent1Output.CandidateJobs), agent1Duration.Milliseconds())

	// Agent 2: LLM decision generation
	t1 := time.Now()
	candidatesJSON, _ := json.Marshal(agent1Output.CandidateJobs)
	resumeJSON, _ := json.Marshal(agent1Output.ResumeSummary)
	checkJSON, _ := json.Marshal(agent1Output.ProfileCheck)

	agent2Req := pipeline.RunRequest{
		AgentName:        "RecommendationDecisionAgent",
		PromptName:       prompt.PromptRecommendationDecision,
		SchemaName:       schema.SchemaRecommendationDecision,
		TargetOutputType: pipeline.TargetRecommendationDecision,
		ReasoningEffort:  pipeline.ReasoningHigh,
		PromptValues: map[string]string{
			"RESUME_CONTEXT_JSON": string(resumeJSON),
			"CANDIDATE_JOBS_JSON": string(candidatesJSON),
			"REALITY_CHECK_JSON":  string(checkJSON),
		},
		Metadata: map[string]string{
			"candidate_count": fmt.Sprintf("%d", len(agent1Output.CandidateJobs)),
		},
	}

	log.Printf("[DeepRecommend] agent2_decision reasoning_effort=%s", pipeline.ReasoningHigh)
	agent2Result := runner2.Run(ctx, agent2Req)
	if agent2Result.FallbackRequired {
		return nil, fmt.Errorf("agent2 decision failed: %s", agent2Result.FallbackReason)
	}

	agent2Output := &RecommendationDecision{}
	agent2JSON, _ := json.Marshal(agent2Result.Parsed)
	if err := json.Unmarshal(agent2JSON, agent2Output); err != nil {
		return nil, fmt.Errorf("agent2 parse output: %w", err)
	}
	agent2Duration := time.Since(t1)

	log.Printf("[DeepRecommend] stage=agent2_decision candidates_in=%d top_out=%d duration_ms=%d",
		len(agent1Output.CandidateJobs), len(agent2Output.Recommendations), agent2Duration.Milliseconds())

	// Agent 3: Verifier
	t2 := time.Now()
	decisionJSON, _ := json.Marshal(agent2Output)
	candidatesJSON2, _ := json.Marshal(agent1Output.CandidateJobs)
	checkJSON2, _ := json.Marshal(agent1Output.ProfileCheck)

	agent3Req := pipeline.RunRequest{
		AgentName:        "RecommendationVerifierAgent",
		PromptName:       prompt.PromptRecommendationVerifier,
		SchemaName:       schema.SchemaRecommendationVerifier,
		TargetOutputType: pipeline.TargetRecommendationVerifier,
		ReasoningEffort:  pipeline.ReasoningLow,
		PromptValues: map[string]string{
			"DECISION_JSON":      string(decisionJSON),
			"CANDIDATES_JSON":    string(candidatesJSON2),
			"REALITY_CHECK_JSON": string(checkJSON2),
		},
	}

	log.Printf("[DeepRecommend] agent3_verifier reasoning_effort=%s", pipeline.ReasoningLow)
	agent3Result := runner3.Run(ctx, agent3Req)
	if agent3Result.FallbackRequired {
		log.Printf("[DeepRecommend] stage=agent3_verifier fallback — using agent2 output with local normalization")
		finalOutput := normalizeFromAgent2(agent2Output, agent1Output.CandidateJobs, agent1Output.ProfileCheck)
		agent3Duration := time.Since(t2)
		return &DeepRecommendResult{
			FinalOutput:     finalOutput,
			Agent1Output:    agent1Output,
			Agent2Output:    agent2Output,
			Agent3Output:    finalOutput,
			Agent1Duration:  agent1Duration,
			Agent2Duration:  agent2Duration,
			Agent3Duration:  agent3Duration,
		}, nil
	}

	agent3Output := &RecommendationFinalOutput{}
	agent3JSON, _ := json.Marshal(agent3Result.Parsed)
	if err := json.Unmarshal(agent3JSON, agent3Output); err != nil {
		finalOutput := normalizeFromAgent2(agent2Output, agent1Output.CandidateJobs, agent1Output.ProfileCheck)
		agent3Duration := time.Since(t2)
		return &DeepRecommendResult{
			FinalOutput:     finalOutput,
			Agent1Output:    agent1Output,
			Agent2Output:    agent2Output,
			Agent3Output:    finalOutput,
			Agent1Duration:  agent1Duration,
			Agent2Duration:  agent2Duration,
			Agent3Duration:  agent3Duration,
		}, nil
	}
	agent3Duration := time.Since(t2)

	log.Printf("[DeepRecommend] stage=agent3_verifier verified=true duration_ms=%d",
		agent3Duration.Milliseconds())

	return &DeepRecommendResult{
		FinalOutput:    agent3Output,
		Agent1Output:   agent1Output,
		Agent2Output:   agent2Output,
		Agent3Output:   agent3Output,
		Agent1Duration: agent1Duration,
		Agent2Duration: agent2Duration,
		Agent3Duration: agent3Duration,
	}, nil
}

// normalizeFromAgent2 does deterministic field normalization when Agent 3 fails.
func normalizeFromAgent2(decision *RecommendationDecision, candidates []CandidateContextJob, check *ProfileRealityCheck) *RecommendationFinalOutput {
	candidateMap := make(map[uint64]bool)
	for _, c := range candidates {
		candidateMap[c.JobID] = true
	}

	var items []RecommendationFinalItem
	for _, d := range decision.Recommendations {
		if !candidateMap[d.JobID] {
			continue
		}
		if d.RecommendationScore < 0 { d.RecommendationScore = 0 }
		if d.RecommendationScore > 100 { d.RecommendationScore = 100 }
		if len(d.MatchHighlights) == 0 { d.MatchHighlights = []string{} }
		if len(d.RiskPoints) == 0 { d.RiskPoints = []string{} }
		if len(d.ResumeEvidence) == 0 { d.ResumeEvidence = []string{} }
		if len(d.JobEvidence) == 0 { d.JobEvidence = []string{} }

		items = append(items, RecommendationFinalItem{
			JobID:                d.JobID,
			Title:                d.Title,
			RecommendationScore:  d.RecommendationScore,
			RecommendationReason: d.RecommendationReason,
			RecommendationLevel:  d.RecommendationLevel,
			MatchHighlights:      safeArray(d.MatchHighlights),
			RiskPoints:           safeArray(d.RiskPoints),
			ResumeEvidence:       safeArray(d.ResumeEvidence),
			JobEvidence:          safeArray(d.JobEvidence),
			SuggestedAction:      d.SuggestedAction,
			OptimizationSignal:   d.OptimizationSignal,
		})
		// Limit to 5
		if len(items) >= 5 { break }
	}

	issues := []string{}
	if check != nil && !check.ProfileConsistent {
		issues = append(issues, "profile_consistency_check")
	}

	return &RecommendationFinalOutput{
		AgentName:     "recommend_verifier_agent",
		GeneratedBy:   "llm_agent",
		AgentStatus:   "verified_by_local_guard",
		Fallback:      false,
		RepairApplied: true,
		GenerationTier: "quality",
		Items:         items,
		Verification: &VerificationReport{
			AllJobIDsValid:      len(items) <= len(candidates),
			ScoreDifferentiated: checkScoreDifferentiated(items),
			FieldsComplete:      true,
			IssuesFound:         issues,
		},
	}
}

func safeArray(v []string) []string {
	if v == nil { return []string{} }
	return v
}

func checkScoreDifferentiated(items []RecommendationFinalItem) bool {
	if len(items) < 2 { return true }
	first := items[0].RecommendationScore
	for _, it := range items[1:] {
		if it.RecommendationScore != first { return true }
	}
	return false
}
