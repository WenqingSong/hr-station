package service

import (
	"context"
	"encoding/json"
	apperror "offer-catcher/server/internal/errors"
	"fmt"
	"log"
	"math"
	"strings"
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
	// K9-1-U-C: result persistence metadata
	ResultID       uint64 `json:"result_id,omitempty"`
	GenerationTier string `json:"generation_tier,omitempty"`
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
	fallbackEnable     bool
	enableMockAgent    bool
	profileSvc         *ResumeProfileService
	tagEvidence        *TagEvidenceService // K13-Fix-B: Fast 统一使用 Tag/Evidence
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
	fallbackEnable bool,
	enableMockAgent bool,
	profileSvc *ResumeProfileService,
	tagEvidence *TagEvidenceService,
) *RecommendationService {
	return &RecommendationService{
		resumeRepo:         resumeRepo,
		jobRepo:            jobRepo,
		recommendationRepo: recommendationRepo,
		mockAgent:          mockAgent,
		fastRunner:         fastRunner,
		qualityRunner:      qualityRunner,
		llmEnabled:         llmEnabled,
		fallbackEnable:     fallbackEnable,
		enableMockAgent:    enableMockAgent,
		profileSvc:         profileSvc,
		tagEvidence:        tagEvidence,
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

	// K9-1-Regression-C: quality profile not ready → fallback to fast profile
	profileStatus := "ready"
	var profileDominant string
	var profileSource string
	if s.profileSvc != nil && req.ResumeID > 0 {
		profile, st, err := s.profileSvc.EnsureProfile(ctx, userID, req.ResumeID, profileType)
		if err != nil {
			log.Printf("[RecommendationService] ensure profile failed resume=%d tier=%s err=%v", req.ResumeID, profileType, err)
		} else if profile != nil && st == model.ProfileStatusReady {
			profileDominant = profile.DominantProfile
			profileSource = profile.Source
			profileStatus = "ready"
		} else if req.GenerationTier == model.GenerationTierQuality {
			// quality profile not ready → fallback to fast profile
			log.Printf("[RecommendationService] quality profile status=%s resume=%d, falling back to fast profile", st, req.ResumeID)
			fastProfile, fastSt, fastErr := s.profileSvc.EnsureProfile(ctx, userID, req.ResumeID, "fast")
			if fastErr != nil || fastProfile == nil || fastSt != model.ProfileStatusReady {
				log.Printf("[RecommendationService] fast profile fallback also failed resume=%d st=%s err=%v", req.ResumeID, fastSt, fastErr)
				profileStatus = st
			} else {
				profileDominant = fastProfile.DominantProfile
				profileSource = fastProfile.Source
				profileStatus = "ready"
				profileType = "fast" // using fast profile for quality generation
			}
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

	// K13-Fix-B: Fast tier uses TagEvidenceService (no LLM)
	isFast := req.GenerationTier != model.GenerationTierQuality

	if isFast && s.tagEvidence != nil {
		log.Printf("[RecommendationService] use_fast_tag_engine tier=%s resume=%d", req.GenerationTier, req.ResumeID)
		resp, err := s.fastRecommendWithTags(ctx, userID, resume, jobs, req.Limit, profileType, profileDominant, profileSource)
		if err == nil {
			return resp, nil
		}
		log.Printf("[RecommendationService] fast tag engine failed: %v, falling back to mock", err)
	}

	// K13-Fix-B: Deep/Quality tier uses LLM pipeline
	runner := s.selectRunner(req.GenerationTier)
	isQuality := req.GenerationTier == model.GenerationTierQuality

	if isQuality && s.llmEnabled && runner != nil {
		// K13-Fix-C1: Try 3-agent pipeline first (for true Deep recommendation)
		if s.tagEvidence != nil {
			log.Printf("[RecommendationService] use_deep_3agent tier=%s resume=%d", req.GenerationTier, req.ResumeID)
			resp, err3a := s.recommendWith3Agents(ctx, userID, resume, jobs, req.Limit, runner, profileType, profileDominant, profileSource)
			if err3a == nil {
				return resp, nil
			}
			log.Printf("[RecommendationService] 3-agent pipeline failed: %v, falling back to old LLM path", err3a)
		}

		// Fall back to old single-agent LLM pipeline
		log.Printf("[RecommendationService] use_deep_llm tier=%s", req.GenerationTier)
		resp, fallbackErr := s.recommendWithLLM(ctx, userID, resume, jobs, req.Limit, runner)
		if fallbackErr == nil {
			s.fillProfileFields(resp, profileType, profileDominant, profileSource)
			return resp, nil
		}
		log.Printf("[RecommendationService] LLM pipeline fallback: %v", fallbackErr)
		if !s.fallbackEnable {
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed and fallback disabled: %s", fallbackErr.FallbackReason), 500)
		}
		return s.recommendWithMock(ctx, userID, resume, jobs, req.Limit, fallbackErr, false, profileType, profileDominant, profileSource)
	}

	// Fallback to mock agent
	if !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "no generation path available", 500)
	}
	log.Printf("[RecommendationService] use_mock_fallback tier=%s", req.GenerationTier)
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

	// K9-1-U-C: Save with full metadata
	if len(recommendedJobs) > 0 {
		resultJSON, marshalErr := json.Marshal(recommendedJobs)
		if marshalErr == nil {
			recResult := &model.RecommendationResult{
				UserID:         userID,
				ResumeID:       resume.ID,
				GenerationTier: "quality",
				GeneratedBy:    "llm_agent",
				Status:         "ready",
				ResultJSON:     resultJSON,
			}
			if createErr := s.recommendationRepo.Create(ctx, recResult); createErr != nil {
				_ = createErr
			}
		}
	}

	return resp, nil
}

// recommendWithMock calls the Mock Agent and builds the response.
func (s *RecommendationService) recommendWithMock(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int, fbErr *agent.RecFallbackError, qualityMode bool, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	var recommendedJobs []model.RecommendedJob
	var err error

	// K9-1-Regression-D: quality mode uses Feature Card + Hard Guard pipeline
	if qualityMode {
		recommendedJobs = s.mockAgent.RecommendQualityWithFeatureCards(resume, jobs, profileDominant)
	} else {
		// Fast mode: use stored profile tags if available
		usedProfile := false
		if s.profileSvc != nil && resume.ID > 0 && profileDominant != "" {
			profile, findErr := s.profileSvc.GetProfileStatus(ctx, userID, resume.ID)
			if findErr == nil && profile.FastStatus == "ready" {
				candTags := convertProfileTags(profile.PrimaryTags)
				if len(candTags) > 0 {
					recommendedJobs, err = s.mockAgent.RecommendWithStoredProfile(resume, jobs, profileDominant, candTags, nil)
					usedProfile = true
				}
			}
		}
		if !usedProfile {
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

	// K9-1-U-C: Save recommendation result snapshot with full metadata
	if len(recommendedJobs) > 0 {
		resultJSON, err := json.Marshal(recommendedJobs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal recommendation result: %w", err)
		}

		tier := "fast"
		if qualityMode {
			tier = "quality"
		}
		recSnapshot := &model.RecommendationResult{
			UserID:                 userID,
			ResumeID:               resume.ID,
			GenerationTier:         tier,
			GeneratedBy:            generatedBy,
			Status:                 "ready",
			ProfileType:            profileType,
			ProfileDominantProfile: profileDominant,
			ResultJSON:             resultJSON,
		}

		if createErr := s.recommendationRepo.Create(ctx, recSnapshot); createErr != nil {
			_ = createErr
		} else {
			resp.ResultID = recSnapshot.ID
			resp.GenerationTier = tier
		}
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

// GetLatestRecommendation returns the latest ready recommendation result for the given user+resume.
func (s *RecommendationService) GetLatestRecommendation(ctx context.Context, userID, resumeID uint64) (*model.RecommendationLatestResult, error) {
	if resumeID <= 0 {
		return nil, fmt.Errorf("invalid resume_id")
	}
	// Verify resume belongs to user
	resume, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify resume: %w", err)
	}
	if resume == nil {
		return nil, apperror.ErrResumeNotFound
	}

	latest, err := s.recommendationRepo.FindLatestByResumeID(ctx, userID, resumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest recommendation: %w", err)
	}

	if latest == nil {
		return &model.RecommendationLatestResult{HasResult: false}, nil
	}

	// Parse result_json
	var resultData interface{}
	if len(latest.ResultJSON) > 0 {
		if err := json.Unmarshal(latest.ResultJSON, &resultData); err != nil {
			return nil, fmt.Errorf("failed to parse saved result: %w", err)
		}
	}

	return &model.RecommendationLatestResult{
		HasResult:              true,
		ResultID:               latest.ID,
		ResumeID:               latest.ResumeID,
		GenerationTier:         latest.GenerationTier,
		GeneratedBy:            latest.GeneratedBy,
		Status:                 latest.Status,
		ProfileType:            latest.ProfileType,
		ProfileDominantProfile: latest.ProfileDominantProfile,
		CreatedAt:              latest.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Result:                 resultData,
	}, nil
}

// ─── K13-Fix-B: Fast Recommendation via TagEvidenceService ───────────────

// fastRecommendWithTags generates Fast-tier recommendations using TagEvidenceService.
// No LLM call — uses RecallJobsByTags + tag matrix scoring.
func (s *RecommendationService) fastRecommendWithTags(ctx context.Context, userID uint64, resume *model.ResumeVersion, allJobs []*model.Job, limit int, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	// 1. Ensure tag profile for this resume
	if s.tagEvidence == nil {
		return nil, fmt.Errorf("tag evidence service not available")
	}
	_, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resume.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("fast recommend: %w", err)
	}

	// 2. Build job map for title lookup
	jobMap := make(map[uint64]*model.Job)
	for _, j := range allJobs {
		jobMap[j.ID] = j
	}

	// 3. Recall top candidates
	candidates, err := s.tagEvidence.RecallJobsByTags(ctx, resume.ID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("fast recommend: recall: %w", err)
	}

	// 4. Build RecommendedJob entries from candidates
	var recommended []*model.RecommendedJob
	for i, c := range candidates {
		if i >= limit {
			break
		}
		// Fill title/department/city from job lookup
		title, dept, city := "", "", ""
		if j, ok := jobMap[c.JobID]; ok {
			title, dept, city = j.Title, j.Department, j.City
		}
		// Scale tagScore (count-based, typically 1-50) to 70-95 range
		scaledScore := int(math.Min(95, math.Max(70, 70 + c.TagScore*0.5)))
		rec := &model.RecommendedJob{
			JobID:               c.JobID,
			Title:               title,
			Department:          dept,
			City:                city,
			Category:            c.Category,
			JobType:             "",
			RecommendationScore: scaledScore,
			RecommendationReason:  buildFastOneSentenceReason(c),
			RecommendationLabel:  buildFastLabel(c.TagScore),
			RecommendationPriority: buildFastPriority(c.TagScore),
			MatchHighlights:      buildFastHighlights(c),
			RiskPoints:           buildFastRiskReminders(c),
			SuggestedAction:      buildFastSuggestedAction(c),
			WhyRankedHere:        buildFastRecommendReasonExplainer(c),
			GeneratedBy:         "fast_tag_engine",
			AgentStatus:         "success",
			FallbackReason:      "none",
			RepairApplied:       false,
			DataQuality:         "rule",
			ResumeEvidence:      c.EvidenceRefs,
		}
		recommended = append(recommended, rec)
	}

	// 4. Build response
	summary := fmt.Sprintf("基于 %d 个岗位标签匹配，共推荐 %d 个岗位", len(allJobs), len(recommended))
	meta := model.AgentResponseMeta{
		AgentName:      "FastRecommendationEngine",
		GeneratedBy:    "fast_tag_engine",
		AgentStatus:    "success",
		FallbackReason: "none",
		RepairApplied:  false,
		SchemaValid:    true,
	}

	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	resp := &RecommendationResponse{
		ResumeVersionID:       resume.ID,
		ResumeID:              resume.ID,
		GeneratedBy:           "fast_tag_engine",
		RecommendedJobs:       recommended,
		Summary:               summary,
		CreatedAt:             now,
		Meta:                  meta,
		GenerationTier:        "fast",
		PipelineMode:          "fast_tag_engine",
		ProfileStatus:         "ready",
		ProfileType:           profileType,
		ProfileSource:         profileSource,
		ProfileDominantProfile: profileDominant,
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)
	return resp, nil
}

// ─── K13-Fix-C1: Deep Recommendation via 3-Agent Pipeline ───────────────

// recommendWith3Agents implements the Deep recommendation using 3 physical agents:
// Agent 1: Context/Tag/Evidence (rule-based)
// Agent 2: Decision/Generation (LLM)
// Agent 3: Verifier/Frontend-Contract (LLM)
func (s *RecommendationService) recommendWith3Agents(ctx context.Context, userID uint64, resume *model.ResumeVersion, allJobs []*model.Job, limit int, runner agent.PipelineRunner, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	// Prepare TagEvidenceService data
	rp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, resume.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("3agent: resume profile: %w", err)
	}

	jps, err := s.tagEvidence.GetAllJobTagProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("3agent: job profiles: %w", err)
	}

	// Build job map for title lookup
	jobMap := make(map[uint64]*model.Job)
	for _, j := range allJobs {
		jobMap[j.ID] = j
	}

	// Run 3-agent pipeline
	result, err := agent.RunDeepRecommendation3Agents(ctx, rp, jps, jobMap, runner, runner)
	if err != nil {
		return nil, fmt.Errorf("3agent: %w", err)
	}

	// Build RecommendedJob entries from final output
	var recommended []*model.RecommendedJob
	for _, item := range result.FinalOutput.Items {
		if len(recommended) >= limit {
			break
		}
		title, dept, city := item.Title, "", ""
		if j, ok := jobMap[item.JobID]; ok {
			title, dept, city = j.Title, j.Department, j.City
		}
		rec := &model.RecommendedJob{
			JobID:                item.JobID,
			Title:                title,
			Department:           dept,
			City:                 city,
			RecommendationScore:  item.RecommendationScore,
			RecommendationReason: item.RecommendationReason,
			RecommendationLabel:  item.RecommendationLevel,
			RecommendationPriority: item.RecommendationLevel,
			MatchHighlights:      item.MatchHighlights,
			RiskPoints:           item.RiskPoints,
			ResumeEvidence:       item.ResumeEvidence,
			SuggestedAction:      item.SuggestedAction,
			ApplyReadiness:       item.OptimizationSignal,
			GeneratedBy:          result.FinalOutput.GeneratedBy,
			AgentStatus:          result.FinalOutput.AgentStatus,
			RepairApplied:        result.FinalOutput.RepairApplied,
			DataQuality:          "llm",
		}
		recommended = append(recommended, rec)
	}

	meta := model.AgentResponseMeta{
		AgentName:   "DeepRecommendationEngine",
		GeneratedBy: result.FinalOutput.GeneratedBy,
		AgentStatus: result.FinalOutput.AgentStatus,
		RepairApplied: result.FinalOutput.RepairApplied,
	}

	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	resp := &RecommendationResponse{
		ResumeVersionID:        resume.ID,
		ResumeID:               resume.ID,
		GeneratedBy:            result.FinalOutput.GeneratedBy,
		RecommendedJobs:        recommended,
		Summary:                "",
		CreatedAt:              now,
		Meta:                   meta,
		GenerationTier:         "quality",
		PipelineMode:           "deep_3agent",
		ProfileStatus:          "ready",
		ProfileType:            profileType,
		ProfileSource:          profileSource,
		ProfileDominantProfile: profileDominant,
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)
	return resp, nil
}

func buildFastRecommendReason(c model.JobCandidate) string {
	if len(c.MatchedTags) > 0 {
		return fmt.Sprintf("标签匹配：%s", strings.Join(c.MatchedTags[:min(3, len(c.MatchedTags))], "、"))
	}
	return "该岗位可能与你的简历方向存在潜在匹配"
}

func buildFastLabel(score float64) string {
	if score >= 15 { return "强匹配" }
	if score >= 8 { return "较高匹配" }
	return "可关注"
}

func buildFastPriority(score float64) string {
	if score >= 15 { return "strong" }
	if score >= 8 { return "high" }
	return "try"
}

func convertProfileTags(tags []model.ProfileTag) []agent.StoredTag {
	result := make([]agent.StoredTag, len(tags))
	for i, t := range tags {
		result[i] = agent.StoredTag{Tag: t.Tag, Group: t.Group, Confidence: t.Confidence, Evidence: t.Evidence}
	}
	return result
}
