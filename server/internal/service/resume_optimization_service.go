package service

import (
	"context"
	"encoding/json"
	"fmt"

	apperror "offer-catcher/server/internal/errors"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ResumeOptimizationRequest represents a request for resume optimization.
type ResumeOptimizationRequest struct {
	// K8-5-48-D: resume_id + job_id are the primary inputs; resume_text is deprecated optional fallback.
	ResumeID       uint64                                 `json:"resume_id,omitempty"`
	JobID          uint64                                 `json:"job_id,omitempty"`
	ResumeText     string                                 `json:"resume_text,omitempty"` // @deprecated use resume_id
	TargetJob      *model.ResumeOptimizationTargetJob     `json:"target_job,omitempty"`
	TargetSection  string                                 `json:"target_section,omitempty"`
	UserIntention  *model.ResumeOptimizationUserIntention `json:"user_intention,omitempty"`
	Mode           string                                 `json:"mode,omitempty"`           // "core" — use Core Agent
	GenerationTier model.GenerationTier                   `json:"generation_tier,omitempty"` // "fast" or "quality"
}

// ResumeOptimizationResponse represents the response for resume optimization.
type ResumeOptimizationResponse struct {
	OptimizationID uint64                          `json:"optimization_id"`
	GeneratedBy    string                          `json:"generated_by"` // deprecated: use meta.generated_by
	Result         *model.ResumeOptimizationResult `json:"result"`
	CreatedAt      string                          `json:"created_at"`
	Meta           model.AgentResponseMeta         `json:"meta"`
}

// ResumeOptimizationService handles resume optimization business logic.
// It supports LLM Pipeline priority with Mock Agent fallback.
// Supports generation_tier: "fast" uses fastRunner, "quality" uses qualityRunner.
type ResumeOptimizationService struct {
	resumeRepo       *repository.ResumeRepository
	jobRepo          *repository.JobRepository
	optimizationRepo *repository.ResumeOptimizationRepository
	mockAgent        interface{}
	fastRunner       agent.PipelineRunner
	qualityRunner    agent.PipelineRunner
	llmEnabled       bool
	fallbackEnable   bool
	enableMockAgent  bool
	tagEvidence      *TagEvidenceService // K13-Fix-C3
}

// NewResumeOptimizationService creates a new ResumeOptimizationService.
func NewResumeOptimizationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	optimizationRepo *repository.ResumeOptimizationRepository,
	mockAgent interface{},
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
	fallbackEnable bool,
	enableMockAgent bool,
	tagEvidence *TagEvidenceService,
) *ResumeOptimizationService {
	return &ResumeOptimizationService{
		resumeRepo:       resumeRepo,
		jobRepo:          jobRepo,
		optimizationRepo: optimizationRepo,
		mockAgent:        mockAgent,
		fastRunner:       fastRunner,
		qualityRunner:    qualityRunner,
		llmEnabled:       llmEnabled,
		fallbackEnable:   fallbackEnable,
		enableMockAgent:  enableMockAgent,
		tagEvidence:      tagEvidence,
	}
}

// selectRunner returns the appropriate pipeline runner for the given tier.
func (s *ResumeOptimizationService) selectRunner(tier model.GenerationTier) agent.PipelineRunner {
	// tier validated by handler
	if tier == model.GenerationTierQuality && s.qualityRunner != nil {
		return s.qualityRunner
	}
	return s.fastRunner
}

// Optimize generates a resume optimization for the given user and request.
// When LLM is enabled and Pipeline Runner is available, it uses the LLM Pipeline
// first. If the Pipeline returns fallback_required or mapping fails, it falls
// back to the Mock Agent. When LLM is disabled, it uses the Mock Agent directly.
func (s *ResumeOptimizationService) Optimize(ctx context.Context, userID uint64, req ResumeOptimizationRequest) (*ResumeOptimizationResponse, error) {
	// K8-5-48-D: resolve resume text from resume_id if provided
	if req.ResumeText == "" && req.ResumeID > 0 {
		resume, err := s.resumeRepo.GetByID(ctx, userID, req.ResumeID)
		if err != nil {
			return nil, fmt.Errorf("failed to get resume: %w", err)
		}
		if resume == nil {
			return nil, apperror.ErrResumeNotFound
		}
		req.ResumeText = resume.RawText
		// K8-5-48-D: also resolve job from job_id for target_job info
		if req.JobID > 0 && req.TargetJob == nil {
			job, jErr := s.jobRepo.FindByID(ctx, req.JobID)
			if jErr == nil && job != nil {
				req.TargetJob = &model.ResumeOptimizationTargetJob{
					Title:  job.Title,
					JDText: job.Description,
				}
			}
		}
	}

	if err := validateOptimizeRequest(req); err != nil {
		return nil, err
	}

	runner := s.selectRunner(req.GenerationTier)
	log.Printf("[ResumeOptimizationService] generation_tier=%s mode=%s", string(req.GenerationTier), req.Mode)

	// Dispatch to Core Agent if mode=core
	if req.Mode == "core" {
		return s.optimizeCore(ctx, userID, req, runner)
	}

	input := &agent.OptimizeInput{
		ResumeText:    req.ResumeText,
		TargetSection: req.TargetSection,
	}

	if req.TargetJob != nil {
		input.JobTitle = req.TargetJob.Title
		input.JDText = req.TargetJob.JDText
	}

	if req.UserIntention != nil {
		input.AcceptRelocation = req.UserIntention.AcceptRelocation
	}

	// Try LLM Pipeline if enabled
	if s.llmEnabled && runner != nil {
		log.Printf("[ResumeOptimizationService] use_llm=true tier=%s", string(req.GenerationTier))
		resp, fallbackErr := s.optimizeWithLLM(ctx, userID, input, req, runner)
		if fallbackErr == nil {
			return resp, nil
		}

		// Log the fallback reason
		log.Printf("[ResumeOptimizationService] LLM pipeline fallback: %v", fallbackErr)

		// K13-Fix-1: Check fallback enable flag
		if !s.fallbackEnable {
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed and fallback disabled: %s", fallbackErr.FallbackReason), 500)
		}

		// Fall back to Mock Agent
		return s.optimizeWithMock(ctx, userID, input, req, fallbackErr)
	}

	// LLM disabled — check if mock is allowed
	if !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "LLM is disabled and ENABLE_MOCK_AGENT=false, no generation path available", 500)
	}

	// LLM disabled — use Mock Agent directly
	log.Printf("[ResumeOptimizationService] use_llm=false use_mock=true tier=%s", string(req.GenerationTier))
	return s.optimizeWithMock(ctx, userID, input, req, nil)
}

// optimizeWithLLM calls the ResumeOptimizationAgent (LLM Pipeline) and builds the response.
func (s *ResumeOptimizationService) optimizeWithLLM(ctx context.Context, userID uint64, input *agent.OptimizeInput, req ResumeOptimizationRequest, runner agent.PipelineRunner) (*ResumeOptimizationResponse, *agent.ResumeOptFallbackError) {
	llmAgent := agent.NewResumeOptimizationAgent(runner)
	result, err := llmAgent.Optimize(ctx, input)
	if err != nil {
		if fbErr, ok := err.(*agent.ResumeOptFallbackError); ok {
			return nil, fbErr
		}
		// Unexpected error — wrap as fallback
		return nil, &agent.ResumeOptFallbackError{
			FallbackReason: "llm_agent_error",
			ErrorCode:      "LLM_AGENT_ERROR",
			ErrorMessage:   err.Error(),
		}
	}

	now := time.Now()
	resp := &ResumeOptimizationResponse{
		GeneratedBy: result.GeneratedBy,
		Result:      result,
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
		Meta: model.DeriveMetaFromResult(
			model.AgentNameResumeOptimization,
			result.GeneratedBy,
			result.AgentStatus,
			result.FallbackReason,
			result.ErrorCode,
			result.RepairApplied,
		),
	}

	// Save optimization snapshot (best effort)
	s.saveSnapshot(ctx, userID, req, result, resp)

	return resp, nil
}

// optimizeWithMock calls the Mock Agent and builds the response.
// If fbErr is non-nil, it means we're in fallback mode from a failed LLM attempt.
func (s *ResumeOptimizationService) optimizeWithMock(
	ctx context.Context,
	userID uint64,
	input *agent.OptimizeInput,
	req ResumeOptimizationRequest,
	fbErr *agent.ResumeOptFallbackError,
) (*ResumeOptimizationResponse, error) {
	var result *model.ResumeOptimizationResult
	var err error
	switch ma := s.mockAgent.(type) {
	case agent.ResumeOptimizer:
		result, err = ma.Optimize(ctx, input)
	default:
		return nil, fmt.Errorf("mock agent does not support Optimize")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate resume optimization: %w", err)
	}

	// Set metadata for mock result
	result.GeneratedBy = "mock_agent"
	var meta model.AgentResponseMeta
	if fbErr != nil {
		result.AgentStatus = "fallback"
		result.FallbackReason = fbErr.FallbackReason
		result.ErrorCode = fbErr.ErrorCode
		result.RepairApplied = false
		meta = model.DeriveMetaFromResult(
			model.AgentNameResumeOptimization,
			"mock_agent",
			"fallback",
			fbErr.FallbackReason,
			fbErr.ErrorCode,
			false,
		)
	} else {
		result.AgentStatus = model.AgentStatusMockSuccess
		meta = model.NewAgentResponseMeta(model.AgentNameResumeOptimization)
	}

	// Handle short resume case
	if len(req.ResumeText) < 50 {
		result.ApplyReadiness = "insufficient_information"
		result.OptimizationStopSignal = "needs_more_evidence"
	}

	now := time.Now()
	resp := &ResumeOptimizationResponse{
		GeneratedBy: "mock_agent",
		Result:      result,
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
	}

	// Save optimization snapshot (best effort)
	s.saveSnapshot(ctx, userID, req, result, resp)

	return resp, nil
}

// saveSnapshot saves the optimization result to the repository (best effort).
func (s *ResumeOptimizationService) saveSnapshot(
	ctx context.Context,
	userID uint64,
	req ResumeOptimizationRequest,
	result *model.ResumeOptimizationResult,
	resp *ResumeOptimizationResponse,
) {
	if s.optimizationRepo == nil {
		return
	}

	resultJSON, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return
	}

	jobID := uint64(0)
	snapshot := &model.ResumeOptimizationSnapshot{
		UserID:     userID,
		JobID:      jobID,
		ResultJSON: resultJSON,
	}
	if createErr := s.optimizationRepo.Create(ctx, snapshot); createErr != nil {
		log.Printf("[ResumeOptimizationService] Failed to save snapshot: %v", createErr)
	} else {
		resp.OptimizationID = snapshot.ID
	}
}

// GetLatestOptimization returns the most recent optimization result for a user.
// Returns nil if no previous results exist.
func (s *ResumeOptimizationService) GetLatestOptimization(ctx context.Context, userID uint64) (*ResumeOptimizationResponse, error) {
	if s.optimizationRepo == nil {
		return nil, nil
	}
	snapshot, err := s.optimizationRepo.GetLatestByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest optimization: %w", err)
	}
	if snapshot == nil {
		return nil, nil
	}

	var result model.ResumeOptimizationResult
	if err := json.Unmarshal(snapshot.ResultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal optimization result: %w", err)
	}

	return &ResumeOptimizationResponse{
		OptimizationID: snapshot.ID,
		GeneratedBy:    "stored",
		Result:         &result,
		CreatedAt:      snapshot.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Meta: model.NewAgentResponseMeta(model.AgentNameResumeOptimization),
	}, nil
}

// optimizeCore uses the Core Agent (LLM → Mock fallback).
func (s *ResumeOptimizationService) optimizeCore(ctx context.Context, userID uint64, req ResumeOptimizationRequest, runner agent.PipelineRunner) (*ResumeOptimizationResponse, error) {
	input := &agent.OptimizeInput{ResumeText: req.ResumeText}
	if req.TargetJob != nil {
		input.JobTitle = req.TargetJob.Title
		input.JDText = req.TargetJob.JDText
	}

	now := time.Now()
	meta := model.NewAgentResponseMeta(model.AgentNameResumeOptimization)

	// Try LLM Core Agent
	if s.llmEnabled && runner != nil {
		coreAgent := agent.NewCoreLLMAgent(runner)
		coreResult, fbErr := coreAgent.Optimize(ctx, input)
		if fbErr == nil {
			return &ResumeOptimizationResponse{
				GeneratedBy: coreResult.GeneratedBy,
				CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
				Meta: model.DeriveMetaFromResult(model.AgentNameResumeOptimization,
					coreResult.GeneratedBy, coreResult.AgentStatus, "", "", false),
				Result: &model.ResumeOptimizationResult{
					Summary: coreResult.OptimizationSummary,
					ApplyReadiness: coreResult.ApplyReadiness,
					OptimizationStopSignal: coreResult.OptimizationStopSignal,
					TopFixItems: mapCoreFixToModel(coreResult.TopFixItems),
					JDEvidenceChain: mapCoreKeywordsToEvidence(coreResult.MissingKeywords),
					VerificationNotice: coreResult.AuthenticityNotice,
					GeneratedBy: coreResult.GeneratedBy,
					AgentStatus: coreResult.AgentStatus,
				},
			}, nil
		}
		log.Printf("[optimizeCore] LLM fallback: %v", fbErr)

		// K13-Fix-1: Check fallback enable flag
		if !s.fallbackEnable {
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("Core LLM failed and fallback disabled: %v", fbErr), 500)
		}
	} else if !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "LLM is disabled and ENABLE_MOCK_AGENT=false, no generation path available", 500)
	}

	// Fallback to Core Mock
	log.Printf("[optimizeCore] use_llm=false use_mock=true")
	mockResult := agent.NewCoreMockAgent().Optimize(input)
	return &ResumeOptimizationResponse{
		GeneratedBy: "mock_agent",
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:        meta,
		Result: &model.ResumeOptimizationResult{
			Summary:           mockResult.OptimizationSummary,
			ApplyReadiness:    mockResult.ApplyReadiness,
			TopFixItems:       mapCoreFixToModel(mockResult.TopFixItems),
			VerificationNotice: mockResult.AuthenticityNotice,
			GeneratedBy:       "mock_agent",
			AgentStatus:       "mock_success",
			LongTermImprovement: mapCoreChecklistToLongTerm(mockResult.FinalChecklist),
		},
	}, nil
}

func mapCoreFixToModel(items []agent.CoreFixItem) []model.ResumeOptimizationFixItem {
	var out []model.ResumeOptimizationFixItem
	for _, it := range items {
		out = append(out, model.ResumeOptimizationFixItem{
			Issue:            it.Issue,
			WhyItMatters:     it.WhyItMatters,
			Action:           it.Action,
			ExampleReference: it.RewriteExample,
			SelfCheckQuestion: it.JDEvidence + " | " + it.ResumeGap,
		})
	}
	return out
}

func mapCoreKeywordsToEvidence(keywords []string) []model.ResumeOptimizationJDEvidenceItem {
	var out []model.ResumeOptimizationJDEvidenceItem
	for _, kw := range keywords {
		out = append(out, model.ResumeOptimizationJDEvidenceItem{
			JDRequirement:  kw,
			ResumeEvidence: "简历中暂未看到直接证据",
			EvidenceStrength: "missing",
			MatchJudgement: "missing",
		})
	}
	return out
}

func mapCoreChecklistToLongTerm(items []string) []model.ResumeOptimizationLongTermItem {
	var out []model.ResumeOptimizationLongTermItem
	for _, item := range items {
		out = append(out, model.ResumeOptimizationLongTermItem{Direction: item})
	}
	return out
}

// validateOptimizeRequest validates the resume optimization request.
func validateOptimizeRequest(req ResumeOptimizationRequest) error {
	if strings.TrimSpace(req.ResumeText) == "" {
		return apperror.New(apperror.CodeInvalidParam, "resume_text is required", 400)
	}

	if req.TargetJob == nil {
		return apperror.New(apperror.CodeInvalidParam, "target_job is required", 400)
	}

	hasTitle := strings.TrimSpace(req.TargetJob.Title) != ""
	hasJDText := strings.TrimSpace(req.TargetJob.JDText) != ""

	if !hasTitle && !hasJDText {
		return apperror.New(apperror.CodeInvalidParam, "target_job.title or target_job.jd_text is required", 400)
	}

	return nil
}

// ─── K9-3-C: New V2 Optimization API ─────────────────────────────────

type OptimizationV2Request struct {
	OptimizationMode string              `json:"optimization_mode"`
	TargetDirection  string              `json:"target_direction,omitempty"`
	JobID            uint64              `json:"job_id,omitempty"`
	GenerationTier   model.GenerationTier `json:"generation_tier,omitempty"`
}

func (s *ResumeOptimizationService) OptimizeV2(ctx context.Context, userID, resumeID uint64, req OptimizationV2Request) (*model.OptimizationResultV2, error) {
	// Validate
	mode := req.OptimizationMode
	if mode == "" { mode = "general" }
	if mode != "general" && mode != "target_direction" && mode != "target_job" {
		return nil, apperror.New(apperror.CodeInvalidParam, "optimization_mode must be general/target_direction/target_job", 400)
	}
	if mode == "target_direction" && req.TargetDirection == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "target_direction is required for target_direction mode", 400)
	}
	if mode == "target_job" && req.JobID == 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "job_id is required for target_job mode", 400)
	}

	// Get resume
	resume, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resume: %w", err)
	}
	if resume == nil {
		return nil, apperror.ErrResumeNotFound
	}

	// Get job if target_job mode
	var job *model.Job
	if mode == "target_job" && req.JobID > 0 {
		j, jErr := s.jobRepo.FindByID(ctx, req.JobID)
		if jErr != nil || j == nil {
			return nil, apperror.ErrJobNotFound
		}
		job = j
	}

	genTier := string(req.GenerationTier)
	if genTier == "" { genTier = "fast" }

	// K13-Fix-C3: Fast tier → TagEvidence engine (no LLM)
	qualityMode := req.GenerationTier == model.GenerationTierQuality
	runner := s.selectRunner(req.GenerationTier)

	if !qualityMode && s.tagEvidence != nil {
		log.Printf("[OptimizeV2] use_fast_tag_engine tier=%s resume_id=%d", genTier, resumeID)
		result, err := s.fastOptimizeWithTags(ctx, userID, resumeID, resume, mode, req.TargetDirection, req.JobID, job)
		if err == nil {
			filterEducationOptimizationAdvice(result, resume, job)
			return result, nil
		}
		log.Printf("[OptimizeV2] fast tag optimize failed: %v, falling back to mock", err)
	}

	// K13-Fix-C3: Quality tier → 3-agent pipeline
	if qualityMode && s.llmEnabled && runner != nil {
		if s.tagEvidence != nil {
			log.Printf("[OptimizeV2] use_deep_3agent tier=%s resume_id=%d", genTier, resumeID)
			result, err3a := s.optimizeWith3Agents(ctx, userID, resumeID, resume, mode, req.TargetDirection, req.JobID, job, runner)
			if err3a == nil {
				filterEducationOptimizationAdvice(result, resume, job)
				return result, nil
			}
			log.Printf("[OptimizeV2] 3-agent optimize failed: %v, falling back to old LLM", err3a)
		}
		// Fallback: old V1 LLM pipeline
		input := &agent.OptimizeInput{ResumeText: resume.RawText}
		if job != nil { input.JobTitle = job.Title; input.JDText = job.Description }
		llmAgent := agent.NewResumeOptimizationAgent(runner)
		llmResult, fbErr := llmAgent.Optimize(ctx, input)
		if fbErr == nil {
			result := mapOptimizeV1ToV2(llmResult, resumeID, mode, req.TargetDirection, req.JobID, genTier)
			filterEducationOptimizationAdvice(result, resume, job)
			s.saveOptV2Snapshot(ctx, userID, resumeID, req.JobID, mode, genTier, result.GeneratedBy, result.ProfileType, result.ProfileDominantProfile, req.TargetDirection, result)
			return result, nil
		}
		log.Printf("[OptimizeV2] LLM fallback: %v", fbErr)
		if !s.fallbackEnable {
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed: %v", fbErr), 500)
		}
	} else if qualityMode && !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "quality mode requested but LLM disabled and mock disabled", 500)
	}

	// Fallback to mock agent
	log.Printf("[OptimizeV2] use_mock_fallback tier=%s resume_id=%d", genTier, resumeID)
	v2Agent, ok := s.mockAgent.(*agent.ResumeOptimizationMockAgentV2)
	if !ok { return nil, fmt.Errorf("v2 optimization agent not available") }
	mockResult, mockErr := v2Agent.OptimizeV2(ctx, userID, resumeID, resume, mode, req.TargetDirection, req.JobID, job, genTier)
	if mockErr != nil {
		return nil, fmt.Errorf("failed to generate optimization: %w", mockErr)
	}
	filterEducationOptimizationAdvice(mockResult, resume, job)
	s.saveOptV2Snapshot(ctx, userID, resumeID, req.JobID, mode, genTier, mockResult.GeneratedBy, mockResult.ProfileType, mockResult.ProfileDominantProfile, req.TargetDirection, mockResult)
	return mockResult, nil
}

func (s *ResumeOptimizationService) saveOptV2Snapshot(ctx context.Context, userID, resumeID uint64, jobID uint64, mode, tier, generatedBy, profileType, profileDominant, targetDir string, result *model.OptimizationResultV2) {
	resultJSON, _ := json.Marshal(result)
	opt := &model.ResumeOptimizationSnapshot{
		UserID: userID, ResumeID: resumeID, JobID: jobID,
		OptimizationMode: mode, GenerationTier: tier,
		GeneratedBy: generatedBy, Status: "ready",
		ProfileType: profileType, ProfileDominantProfile: profileDominant,
		TargetDirection: targetDir, ResultJSON: resultJSON,
	}
	if createErr := s.optimizationRepo.CreateV2(ctx, opt); createErr != nil {
		log.Printf("[OptimizeV2] failed to save snapshot: %v", createErr)
	}
	result.OptimizationID = opt.ID
}

// ─── K13-Fix-C3: fastOptimizeWithTags ─────────────────────────────────

func (s *ResumeOptimizationService) fastOptimizeWithTags(ctx context.Context, userID, resumeID uint64, resume *model.Resume, mode, targetDir string, jobID uint64, job *model.Job) (*model.OptimizationResultV2, error) {
	rp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resumeID, userID)
	if err != nil {
		return nil, fmt.Errorf("fast optimize: %w", err)
	}

	var jp *model.JobTagProfile
	if job != nil { jp, _ = s.tagEvidence.EnsureJobTagProfile(ctx, jobID) }

	gaps := buildFastOptGaps(rp, jp)
	score := fastOptScore(len(gaps))

	result := &model.OptimizationResultV2{
		ResumeID:               resumeID,
		JobID:                  jobID,
		OptimizationMode:       mode,
		TargetDirection:        targetDir,
		GenerationTier:         "fast",
		GeneratedBy:            "fast_tag_engine",
		Status:                 "ready",
		ProfileType:            "fast",
		ProfileDominantProfile: rp.DominantProfile,
		OptimizationScore:      score,
		OptimizationLevel:      fastOptLevel(score),
		OneSentenceSummary:     fmt.Sprintf("基于%d项标签分析，生成优化建议", len(rp.PrimaryTags)),
		QuickFixItems:          gapsToQuickFixes(gaps, rp),
		RiskPoints:             gapsToRiskPoints(gaps),
		RewriteExamples: []model.OptimizationRewriteExample{
			{Section: "项目经历", Before: "建议优化表述", After: "使用STAR法则描述：场景-任务-行动-结果", Reason: "提升简历可读性"},
		},
		DeliveryFocus:          "检查专业技能描述是否完整、项目经历是否使用STAR法则",
		LongTermImprovements:  []model.OptimizationLongTermItem{{Direction: "持续完善", Note: "随项目和实习经历增加，定期更新简历"}},
		StructureFixItems:     []model.OptimizationFixItem{},
		ContentFixItems:       []model.OptimizationFixItem{},
		KeywordFixItems:       []model.OptimizationFixItem{},
		EvidenceFixItems:      []model.OptimizationFixItem{},
	}
	filterEducationOptimizationAdvice(result, resume, job)
	s.saveOptV2Snapshot(ctx, userID, resumeID, jobID, mode, "fast", "fast_tag_engine", "fast", rp.DominantProfile, targetDir, result)
	return result, nil
}

// ─── K13-Fix-C3: optimizeWith3Agents ──────────────────────────────────

func (s *ResumeOptimizationService) optimizeWith3Agents(ctx context.Context, userID, resumeID uint64, resume *model.Resume, mode, targetDir string, jobID uint64, job *model.Job, runner agent.PipelineRunner) (*model.OptimizationResultV2, error) {
	rp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resumeID, userID)
	if err != nil { return nil, fmt.Errorf("3agent opt: %w", err) }

	var jp *model.JobTagProfile
	if job != nil { jp, _ = s.tagEvidence.EnsureJobTagProfile(ctx, jobID) }

	final, _, err := agent.RunDeepOptimization3Agents(ctx, rp, jp, mode, runner, runner)
	if err != nil { return nil, fmt.Errorf("3agent opt: %w", err) }

	result := &model.OptimizationResultV2{
		ResumeID:               resumeID, JobID: jobID,
		OptimizationMode:       mode, TargetDirection: targetDir,
		GenerationTier:         "quality",
		GeneratedBy:            final.GeneratedBy,
		Status:                 "ready",
		ProfileType:            "quality",
		ProfileDominantProfile: rp.DominantProfile,
		OptimizationScore:      final.OptimizationScore,
		OptimizationLevel:      fastOptLevel(final.OptimizationScore),
		OneSentenceSummary:     "深度优化报告已完成",
		QuickFixItems:          final.QuickFixItems,
		StructureFixItems:      final.StructureFixItems,
		ContentFixItems:        final.ContentFixItems,
		KeywordFixItems:        final.KeywordFixItems,
		EvidenceFixItems:       final.EvidenceFixItems,
		RiskPoints:             final.RiskPoints,
		RewriteExamples:        final.RewriteExamples,
		DeliveryFocus:          "请根据深度优化建议调整简历",
		LongTermImprovements:   []model.OptimizationLongTermItem{},
	}
	filterEducationOptimizationAdvice(result, resume, job)
	s.saveOptV2Snapshot(ctx, userID, resumeID, jobID, mode, "quality", final.GeneratedBy, "quality", rp.DominantProfile, targetDir, result)
	return result, nil
}

// ─── helpers ─────────────────────────────────────────────────────────

func buildFastOptGaps(rp *model.ResumeTagProfile, jp *model.JobTagProfile) []string {
	var gaps []string
	if rp.Evidence.ProjectCount < 2 { gaps = append(gaps, "建议增加项目经历") }
	if rp.Evidence.InternshipCount < 1 { gaps = append(gaps, "建议补充实习经历") }
	if jp != nil {
		_, missing, _ := computeOptTagMatch(rp, jp)
		for _, m := range missing[:min(3, len(missing))] { gaps = append(gaps, "缺失标签："+m) }
	}
	return gaps
}

func fastOptScore(gapCount int) int {
	s := 90 - gapCount*10
	if s < 50 { s = 50 }
	return s
}

func fastOptLevel(score int) string {
	if score >= 80 { return "建议直接投递" }
	if score >= 65 { return "建议补强后投递" }
	return "需重点优化后再投递"
}

func gapsToQuickFixes(gaps []string, rp *model.ResumeTagProfile) []model.OptimizationFixItem {
	var fixes []model.OptimizationFixItem
	for i, g := range gaps {
		if i >= 4 { break }
		fixes = append(fixes, model.OptimizationFixItem{Section: "整体", Issue: g, Action: "请根据建议调整"})
	}
	return fixes
}

func gapsToRiskPoints(gaps []string) []string {
	r := make([]string, len(gaps))
	for i, g := range gaps { r[i] = g }
	return r
}

func computeOptTagMatch(rp *model.ResumeTagProfile, jp *model.JobTagProfile) (matched, missing []string, score float64) {
	resumeSet := make(map[string]bool)
	for _, t := range rp.PrimaryTags { resumeSet[t.Tag] = true }
	for _, t := range rp.SkillTags { resumeSet[t.Tag] = true }
	jobSet := make(map[string]bool)
	for _, t := range jp.PrimaryTags { jobSet[t.Tag] = true }
	for _, t := range jp.SkillTags { jobSet[t.Tag] = true }
	for _, t := range jp.RequirementTags { jobSet[t.Tag] = true }
	for jt := range jobSet {
		if resumeSet[jt] { matched = append(matched, jt); score += 1 }
	}
	return matched, missing, score
}

// mapOptimizeV1ToV2 maps V1 LLM result to V2 optimization format.
func mapOptimizeV1ToV2(v1 *model.ResumeOptimizationResult, resumeID uint64, mode, targetDir string, jobID uint64, genTier string) *model.OptimizationResultV2 {
	result := &model.OptimizationResultV2{
		ResumeID:               resumeID,
		OptimizationMode:       mode,
		GenerationTier:         genTier,
		GeneratedBy:            v1.GeneratedBy,
		Status:                 "ready",
		ProfileType:            genTier,
		ProfileDominantProfile: genTier,
		OptimizationScore:      0,
		OptimizationLevel:      v1.ApplyReadiness,
		OneSentenceSummary:     v1.Summary,
	}
	// Map fix items
	for _, f := range v1.TopFixItems {
		result.QuickFixItems = append(result.QuickFixItems, model.OptimizationFixItem{
			Section: "修复项",
			Issue:   f.Issue,
			Action:  f.Action,
		})
	}
	if mode == "target_direction" { result.TargetDirection = targetDir }
	if mode == "target_job" { result.JobID = jobID }
	return result
}

// ─── Education advice filtering ──────────────────────────────────────

// isEducationAdviceText checks whether the given text contains education-related advice.
// Only strong education keywords are matched to avoid false positives on tech terms.
func isEducationAdviceText(text string) bool {
	lower := strings.ToLower(text)
	strongKws := []string{
		"学历", "学位", "硕士", "博士", "研究生",
		"985", "211", "双一流", "名校",
		"教育经历", "教育背景", "学历层次",
		"学历要求", "学历门槛", "学历风险",
	}
	for _, kw := range strongKws {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	// "本科" only matches when combined with degree/school context (not tech background)
	if strings.Contains(lower, "本科") {
		return strings.Contains(lower, "学历") || strings.Contains(lower, "学位") ||
			strings.Contains(lower, "本科及以上") || strings.Contains(lower, "本科以上") ||
			strings.Contains(lower, "本科或以上") || strings.Contains(lower, "本科或以上学历")
	}
	if strings.Contains(lower, "学校背景") {
		return true
	}
	return false
}

// hasResumeEducation checks whether the resume has identifiable education records.
func hasResumeEducation(resume *model.Resume) bool {
	if resume == nil {
		return false
	}
	// Check Educations array
	for _, e := range resume.Educations {
		if e.School != "" || e.Degree != "" || e.Major != "" {
			return true
		}
	}
	// Check shim fields
	if resume.School != "" || resume.Degree != "" || resume.Major != "" {
		return true
	}
	// Check raw_text for education signals
	rawLower := strings.ToLower(resume.RawText)
	eduSignals := []string{"本科", "硕士", "博士", "学士", "研究生", "学历", "教育经历", "gpa"}
	for _, s := range eduSignals {
		if strings.Contains(rawLower, s) {
			return true
		}
	}
	return false
}

// shouldAllowEducationAdvice decides whether education-related advice should be allowed
// based on the job JD requirements and the resume's education data.
func shouldAllowEducationAdvice(resume *model.Resume, job *model.Job) bool {
	if job == nil {
		return false
	}
	// Build JD text from all available fields
	jdText := strings.ToLower(job.Title + " " + job.Direction + " " + job.Category + " " +
		job.Description + " " + job.Requirements)

	// Strong triggers: JD explicitly demands higher education
	strongTriggers := []string{
		"硕士及以上", "研究生及以上", "博士", "博士优先",
		"优秀硕士", "硕博", "985", "211", "双一流",
		"名校", "重点院校", "科研经历", "论文", "顶会",
		"实验室", "研究背景",
	}
	for _, kw := range strongTriggers {
		if strings.Contains(jdText, kw) {
			return true
		}
	}

	// High-education-oriented job types
	highEduRoles := []string{
		"算法研究", "ai research", "技术研究", "研究科学家",
		"机器学习研究员", "多模态算法", "nlp 算法研究",
		"计算机视觉研究", "高性能计算研究", "实验室研究",
	}
	for _, role := range highEduRoles {
		if strings.Contains(jdText, role) {
			return true
		}
	}

	// If JD only has generic "本科及以上", still check if resume has NO education at all
	if strings.Contains(jdText, "本科及以上") || strings.Contains(jdText, "本科以上") ||
		strings.Contains(jdText, "本科或以上") {
		// Only allow if resume has zero education evidence
		if !hasResumeEducation(resume) {
			return true
		}
		return false
	}

	// No valid trigger found — do not allow education advice
	return false
}

// filterEducationOptimizationAdvice removes education-related advice from the
// optimization result when shouldAllowEducationAdvice returns false.
func filterEducationOptimizationAdvice(result *model.OptimizationResultV2, resume *model.Resume, job *model.Job) {
	if result == nil {
		return
	}
	if shouldAllowEducationAdvice(resume, job) {
		return
	}

	// Filter RiskPoints
	result.RiskPoints = filterSliceEdu(result.RiskPoints)

	// Filter fix items
	result.QuickFixItems = filterFixItemsEdu(result.QuickFixItems)
	result.StructureFixItems = filterFixItemsEdu(result.StructureFixItems)
	result.ContentFixItems = filterFixItemsEdu(result.ContentFixItems)
	result.KeywordFixItems = filterFixItemsEdu(result.KeywordFixItems)
	result.EvidenceFixItems = filterFixItemsEdu(result.EvidenceFixItems)

	// Filter RewriteExamples
	result.RewriteExamples = filterRewriteExamplesEdu(result.RewriteExamples)

	// Filter LongTermImprovements
	result.LongTermImprovements = filterLongTermEdu(result.LongTermImprovements)

	// Filter DeliveryFocus
	if isEducationAdviceText(result.DeliveryFocus) {
		result.DeliveryFocus = "优先强化与目标岗位相关的项目经历、实习经历、技术栈关键词和量化成果表达。"
	}
}

func filterSliceEdu(items []string) []string {
	if len(items) == 0 {
		return items
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if !isEducationAdviceText(item) {
			out = append(out, item)
		}
	}
	return out
}

func filterFixItemsEdu(items []model.OptimizationFixItem) []model.OptimizationFixItem {
	if len(items) == 0 {
		return items
	}
	out := make([]model.OptimizationFixItem, 0, len(items))
	for _, item := range items {
		combined := item.Section + " " + item.Issue + " " + item.Action
		if !isEducationAdviceText(combined) {
			out = append(out, item)
		}
	}
	return out
}

func filterRewriteExamplesEdu(items []model.OptimizationRewriteExample) []model.OptimizationRewriteExample {
	if len(items) == 0 {
		return items
	}
	out := make([]model.OptimizationRewriteExample, 0, len(items))
	for _, item := range items {
		combined := item.Section + " " + item.Before + " " + item.After + " " + item.Reason
		if !isEducationAdviceText(combined) {
			out = append(out, item)
		}
	}
	return out
}

func filterLongTermEdu(items []model.OptimizationLongTermItem) []model.OptimizationLongTermItem {
	if len(items) == 0 {
		return items
	}
	out := make([]model.OptimizationLongTermItem, 0, len(items))
	for _, item := range items {
		combined := item.Direction + " " + item.Note
		if !isEducationAdviceText(combined) {
			out = append(out, item)
		}
	}
	return out
}

// GetLatestOptimizationV2 returns the latest ready optimization for user+resume+mode.
func (s *ResumeOptimizationService) GetLatestOptimizationV2(ctx context.Context, userID, resumeID uint64, mode string, targetDir string, jobID uint64) (*model.OptimizationLatestResult, error) {
	if resumeID <= 0 { return nil, apperror.ErrInvalidParam }
	if mode == "" { mode = "general" }
	if mode != "general" && mode != "target_direction" && mode != "target_job" {
		return nil, apperror.New(apperror.CodeInvalidParam, "optimization_mode must be general/target_direction/target_job", 400)
	}
	if mode == "target_direction" && targetDir == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "target_direction is required", 400)
	}
	if mode == "target_job" && jobID == 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "job_id is required", 400)
	}
	// Verify resume belongs to user
	resume, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil { return nil, fmt.Errorf("failed to verify resume: %w", err) }
	if resume == nil { return nil, apperror.ErrResumeNotFound }

	// Verify job exists for target_job mode
	if mode == "target_job" && jobID > 0 {
		j, jErr := s.jobRepo.FindByID(ctx, jobID)
		if jErr != nil { return nil, fmt.Errorf("failed to verify job: %w", jErr) }
		if j == nil { return nil, apperror.ErrJobNotFound }
	}

	latest, err := s.optimizationRepo.FindLatestByMode(ctx, userID, resumeID, mode, targetDir, jobID)
	if err != nil { return nil, fmt.Errorf("failed to get latest optimization: %w", err) }
	if latest == nil { return &model.OptimizationLatestResult{HasResult: false}, nil }

	var resultData interface{}
	if len(latest.ResultJSON) > 0 {
		if err := json.Unmarshal(latest.ResultJSON, &resultData); err != nil {
			return nil, fmt.Errorf("failed to parse saved result: %w", err)
		}
	}

	return &model.OptimizationLatestResult{
		HasResult:             true,
		OptimizationID:        latest.ID,
		ResumeID:              latest.ResumeID,
		JobID:                 latest.JobID,
		OptimizationMode:      mode,
		TargetDirection:       latest.TargetDirection,
		GenerationTier:        latest.GenerationTier,
		GeneratedBy:           latest.GeneratedBy,
		Status:                latest.Status,
		ProfileType:           latest.ProfileType,
		ProfileDominantProfile: latest.ProfileDominantProfile,
		CreatedAt:             latest.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Result:                resultData,
	}, nil
}
