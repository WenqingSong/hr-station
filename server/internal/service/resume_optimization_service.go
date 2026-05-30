package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ResumeOptimizationRequest represents a request for resume optimization.
type ResumeOptimizationRequest struct {
	ResumeText      string                                 `json:"resume_text"`
	TargetJob       *model.ResumeOptimizationTargetJob     `json:"target_job,omitempty"`
	TargetSection   string                                 `json:"target_section,omitempty"`
	UserIntention   *model.ResumeOptimizationUserIntention `json:"user_intention,omitempty"`
}

// ResumeOptimizationResponse represents the response for resume optimization.
type ResumeOptimizationResponse struct {
	OptimizationID uint64                              `json:"optimization_id"`
	GeneratedBy    string                              `json:"generated_by"`
	Result         *model.ResumeOptimizationResult     `json:"result"`
	CreatedAt      string                              `json:"created_at"`
}

// ResumeOptimizationService handles resume optimization business logic.
// Current stage: Mock Agent only. LLM Pipeline integration will be added in 4-3.8.
type ResumeOptimizationService struct {
	optimizationRepo *repository.ResumeOptimizationRepository
	mockAgent        *agent.ResumeOptimizationMockAgent
}

// NewResumeOptimizationService creates a new ResumeOptimizationService.
func NewResumeOptimizationService(
	optimizationRepo *repository.ResumeOptimizationRepository,
	mockAgent *agent.ResumeOptimizationMockAgent,
) *ResumeOptimizationService {
	return &ResumeOptimizationService{
		optimizationRepo: optimizationRepo,
		mockAgent:        mockAgent,
	}
}

// Optimize generates a resume optimization for the given user and request.
// Current stage: always uses Mock Agent. LLM Pipeline will be added later.
func (s *ResumeOptimizationService) Optimize(ctx context.Context, userID uint64, req ResumeOptimizationRequest) (*ResumeOptimizationResponse, error) {
	// Validate request
	if err := validateOptimizeRequest(req); err != nil {
		return nil, err
	}

	// Build agent input
	input := &agent.OptimizeInput{
		ResumeText:   req.ResumeText,
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

	// Call Mock Agent
	result, err := s.mockAgent.Optimize(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate resume optimization: %w", err)
	}

	// Set metadata for mock result
	result.GeneratedBy = "mock_agent"
	result.AgentStatus = "mock_success"

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
	}

	// Save optimization snapshot (best effort)
	if s.optimizationRepo != nil {
		resultJSON, marshalErr := json.Marshal(result)
		if marshalErr == nil {
			jobID := uint64(0)
			snapshot := &model.ResumeOptimizationSnapshot{
				UserID:     userID,
				JobID:      jobID,
				ResultJSON: resultJSON,
			}
			if createErr := s.optimizationRepo.Create(ctx, snapshot); createErr != nil {
				// Log but don't fail the response if snapshot save fails
				log.Printf("[ResumeOptimizationService] Failed to save snapshot: %v", createErr)
			} else {
				resp.OptimizationID = snapshot.ID
			}
		}
	}

	return resp, nil
}

// validateOptimizeRequest validates the resume optimization request.
func validateOptimizeRequest(req ResumeOptimizationRequest) error {
	if strings.TrimSpace(req.ResumeText) == "" {
		return errors.New("resume_text is required")
	}

	if req.TargetJob == nil {
		return errors.New("target_job is required")
	}

	hasTitle := strings.TrimSpace(req.TargetJob.Title) != ""
	hasJDText := strings.TrimSpace(req.TargetJob.JDText) != ""

	if !hasTitle && !hasJDText {
		return errors.New("target_job.title or target_job.jd_text is required")
	}

	return nil
}
