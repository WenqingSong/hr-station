package service

import (
	"context"
	"encoding/json"
	apperror "offer-catcher/server/internal/errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/llm"
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

// ─── H18-C: Fast Match SEP Context Types ──────────────────────────────────

// FastMatchSEPContextStatus describes the SEP readiness for Fast Match.
type FastMatchSEPContextStatus string

const (
	FastMatchSEPContextReady    FastMatchSEPContextStatus = "ready"
	FastMatchSEPContextMissing  FastMatchSEPContextStatus = "missing"
	FastMatchSEPContextBuilding FastMatchSEPContextStatus = "building"
	FastMatchSEPContextInvalid  FastMatchSEPContextStatus = "invalid"
)

// FastMatchSEPReadiness holds the SEP context readiness check result.
type FastMatchSEPReadiness struct {
	ResumeID int64
	JobID    int64

	ResumeProfileReady bool
	JobProfileReady    bool

	ResumeProfileStatus string
	JobProfileStatus    string

	Status FastMatchSEPContextStatus
	Reason string
}

// MatchService handles job match analysis business logic.
// K13-Fix-B: Fast tier uses TagEvidenceService (no LLM); Deep tier uses LLM pipeline.
// H18-C: Fast tier SEP context gate — tag fallback removed from Fast primary path.
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
	hy3Client      llm.Client // JM-R3-R: Hy3 LLM client for 5-module Deep pipeline
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
	hy3Client llm.Client,
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
		hy3Client:      hy3Client,
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

// ─── H18-C: Fast Match SEP Context Gate ──────────────────────────────────

// checkFastSEPContext checks whether both resume and job SEP profiles are ready for Fast Match.
// Returns readiness status and the reason if not ready.
func (s *MatchService) checkFastSEPContext(ctx context.Context, resumeID, jobID int64) *FastMatchSEPReadiness {
	result := &FastMatchSEPReadiness{
		ResumeID: resumeID,
		JobID:    jobID,
		Status:   FastMatchSEPContextReady,
	}

	if s.semanticRepo == nil {
		result.Status = FastMatchSEPContextMissing
		result.Reason = "semantic profile repository not available"
		return result
	}

	// Check resume SEP profile
	rp, rerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, resumeID)
	if rerr != nil || rp == nil {
		result.ResumeProfileStatus = "missing"
		result.Status = FastMatchSEPContextMissing
		result.Reason = fmt.Sprintf("resume SEP profile not found: resume=%d", resumeID)
		return result
	}
	result.ResumeProfileStatus = rp.Status

	switch rp.Status {
	case model.SemanticProfileStatusVerified:
		result.ResumeProfileReady = true
	case model.SemanticProfileStatusPending, model.SemanticProfileStatusProcessing:
		result.Status = FastMatchSEPContextBuilding
		result.Reason = fmt.Sprintf("resume SEP profile still %s: resume=%d", rp.Status, resumeID)
		return result
	case model.SemanticProfileStatusFailed, model.SemanticProfileStatusStale:
		result.Status = FastMatchSEPContextInvalid
		result.Reason = fmt.Sprintf("resume SEP profile status=%s: resume=%d", rp.Status, resumeID)
		return result
	default:
		result.Status = FastMatchSEPContextMissing
		result.Reason = fmt.Sprintf("resume SEP profile status=%s: resume=%d", rp.Status, resumeID)
		return result
	}

	// Check job SEP profile
	jp, jerr := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, jobID)
	if jerr != nil || jp == nil {
		result.JobProfileStatus = "missing"
		result.Status = FastMatchSEPContextMissing
		result.Reason = fmt.Sprintf("job SEP profile not found: job=%d", jobID)
		return result
	}
	result.JobProfileStatus = jp.Status

	switch jp.Status {
	case model.SemanticProfileStatusVerified:
		result.JobProfileReady = true
	case model.SemanticProfileStatusPending, model.SemanticProfileStatusProcessing:
		result.Status = FastMatchSEPContextBuilding
		result.Reason = fmt.Sprintf("job SEP profile still %s: job=%d", jp.Status, jobID)
		return result
	case model.SemanticProfileStatusFailed, model.SemanticProfileStatusStale:
		result.Status = FastMatchSEPContextInvalid
		result.Reason = fmt.Sprintf("job SEP profile status=%s: job=%d", jp.Status, jobID)
		return result
	default:
		result.Status = FastMatchSEPContextMissing
		result.Reason = fmt.Sprintf("job SEP profile status=%s: job=%d", jp.Status, jobID)
		return result
	}

	return result
}

// safeMatchFallbackCtx returns a short-lived background context for mock fallback DB persistence.
// Uses context.Background to ensure the save succeeds even if the request context is cancelled.
func safeMatchFallbackCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// Analyze generates a match analysis for the given user and job.
// H18-C: Fast tier primary path uses SEP only — tag fallback removed.
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

	// H18-C: Fast tier — SEP is the only data source, no tag fallback
	isFast := genTier == "fast" || req.GenerationTier != model.GenerationTierQuality
	if isFast {
		sepCtx := s.checkFastSEPContext(ctx, int64(resume.ID), int64(job.ID))
		if sepCtx.Status == FastMatchSEPContextReady {
			sepResp, sepErr := s.fastMatchWithSEP(ctx, userID, resume, job)
			if sepErr == nil {
				return sepResp, nil
			}
			log.Printf("[MatchService] SEP fast match eng error despite ready context: %v, returning profile_missing", sepErr)
		}
		// SEP not ready → return profile_pending, no tag fallback
		log.Printf("[MatchService] fast_sep_profile_pending status=%s reason=%s resume=%d job=%d",
			sepCtx.Status, sepCtx.Reason, resume.ID, job.ID)
		return s.buildFastSEPPendingResponse(resume.ID, job.ID, sepCtx), nil
	}

	// JM-R3: Deep/Quality tier — screening_match_v1 (5-module Hy3 + Judge)
	runner := s.selectRunner(req.GenerationTier)
	isQuality := genTier == "quality" || req.GenerationTier == model.GenerationTierQuality
	if isQuality && s.llmEnabled && runner != nil {
		log.Printf("[MatchService] use_screening_match_v1 tier=%s resume=%d job=%d", genTier, req.ResumeID, jobID)
		smResp, smErr := s.runScreeningMatchEntry(ctx, resume, job)
		if smErr == nil {
			return smResp, nil
		}
		log.Printf("[MatchService] screening match v1 failed: %v, trying legacy fallback", smErr)
		// Legacy: 3-agent fallback
		if s.tagEvidence != nil {
			resp, err3a := s.matchWith3Agents(ctx, userID, resume, job, runner)
			if err3a == nil { markLegacyFallback(resp); return resp, nil }
			log.Printf("[MatchService] 3-agent match failed: %v", err3a)
		}
		log.Printf("[MatchService] use_deep_llm tier=%s", genTier)
		resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
		if fallbackErr == nil { markLegacyFallback(resp); return resp, nil }
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
			GenerationTier:         "fast",
			GeneratedBy:            "fast_sep_match_engine",
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

	// Save with full metadata (K15-P0-R2: use safe context, request ctx may be cancelled)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		log.Printf("[MatchService] failed to marshal fallback result: %v", err)
		return resp, nil // return the already-built response, don't fail
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

	saveCtx, saveCancel := safeMatchFallbackCtx()
	if createErr := s.matchRepo.Create(saveCtx, analysis); createErr != nil {
		log.Printf("[MatchService] failed to save fallback result (non-fatal): %v", createErr)
	} else {
		resp.AnalysisID = analysis.ID
	}
	saveCancel()

	return resp, nil
}

// GetLatestMatchAnalysis returns the latest ready match analysis for the given user+resume+job.
// If generationTier is empty, defaults to "fast".
func (s *MatchService) GetLatestMatchAnalysis(ctx context.Context, userID, resumeID, jobID uint64, generationTier string) (*model.MatchLatestResult, error) {
	if resumeID <= 0 {
		return nil, apperror.ErrInvalidParam
	}
	if jobID <= 0 {
		return nil, apperror.ErrInvalidParam
	}

	// Normalise tier
	tier, err := normalizeMatchGenerationTier(generationTier)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidParam, err.Error(), 400)
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

	latest, err := s.matchRepo.FindLatestByResumeAndJobAndTier(ctx, userID, resumeID, jobID, tier)
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

// normalizeMatchGenerationTier normalises the tier for match latest queries.
// Empty → "fast", "fast" → "fast", "quality"/"deep" → "quality", other → error.
func normalizeMatchGenerationTier(input string) (string, error) {
	switch input {
	case "", "fast":
		return "fast", nil
	case "quality", "deep":
		return "quality", nil
	default:
		return "", fmt.Errorf("invalid generation_tier: %q (expected fast or quality)", input)
	}
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

// ─── H18-C: Fast SEP Profile Pending Response ─────────────────────────────

// buildFastSEPPendingResponse constructs a profile_pending response when SEP profiles are not ready.
// No tag-based score fallback — returns score 0 with "待分析" level.
func (s *MatchService) buildFastSEPPendingResponse(resumeID, jobID uint64, sepCtx *FastMatchSEPReadiness) *MatchResponse {
	result := &model.MatchAnalysisResult{
		MatchScore:              0,
		MatchLevel:              "待分析",
		MatchLabel:              "待分析",
		ScreeningResult:         "profile_pending",
		EvidenceChain:           []model.MatchEvidence{},
		MatchHighlights:         []string{},
		GapPoints:               []string{},
		RiskPoints:              []string{},
		OptimizationSuggestions: []model.MatchSuggestion{},
		NextSteps:               []string{"等待语义画像生成完成后重新分析"},
		FitPoints:               []string{},
		ResumeEvidence:          []string{},
		JobRequirements:         []string{},
		GeneratedBy:             "fast_sep_profile_pending",
		AgentStatus:             "profile_missing",
		FallbackReason:          "resume_or_job_sep_profile_not_ready",
		RepairApplied:           false,
		DataQuality:             "sep",
		Summary:                 "语义画像尚未准备完成，暂不能给出可靠的岗位匹配分。请稍后重新生成。",
	}

	agentStatus := "profile_missing"
	switch sepCtx.Status {
	case FastMatchSEPContextBuilding:
		agentStatus = "profile_building"
	case FastMatchSEPContextInvalid:
		agentStatus = "profile_invalid"
	}
	result.AgentStatus = agentStatus

	meta := model.AgentResponseMeta{
		AgentName:      "FastSEPMatchEngine",
		GeneratedBy:    "fast_sep_profile_pending",
		AgentStatus:    agentStatus,
		FallbackReason: "resume_or_job_sep_profile_not_ready",
		RepairApplied:  false,
		SchemaValid:    true,
	}

	return &MatchResponse{
		AnalysisID:  0,
		GeneratedBy: "fast_sep_profile_pending",
		Result:      result,
		AnalysisBasis: &model.AnalysisBasis{
			JobDescriptionUsed:  false,
			JobRequirementsUsed: false,
			DataQuality:         "pending",
		},
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Meta:      meta,
	}
}

// ─── Legacy: Fast Match via TagEvidenceService ────────────────────────────
// H18-C: Fast primary path no longer calls this. Kept for Deep 3-agent fallback (matchWith3Agents).
// Will be removed in H19 when Deep migrates away from TagEvidenceService.

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

// ─── JM-R3: Screening Match V1 Entry ─────────────────────────────────────

// runScreeningMatchEntry builds context, runs 5-module pipeline, returns response.
// JM-R3-R: Quality tier now uses real Hy3 5-module LLM pipeline; rule-based is fallback.
func (s *MatchService) runScreeningMatchEntry(
	ctx context.Context, resume *model.ResumeVersion, job *model.Job,
) (*MatchResponse, error) {
	smCtx, err := s.buildScreeningMatchContext(ctx, resume, job)
	if err != nil {
		// SEP not ready → try light text-based fallback before returning profile_pending
		if smCtx == nil {
			// JM-R5: Attempt light match using raw resume text when SEP is missing
			if lightCtx := s.buildLightScreeningContext(resume, job); lightCtx != nil {
				report := runScreeningMatchPipeline(lightCtx)
				resp := screeningMatchToResponse(report)
				resp.Meta.FallbackReason = "se_profile_missing_light_match"
				resp.Meta.GeneratedBy = "rule_screening_match_fallback"
				resp.Meta.AgentStatus = "fallback"
				resp.Result.GeneratedBy = "rule_screening_match_fallback"
				resp.Result.AgentStatus = "fallback"
				return resp, nil
			}
			result := &model.MatchAnalysisResult{
				MatchScore:      0, MatchLevel: "待分析", ScreeningResult: "profile_pending",
				GeneratedBy: "deep_match_profile_pending", AgentStatus: "profile_missing",
				FallbackReason: "resume_or_job_sep_profile_not_ready",
				Summary:        "语义画像尚未准备完成，暂不能给出深度匹配分析。",
				MatchHighlights: []string{}, GapPoints: []string{}, RiskPoints: []string{},
				NextSteps: []string{"等待语义画像生成完成后重新分析"},
				FitPoints: []string{}, EvidenceChain: []model.MatchEvidence{},
			}
			return &MatchResponse{
				GeneratedBy: "deep_match_profile_pending", Result: result,
				CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
				Meta: model.AgentResponseMeta{AgentName: "ScreeningMatch", GeneratedBy: "deep_match_profile_pending",
					AgentStatus: "profile_missing", FallbackReason: "resume_or_job_sep_profile_not_ready"},
			}, nil
		}
		return nil, err
	}

	// JM-R3-R: Quality tier — use real Hy3 5-module pipeline when client is available
	if s.hy3Client != nil {
		log.Printf("[ScreeningMatch] attempting Hy3 5-module pipeline for resume=%d job=%d", resume.ID, job.ID)
		hy3Report, hy3Status, hy3GenBy := runHy3ScreeningPipeline(ctx, smCtx, s.hy3Client)
		if hy3Report != nil {
			resp := screeningMatchToResponse(hy3Report)
			resp.Meta.AgentStatus = hy3Status
			resp.Meta.GeneratedBy = hy3GenBy
			resp.Result.GeneratedBy = hy3GenBy
			resp.Result.AgentStatus = hy3Status
			return resp, nil
		}
		log.Printf("[ScreeningMatch] Hy3 pipeline failed, falling back to rule-based for resume=%d job=%d", resume.ID, job.ID)
	}

	// JM-R3-R: Rule-based fallback (also used when hy3Client is nil)
	report := runScreeningMatchPipeline(smCtx)
	resp := screeningMatchToResponse(report)
	if s.hy3Client == nil {
		// No Hy3 client at all — this is the only path
		resp.Meta.FallbackReason = "rule_screening_match_no_hy3"
		resp.Result.GeneratedBy = "rule_screening_match_fallback"
		resp.Result.AgentStatus = "fallback"
	} else {
		// Had Hy3 client but it failed
		resp.Meta.FallbackReason = "hy3_failed_rule_fallback"
		resp.Result.GeneratedBy = "rule_screening_match_fallback"
		resp.Result.AgentStatus = "fallback"
	}
	return resp, nil
}

// buildLightScreeningContext creates a minimal context from raw resume text
// when SEP profiles are unavailable. Used for light match fallback.
func (s *MatchService) buildLightScreeningContext(resume *model.ResumeVersion, job *model.Job) *ScreeningMatchContext {
	// Only proceed if we have raw resume text to work with
	rawText := resume.RawText
	if rawText == "" {
		return nil
	}

	// Basic keyword match scoring against job direction/category
	resumeLower := strings.ToLower(rawText)
	jobLower := strings.ToLower(job.Title + " " + job.Direction + " " + job.Category)

	// Detect basic matches from keyword overlap
	keywords := []string{"java", "go", "python", "mysql", "redis", "spring", "gin",
		"后端", "后台", "开发", "数据库", "接口", "docker", "linux"}
	matchCount := 0
	for _, kw := range keywords {
		if strings.Contains(resumeLower, kw) && strings.Contains(jobLower, kw) {
			matchCount++
		}
	}

	// Also count resume skills present in a broader job context
	skillKeywords := []string{"java", "sql", "mysql", "redis", "spring", "docker", "linux", "git"}
	skillMatchCount := 0
	for _, kw := range skillKeywords {
		if strings.Contains(resumeLower, kw) {
			skillMatchCount++
		}
	}
	// Combined: shared keywords + skill presence in resume
	effectiveMatch := matchCount + minInt(skillMatchCount/2, 3)

	// Detect internship/project presence (check for positive evidence, not negation)
	hasNoInternship := strings.Contains(resumeLower, "没有") && strings.Contains(resumeLower, "实习")
	hasNoProject := strings.Contains(resumeLower, "没有") && strings.Contains(resumeLower, "项目")
	hasInternship := strings.Contains(resumeLower, "实习") && !hasNoInternship
	hasProject := (strings.Contains(resumeLower, "项目经历") || strings.Contains(resumeLower, "项目")) && !hasNoProject

	// Derive a rough score
	baseScore := 35
	if effectiveMatch >= 3 {
		baseScore = 42
	}
	if effectiveMatch >= 5 {
		baseScore = 50
	}
	if hasInternship {
		baseScore += 5
	}
	if hasProject {
		baseScore += 5
	}
	// Floor: if direction is related and there's some skill overlap, minimum 40
	if hasDirectionOverlap(resumeLower, jobLower) && skillMatchCount >= 2 {
		if baseScore < 40 {
			baseScore = 40
		}
	}
	if baseScore > 60 {
		baseScore = 60
	}
	if baseScore < 25 {
		baseScore = 25
	}

	// Build highlights
	highlights := []string{}
	if hasInternship {
		highlights = append(highlights, "有实习相关描述")
	}
	if hasProject {
		highlights = append(highlights, "有项目相关描述")
	}
	if matchCount >= 3 {
		highlights = append(highlights, "具备部分岗位方向相关技能基础")
	}

	gapPoints := []string{}
	if !hasInternship && !hasProject {
		gapPoints = append(gapPoints, "简历缺少项目和实习经历，建议先补充实践证据")
	} else if !hasInternship {
		gapPoints = append(gapPoints, "缺少实习经历，建议补充企业实习或真实项目经验")
	}
	if matchCount < 5 {
		gapPoints = append(gapPoints, "技能覆盖面有限，建议补充后端开发核心技术栈的项目实践")
	}

	// Determine risk
	riskName := "简历证据不足，缺少真实项目支撑"
	riskLevel := "high"
	if hasProject || hasInternship {
		riskName = "项目深度和工程证据不足"
		riskLevel = "medium"
	}
	structuredRisks := []MatchRiskPoint{
		{
			RiskName:      riskName,
			RiskLevel:     riskLevel,
			WhyItMatters:  "实习生岗位筛选时，HR会优先关注有真实项目或实习经历的简历。仅有基础技能列表而缺少实践证据，初筛通过率很低。",
			EvidenceGap:   "简历以基础技能描述为主，缺少完整的项目开发、接口设计、数据库操作等后端工程实践证据。",
			FixSuggestion: "建议先完成一个较完整的后端项目（含接口开发、数据库设计、登录鉴权、缓存使用和部署验证），把项目作为简历核心证据，再投递该岗位。",
		},
	}

	// Build light FastResult
	lightResult := &model.MatchAnalysisResult{
		MatchScore:      baseScore,
		MatchLevel:      matchLevel(baseScore),
		MatchHighlights: highlights,
		GapPoints:       gapPoints,
		RiskPoints:      []string{riskName},
		Summary:         fmt.Sprintf("基础匹配分析：简历与%s方向匹配度%s（基于基础文本分析，非完整语义画像）。", job.Title, matchLevel(baseScore)),
		StructuredRisks: structuredRisks,
		CoreAdvantages:  highlights,
		MainGaps:        gapPoints,
	}

	lightCtx := &ScreeningMatchContext{
		ResumeID:        int64(resume.ID),
		JobID:           int64(job.ID),
		ResumeName:      resume.Name,
		ResumeRawText:   rawText,
		JobTitle:        job.Title,
		JobDirection:    job.Direction,
		JobCategory:     job.Category,
		JobCity:         job.City,
		JobRawJD:        job.Description,
		FastResult:      lightResult,
		ResumeSEPProfile: nil,
		JobSEPProfile:    nil,
	}
	return lightCtx
}

// hasDirectionOverlap checks if the resume text and job context share direction-related terms.
func hasDirectionOverlap(resumeLower, jobLower string) bool {
	dirTerms := []string{"后端", "后台", "开发", "软件", "java", "go", "python", "服务端"}
	for _, term := range dirTerms {
		if strings.Contains(resumeLower, term) && strings.Contains(jobLower, term) {
			return true
		}
	}
	// Broad match: if resume mentions "开发" and job involves software/backend
	if strings.Contains(resumeLower, "开发") &&
		(strings.Contains(jobLower, "开发") || strings.Contains(jobLower, "软件") || strings.Contains(jobLower, "后端") || strings.Contains(jobLower, "后台")) {
		return true
	}
	return false
}

// screeningMatchToResponse converts ScreeningMatchV1 to MatchResponse.
func screeningMatchToResponse(report *ScreeningMatchV1) *MatchResponse {
	result := &model.MatchAnalysisResult{
		MatchScore:      report.ScreeningDecision.Score,
		MatchLevel:      report.ScreeningDecision.Level,
		ScreeningResult: "pass",
		GeneratedBy:     "deep_match_agent",
		AgentStatus:     "screening_match_verified",
		FallbackReason:  "none",
		Summary:         report.ScreeningDecision.Summary,
		ReportVersion:   report.Version,
		ApplicationDecision: report.ScreeningDecision.Decision,
		DecisionLabel:       report.ScreeningDecision.DecisionLabel,
		OneSentenceConclusion: report.ScreeningDecision.Summary,
		// Set legacy arrays for backward compat
		MatchHighlights: []string{},
		GapPoints:       []string{},
		RiskPoints:      []string{},
		NextSteps:       []string{report.JobStrategy.RecommendedAction},
		FitPoints:       []string{},
		EvidenceChain:   []model.MatchEvidence{},
		// JM-R3-R: Propagate all 5 Hy3 module outputs
		ScreeningPassPoints:      report.ScreeningPassPoints,
		ScreeningEliminationRisks: report.ScreeningEliminationRisks,
		ResumeRewriteCommands:    report.ResumeRewriteCommands,
		JobStrategy:              report.JobStrategy,
	}
	for _, p := range report.ScreeningPassPoints {
		result.MatchHighlights = append(result.MatchHighlights, p.Title)
		result.FitPoints = append(result.FitPoints, p.ResumeEvidence)
	}
	for _, r := range report.ScreeningEliminationRisks {
		result.RiskPoints = append(result.RiskPoints, r.RiskTitle)
		result.GapPoints = append(result.GapPoints, r.RiskTitle)
	}

	meta := model.AgentResponseMeta{
		AgentName: "ScreeningMatchEngine", GeneratedBy: "deep_match_agent",
		AgentStatus: "screening_match_verified", FallbackReason: "none",
		RepairApplied: false, SchemaValid: true,
	}

	return &MatchResponse{
		AnalysisID:  0, GeneratedBy: "deep_match_agent", Result: result,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"), Meta: meta,
	}
}

// ─── H19-B: Deep Match V3 Entry (legacy) ─────────────────────────────────

// runDeepMatchV3Entry builds context, runs v3 pipeline, converts to MatchResponse.
func (s *MatchService) runDeepMatchV3Entry(
	ctx context.Context, userID uint64, resume *model.ResumeVersion, job *model.Job,
) (*MatchResponse, error) {
	v3ctx, err := s.buildDeepMatchContextV3(ctx, userID, resume, job)
	if err != nil && v3ctx == nil {
		return nil, err
	}
	// If SEP missing, return profile_pending
	if v3ctx.Status != "ready" {
		result := &model.MatchAnalysisResult{
			MatchScore:      0,
			MatchLevel:      "待分析",
			ScreeningResult: "profile_pending",
			GeneratedBy:     "deep_match_profile_pending",
			AgentStatus:     "profile_missing",
			FallbackReason:  "resume_or_job_sep_profile_not_ready",
			Summary:         "语义画像尚未准备完成，暂不能给出深度匹配分析。请稍后重新生成。",
			MatchHighlights: []string{},
			GapPoints:       []string{},
			RiskPoints:      []string{},
			NextSteps:       []string{"等待语义画像生成完成后重新分析"},
			FitPoints:       []string{},
			EvidenceChain:   []model.MatchEvidence{},
		}
		return &MatchResponse{
			GeneratedBy: "deep_match_profile_pending",
			Result:      result,
			CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
			Meta: model.AgentResponseMeta{
				AgentName: "DeepMatchAgent", GeneratedBy: "deep_match_profile_pending",
				AgentStatus: "profile_missing", FallbackReason: "resume_or_job_sep_profile_not_ready",
			},
		}, nil
	}

	report := runDeepMatchV3(v3ctx)
	return deepV3ReportToResponse(report, userID, resume, job), nil
}

// deepV3ReportToResponse converts DeepMatchV3Report to MatchResponse + fills legacy model fields.
func deepV3ReportToResponse(report *DeepMatchV3Report, userID uint64, resume *model.ResumeVersion, job *model.Job) *MatchResponse {
	result := &model.MatchAnalysisResult{
		MatchScore:              report.MatchScore,
		MatchLevel:              report.MatchLevel,
		MatchLabel:              report.MatchLevel,
		ScreeningResult:         "pass",
		GeneratedBy:             report.GeneratedBy,
		AgentStatus:             report.AgentStatus,
		FallbackReason:          "none",
		RepairApplied:           false,
		Summary:                 report.OneSentenceConclusion,
		ReportVersion:           report.ReportVersion,
		ApplicationDecision:     report.ApplicationDecision,
		DecisionLabel:           report.DecisionLabel,
		OneSentenceConclusion:   report.OneSentenceConclusion,
		JobUnderstanding:        report.JobUnderstanding,
		RequirementCoverage:     report.RequirementCoverage,
		ResumeEvidenceMapping:   report.ResumeEvidenceMapping,
		CoreMatchPoints:         report.CoreMatchPoints,
		KeyRisks:                report.KeyRisks,
		ResumeRewriteStrategy:   report.ResumeRewriteStrategy,
		InterviewRiskQuestions:  report.InterviewRiskQuestions,
		ActionPlan:              report.ActionPlan,
		AgentTrace:              report.AgentTrace,
		ScoreBreakdownObj:       report.ScoreBreakdown,
		StructuredRisks:         convertV3RisksToMatchRiskPoints(report.KeyRisks),
		EvidenceChain:           []model.MatchEvidence{},
		MatchHighlights:         []string{},
		GapPoints:               []string{},
		RiskPoints:              []string{},
		OptimizationSuggestions: []model.MatchSuggestion{},
		NextSteps:               flattenActionPlan(report.ActionPlan),
		FitPoints:               []string{},
		ResumeEvidence:          []string{},
		DataQuality:             "sep",
	}

	// Fill legacy arrays
	for _, cp := range report.CoreMatchPoints {
		result.MatchHighlights = append(result.MatchHighlights, cp.PointName)
	}
	for _, r := range report.KeyRisks {
		result.RiskPoints = append(result.RiskPoints, r.RiskName)
	}
	for _, ev := range report.ResumeEvidenceMapping {
		result.ResumeEvidence = append(result.ResumeEvidence, ev.EvidenceText)
	}

	meta := model.AgentResponseMeta{
		AgentName:      "DeepMatchAgent",
		GeneratedBy:    report.GeneratedBy,
		AgentStatus:    report.AgentStatus,
		FallbackReason: "none",
		RepairApplied:  false,
		SchemaValid:    true,
	}

	return &MatchResponse{
		AnalysisID:  0,
		GeneratedBy: report.GeneratedBy,
		Result:      result,
		CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
	}
}

func convertV3RisksToMatchRiskPoints(risks []DeepKeyRisk) []MatchRiskPoint {
	out := make([]MatchRiskPoint, len(risks))
	for i, r := range risks {
		out[i] = MatchRiskPoint{
			RiskName:      r.RiskName,
			RiskLevel:     r.RiskLevel,
			WhyItMatters:  r.WhyItMatters,
			EvidenceGap:   r.EvidenceGap,
			FixSuggestion: r.FixSuggestion,
		}
	}
	return out
}

func flattenActionPlan(plan []map[string]any) []string {
	out := []string{}
	for _, p := range plan {
		if s, ok := p["step"].(string); ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		out = append(out, "基于深度分析结果优化简历后再投递")
	}
	return out
}

// ─── H19-B: Legacy Fallback Marking ────────────────────────────────────────

// markLegacyFallback marks a response as legacy fallback (not deep_match_agent verified).
func markLegacyFallback(resp *MatchResponse) {
	resp.GeneratedBy = "deep_match_legacy_fallback"
	resp.Meta.GeneratedBy = "deep_match_legacy_fallback"
	resp.Meta.AgentStatus = "legacy_fallback"
	if resp.Result != nil {
		resp.Result.GeneratedBy = "deep_match_legacy_fallback"
		resp.Result.AgentStatus = "legacy_fallback"
		resp.Result.FallbackReason = "deep_match_v3_failed"
		resp.Result.ReportVersion = "deep_match_legacy"
	}
}

// ─── H19-B: Legacy Sanitize (avoid overwriting fallback markers) ──────────

// sanitizeDeepResponseLegacy only sanitizes null arrays, does NOT overwrite generated_by.
func sanitizeDeepResponseLegacy(resp *MatchResponse) *MatchResponse {
	if resp == nil || resp.Result == nil { return resp }
	r := resp.Result
	if r.MatchHighlights == nil { r.MatchHighlights = []string{} }
	if r.GapPoints == nil { r.GapPoints = []string{} }
	if r.RiskPoints == nil { r.RiskPoints = []string{} }
	if r.EvidenceChain == nil { r.EvidenceChain = []model.MatchEvidence{} }
	if r.NextSteps == nil { r.NextSteps = []string{} }
	if r.OptimizationSuggestions == nil { r.OptimizationSuggestions = []model.MatchSuggestion{} }
	if r.FitPoints == nil { r.FitPoints = []string{} }
	if r.ResumeEvidence == nil { r.ResumeEvidence = []string{} }
	return resp
}

// ─── Legacy: sanitizeFallbackResult (H18-C: kept for remaining legacy callers, not called from Fast primary path) ──
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
