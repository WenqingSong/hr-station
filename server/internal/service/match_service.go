package service

import (
	"context"
	"encoding/json"
	apperror "offer-catcher/server/internal/errors"
	"fmt"
	"log"
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
// It supports LLM Pipeline priority with Mock Agent fallback.
// Supports generation_tier: "fast" uses fastRunner, "quality" uses qualityRunner.
type MatchService struct {
	resumeRepo    *repository.ResumeRepository
	jobRepo       *repository.JobRepository
	matchRepo     *repository.MatchRepository
	mockAgent     *agent.MatchMockAgent
	fastRunner    agent.PipelineRunner
	qualityRunner agent.PipelineRunner
	llmEnabled    bool
}

// NewMatchService creates a new MatchService.
// If both runners are nil or llmEnabled is false, it falls back to Mock Agent only.
func NewMatchService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	matchRepo *repository.MatchRepository,
	mockAgent *agent.MatchMockAgent,
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
) *MatchService {
	return &MatchService{
		resumeRepo:    resumeRepo,
		jobRepo:       jobRepo,
		matchRepo:     matchRepo,
		mockAgent:     mockAgent,
		fastRunner:    fastRunner,
		qualityRunner: qualityRunner,
		llmEnabled:    llmEnabled,
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
	runner := s.selectRunner(req.GenerationTier)
	log.Printf("[MatchService] generation_tier=%s", string(req.GenerationTier))

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

	// Try LLM Pipeline if enabled
	if s.llmEnabled && runner != nil {
		resp, fallbackErr := s.analyzeWithLLM(ctx, resume, job, runner)
		if fallbackErr == nil {
			return resp, nil
		}

		// Log the fallback reason
		log.Printf("[MatchService] LLM pipeline fallback: %v", fallbackErr)

		// Fall back to Mock Agent
		return s.analyzeWithMock(ctx, userID, resume, job, fallbackErr)
	}

	// LLM disabled — use Mock Agent directly
	return s.analyzeWithMock(ctx, userID, resume, job, nil)
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
			UserID:          resume.UserID,
			ResumeID: resume.ID,
			JobID:           job.ID,
			Score:           result.MatchScore,
			ResultJSON:      resultJSON,
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
// If fbErr is non-nil, it means we're in fallback mode from a failed LLM attempt.
func (s *MatchService) analyzeWithMock(
	ctx context.Context,
	userID uint64,
	resume *model.ResumeVersion,
	job *model.Job,
	fbErr *agent.MatchFallbackError,
) (*MatchResponse, error) {
	result, err := s.mockAgent.Analyze(ctx, resume, job)
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
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
			model.AgentNameMatchAnalysis,
			"mock_agent",
			"fallback",
			fbErr.FallbackReason,
			fbErr.ErrorCode,
			false,
		)
	} else {
		result.AgentStatus = model.AgentStatusMockSuccess
		meta = model.NewAgentResponseMeta(model.AgentNameMatchAnalysis)
	}

	now := time.Now()
	resp := &MatchResponse{
		GeneratedBy: "mock_agent",
		Result:      result,
		AnalysisBasis: &model.AnalysisBasis{
			JobDescriptionUsed:  job.Description != "",
			JobRequirementsUsed: job.Requirements != "",
			DataQuality:         job.DataQuality,
		},
		CreatedAt: now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:      meta,
	}

	// Save match analysis snapshot
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal match result: %w", err)
	}

	analysis := &model.MatchAnalysis{
		UserID:          userID,
		ResumeID: resume.ID,
		JobID:           job.ID,
		Score:           result.MatchScore,
		ResultJSON:      resultJSON,
	}

	if err := s.matchRepo.Create(ctx, analysis); err != nil {
		// Log but don't fail the response if snapshot save fails
		_ = err
	}

	resp.AnalysisID = analysis.ID

	return resp, nil
}
