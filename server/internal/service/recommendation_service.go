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

// RecommendationRequest represents a request for job recommendations.
type RecommendationRequest struct {
	ResumeID uint64 `json:"resume_id"`
	Limit           int                `json:"limit"`
	GenerationTier  model.GenerationTier `json:"generation_tier,omitempty"`
}

// RecommendationResponse represents the response for job recommendations.
type RecommendationResponse struct {
	ResumeVersionID uint64                        `json:"resume_version_id"`
	ResumeID uint64 `json:"resume_id"`
	GeneratedBy     string                        `json:"generated_by"`     // deprecated: use meta.generated_by
	RecommendedJobs []*model.RecommendedJob       `json:"recommended_jobs"` // deprecated: use result.recommended_jobs
	Result          *RecommendationBusinessResult `json:"result"`
	Summary         string                        `json:"summary"`
	CreatedAt       string                        `json:"created_at"`
	Meta            model.AgentResponseMeta       `json:"meta"`
}

// RecommendationBusinessResult wraps the business result for recommendation.
type RecommendationBusinessResult struct {
	RecommendedJobs []*model.RecommendedJob `json:"recommended_jobs"`
}

// RecommendationService handles job recommendation business logic.
// It supports LLM Pipeline priority with Mock Agent fallback.
// Supports generation_tier: "fast" uses fastRunner, "quality" uses qualityRunner.
type RecommendationService struct {
	resumeRepo         *repository.ResumeRepository
	jobRepo            *repository.JobRepository
	recommendationRepo *repository.RecommendationRepository
	mockAgent          *agent.RecommendationMockAgent
	fastRunner         agent.PipelineRunner
	qualityRunner      agent.PipelineRunner
	llmEnabled         bool
}

// NewRecommendationService creates a new RecommendationService.
// If both runners are nil or llmEnabled is false, it falls back to Mock Agent only.
func NewRecommendationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	recommendationRepo *repository.RecommendationRepository,
	mockAgent *agent.RecommendationMockAgent,
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
) *RecommendationService {
	return &RecommendationService{
		resumeRepo:         resumeRepo,
		jobRepo:            jobRepo,
		recommendationRepo: recommendationRepo,
		mockAgent:          mockAgent,
		fastRunner:         fastRunner,
		qualityRunner:      qualityRunner,
		llmEnabled:         llmEnabled,
	}
}

// SelectRunner returns the appropriate pipeline runner for the given tier.
func (s *RecommendationService) selectRunner(tier model.GenerationTier) agent.PipelineRunner {
	// tier validated by handler
	if tier == model.GenerationTierQuality && s.qualityRunner != nil {
		return s.qualityRunner
	}
	return s.fastRunner
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

	// Select runner based on generation_tier
	runner := s.selectRunner(req.GenerationTier)
	log.Printf("[RecommendationService] generation_tier=%s", string(req.GenerationTier))

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
		return nil, apperror.ErrJobNotFound
	}

	// Try LLM Pipeline if enabled
	if s.llmEnabled && runner != nil {
		resp, fallbackErr := s.recommendWithLLM(ctx, userID, resume, jobs, req.Limit, runner)
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

// recommendWithLLM calls the RecommendationAgent (LLM Pipeline) and builds the response.
func (s *RecommendationService) recommendWithLLM(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int, runner agent.PipelineRunner) (*RecommendationResponse, *agent.RecFallbackError) {
	llmAgent := agent.NewRecommendationAgent(runner)
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

	// Derive meta from the first recommended job's metadata, or default to LLM success
	var meta model.AgentResponseMeta
	if len(jobPtrs) > 0 && jobPtrs[0] != nil {
		meta = model.DeriveMetaFromResult(
			model.AgentNameRecommendation,
			jobPtrs[0].GeneratedBy,
			jobPtrs[0].AgentStatus,
			jobPtrs[0].FallbackReason,
			jobPtrs[0].ErrorCode,
			jobPtrs[0].RepairApplied,
		)
	} else {
		meta = model.NewAgentResponseMeta(model.AgentNameRecommendation)
		meta.GeneratedBy = model.GeneratedByLLMAgent
		meta.AgentStatus = model.AgentStatusSuccess
	}

	recResult := &RecommendationBusinessResult{RecommendedJobs: jobPtrs}

	resp := &RecommendationResponse{
		ResumeID: resume.ID,
		GeneratedBy:     "llm_agent",
		RecommendedJobs: jobPtrs,
		Result:          recResult,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:            meta,
	}

	// Save recommendation result snapshot (same as mock path)
	resultJSON, marshalErr := json.Marshal(recommendedJobs)
	if marshalErr == nil {
		recResult := &model.RecommendationResult{
			UserID:          userID,
			ResumeID: resume.ID,
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
	var meta model.AgentResponseMeta
	for i := range recommendedJobs {
		recommendedJobs[i].GeneratedBy = "mock_agent"
		if fbErr != nil {
			recommendedJobs[i].AgentStatus = "fallback"
			recommendedJobs[i].FallbackReason = fbErr.FallbackReason
			recommendedJobs[i].ErrorCode = fbErr.ErrorCode
			recommendedJobs[i].RepairApplied = false
		} else {
			recommendedJobs[i].AgentStatus = model.AgentStatusMockSuccess
		}
	}

	// Build meta from first job or fallback info
	if fbErr != nil {
		meta = model.DeriveMetaFromResult(
			model.AgentNameRecommendation,
			"mock_agent",
			"fallback",
			fbErr.FallbackReason,
			fbErr.ErrorCode,
			false,
		)
	} else {
		meta = model.NewAgentResponseMeta(model.AgentNameRecommendation)
	}

	// Convert to pointer slice for response
	jobPtrs := make([]*model.RecommendedJob, len(recommendedJobs))
	for i := range recommendedJobs {
		jobPtrs[i] = &recommendedJobs[i]
	}

	summary := generateSummary(resume, jobPtrs)
	now := time.Now()

	recResult := &RecommendationBusinessResult{RecommendedJobs: jobPtrs}

	resp := &RecommendationResponse{
		ResumeID: resume.ID,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: jobPtrs,
		Result:          recResult,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:            meta,
	}

	// Save recommendation result snapshot
	resultJSON, err := json.Marshal(recommendedJobs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recommendation result: %w", err)
	}

	recSnapshot := &model.RecommendationResult{
		UserID:          userID,
		ResumeID: resume.ID,
		ResultJSON:      resultJSON,
	}

	if err := s.recommendationRepo.Create(ctx, recSnapshot); err != nil {
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

	targetDesc := fmt.Sprintf("基于简历「%s」，", resume.Name)

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
