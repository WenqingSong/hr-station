package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// RecommendationRequest represents a request for job recommendations.
type RecommendationRequest struct {
	ResumeVersionNo int `json:"resume_version_no"`
	Limit           int `json:"limit"`
}

// RecommendationResponse represents the response for job recommendations.
type RecommendationResponse struct {
	ResumeVersionID uint64                  `json:"resume_version_id"`
	ResumeVersionNo int                     `json:"resume_version_no"`
	GeneratedBy     string                  `json:"generated_by"`
	RecommendedJobs []*model.RecommendedJob `json:"recommended_jobs"`
	Summary         string                  `json:"summary"`
	CreatedAt       string                  `json:"created_at"`
}

// RecommendationService handles job recommendation business logic.
// It supports LLM Pipeline priority with Mock Agent fallback.
type RecommendationService struct {
	resumeRepo         *repository.ResumeRepository
	jobRepo            *repository.JobRepository
	recommendationRepo *repository.RecommendationRepository
	mockAgent          *agent.RecommendationMockAgent
	pipelineRunner     agent.PipelineRunner
	llmEnabled         bool
}

// NewRecommendationService creates a new RecommendationService.
// If pipelineRunner is nil or llmEnabled is false, it falls back to Mock Agent only.
func NewRecommendationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	recommendationRepo *repository.RecommendationRepository,
	mockAgent *agent.RecommendationMockAgent,
	pipelineRunner agent.PipelineRunner,
	llmEnabled bool,
) *RecommendationService {
	return &RecommendationService{
		resumeRepo:         resumeRepo,
		jobRepo:            jobRepo,
		recommendationRepo: recommendationRepo,
		mockAgent:          mockAgent,
		pipelineRunner:     pipelineRunner,
		llmEnabled:         llmEnabled,
	}
}

// Recommend generates job recommendations for the given user based on their resume.
// When LLM is enabled and Pipeline Runner is available, it uses the LLM Pipeline
// first. If the Pipeline returns fallback_required or mapping fails, it falls
// back to the Mock Agent. When LLM is disabled, it uses the Mock Agent directly.
func (s *RecommendationService) Recommend(ctx context.Context, userID uint64, req RecommendationRequest) (*RecommendationResponse, error) {
	// Apply defaults
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 5 {
		req.Limit = 5
	}

	// Get resume version
	resume, err := s.getResume(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Get all jobs as candidates (local_mock + tencent_join, prioritizing full_jd and featured)
	jobs, err := s.jobRepo.ListRecommendCandidates(ctx, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil, errors.New("no jobs available")
	}

	// Try LLM Pipeline if enabled
	if s.llmEnabled && s.pipelineRunner != nil {
		resp, fallbackErr := s.recommendWithLLM(ctx, userID, resume, jobs, req.Limit)
		if fallbackErr == nil {
			return resp, nil
		}

		// Log the fallback reason
		log.Printf("[RecommendationService] LLM pipeline fallback: %v", fallbackErr)

		// Fall back to Mock Agent
		return s.recommendWithMock(ctx, userID, resume, jobs, req.Limit, fallbackErr)
	}

	// LLM disabled — use Mock Agent directly
	return s.recommendWithMock(ctx, userID, resume, jobs, req.Limit, nil)
}

// getResume retrieves the resume version based on the request.
func (s *RecommendationService) getResume(ctx context.Context, userID uint64, req RecommendationRequest) (*model.ResumeVersion, error) {
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
	return resume, nil
}

// recommendWithLLM calls the RecommendationAgent (LLM Pipeline) and builds the response.
func (s *RecommendationService) recommendWithLLM(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int) (*RecommendationResponse, *agent.RecFallbackError) {
	llmAgent := agent.NewRecommendationAgent(s.pipelineRunner)
	recommendedJobs, err := llmAgent.Recommend(ctx, resume, jobs)
	if err != nil {
		if fbErr, ok := err.(*agent.RecFallbackError); ok {
			return nil, fbErr
		}
		// Unexpected error — wrap as fallback
		return nil, &agent.RecFallbackError{
			FallbackReason: "llm_agent_error",
			ErrorCode:      "LLM_AGENT_ERROR",
			ErrorMessage:   err.Error(),
		}
	}

	// Apply limit
	if len(recommendedJobs) > limit {
		recommendedJobs = recommendedJobs[:limit]
	}

	// Convert to pointer slice for response
	jobPtrs := make([]*model.RecommendedJob, len(recommendedJobs))
	for i := range recommendedJobs {
		jobPtrs[i] = &recommendedJobs[i]
	}

	summary := generateSummary(resume, jobPtrs)
	now := time.Now()

	resp := &RecommendationResponse{
		ResumeVersionID: resume.ID,
		ResumeVersionNo: resume.VersionNo,
		GeneratedBy:     "llm_agent",
		RecommendedJobs: jobPtrs,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Save recommendation result snapshot (same as mock path)
	resultJSON, marshalErr := json.Marshal(recommendedJobs)
	if marshalErr == nil {
		recResult := &model.RecommendationResult{
			UserID:          userID,
			ResumeVersionID: resume.ID,
			ResultJSON:      resultJSON,
		}
		if createErr := s.recommendationRepo.Create(ctx, recResult); createErr != nil {
			// Log but don't fail the response if snapshot save fails
			_ = createErr
		}
	}

	return resp, nil
}

// recommendWithMock calls the Mock Agent and builds the response.
// If fbErr is non-nil, it means we're in fallback mode from a failed LLM attempt.
func (s *RecommendationService) recommendWithMock(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int, fbErr *agent.RecFallbackError) (*RecommendationResponse, error) {
	recommendedJobs, err := s.mockAgent.Recommend(ctx, resume, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Apply limit
	if len(recommendedJobs) > limit {
		recommendedJobs = recommendedJobs[:limit]
	}

	// Set metadata for mock results
	for i := range recommendedJobs {
		recommendedJobs[i].GeneratedBy = "mock_agent"
		if fbErr != nil {
			recommendedJobs[i].AgentStatus = "fallback"
			recommendedJobs[i].FallbackReason = fbErr.FallbackReason
			recommendedJobs[i].ErrorCode = fbErr.ErrorCode
			recommendedJobs[i].RepairApplied = false
		} else {
			recommendedJobs[i].AgentStatus = "mock"
		}
	}

	// Convert to pointer slice for response
	jobPtrs := make([]*model.RecommendedJob, len(recommendedJobs))
	for i := range recommendedJobs {
		jobPtrs[i] = &recommendedJobs[i]
	}

	summary := generateSummary(resume, jobPtrs)
	now := time.Now()

	resp := &RecommendationResponse{
		ResumeVersionID: resume.ID,
		ResumeVersionNo: resume.VersionNo,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: jobPtrs,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Save recommendation result snapshot
	resultJSON, err := json.Marshal(recommendedJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recommendation result: %w", err)
	}

	recResult := &model.RecommendationResult{
		UserID:          userID,
		ResumeVersionID: resume.ID,
		ResultJSON:      resultJSON,
	}

	if err := s.recommendationRepo.Create(ctx, recResult); err != nil {
		// Log but don't fail the response if snapshot save fails
		_ = err
	}

	return resp, nil
}

// generateSummary creates a summary of the recommendations.
func generateSummary(resume *model.ResumeVersion, jobs []*model.RecommendedJob) string {
	if len(jobs) == 0 {
		return "暂无合适的推荐岗位。"
	}

	var categories []string
	seen := make(map[string]bool)
	hasTencent := false
	for _, j := range jobs {
		if !seen[j.Category] {
			categories = append(categories, j.Category)
			seen[j.Category] = true
		}
		if j.ExternalSource == "tencent_join" {
			hasTencent = true
		}
	}

	targetDesc := ""
	if resume.TargetRole != "" {
		targetDesc = fmt.Sprintf("基于当前简历版本（目标岗位：%s），", resume.TargetRole)
	} else {
		targetDesc = "基于当前简历版本，"
	}

	sourceDesc := "系统优先推荐%s方向岗位。"
	if hasTencent {
		sourceDesc = "系统从本地岗位库与腾讯招聘公开岗位中优先推荐%s方向岗位。"
	}

	return fmt.Sprintf(targetDesc+sourceDesc, stringsJoin(categories, "、"))
}

func stringsJoin(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	result := ss[0]
	for i := 1; i < len(ss); i++ {
		result += sep + ss[i]
	}
	return result
}
