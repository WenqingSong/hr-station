// ─── Deep Recommendation Cascaded Pipeline ──────────────────────────────────
//
// Cascaded Deep Recommendation Pipeline
// 级联式深度岗位推荐管线
//
// Layers:
//   0: SemanticProfileGate (pre-existing)
//   1: Candidate Discovery Agent    — Hy3 high ×1
//   2: Semantic Rerank Agent        — Hy3 high ×1
//   3: Per-Job Deep Synthesis       — Hy3 high ×5
//   4: Cross-Job Calibration        — Hy3 low  ×1

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ─── Step Types ─────────────────────────────────────────────────────────────

// DeepCascadeStep records the result of each pipeline stage for trace/debug.
type DeepCascadeStep struct {
	Layer      string `json:"layer"`
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Tokens     int    `json:"tokens,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

// ─── Layer 1-4 Output Types ──────────────────────────────────────────────────

type DeepCandidateDiscoveryOutput struct {
	Candidates []DeepCandidateJob `json:"candidates"`
	Summary    string             `json:"summary"`
}

type DeepCandidateJob struct {
	JobID           int64    `json:"job_id"`
	JobTitle        string   `json:"job_title"`
	CandidateReason string   `json:"candidate_reason"`
	MatchedSignals  []string `json:"matched_signals"`
	RiskSignals     []string `json:"risk_signals"`
	InitialScore    int      `json:"initial_score"`
}

type DeepSemanticRerankOutput struct {
	TopJobs          []DeepRerankedJob `json:"top_jobs"`
	ExcludedJobs     []DeepExcludedJob `json:"excluded_jobs,omitempty"`
	RankingRationale string            `json:"ranking_rationale"`
}

type DeepRerankedJob struct {
	JobID          int64    `json:"job_id"`
	Rank           int      `json:"rank"`
	Score          int      `json:"score"`
	FitLevel       string   `json:"fit_level"`
	WhyThisRank    string   `json:"why_this_rank"`
	CoreEvidence   []string `json:"core_evidence"`
	MainRisks      []string `json:"main_risks"`
	SynthesisFocus []string `json:"synthesis_focus"`
}

type DeepExcludedJob struct {
	JobID  int64  `json:"job_id"`
	Reason string `json:"reason"`
}

type DeepPerJobSynthesisOutput struct {
	JobID                 int64    `json:"job_id"`
	FinalScore            int      `json:"final_score"`
	FitLevel              string   `json:"fit_level"`
	OneSentenceConclusion string   `json:"one_sentence_conclusion"`
	WhyRecommended        []string `json:"why_recommended"`
	ResumeEvidence        []string `json:"resume_evidence"`
	JobRequirementMapping []string `json:"job_requirement_mapping"`
	RiskPoints            []string `json:"risk_points"`
	ApplicationStrategy   string   `json:"application_strategy"`
	ResumeImprovementHint string   `json:"resume_improvement_hint"`
	DisplayTags           []string `json:"display_tags"`
	WhyRankedHere         string   `json:"why_ranked_here"`
}

type DeepCalibrationOutput struct {
	FinalOrder      []DeepCalibratedJob `json:"final_order"`
	GlobalSummary   string              `json:"global_summary"`
	QualityStatus   string              `json:"quality_status"`
	QualityWarnings []string            `json:"quality_warnings,omitempty"`
}

type DeepCalibratedJob struct {
	JobID             int64  `json:"job_id"`
	FinalRank         int    `json:"final_rank"`
	CalibratedScore   int    `json:"calibrated_score"`
	CalibrationReason string `json:"calibration_reason"`
}

// ─── Pipeline Configuration ──────────────────────────────────────────────────

// DeepCascadeConfig holds the LLM client and related services for the pipeline.
type DeepCascadeConfig struct {
	LLMClient    llm.Client
	QualityModel string
	SemanticGate *SemanticProfileGate
	SemanticRepo *repository.SemanticProfileRepository
	ResumeRepo   *repository.ResumeRepository
	JobRepo      *repository.JobRepository
}

// ─── Deep Cascade Runner ────────────────────────────────────────────────────

// RunDeepRecommendationCascade executes the full cascaded pipeline.
func RunDeepRecommendationCascade(ctx context.Context, cfg *DeepCascadeConfig, resumeID int64, jobs []*model.Job, limit int) ([]*model.RecommendedJob, []DeepCascadeStep, string, error) {
	var trace []DeepCascadeStep

	// ── Layer 0: SemanticProfileGate ──
	t0 := time.Now()
	resumeResult, err := cfg.SemanticGate.EnsureResumeProfile(ctx, resumeID)
	if err != nil || resumeResult == nil || resumeResult.Status != "verified" {
		return nil, trace, "", fmt.Errorf("layer0: resume SEP not verified (status=%s err=%v)", gateStatus(resumeResult), err)
	}
	resumeTags, err := cfg.SemanticRepo.ListTagsByProfileID(ctx, resumeResult.ProfileID)
	if err != nil || len(resumeTags) < 3 {
		return nil, trace, "", fmt.Errorf("layer0: resume tags insufficient (%d)", len(resumeTags))
	}

	// Gate job SEPs
	jobSEP := make(map[int64]*gateJobSep)
	skipped := 0
	for _, j := range jobs {
		jid := int64(j.ID)
		jr, err := cfg.SemanticGate.EnsureJobProfile(ctx, jid)
		if err != nil || jr == nil || jr.Status != "verified" {
			skipped++
			continue
		}
		jt, err := cfg.SemanticRepo.ListTagsByProfileID(ctx, jr.ProfileID)
		if err != nil || len(jt) < 3 {
			skipped++
			continue
		}
		jobSEP[jid] = &gateJobSep{
			profileResult: jr,
			tags:          jt,
			title:         j.Title,
			direction:     j.Direction,
			category:      j.Category,
			description:   truncateStr(j.Description, 2000),
			requirements:  truncateStr(j.Requirements, 1500),
		}
	}
	trace = append(trace, DeepCascadeStep{
		Layer: "0_gate", Status: "ok", DurationMs: time.Since(t0).Milliseconds(),
		Detail: fmt.Sprintf("resume=%d tags=%d jobs_ready=%d skipped=%d", resumeID, len(resumeTags), len(jobSEP), skipped),
	})
	if len(jobSEP) < 5 {
		return nil, trace, "", fmt.Errorf("layer0: only %d jobs have SEP profiles (need >=5)", len(jobSEP))
	}
	log.Printf("[DeepCascade] L0: resume=%d tags=%d jobs=%d/%d", resumeID, len(resumeTags), len(jobSEP), len(jobs))

	// ── Layer 1: Candidate Discovery ──
	l1Result, l1Trace, err := runLayer1Discovery(ctx, cfg, resumeID, resumeResult, resumeTags, jobSEP)
	trace = append(trace, l1Trace)
	if err != nil {
		return nil, trace, "", fmt.Errorf("layer1: %w", err)
	}

	// ── Layer 2: Semantic Rerank ──
	l2Result, l2Trace, err := runLayer2Rerank(ctx, cfg, resumeID, resumeResult, resumeTags, l1Result, jobSEP)
	trace = append(trace, l2Trace)
	if err != nil {
		return nil, trace, "", fmt.Errorf("layer2: %w", err)
	}
	top5 := l2Result.TopJobs
	if len(top5) < 5 {
		return nil, trace, "", fmt.Errorf("layer2: only %d TopJobs (need 5)", len(top5))
	}

	// ── Layer 3: Per-Job Deep Synthesis ×5 ──
	var l3Results []*DeepPerJobSynthesisOutput
	for _, tj := range top5 {
		jsep, ok := jobSEP[tj.JobID]
		if !ok {
			log.Printf("[DeepCascade] L3: job %d not in SEP pool, skipping", tj.JobID)
			continue
		}
		l3o, l3t, err := runLayer3Synthesis(ctx, cfg, resumeID, resumeResult, resumeTags, tj, jsep)
		trace = append(trace, l3t)
		if err != nil {
			log.Printf("[DeepCascade] L3: job %d failed: %v", tj.JobID, err)
			continue
		}
		l3Results = append(l3Results, l3o)
	}
	if len(l3Results) < 4 {
		return nil, trace, "", fmt.Errorf("layer3: only %d successful syntheses (need >=4)", len(l3Results))
	}

	// ── Layer 4: Cross-Job Calibration ──
	l4Result, l4Trace, err := runLayer4Calibration(ctx, cfg, resumeID, resumeResult, l2Result, l3Results)
	trace = append(trace, l4Trace)
	if err != nil {
		log.Printf("[DeepCascade] L4 calibration failed, using L3 raw order: %v", err)
	}

	// ── Build Response ──
	jobs0 := buildRecommendedJobsFromCascade(l3Results, l4Result, top5, jobSEP, limit)
	globalSummary := ""
	if l4Result != nil { globalSummary = l4Result.GlobalSummary }
	if globalSummary == "" { globalSummary = resumeResult.Summary }
	return jobs0, trace, globalSummary, nil
}

// ─── Layer 1: Candidate Discovery ──────────────────────────────────────────

type gateJobSep struct {
	profileResult *SemanticProfileGateResult
	tags          []*model.SemanticProfileTag
	title         string
	direction     string
	category      string
	description   string
	requirements  string
}

func runLayer1Discovery(ctx context.Context, cfg *DeepCascadeConfig, resumeID int64, resumeResult *SemanticProfileGateResult, resumeTags []*model.SemanticProfileTag, jobSEP map[int64]*gateJobSep) (*DeepCandidateDiscoveryOutput, DeepCascadeStep, error) {
	sysPrompt, usrPrompt := buildLayer1DiscoveryPrompt(resumeID, resumeResult.Summary, resumeTags, jobSEP)
	t0 := time.Now()
	resp, err := callHy3(ctx, cfg, "deep_cascade_l1", sysPrompt, usrPrompt)
	elapsed := time.Since(t0).Milliseconds()
	if err != nil {
		return nil, DeepCascadeStep{Layer: "1_discovery", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	content := extractJSON(resp.RawContent)
	var out DeepCandidateDiscoveryOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, DeepCascadeStep{Layer: "1_discovery", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("json: %v", err)}, err
	}
	if err := validateLayer1(&out, jobSEP); err != nil {
		return nil, DeepCascadeStep{Layer: "1_discovery", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	log.Printf("[DeepCascade] L1: discovered=%d candidates tokens=%d time=%dms", len(out.Candidates), resp.Usage.TotalTokens, elapsed)
	return &out, DeepCascadeStep{Layer: "1_discovery", Status: "ok", DurationMs: elapsed, Tokens: resp.Usage.TotalTokens}, nil
}

// ─── Layer 2: Semantic Rerank ──────────────────────────────────────────────

func runLayer2Rerank(ctx context.Context, cfg *DeepCascadeConfig, resumeID int64, resumeResult *SemanticProfileGateResult, resumeTags []*model.SemanticProfileTag, l1 *DeepCandidateDiscoveryOutput, jobSEP map[int64]*gateJobSep) (*DeepSemanticRerankOutput, DeepCascadeStep, error) {
	sysPrompt, usrPrompt := buildLayer2RerankPrompt(resumeID, resumeResult.Summary, resumeTags, l1, jobSEP)
	t0 := time.Now()
	resp, err := callHy3(ctx, cfg, "deep_cascade_l2", sysPrompt, usrPrompt)
	elapsed := time.Since(t0).Milliseconds()
	if err != nil {
		return nil, DeepCascadeStep{Layer: "2_rerank", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	content := extractJSON(resp.RawContent)
	var out DeepSemanticRerankOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, DeepCascadeStep{Layer: "2_rerank", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("json: %v", err)}, err
	}
	if err := validateLayer2(&out, l1); err != nil {
		return nil, DeepCascadeStep{Layer: "2_rerank", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	log.Printf("[DeepCascade] L2: top5=%v tokens=%d time=%dms", top5JobIDs(out.TopJobs), resp.Usage.TotalTokens, elapsed)
	return &out, DeepCascadeStep{Layer: "2_rerank", Status: "ok", DurationMs: elapsed, Tokens: resp.Usage.TotalTokens}, nil
}

// ─── Layer 3: Per-Job Deep Synthesis ────────────────────────────────────────

func runLayer3Synthesis(ctx context.Context, cfg *DeepCascadeConfig, resumeID int64, resumeResult *SemanticProfileGateResult, resumeTags []*model.SemanticProfileTag, tj DeepRerankedJob, jsep *gateJobSep) (*DeepPerJobSynthesisOutput, DeepCascadeStep, error) {
	sysPrompt, usrPrompt := buildLayer3SynthesisPrompt(resumeID, resumeResult.Summary, resumeTags, tj, jsep)
	t0 := time.Now()
	resp, err := callHy3(ctx, cfg, "deep_cascade_l3", sysPrompt, usrPrompt)
	elapsed := time.Since(t0).Milliseconds()
	if err != nil {
		return nil, DeepCascadeStep{Layer: "3_synthesis", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("job=%d: %v", tj.JobID, err)}, err
	}
	content := extractJSON(resp.RawContent)
	var out DeepPerJobSynthesisOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, DeepCascadeStep{Layer: "3_synthesis", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("job=%d json: %v", tj.JobID, err)}, err
	}
	out.JobID = tj.JobID
	if err := validateLayer3(&out); err != nil {
		return nil, DeepCascadeStep{Layer: "3_synthesis", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("job=%d: %v", tj.JobID, err)}, err
	}
	log.Printf("[DeepCascade] L3: job=%d score=%d tokens=%d time=%dms", tj.JobID, out.FinalScore, resp.Usage.TotalTokens, elapsed)
	return &out, DeepCascadeStep{Layer: "3_synthesis", Status: "ok", DurationMs: elapsed, Tokens: resp.Usage.TotalTokens, Detail: fmt.Sprintf("job=%d", tj.JobID)}, nil
}

// ─── Layer 4: Cross-Job Calibration ────────────────────────────────────────

func runLayer4Calibration(ctx context.Context, cfg *DeepCascadeConfig, resumeID int64, resumeResult *SemanticProfileGateResult, l2 *DeepSemanticRerankOutput, l3Results []*DeepPerJobSynthesisOutput) (*DeepCalibrationOutput, DeepCascadeStep, error) {
	sysPrompt, usrPrompt := buildLayer4CalibrationPrompt(resumeID, resumeResult.Summary, l2, l3Results)
	t0 := time.Now()
	resp, err := callHy3Low(ctx, cfg, "deep_cascade_l4", sysPrompt, usrPrompt)
	elapsed := time.Since(t0).Milliseconds()
	if err != nil {
		return nil, DeepCascadeStep{Layer: "4_calibration", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	content := extractJSON(resp.RawContent)
	var out DeepCalibrationOutput
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, DeepCascadeStep{Layer: "4_calibration", Status: "failed", DurationMs: elapsed, Detail: fmt.Sprintf("json: %v", err)}, err
	}
	if err := validateLayer4(&out, l3Results); err != nil {
		return nil, DeepCascadeStep{Layer: "4_calibration", Status: "failed", DurationMs: elapsed, Detail: err.Error()}, err
	}
	log.Printf("[DeepCascade] L4: calibration status=%s tokens=%d time=%dms", out.QualityStatus, resp.Usage.TotalTokens, elapsed)
	return &out, DeepCascadeStep{Layer: "4_calibration", Status: "ok", DurationMs: elapsed, Tokens: resp.Usage.TotalTokens}, nil
}

// ─── LLM Helpers ────────────────────────────────────────────────────────────

func callHy3(ctx context.Context, cfg *DeepCascadeConfig, label, sys, usr string) (*llm.ChatResponse, error) {
	return cfg.LLMClient.Chat(ctx, llm.ChatRequest{
		Model:           cfg.QualityModel,
		Messages:        []llm.Message{{Role: llm.RoleSystem, Content: sys}, {Role: llm.RoleUser, Content: usr}},
		Temperature:     0.3,
		MaxTokens:       4096,
		ReasoningEffort: "high",
	})
}

func callHy3Low(ctx context.Context, cfg *DeepCascadeConfig, label, sys, usr string) (*llm.ChatResponse, error) {
	return cfg.LLMClient.Chat(ctx, llm.ChatRequest{
		Model:           cfg.QualityModel,
		Messages:        []llm.Message{{Role: llm.RoleSystem, Content: sys}, {Role: llm.RoleUser, Content: usr}},
		Temperature:     0.3,
		MaxTokens:       4096,
		ReasoningEffort: "low",
	})
}

// ─── JSON Extraction ────────────────────────────────────────────────────────

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if idx := strings.Index(raw, "```json"); idx >= 0 {
		raw = raw[idx+7:]
		if j := strings.LastIndex(raw, "```"); j >= 0 { raw = raw[:j] }
	} else if idx := strings.Index(raw, "```"); idx >= 0 {
		raw = raw[idx+3:]
		if j := strings.LastIndex(raw, "```"); j >= 0 { raw = raw[:j] }
	}
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		if idx := strings.Index(raw, "{"); idx >= 0 { raw = raw[idx:] }
	}
	if !strings.HasSuffix(raw, "}") {
		if idx := strings.LastIndex(raw, "}"); idx >= 0 { raw = raw[:idx+1] }
	}
	return raw
}

// ─── Build Final RecommendedJob ─────────────────────────────────────────────

func buildRecommendedJobsFromCascade(l3Results []*DeepPerJobSynthesisOutput, l4Result *DeepCalibrationOutput, top5 []DeepRerankedJob, jobSEP map[int64]*gateJobSep, limit int) []*model.RecommendedJob {
	// Build calibration lookup
	calLookup := make(map[int64]*DeepCalibratedJob)
	if l4Result != nil {
		for i := range l4Result.FinalOrder {
			c := &l4Result.FinalOrder[i]
			calLookup[c.JobID] = c
		}
	}
	// Build L3 lookup
	l3Lookup := make(map[int64]*DeepPerJobSynthesisOutput)
	for _, l3 := range l3Results {
		l3Lookup[l3.JobID] = l3
	}

	var result []*model.RecommendedJob
	for i, l3 := range l3Results {
		if i >= limit { break }
		jsep := jobSEP[l3.JobID]
		title, dept, city, direction, category := "", "", "", "", ""
		if jsep != nil {
			title = jsep.title
			direction = jsep.direction
			category = jsep.category
		}
		score := l3.FinalScore
		whyRanked := fmt.Sprintf("深度分析排名第%d位：%s", i+1, l3.OneSentenceConclusion)
		if cal, ok := calLookup[l3.JobID]; ok {
			score = cal.CalibratedScore
			whyRanked = cal.CalibrationReason
		}

		rec := &model.RecommendedJob{
			JobID:                uint64(l3.JobID),
			Title:                title,
			Department:           dept,
			City:                 city,
			Category:             category,
			JobType:              direction,
			RecommendationScore:  score,
			RecommendationReason: l3.OneSentenceConclusion,
			RecommendationLabel:  l3.FitLevel,
			RecommendationPriority: cascadeFitToPriority(l3.FitLevel),
			MatchHighlights:      l3.WhyRecommended,
			RiskPoints:           l3.RiskPoints,
			SuggestedAction:      l3.ApplicationStrategy,
			WhyRankedHere:        whyRanked,
			JobSpecificReason:    l3.OneSentenceConclusion,
			JobSpecificRisks:     l3.RiskPoints,
			ResumeEvidence:       l3.ResumeEvidence,
			GeneratedBy:          "llm_agent",
			AgentStatus:          "cascade_verified",
			FallbackReason:       "none",
			RepairApplied:        false,
			DataQuality:          "llm",
		}
		result = append(result, rec)
	}
	return result
}

func cascadeFitToPriority(fit string) string {
	switch fit {
	case "优先投递": return "strong"
	case "重点关注", "建议投递": return "high"
	case "可以尝试", "可以考虑", "可投递", "中等匹配": return "try"
	case "谨慎投递", "一般匹配": return "backup"
	default: return "backup"
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func gateStatus(r *SemanticProfileGateResult) string {
	if r == nil { return "nil" }
	return r.Status
}

func top5JobIDs(jobs []DeepRerankedJob) string {
	var s []string
	for _, j := range jobs {
		s = append(s, fmt.Sprintf("%d", j.JobID))
	}
	return strings.Join(s, ",")
}
