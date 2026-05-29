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
type RecommendationService struct {
	resumeRepo         *repository.ResumeRepository
	jobRepo            *repository.JobRepository
	recommendationRepo *repository.RecommendationRepository
	mockAgent          *agent.RecommendationMockAgent
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	recommendationRepo *repository.RecommendationRepository,
	mockAgent *agent.RecommendationMockAgent,
) *RecommendationService {
	return &RecommendationService{
		resumeRepo:         resumeRepo,
		jobRepo:            jobRepo,
		recommendationRepo: recommendationRepo,
		mockAgent:          mockAgent,
	}
}

// Recommend generates job recommendations for the given user based on their resume.
func (s *RecommendationService) Recommend(ctx context.Context, userID uint64, req RecommendationRequest) (*RecommendationResponse, error) {
	// Apply defaults
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 5 {
		req.Limit = 5
	}

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

	// Get all jobs as candidates (local_mock + tencent_join, prioritizing full_jd and featured)
	jobs, err := s.jobRepo.ListRecommendCandidates(ctx, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil, errors.New("no jobs available")
	}

	// Generate recommendations via mock agent
	recommendedJobs, err := s.mockAgent.Recommend(resume, jobs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Apply limit
	if len(recommendedJobs) > req.Limit {
		recommendedJobs = recommendedJobs[:req.Limit]
	}

	// Convert to pointer slice for response
	jobPtrs := make([]*model.RecommendedJob, len(recommendedJobs))
	for i := range recommendedJobs {
		jobPtrs[i] = &recommendedJobs[i]
	}

	// Build summary
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
		// In production, this should be logged properly
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
