package service

import (
	"context"
	"encoding/json"
	apperror "offer-catcher/server/internal/errors"
	"fmt"
	"log"
	"math"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// MatchRequest represents a request for job match analysis.
type MatchRequest struct {
	ResumeID uint64 `json:"resume_id"`
	GenerationTier  model.GenerationTier `json:"generation_tier,omitempty"`
}

// MatchResponse represents the response for job match analysis.
type MatchResponse struct {
	AnalysisID    uint64                     `json:"analysis_id"`
	GeneratedBy   string                     `json:"generated_by"` // deprecated: use meta.generated_by
	Result        *model.MatchAnalysisResult `json:"result"`
	AnalysisBasis *model.AnalysisBasis       `json:"analysis_basis"`
	CreatedAt     string                     `json:"created_at"`
	Meta          model.AgentResponseMeta    `json:"meta"`
}

// MatchService handles job match analysis business logic.
// K13-Fix-B: Fast tier uses TagEvidenceService (no LLM); Deep tier uses LLM pipeline.
// H18-C: Fast tier prefers SEP profile context.
type MatchService struct {
	resumeRepo    *repository.ResumeRepository
	jobRepo       *repository.JobRepository
	matchRepo     *repository.MatchRepository
	mockAgent     *agent.MatchMockAgent
	fastRunner    agent.PipelineRunner
	qualityRunner agent.PipelineRunner
	llmEnabled    bool
	fallbackEnable bool
	enableMockAgent bool
	tagEvidence    *TagEvidenceService // K13-Fix-B
	semanticRepo   *repository.SemanticProfileRepository // H18-C: SEP profile access
}

// NewMatchService creates a new MatchService.
func NewMatchService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	matchRepo *repository.MatchRepository,
	mockAgent *agent.MatchMockAgent,
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
	fallbackEnable bool,
	enableMockAgent bool,
	tagEvidence *TagEvidenceService,
	semanticRepo *repository.SemanticProfileRepository,
) *MatchService {
	return &MatchService{
		resumeRepo:    resumeRepo,
		jobRepo:       jobRepo,
		matchRepo:     matchRepo,
		mockAgent:     mockAgent,
		fastRunner:    fastRunner,
		qualityRunner: qualityRunner,
		llmEnabled:    llmEnabled,
		fallbackEnable: fallbackEnable,
		enableMockAgent: enableMockAgent,
		tagEvidence:    tagEvidence,
		semanticRepo:   semanticRepo,
	}
}

// selectRunner returns the appropriate pipeline runner for the given tier.
func (s *MatchService) selectRunner(tier model.GenerationTier) agent.PipelineRunner {
	// tier validated by handler
	if tier == model.GenerationTierQuality && s.qualityRunner != nil {
		return s.qualityRunner
	}
	return s.fastRunner
}

// Analyze generates a match analysis for the given user and job.
// When LLM is enabled and Pipeline Runner is available, it uses the LLM Pipeline
// first. If the Pipeline returns fallback_required or mapping fails, it falls
// back to the Mock Agent. When LLM is disabled, it uses the Mock Agent directly.
func (s *MatchService) Analyze(ctx context.Context, userID uint64, jobID uint64, req MatchRequest) (*MatchResponse, error) {
	log.Printf("[MatchService] generation_tier=%s", string(req.GenerationTier))

	genTier := string(req.GenerationTier)
	if genTier == "" {
		genTier = "fast"
	}

	// Get resume version
	resume, err := s.getResume(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Get job
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, apperror.ErrJobNotFound
	}

	// H18-C: Fast tier — try SEP profile first, fallback to tag engine
	isFast := genTier == "fast" || req.GenerationTier != model.GenerationTierQuality
	if isFast && s.tagEvidence != nil {
		// Try SEP-based match first
		if s.semanticRepo != nil {
			sepResp, sepErr := s.fastMatchWithSEP(ctx, userID, resume, job)
			if sepErr == nil {
				return sepResp, nil
			}
			log.Printf("[MatchService] SEP fast match unavailable: %v, falling back to tag engine", sepErr)
		}
		// Fallback to tag engine
		log.Printf("[MatchService] use_fast_tag_fallback tier=%s resume=%d job=%d", genTier, req.ResumeID, jobID)
		resp, err := s.fastMatchWithTags(ctx, userID, resume, job)
		if err == nil {
			// H18-C: sanitize fallback output
			sanitizeFallbackResult(resp)
			return resp, nil
		}
		log.Printf("[MatchService] fast tag match failed: %v, falling back to mock", err)
	}

	// H19-B: Deep/Quality tier — new SEP-based Deep Match Agent
	runner := s.selectRunner(req.GenerationTier)
	isQuality := genTier == "quality" || req.GenerationTier == model.GenerationTierQuality
	if isQuality && s.llmEnabled && runner != nil {
		// Try H19-B Deep Match SEP Agent
		log.Printf("[MatchService] use_deep_sep_agent tier=%s resume=%d job=%d", genTier, req.ResumeID, jobID)
		resp, errDeep := s.runDeepMatchSEP(ctx, userID, resume, job, runner)
		if errDeep == nil {
			return resp, nil
		}
		log.Printf("[MatchService] deep SEP agent failed: %v, trying 3-agent fallback", errDeep)
		// Fallback to old 3-agent
		if s.tagEvidence != nil {
			resp, err3a := s.matchWith3Agents(ctx, userID, resume, job, runner)
			if err3a == nil { return resp, nil }
			log.Printf("[MatchService] 3-agent match failed: %v", err3a)
		}
		// Fallback to single-agent LLM
		log.Printf("[MatchService] use_deep_llm tier=%s", genTier)
		resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
		if fallbackErr == nil { return resp, nil }
		log.Printf("[MatchService] LLM pipeline fallback: %v", fallbackErr)
		if !s.fallbackEnable {
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed and fallback disabled: %s", fallbackErr.FallbackReason), 500)
		}
		return s.analyzeWithMock(ctx, userID, resume, job, fallbackErr, genTier)
	}

	// Fallback
	if !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "no generation path available", 500)
	}
	log.Printf("[MatchService] use_mock_fallback tier=%s", genTier)
	return s.analyzeWithMock(ctx, userID, resume, job, nil, genTier)
}

// getResume retrieves the resume version based on the request.
func (s *MatchService) getResume(ctx context.Context, userID uint64, req MatchRequest) (*model.ResumeVersion, error) {
	var resume *model.ResumeVersion
	var err error

	if req.ResumeID > 0 {
		resume, err = s.resumeRepo.FindByUserIDAndID(ctx, userID, req.ResumeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resume version: %w", err)
		}
		if resume == nil {
			return nil, apperror.ErrResumeNotFound
		}
	} else {
		resume, err = s.resumeRepo.GetCurrentByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current resume: %w", err)
		}
		if resume == nil {
			return nil, apperror.ErrResumeNotAvailable
		}
	}
	return resume, nil
}

// analyzeWithLLM calls the MatchAnalysisAgent (LLM Pipeline) and builds the response.
func (s *MatchService) analyzeWithLLM(ctx context.Context, resume *model.ResumeVersion, job *model.Job, runner agent.PipelineRunner) (*MatchResponse, *agent.MatchFallbackError) {
	llmAgent := agent.NewMatchAnalysisAgent(runner)
	result, err := llmAgent.Analyze(ctx, resume, job)
	if err != nil {
		if fbErr, ok := err.(*agent.MatchFallbackError); ok {
			return nil, fbErr
		}
		// Unexpected error — wrap as fallback
		return nil, &agent.MatchFallbackError{
			FallbackReason: "llm_agent_error",
			ErrorCode:      "LLM_AGENT_ERROR",
			ErrorMessage:   err.Error(),
		}
	}

	now := time.Now()
	resp := &MatchResponse{
		GeneratedBy: result.GeneratedBy,
		Result:      result,
		AnalysisBasis: &model.AnalysisBasis{
			JobDescriptionUsed:  job.Description != "",
			JobRequirementsUsed: job.Requirements != "",
			DataQuality:         job.DataQuality,
		},
		CreatedAt: now.Format("2006-01-02T15:04:05Z07:00"),
		Meta: model.DeriveMetaFromResult(
			model.AgentNameMatchAnalysis,
			result.GeneratedBy,
			result.AgentStatus,
			result.FallbackReason,
			result.ErrorCode,
			result.RepairApplied,
		),
	}

	// Save match analysis snapshot (same as mock path)
	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr == nil {
		analysis := &model.MatchAnalysis{
			UserID:                 resume.UserID,
			ResumeID:               resume.ID,
			JobID:                  job.ID,
			ProfileType:            result.ProfileType,
			ProfileDominantProfile: result.ProfileDominantProfile,
			Score:                  result.MatchScore,
			ResultJSON:             resultJSON,
		}
		if createErr := s.matchRepo.Create(ctx, analysis); createErr != nil {
			// Log but don't fail the response if snapshot save fails
			_ = createErr
		}
		resp.AnalysisID = analysis.ID
	}

	return resp, nil
}

// analyzeWithMock calls the Mock Agent and builds the response.
// K9-2-Regression-B: Dispatches to AnalyzeFast or AnalyzeQuality based on genTier.
func (s *MatchService) analyzeWithMock(
	ctx context.Context,
	userID uint64,
	resume *model.ResumeVersion,
	job *model.Job,
	fbErr *agent.MatchFallbackError,
	genTier string,
) (*MatchResponse, error) {
	var result *model.MatchAnalysisResult
	var err error

	if genTier == "quality" {
		result, err = s.mockAgent.AnalyzeQuality(ctx, resume, job)
	} else {
		result, err = s.mockAgent.AnalyzeFast(ctx, resume, job)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
	}

	genBy := "mock_agent"
	if genTier == "quality" {
		genBy = "mock_quality"
	}
	result.GeneratedBy = genBy
	var meta model.AgentResponseMeta
	if fbErr != nil {
		result.AgentStatus = "fallback"
		result.FallbackReason = fbErr.FallbackReason
		result.ErrorCode = fbErr.ErrorCode
		result.RepairApplied = false
		meta = model.DeriveMetaFromResult(model.AgentNameMatchAnalysis, "mock_agent", "fallback", fbErr.FallbackReason, fbErr.ErrorCode, false)
	} else {
		result.AgentStatus = model.AgentStatusMockSuccess
		meta = model.NewAgentResponseMeta(model.AgentNameMatchAnalysis)
	}

	now := time.Now()
	resp := &MatchResponse{
		GeneratedBy: genBy,
		Result:      result,
		AnalysisBasis: &model.AnalysisBasis{JobDescriptionUsed: job.Description != "", JobRequirementsUsed: job.Requirements != "", DataQuality: job.DataQuality},
		CreatedAt:    now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:         meta,
	}

	// Save with full metadata
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal match result: %w", err)
	}

	analysis := &model.MatchAnalysis{
		UserID:                 userID,
		ResumeID:               resume.ID,
		JobID:                  job.ID,
		GenerationTier:         genTier,
		GeneratedBy:            genBy,
		Status:                 "ready",
		ProfileType:            result.ProfileType,
		ProfileDominantProfile: result.ProfileDominantProfile,
		Score:                  result.MatchScore,
		ResultJSON:             resultJSON,
	}

	if err := s.matchRepo.Create(ctx, analysis); err != nil {
		_ = err
	}

	resp.AnalysisID = analysis.ID

	return resp, nil
}

// GetLatestMatchAnalysis returns the latest ready match analysis for the given user+resume+job.
func (s *MatchService) GetLatestMatchAnalysis(ctx context.Context, userID, resumeID, jobID uint64) (*model.MatchLatestResult, error) {
	if resumeID <= 0 {
		return nil, apperror.ErrInvalidParam
	}
	if jobID <= 0 {
		return nil, apperror.ErrInvalidParam
	}
	// Verify resume belongs to user
	resume, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify resume: %w", err)
	}
	if resume == nil {
		return nil, apperror.ErrResumeNotFound
	}
	// Verify job exists
	if _, err := s.jobRepo.FindByID(ctx, jobID); err != nil {
		return nil, apperror.ErrJobNotFound
	}

	latest, err := s.matchRepo.FindLatestByResumeAndJob(ctx, userID, resumeID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest match analysis: %w", err)
	}

	if latest == nil {
		return &model.MatchLatestResult{HasResult: false}, nil
	}

	var resultData interface{}
	if len(latest.ResultJSON) > 0 {
		if err := json.Unmarshal(latest.ResultJSON, &resultData); err != nil {
			return nil, fmt.Errorf("failed to parse saved match result: %w", err)
		}
	}

	return &model.MatchLatestResult{
		HasResult:             true,
		AnalysisID:            latest.ID,
		ResumeID:              latest.ResumeID,
		JobID:                 latest.JobID,
		GenerationTier:        latest.GenerationTier,
		GeneratedBy:           latest.GeneratedBy,
		Status:                latest.Status,
		ProfileType:           latest.ProfileType,
		ProfileDominantProfile: latest.ProfileDominantProfile,
		CreatedAt:             latest.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Result:                resultData,
	}, nil
}

// ─── H18-C: Fast Match via SEP Profile ────────────────────────────────────

// fastMatchWithSEP generates Fast-tier match using SEP profiles.
// Reads resume + job semantic_profiles. Falls back to tag engine if profiles missing.
func (s *MatchService) fastMatchWithSEP(ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job) (*MatchResponse, error) {
	resumeID := int64(resume.ID)
	jobID := int64(job.ID)

	// Read resume SEP profile
	rp, err := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, resumeID)
	if err != nil || rp == nil || rp.Status != "verified" {
		return nil, fmt.Errorf("resume SEP not ready: resume=%d", resumeID)
	}
	// Read job SEP profile
	jp, err := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, jobID)
	if err != nil || jp == nil || jp.Status != "verified" {
		return nil, fmt.Errorf("job SEP not ready: job=%d", jobID)
	}

	// Read SEP tags
	rtags, err := s.semanticRepo.ListTagsByProfileID(ctx, rp.ID)
	if err != nil || len(rtags) < 3 {
		return nil, fmt.Errorf("resume SEP tags insufficient: resume=%d count=%d", resumeID, len(rtags))
	}
	jtgs, err := s.semanticRepo.ListTagsByProfileID(ctx, jp.ID)
	if err != nil || len(jtgs) < 3 {
		return nil, fmt.Errorf("job SEP tags insufficient: job=%d count=%d", jobID, len(jtgs))
	}

	// H18-D: Full SEP Evidence-Grounded Fast Match Engine
	sepCtx := &MatchSEPContext{
		ResumeProfile: rp,
		JobProfile:    jp,
		ResumeTags:    rtags,
		JobTags:       jtgs,
		JobTitle:      job.Title,
		ResumeID:      resumeID,
		JobID:         jobID,
	}
	result := runFastSEPMatch(sepCtx)

	meta := model.AgentResponseMeta{
		AgentName:      "FastSEPMatchEngine",
		GeneratedBy:    result.GeneratedBy,
		AgentStatus:    result.AgentStatus,
		FallbackReason: result.FallbackReason,
		RepairApplied:  result.RepairApplied,
		SchemaValid:    true,
	}

	return &MatchResponse{
		AnalysisID:  0,
		GeneratedBy: result.GeneratedBy,
		Result:      result,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
	}, nil
}

// ─── K13-Fix-B: Fast Match via TagEvidenceService ─────────────────────────

// fastMatchWithTags generates Fast-tier match analysis using TagEvidenceService.
// No LLM call — compares resume tags + job tags + evidence.
func (s *MatchService) fastMatchWithTags(ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job) (*MatchResponse, error) {
	if s.tagEvidence == nil {
		return nil, fmt.Errorf("tag evidence service not available")
	}

	rp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resume.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("fast match: %w", err)
	}
	jp, err := s.tagEvidence.EnsureJobTagProfile(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("fast match: %w", err)
	}

	matched, missing, tagScore := computeTagMatch(rp, jp)
	score := int(math.Min(98, 60 + tagScore*0.8)) // scale count-based score to 60-98
	level := fastMatchLevel(score)

	evidenceChain := []model.MatchEvidence{}
	for _, m := range matched[:min(3, len(matched))] {
		evidenceChain = append(evidenceChain, model.MatchEvidence{Requirement: m, ResumePoint: "已匹配", Judgement: "match"})
	}
	for _, m := range missing[:min(3, len(missing))] {
		evidenceChain = append(evidenceChain, model.MatchEvidence{Requirement: m, ResumePoint: "未匹配", Judgement: "gap"})
	}

	result := &model.MatchAnalysisResult{
		MatchScore:              score,
		MatchLevel:              level,
		MatchLabel:              level,
		ScreeningResult:         "pass",
		EvidenceChain:           evidenceChain,
		MatchHighlights:         matched,
		GapPoints:               missing,
		RiskPoints:              missing,
		OptimizationSuggestions: []model.MatchSuggestion{{Priority: "P0", Title: "查漏补缺", Reason: "标签匹配发现缺口", Action: "针对缺失标签补充对应经历或技能描述"}},
		NextSteps:               []string{"针对缺失标签补充对应经历或技能描述"},
		GeneratedBy:             "fast_tag_fallback_engine",
		AgentStatus:             "success",
		FallbackReason:          "sep_not_ready",
		RepairApplied:           false,
		DataQuality:             "rule",
		Summary:                 level,
	}

	meta := model.AgentResponseMeta{
		AgentName:      "FastMatchEngine",
		GeneratedBy:    "fast_tag_fallback_engine",
		AgentStatus:    "success",
		FallbackReason: "sep_not_ready",
		RepairApplied:  false,
		SchemaValid:    true,
	}

	return &MatchResponse{
		AnalysisID:  0,
		GeneratedBy: "fast_tag_fallback_engine",
		Result:      result,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
	}, nil
}

// H18-C: sanitizeFallbackResult ensures fallback output has no null arrays or raw tag codes.
func sanitizeFallbackResult(resp *MatchResponse) {
	if resp == nil || resp.Result == nil {
		return
	}
	r := resp.Result
	if r.MatchHighlights == nil { r.MatchHighlights = []string{} }
	if r.GapPoints == nil { r.GapPoints = []string{} }
	if r.RiskPoints == nil { r.RiskPoints = []string{} }
	if r.EvidenceChain == nil { r.EvidenceChain = []model.MatchEvidence{} }
	if r.NextSteps == nil { r.NextSteps = []string{} }
	if r.OptimizationSuggestions == nil { r.OptimizationSuggestions = []model.MatchSuggestion{} }
}

// ─── K13-Fix-C2: Deep Match with 3 Agents ───────────────────────────────

// matchWith3Agents runs the 3-agent deep match pipeline.
func (s *MatchService) matchWith3Agents(ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job, runner agent.PipelineRunner) (*MatchResponse, error) {
	if s.tagEvidence == nil {
		return nil, fmt.Errorf("3agent match: tag evidence service not available")
	}

	rp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resume.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("3agent match: resume profile: %w", err)
	}
	jp, err := s.tagEvidence.EnsureJobTagProfile(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("3agent match: job profile: %w", err)
	}

	final, trace, err := agent.RunDeepMatch3Agents(ctx, rp, jp, resume.Name, job.Title, runner, runner)
	if err != nil {
		return nil, fmt.Errorf("3agent match: %w", err)
	}

	result := &model.MatchAnalysisResult{
		MatchScore:              final.Score,
		MatchLevel:              final.Level,
		MatchLabel:              final.Level,
		ScreeningResult:         "pass",
		MatchHighlights:         extractTitles(final.MatchedPoints),
		GapPoints:               extractTitles(final.GapPoints),
		RiskPoints:              extractRiskTitles(final.RiskPoints),
		Summary:                 final.Summary,
		GeneratedBy:             final.GeneratedBy,
		AgentStatus:             final.AgentStatus,
		RepairApplied:           final.RepairApplied,
		FallbackReason:          final.FallbackReason,
		DataQuality:             "llm",
		OptimizationSuggestions: convertImprovements(final.ImprovementSuggestions),
		NextSteps:               extractActions(final.ImprovementSuggestions),
	}

	meta := model.AgentResponseMeta{
		AgentName:      "DeepMatchEngine",
		GeneratedBy:    final.GeneratedBy,
		AgentStatus:    final.AgentStatus,
		FallbackReason: final.FallbackReason,
		RepairApplied:  final.RepairApplied,
		SchemaValid:    final.AgentStatus != "fallback",
	}

	_ = trace // available for logging/debug

	return &MatchResponse{
		AnalysisID:  0,
		GeneratedBy: final.GeneratedBy,
		Result:      result,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
	}, nil
}

func extractTitles(pts []agent.MatchDecisionPoint) []string {
	if len(pts) == 0 { return []string{} }
	r := make([]string, len(pts))
	for i, p := range pts { r[i] = p.Title }
	return r
}

func extractRiskTitles(pts []agent.MatchDecisionRisk) []string {
	if len(pts) == 0 { return []string{} }
	r := make([]string, len(pts))
	for i, p := range pts { r[i] = p.Title }
	return r
}

func convertImprovements(items []agent.MatchImprovement) []model.MatchSuggestion {
	result := make([]model.MatchSuggestion, len(items))
	for i, item := range items {
		result[i] = model.MatchSuggestion{
			Priority: item.Priority,
			Title:    item.Title,
			Action:   item.Action,
		}
	}
	return result
}

func extractActions(items []agent.MatchImprovement) []string {
	r := make([]string, len(items))
	for i, item := range items {
		r[i] = item.Action
	}
	return r
}

func fastMatchLevel(score int) string {
	if score >= 80 { return "建议投递" }
	if score >= 60 { return "建议补强后投递" }
	return "待进一步分析"
}
