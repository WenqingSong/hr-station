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
	ResumeText    string                                 `json:"resume_text"`
	TargetJob     *model.ResumeOptimizationTargetJob     `json:"target_job,omitempty"`
	TargetSection string                                 `json:"target_section,omitempty"`
	UserIntention *model.ResumeOptimizationUserIntention `json:"user_intention,omitempty"`
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
type ResumeOptimizationService struct {
	optimizationRepo *repository.ResumeOptimizationRepository
	mockAgent        *agent.ResumeOptimizationMockAgent
	pipelineRunner   agent.PipelineRunner
	llmEnabled       bool
}

// NewResumeOptimizationService creates a new ResumeOptimizationService.
// If pipelineRunner is nil or llmEnabled is false, it falls back to Mock Agent only.
func NewResumeOptimizationService(
	optimizationRepo *repository.ResumeOptimizationRepository,
	mockAgent *agent.ResumeOptimizationMockAgent,
	pipelineRunner agent.PipelineRunner,
	llmEnabled bool,
) *ResumeOptimizationService {
	return &ResumeOptimizationService{
		optimizationRepo: optimizationRepo,
		mockAgent:        mockAgent,
		pipelineRunner:   pipelineRunner,
		llmEnabled:       llmEnabled,
	}
}

// Optimize generates a resume optimization for the given user and request.
// When LLM is enabled and Pipeline Runner is available, it uses the LLM Pipeline
// first. If the Pipeline returns fallback_required or mapping fails, it falls
// back to the Mock Agent. When LLM is disabled, it uses the Mock Agent directly.
func (s *ResumeOptimizationService) Optimize(ctx context.Context, userID uint64, req ResumeOptimizationRequest) (*ResumeOptimizationResponse, error) {
	// Validate request
	if err := validateOptimizeRequest(req); err != nil {
		return nil, err
	}

	// Build agent input
	input := &agent.OptimizeInput{
		ResumeText:    req.ResumeText,
		TargetSection: req.TargetSection,
	}

	if req.TargetJob != nil {
		input.JobTitle = req.TargetJob.Title
		input.JDText = req.TargetJob.JDText
	}

	if req.UserIntention != nil {
		input.TargetRoles = req.UserIntention.TargetRoles
		input.ExpectedLocations = req.UserIntention.ExpectedLocations
		input.AcceptRelocation = req.UserIntention.AcceptRelocation
	}

	// Try LLM Pipeline if enabled
	if s.llmEnabled && s.pipelineRunner != nil {
		resp, fallbackErr := s.optimizeWithLLM(ctx, userID, input, req)
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
func (s *ResumeOptimizationService) optimizeWithLLM(ctx context.Context, userID uint64, input *agent.OptimizeInput, req ResumeOptimizationRequest) (*ResumeOptimizationResponse, *agent.ResumeOptFallbackError) {
	llmAgent := agent.NewResumeOptimizationAgent(s.pipelineRunner)
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
	result, err := s.mockAgent.Optimize(ctx, input)
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
