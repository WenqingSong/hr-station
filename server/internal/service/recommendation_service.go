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
	PipelineMode    string                        `json:"pipeline_mode,omitempty"`
	JudgeResult     *model.RecommendationJudgeResult `json:"judge_result,omitempty"`

	// K9-1-H4-C: profile status fields
	ProfileStatus          string `json:"profile_status,omitempty"`
	ProfileType            string `json:"profile_type,omitempty"`
	ProfileSource          string `json:"profile_source,omitempty"`
	ProfileDominantProfile string `json:"profile_dominant_profile,omitempty"`
	ProfileMessage         string `json:"profile_message,omitempty"`
}

// RecommendationBusinessResult wraps the business result for recommendation.
type RecommendationBusinessResult struct {
	RecommendedJobs []*model.RecommendedJob `json:"recommended_jobs"`
}

// RecommendationService handles job recommendation business logic.
type RecommendationService struct {
	resumeRepo         *repository.ResumeRepository
	jobRepo            *repository.JobRepository
	recommendationRepo *repository.RecommendationRepository
	mockAgent          *agent.RecommendationMockAgent
	fastRunner         agent.PipelineRunner
	qualityRunner      agent.PipelineRunner
	llmEnabled         bool
	profileSvc         *ResumeProfileService
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(
	resumeRepo *repository.ResumeRepository,
	jobRepo *repository.JobRepository,
	recommendationRepo *repository.RecommendationRepository,
	mockAgent *agent.RecommendationMockAgent,
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
	profileSvc *ResumeProfileService,
) *RecommendationService {
	return &RecommendationService{
		resumeRepo:         resumeRepo,
		jobRepo:            jobRepo,
		recommendationRepo: recommendationRepo,
		mockAgent:          mockAgent,
		fastRunner:         fastRunner,
		qualityRunner:      qualityRunner,
		llmEnabled:         llmEnabled,
		profileSvc:         profileSvc,
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
func (s *RecommendationService) Recommend(ctx context.Context, userID uint64, req RecommendationRequest) (*RecommendationResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 5 {
		req.Limit = 5
	}

	// K9-1-H4-C: determine profile type from generation_tier
	profileType := "fast"
	if req.GenerationTier == model.GenerationTierQuality {
		profileType = "quality"
	}

	// K9-1-H4-C: ensure profile is ready before recommending
	profileStatus := "ready"
	var profileDominant string
	var profileSource string
	if s.profileSvc != nil && req.ResumeID > 0 {
		profile, st, err := s.profileSvc.EnsureProfile(ctx, userID, req.ResumeID, profileType)
		if err != nil {
			log.Printf("[RecommendationService] ensure profile failed resume=%d err=%v", req.ResumeID, err)
			profileStatus = "failed"
		} else if profile != nil && st == model.ProfileStatusReady {
			profileStatus = "ready"
			profileDominant = profile.DominantProfile
			profileSource = profile.Source
		} else {
			profileStatus = st
		}
	} else if s.profileSvc == nil {
		profileStatus = "legacy"
	}

	// When profile is not ready, return early with status
	if profileStatus != "ready" && profileStatus != "legacy" {
		msg := "正在生成简历画像，请稍后重试"
		if profileStatus == "failed" {
			msg = "简历画像生成失败，可重新生成画像"
		} else if profileStatus == "stale" {
			msg = "简历画像需要更新，请稍后重试"
		}
		log.Printf("[RecommendationService] profile not ready status=%s resume=%d", profileStatus, req.ResumeID)
		return &RecommendationResponse{
			ResumeID:               req.ResumeID,
			Result:                 &RecommendationBusinessResult{RecommendedJobs: []*model.RecommendedJob{}},
			RecommendedJobs:        []*model.RecommendedJob{},
			Summary:                msg,
			ProfileStatus:          profileStatus,
			ProfileType:            profileType,
			ProfileMessage:         msg,
			CreatedAt:              time.Now().Format("2006-01-02T15:04:05Z07:00"),
		}, nil
	}

	return s.recommendCore(ctx, userID, req, profileType, profileDominant, profileSource)
}

func (s *RecommendationService) recommendCore(ctx context.Context, userID uint64, req RecommendationRequest, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	runner := s.selectRunner(req.GenerationTier)
	resume, err := s.getResume(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	jobs, err := s.jobRepo.ListRecommendCandidates(ctx, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	if len(jobs) == 0 {
		return nil, apperror.ErrJobNotFound
	}

	// Try LLM Pipeline
	if s.llmEnabled && runner != nil {
		resp, fallbackErr := s.recommendWithLLM(ctx, userID, resume, jobs, req.Limit, runner)
		if fallbackErr == nil {
			s.fillProfileFields(resp, profileType, profileDominant, profileSource)
			return resp, nil
		}
		log.Printf("[RecommendationService] LLM pipeline fallback: %v", fallbackErr)
		return s.recommendWithMock(ctx, userID, resume, jobs, req.Limit, fallbackErr, false, profileType, profileDominant, profileSource)
	}

	qualityMode := req.GenerationTier == model.GenerationTierQuality
	return s.recommendWithMock(ctx, userID, resume, jobs, req.Limit, nil, qualityMode, profileType, profileDominant, profileSource)
}

func (s *RecommendationService) fillProfileFields(resp *RecommendationResponse, profileType, dominant, source string) {
	resp.ProfileStatus = "ready"
	resp.ProfileType = profileType
	resp.ProfileDominantProfile = dominant
	resp.ProfileSource = source
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
func (s *RecommendationService) recommendWithMock(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int, fbErr *agent.RecFallbackError, qualityMode bool, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	var recommendedJobs []model.RecommendedJob
	var err error

	// K9-1-H4-C: use stored profile tags if available
	usedProfile := false
	if s.profileSvc != nil && resume.ID > 0 && profileDominant != "" {
		profile, findErr := s.profileSvc.GetProfileStatus(ctx, userID, resume.ID)
		if findErr == nil && profile.FastStatus == "ready" {
			candTags := convertProfileTags(profile.PrimaryTags)
			if len(candTags) > 0 {
				log.Printf("[RecommendationService] using stored profile dominant=%s tags=%d", profileDominant, len(candTags))
				if qualityMode {
					recommendedJobs, err = s.mockAgent.RecommendQualityWithStoredProfile(resume, jobs, profileDominant, candTags, nil)
				} else {
					recommendedJobs, err = s.mockAgent.RecommendWithStoredProfile(resume, jobs, profileDominant, candTags, nil)
				}
				usedProfile = true
			}
		}
	}

	if !usedProfile {
		if qualityMode {
			recommendedJobs, err = s.mockAgent.RecommendQuality(context.Background(), resume, jobs)
		} else {
			recommendedJobs, err = s.mockAgent.Recommend(context.Background(), resume, jobs)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	if len(recommendedJobs) > limit {
		recommendedJobs = recommendedJobs[:limit]
	}

	generatedBy := "mock_agent"
	if qualityMode {
		generatedBy = "mock_quality"
	}

	var meta model.AgentResponseMeta
	for i := range recommendedJobs {
		recommendedJobs[i].GeneratedBy = generatedBy
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

	pipelineMode := "fast_rule"
	if qualityMode {
		pipelineMode = "quality_pipeline"
	}
	resp := &RecommendationResponse{
		ResumeID:        resume.ID,
		GeneratedBy:     "mock_agent",
		RecommendedJobs: jobPtrs,
		Result:          recResult,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:            meta,
		PipelineMode:    pipelineMode,
		JudgeResult:     &model.RecommendationJudgeResult{Pass: true, Severity: "pass", Issues: nil, FixStrategy: "none"},
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
		_ = err
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)
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

func convertProfileTags(tags []model.ProfileTag) []agent.StoredTag {
	result := make([]agent.StoredTag, len(tags))
	for i, t := range tags {
		result[i] = agent.StoredTag{Tag: t.Tag, Group: t.Group, Confidence: t.Confidence, Evidence: t.Evidence}
	}
	return result
}
