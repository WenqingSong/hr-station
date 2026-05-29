package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// MatchRequest represents a request for job match analysis.
type MatchRequest struct {
	ResumeVersionNo int `json:"resume_version_no"`
}

// MatchResponse represents the response for job match analysis.
type MatchResponse struct {
	AnalysisID    uint64                     `json:"analysis_id"`
	GeneratedBy   string                     `json:"generated_by"`
	Result        *model.MatchAnalysisResult `json:"result"`
	AnalysisBasis *model.AnalysisBasis       `json:"analysis_basis"`
	CreatedAt     string                     `json:"created_at"`
}

// MatchService handles job match analysis business logic.
type MatchService struct {
	resumeRepo *repository.ResumeRepository
	jobRepo    *repository.JobRepository
	matchRepo  *repository.MatchRepository
	mockAgent  *agent.MatchMockAgent
}

// NewMatchService creates a new MatchService.
func NewMatchService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	matchRepo *repository.MatchRepository,
	mockAgent *agent.MatchMockAgent,
) *MatchService {
	return &MatchService{
		resumeRepo: resumeRepo,
		jobRepo:    jobRepo,
		matchRepo:  matchRepo,
		mockAgent:  mockAgent,
	}
}

// Analyze generates a match analysis for the given user and job.
func (s *MatchService) Analyze(ctx context.Context, userID uint64, jobID uint64, req MatchRequest) (*MatchResponse, error) {
	// Get resume version
	var resume *model.ResumeVersion
	var err error

	if req.ResumeVersionNo > 0 {
		resume, err = s.resumeRepo.GetByUserIDAndVersionNo(ctx, userID, req.ResumeVersionNo)
		if err != nil {
			return nil, fmt.Errorf("failed to get resume version: %w", err)
		}
		if resume == nil {
			return nil, errors.New("resume version not found")
		}
	} else {
		resume, err = s.resumeRepo.GetCurrentByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current resume: %w", err)
		}
		if resume == nil {
			return nil, errors.New("resume not found")
		}
	}

	// Get job
	job, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, errors.New("job not found")
	}

	// Generate match analysis via mock agent
	result, err := s.mockAgent.Analyze(resume, job)
	if err != nil {
		return nil, fmt.Errorf("failed to generate match analysis: %w", err)
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
	}

	// Save match analysis snapshot
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal match result: %w", err)
	}

	analysis := &model.MatchAnalysis{
		UserID:          userID,
		ResumeVersionID: resume.ID,
		JobID:           jobID,
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
