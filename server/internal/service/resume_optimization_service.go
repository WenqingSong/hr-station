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
}

// NewResumeOptimizationService creates a new ResumeOptimizationService.
// If both runners are nil or llmEnabled is false, it falls back to Mock Agent only.
func NewResumeOptimizationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	optimizationRepo *repository.ResumeOptimizationRepository,
	mockAgent interface{},
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
) *ResumeOptimizationService {
	return &ResumeOptimizationService{
		resumeRepo:       resumeRepo,
		jobRepo:          jobRepo,
		optimizationRepo: optimizationRepo,
		mockAgent:        mockAgent,
		fastRunner:       fastRunner,
		qualityRunner:    qualityRunner,
		llmEnabled:       llmEnabled,
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
		resp, fallbackErr := s.optimizeWithLLM(ctx, userID, input, req, runner)
		if fallbackErr == nil {
			return resp, nil
		}

		// Log the fallback reason
		log.Printf("[ResumeOptimizationService] LLM pipeline fallback: %v", fallbackErr)

		// Fall back to Mock Agent
		return s.optimizeWithMock(ctx, userID, input, req, fallbackErr)
	}

	// LLM disabled — use Mock Agent directly
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
	}

	// Fallback to Core Mock
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

	// Get V2 agent
	v2Agent, ok := s.mockAgent.(*agent.ResumeOptimizationMockAgentV2)
	if !ok {
		// Fallback to old result mapper
		return nil, fmt.Errorf("v2 optimization agent not available")
	}

	result, err := v2Agent.OptimizeV2(ctx, userID, resumeID, resume, mode, req.TargetDirection, req.JobID, job, genTier)
	if err != nil {
		return nil, fmt.Errorf("failed to generate optimization: %w", err)
	}

	// Save to DB
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	opt := &model.ResumeOptimizationSnapshot{
		UserID:                 userID,
		ResumeID:               resumeID,
		JobID:                  req.JobID,
		OptimizationMode:       mode,
		GenerationTier:         genTier,
		GeneratedBy:            result.GeneratedBy,
		Status:                 "ready",
		ProfileType:            result.ProfileType,
		ProfileDominantProfile: result.ProfileDominantProfile,
		TargetDirection:        req.TargetDirection,
		ResultJSON:             resultJSON,
	}
	if createErr := s.optimizationRepo.CreateV2(ctx, opt); createErr != nil {
		log.Printf("[OptimizeV2] failed to save snapshot: %v", createErr)
	}
	result.OptimizationID = opt.ID

	return result, nil
}
