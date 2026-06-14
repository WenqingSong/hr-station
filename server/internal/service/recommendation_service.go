package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	apperror "offer-catcher/server/internal/errors"
	"strings"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// ErrFastRecommendSEPMissing is returned when the current resume SEP profile is not ready.
// The caller (recommendCore) catches this and falls back to TagEvidenceService.
var ErrFastRecommendSEPMissing = errors.New("fast recommend sep profile missing")

// safeFallbackCtx returns a context that is NOT derived from the request context.
// Use this for mock fallback and DB persistence when the original request context
// has been cancelled (e.g. after DeepCascade timeout). 15s is generous for in-memory
// mock + single-row DB insert.
func safeFallbackCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 15*time.Second)
}

// deepCascadeProfileBlockThreshold is the maximum number of jobs with blocked
// (missing / stale / hash-mismatch) SEP profiles that the cascade pipeline will
// tolerate before skipping entirely and going straight to deep_fallback_agent.
// Setting this low prevents the cascade from synchronously generating profiles
// during a single user HTTP request.
const deepCascadeProfileBlockThreshold = 3

// RecommendationRequest represents a request for job recommendations.
type RecommendationRequest struct {
	ResumeID       uint64               `json:"resume_id"`
	Limit          int                  `json:"limit"`
	GenerationTier model.GenerationTier `json:"generation_tier,omitempty"`
}

// RecommendationResponse represents the response for job recommendations.
type RecommendationResponse struct {
	ResumeVersionID uint64                        `json:"resume_version_id"`
	ResumeID        uint64                        `json:"resume_id"`
	GeneratedBy     string                        `json:"generated_by"`     // deprecated: use meta.generated_by
	RecommendedJobs []*model.RecommendedJob       `json:"recommended_jobs"` // deprecated: use result.recommended_jobs
	Result          *RecommendationBusinessResult `json:"result"`
	Summary         string                        `json:"summary"`
	CreatedAt       string                        `json:"created_at"`
	Meta            model.AgentResponseMeta       `json:"meta"`
	// K9-1-U-C: result persistence metadata
	ResultID       uint64                           `json:"result_id,omitempty"`
	GenerationTier string                           `json:"generation_tier,omitempty"`
	PipelineMode   string                           `json:"pipeline_mode,omitempty"`
	JudgeResult    *model.RecommendationJudgeResult `json:"judge_result,omitempty"`

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
	Summary         string                  `json:"summary,omitempty"`
	QualityStatus   string                  `json:"quality_status,omitempty"`
	PipelineTrace   string                  `json:"pipeline_trace,omitempty"`

	// ── L2: Deep Cascade calibration fields ──
	GlobalSummary           string `json:"global_summary,omitempty"`
	CandidatePositioning    string `json:"candidate_positioning,omitempty"`
	TopRecommendationReason string `json:"top_recommendation_reason,omitempty"`
	PortfolioRiskSummary    string `json:"portfolio_risk_summary,omitempty"`
	DeliveryStrategy        string `json:"delivery_strategy,omitempty"`
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
	tagEvidence        *TagEvidenceService                   // K13-Fix-B: Fast 统一使用 Tag/Evidence
	semanticGate       *SemanticProfileGate                  // SEP Gate for SEP-based Fast recommend
	semanticRepo       *repository.SemanticProfileRepository // SEP tag queries
	qualityClient      llm.Client                            // Hy3 client for Deep Cascade
	qualityModel       string                                // model for Deep Cascade
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
	semanticGate *SemanticProfileGate,
	semanticRepo *repository.SemanticProfileRepository,
	qualityClient llm.Client,
	qualityModel string,
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
		semanticGate:       semanticGate,
		semanticRepo:       semanticRepo,
		qualityClient:      qualityClient,
		qualityModel:       qualityModel,
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
			ResumeID:        req.ResumeID,
			Result:          &RecommendationBusinessResult{RecommendedJobs: []*model.RecommendedJob{}},
			RecommendedJobs: []*model.RecommendedJob{},
			Summary:         msg,
			ProfileStatus:   profileStatus,
			ProfileType:     profileType,
			ProfileMessage:  msg,
			CreatedAt:       time.Now().Format("2006-01-02T15:04:05Z07:00"),
		}, nil
	}

	// R3: Quality cache-first — before entering DeepCascade, check if we already
	// have a ready quality result for this resume. If so, return it immediately
	// without triggering any SEP generation or LLM calls.
	if req.GenerationTier == model.GenerationTierQuality {
		if cached, err := s.recommendationRepo.FindLatestByResumeID(ctx, userID, req.ResumeID, "quality"); err == nil && cached != nil && len(cached.ResultJSON) > 0 {
			var jobs []model.RecommendedJob
			if jsonErr := json.Unmarshal(cached.ResultJSON, &jobs); jsonErr == nil && len(jobs) > 0 {
				log.Printf("[RecommendationService] quality cache hit resume=%d result_id=%d generated_by=%s jobs=%d, skip deep cascade",
					cached.ResumeID, cached.ID, cached.GeneratedBy, len(jobs))
				if len(jobs) > req.Limit {
					jobs = jobs[:req.Limit]
				}
				jobPtrs := make([]*model.RecommendedJob, len(jobs))
				for i := range jobs {
					jobPtrs[i] = &jobs[i]
				}
				// Determine display tier: stored "quality" → respond "deep" for cascade compat
				displayTier := cached.GenerationTier
				if displayTier == "quality" {
					displayTier = "deep"
				}
				// Preserve pipeline_mode from the original pipeline if known
				pipeline := pipelineModeFromGeneratedBy(cached.GeneratedBy)
				now := time.Now()
				meta := model.NewAgentResponseMeta(model.AgentNameRecommendation).
					WithGeneratedBy(cached.GeneratedBy).
					WithAgentStatus(model.AgentStatusSuccess)
				return &RecommendationResponse{
					ResumeID:               req.ResumeID,
					GeneratedBy:            cached.GeneratedBy,
					RecommendedJobs:        jobPtrs,
					Result:                 &RecommendationBusinessResult{RecommendedJobs: jobPtrs},
					CreatedAt:              now.Format("2006-01-02T15:04:05Z07:00"),
					Meta:                   meta,
					ResultID:               cached.ID,
					GenerationTier:         displayTier,
					PipelineMode:           pipeline,
					JudgeResult:            &model.RecommendationJudgeResult{Pass: true, Severity: "pass", Issues: nil, FixStrategy: "none"},
					ProfileStatus:          profileStatus,
					ProfileType:            profileType,
					ProfileSource:          profileSource,
					ProfileDominantProfile: profileDominant,
				}, nil
			}
		}
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

	if isFast && s.semanticGate != nil {
		log.Printf("[RecommendationService] use_fast_sep tier=%s resume=%d", req.GenerationTier, req.ResumeID)
		resp, err := s.fastRecommendWithSEP(ctx, userID, resume, jobs, req.Limit, profileType, profileDominant, profileSource)
		if err != nil {
			log.Printf("[RecommendationService] fast SEP failed: %v, trying fast tag engine", err)
			// Fallback to old tag engine if SEP fails
			if s.tagEvidence != nil {
				resp2, err2 := s.fastRecommendWithTags(ctx, userID, resume, jobs, req.Limit, profileType, profileDominant, profileSource)
				if err2 == nil {
					return resp2, nil
				}
				log.Printf("[RecommendationService] fast tag engine also failed: %v", err2)
			}
		} else {
			return resp, nil
		}
	} else if isFast && s.tagEvidence != nil {
		log.Printf("[RecommendationService] use_fast_tag_engine tier=%s resume=%d", req.GenerationTier, req.ResumeID)
		resp, err := s.fastRecommendWithTags(ctx, userID, resume, jobs, req.Limit, profileType, profileDominant, profileSource)
		if err == nil {
			return resp, nil
		}
		log.Printf("[RecommendationService] fast tag engine failed: %v, falling back to mock", err)
	}

	// M3: LLM Capability Gate — quality tier requires LLM to be available
	runner := s.selectRunner(req.GenerationTier)
	isQuality := req.GenerationTier == model.GenerationTierQuality

	// K15-P0: LLM unavailable Deep gate → fallback to Mock Agent, NOT 503
	if isQuality && !s.llmEnabled {
		log.Printf("[RecommendationService] deep tier blocked: llm disabled, falling back to mock")
		if !s.enableMockAgent {
			return nil, apperror.New(apperror.CodeInternal, "深度推荐需要大模型服务，当前模型未启用。你仍可使用快速推荐。", 503)
		}
		qualityMode := true
		safeCtx, safeCancel := safeFallbackCtx()
		defer safeCancel()
		return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, &agent.RecFallbackError{
			FallbackReason: model.FallbackReasonLLMDisabled,
			ErrorCode:      model.ErrorCodeLLMDisabled,
			ErrorMessage:   "深度推荐需要大模型服务，当前模型未启用。显示基于规则的兜底结果。",
		}, qualityMode, profileType, profileDominant, profileSource)
	}
	if isQuality && s.qualityClient == nil {
		log.Printf("[RecommendationService] deep tier blocked: LLM client missing, falling back to mock")
		if !s.enableMockAgent {
			return nil, apperror.New(apperror.CodeInternal, "深度推荐需要连接大模型服务，当前服务未配置。你仍可使用快速推荐。", 503)
		}
		qualityMode := true
		safeCtx, safeCancel := safeFallbackCtx()
		defer safeCancel()
		return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, &agent.RecFallbackError{
			FallbackReason: model.FallbackReasonLLMClientMissing,
			ErrorCode:      model.ErrorCodeLLMClientMissing,
			ErrorMessage:   "深度推荐需要连接大模型服务，当前服务未配置。显示基于规则的兜底结果。",
		}, qualityMode, profileType, profileDominant, profileSource)
	}

	if isQuality && s.llmEnabled && runner != nil {
		// K15-P0-R1: context cancellation guard — if caller disconnected or deadline exceeded,
		// skip all LLM paths and go directly to mock fallback.
		// K15-P0-R2: use safe (background-derived) context for fallback — never pass
		// the cancelled request context to mock or DB persistence.
		if ctx.Err() != nil {
			log.Printf("[RecommendationService] context cancelled/deadline exceeded before quality tier: %v, falling back to mock", ctx.Err())
			if !s.enableMockAgent {
				return nil, apperror.New(apperror.CodeInternal, "request cancelled and mock agent unavailable", 499)
			}
			qualityMode := true
			safeCtx, safeCancel := safeFallbackCtx()
			defer safeCancel()
			return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, &agent.RecFallbackError{
				FallbackReason: fmt.Sprintf("context_cancelled_%v", ctx.Err()),
				ErrorCode:      "CONTEXT_CANCELLED",
				ErrorMessage:   ctx.Err().Error(),
			}, qualityMode, profileType, profileDominant, profileSource)
		}

		// Deep/Quality tier → Cascaded Deep Recommendation Pipeline
		if s.semanticGate != nil && s.semanticRepo != nil {
			// R3: guard against cascade trying to synchronously generate SEP profiles
			// for dozens of jobs during a single HTTP request. Count blocked (missing
			// + stale) profiles — any blocked profile would trigger regeneration in
			// the cascade L0 gate. If blocked > threshold, skip cascade and fallback.
			resumeID := int64(resume.ID)
			blockedCount, breakdown := s.countBlockedJobProfiles(ctx, jobs)
			log.Printf("[RecommendationService] cascade pre-check: total_jobs=%d %s", len(jobs), breakdown)
			if blockedCount > deepCascadeProfileBlockThreshold {
				log.Printf("[RecommendationService] deep cascade skipped: blocked job profiles (%d > %d), using deep_fallback_agent",
					blockedCount, deepCascadeProfileBlockThreshold)
				goto qualityFallback
			}

			log.Printf("[RecommendationService] use_deep_cascade tier=%s resume=%d", req.GenerationTier, resumeID)
			resp, err := s.recommendDeepCascade(ctx, userID, resume, jobs, req.Limit, profileType, profileDominant, profileSource)
			if err == nil {
				return resp, nil
			}
			log.Printf("[RecommendationService] deep cascade failed: %v, trying old 3-agent", err)
			// K15-P0-R1: re-check context before next LLM path
			if ctx.Err() != nil {
				log.Printf("[RecommendationService] context cancelled after deep cascade: %v, skipping to mock", ctx.Err())
			}
		}

		// Fall back to old 3-agent pipeline (only if context still valid)
		if s.tagEvidence != nil && ctx.Err() == nil {
			log.Printf("[RecommendationService] use_deep_3agent tier=%s resume=%d", req.GenerationTier, req.ResumeID)
			resp, err3a := s.recommendWith3Agents(ctx, userID, resume, jobs, req.Limit, runner, profileType, profileDominant, profileSource)
			if err3a == nil {
				return resp, nil
			}
			log.Printf("[RecommendationService] 3-agent pipeline failed: %v, falling back to old LLM path", err3a)
		}

		// Fall back to old single-agent LLM pipeline (only if context still valid)
		if ctx.Err() == nil {
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
			safeCtx, safeCancel := safeFallbackCtx()
			defer safeCancel()
			return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, fallbackErr, true, profileType, profileDominant, profileSource)
		}

	qualityFallback:
		// K15-P0-R1: context cancelled during quality fallback chain — final mock fallback
		// K15-P0-R2: use safe context for fallback, never pass cancelled request context
		cancelReason := "unknown"
		if ctxErr := ctx.Err(); ctxErr != nil {
			cancelReason = fmt.Sprintf("%v", ctxErr)
		}
		log.Printf("[RecommendationService] quality fallback: using deep_fallback_agent (reason=%s)", cancelReason)
		if !s.enableMockAgent {
			return nil, apperror.New(apperror.CodeInternal, "quality recommendation failed: all paths exhausted and mock agent unavailable", 503)
		}
		qualityMode := true
		safeCtx, safeCancel := safeFallbackCtx()
		defer safeCancel()
		return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, &agent.RecFallbackError{
			FallbackReason: fmt.Sprintf("quality_fallback_%s", cancelReason),
			ErrorCode:      "QUALITY_EXHAUSTED",
			ErrorMessage:   "Deep recommendation could not complete; showing fallback results.",
		}, qualityMode, profileType, profileDominant, profileSource)
	}

	// Fallback to mock agent (fast tier or quality tier with no LLM configured at all)
	if !s.enableMockAgent {
		return nil, apperror.New(apperror.CodeInternal, "no generation path available", 500)
	}
	log.Printf("[RecommendationService] use_mock_fallback tier=%s", req.GenerationTier)
	qualityMode := req.GenerationTier == model.GenerationTierQuality
	safeCtx, safeCancel := safeFallbackCtx()
	defer safeCancel()
	return s.recommendWithMock(safeCtx, userID, resume, jobs, req.Limit, nil, qualityMode, profileType, profileDominant, profileSource)
}

func (s *RecommendationService) fillProfileFields(resp *RecommendationResponse, profileType, dominant, source string) {
	resp.ProfileStatus = "ready"
	resp.ProfileType = profileType
	resp.ProfileDominantProfile = dominant
	resp.ProfileSource = source
}

// pipelineModeFromGeneratedBy returns the pipeline_mode corresponding to a cached result's generated_by.
// Used when returning a cache hit to preserve the original pipeline semantics.
func pipelineModeFromGeneratedBy(generatedBy string) string {
	switch generatedBy {
	case "llm_agent":
		return "cascade" // most likely from the cascade pipeline
	case "deep_fallback_agent":
		return "deep_fallback"
	case "mock_quality":
		return "mock_deep_fallback"
	default:
		return ""
	}
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

// ─── Deep Recommendation via Cascaded Pipeline ──────────────────────────

// countBlockedJobProfiles returns the number of jobs with blocked (missing / stale /
// hash-mismatch) SEP profiles. Blocked profiles would trigger synchronous generation
// inside the DeepCascade L0 gate, causing unacceptable latency.
// Uses FindLatestByTarget (read-only) to avoid triggering profile generation.
// Returns total blocked count and a breakdown log string.
func (s *RecommendationService) countBlockedJobProfiles(ctx context.Context, jobs []*model.Job) (totalBlocked int, breakdown string) {
	if s.semanticRepo == nil {
		return len(jobs), fmt.Sprintf("missing=%d stale=0 hash=0 (repo unavailable)", len(jobs))
	}
	missing := 0
	stale := 0
	blocked := 0
	for _, j := range jobs {
		jid := int64(j.ID)
		profile, err := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, jid)
		if err != nil || profile == nil {
			missing++
			continue
		}
		// Stale: profile exists but status is not "verified"
		if profile.Status != model.SemanticProfileStatusVerified {
			stale++
			continue
		}
		// Check for hash mismatch — would cause cascade gate to regenerate
		// Just count tags presence as a proxy for profile completeness
		tagCount, tagErr := s.semanticRepo.CountTagsByProfileID(ctx, profile.ID)
		if tagErr != nil || tagCount < 3 {
			stale++
			continue
		}
	}
	blocked = missing + stale
	breakdownLog := fmt.Sprintf("missing=%d stale=%d total_blocked=%d threshold=%d", missing, stale, blocked, deepCascadeProfileBlockThreshold)
	return blocked, breakdownLog
}

// recommendDeepCascade executes the Cascaded Deep Recommendation Pipeline.
func (s *RecommendationService) recommendDeepCascade(ctx context.Context, userID uint64, resume *model.ResumeVersion, jobs []*model.Job, limit int, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	if s.semanticGate == nil || s.semanticRepo == nil {
		return nil, fmt.Errorf("SEP services not available for deep cascade")
	}
	if s.qualityClient == nil {
		return nil, fmt.Errorf("quality LLM client not available for deep cascade")
	}
	// Deep cascade LLM calls may take ~10-15 minutes; use a generous timeout
	ctx, cancel := context.WithTimeout(ctx, 900*time.Second)
	defer cancel()

	log.Printf("[DeepCascade] starting cascade (qualityClient=%T model=%s jobs=%d)", s.qualityClient, s.qualityModel, len(jobs))

	cfg := &DeepCascadeConfig{
		LLMClient:    s.qualityClient,
		QualityModel: s.qualityModel,
		SemanticGate: s.semanticGate,
		SemanticRepo: s.semanticRepo,
		ResumeRepo:   s.resumeRepo,
		JobRepo:      s.jobRepo,
	}

	resumeID := int64(resume.ID)
	recommendedJobs, l4Result, trace, globalSummary, err := RunDeepRecommendationCascade(ctx, cfg, resumeID, jobs, limit)
	if err != nil {
		return nil, fmt.Errorf("deep cascade: %w", err)
	}

	// Sanitize outputs
	for _, rec := range recommendedJobs {
		rec.RecommendationReason = sanitizeStr(rec.RecommendationReason)
		rec.MatchHighlights = sanitizeStrList(rec.MatchHighlights)
		rec.RiskPoints = sanitizeStrList(rec.RiskPoints)
		rec.SuggestedAction = sanitizeStr(rec.SuggestedAction)
		rec.WhyRankedHere = sanitizeStr(rec.WhyRankedHere)
		rec.JobSpecificReason = sanitizeStr(rec.JobSpecificReason)
		rec.JobSpecificRisks = sanitizeStrList(rec.JobSpecificRisks)
		rec.ResumeEvidence = sanitizeStrList(rec.ResumeEvidence)
		rec.SuggestedImprovement = sanitizeStr(rec.SuggestedImprovement)
		rec.ApplicationStrategy = sanitizeStr(rec.ApplicationStrategy)
		rec.DisplayTags = sanitizeStrList(rec.DisplayTags)
		rec.CandidateCoreStrength = sanitizeStr(rec.CandidateCoreStrength)
		rec.FitConfidenceReason = sanitizeStr(rec.FitConfidenceReason)
	}

	// Build trace bytes
	traceBytes, _ := json.Marshal(trace)
	agentStatus := "cascade_verified"
	if globalSummary == "" {
		agentStatus = "cascade_repaired"
	}

	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	meta := model.AgentResponseMeta{
		AgentName:      "DeepCascadePipeline",
		GeneratedBy:    "llm_agent",
		AgentStatus:    agentStatus,
		FallbackReason: "none",
		RepairApplied:  false,
		SchemaValid:    true,
	}

	// L2: extract L4 calibration fields
	var l4GlobalSummary, candidatePos, topRecReason, portfolioRisk, deliveryStrat string
	if l4Result != nil {
		l4GlobalSummary = l4Result.GlobalSummary
		candidatePos = l4Result.CandidatePositioning
		topRecReason = l4Result.TopRecommendationReason
		portfolioRisk = l4Result.PortfolioRiskSummary
		deliveryStrat = l4Result.DeliveryStrategy
	}
	if l4GlobalSummary == "" {
		l4GlobalSummary = globalSummary
	}

	resp := &RecommendationResponse{
		ResumeVersionID:        resume.ID,
		ResumeID:               resume.ID,
		GeneratedBy:            "llm_agent",
		RecommendedJobs:        recommendedJobs,
		Summary:                l4GlobalSummary,
		CreatedAt:              now,
		Meta:                   meta,
		GenerationTier:         "deep",
		PipelineMode:           "cascade",
		ProfileStatus:          "verified",
		ProfileType:            profileType,
		ProfileSource:          profileSource,
		ProfileDominantProfile: profileDominant,
		Result: &RecommendationBusinessResult{
			RecommendedJobs:         recommendedJobs,
			Summary:                 l4GlobalSummary,
			QualityStatus:           agentStatus,
			PipelineTrace:           string(traceBytes),
			GlobalSummary:           l4GlobalSummary,
			CandidatePositioning:    candidatePos,
			TopRecommendationReason: topRecReason,
			PortfolioRiskSummary:    portfolioRisk,
			DeliveryStrategy:        deliveryStrat,
		},
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)

	// J2: Persist Deep Cascade result to recommendation_results as full recommendedJobs array
	if len(recommendedJobs) > 0 {
		resultJSON, marshalErr := json.Marshal(recommendedJobs)
		if marshalErr == nil {
			recResult := &model.RecommendationResult{
				UserID:                 userID,
				ResumeID:               resume.ID,
				GenerationTier:         "quality",
				GeneratedBy:            "llm_agent",
				Status:                 "ready",
				ProfileType:            profileType,
				ProfileDominantProfile: profileDominant,
				ResultJSON:             resultJSON,
			}
			if createErr := s.recommendationRepo.Create(ctx, recResult); createErr == nil {
				resp.ResultID = recResult.ID
				resp.GenerationTier = "quality"
			}
		}
	}

	return resp, nil
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
		ResumeID:        resume.ID,
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

	// K15-P0: Deep fallback semantics — preserve Deep tier metadata even on Mock fallback
	generatedBy := model.GeneratedByMockAgent
	pipelineMode := "fast_rule"
	responseTier := "fast"
	dbTier := "fast"

	if qualityMode {
		if fbErr != nil {
			// Deep request that fell back → dedicated deep_fallback_agent metadata
			generatedBy = model.GeneratedByDeepFallbackAgent
			pipelineMode = "deep_fallback"
			responseTier = "deep"
			dbTier = "quality"
		} else {
			// Quality mock without explicit error (direct mock, no LLM attempted)
			generatedBy = model.GeneratedByMockQuality
			pipelineMode = "mock_deep_fallback"
			responseTier = "deep"
			dbTier = "quality"
		}
	}

	var meta model.AgentResponseMeta
	for i := range recommendedJobs {
		recommendedJobs[i].GeneratedBy = generatedBy
		if fbErr != nil {
			recommendedJobs[i].AgentStatus = model.AgentStatusFallback
			recommendedJobs[i].FallbackReason = fbErr.FallbackReason
			recommendedJobs[i].ErrorCode = fbErr.ErrorCode
			recommendedJobs[i].RepairApplied = false
		} else {
			recommendedJobs[i].AgentStatus = model.AgentStatusMockSuccess
		}
	}

	// Build meta from first job or fallback info — K15-P0: use deep_fallback_agent for quality+fallback
	if fbErr != nil {
		if qualityMode {
			meta = model.NewAgentResponseMeta(model.AgentNameRecommendation).
				WithDeepFallback(fbErr.FallbackReason, fbErr.ErrorCode)
		} else {
			meta = model.NewAgentResponseMeta(model.AgentNameRecommendation).
				WithFallback(fbErr.FallbackReason, fbErr.ErrorCode)
		}
	} else {
		if qualityMode {
			meta = model.NewAgentResponseMeta(model.AgentNameRecommendation).
				WithGeneratedBy(model.GeneratedByMockQuality).
				WithAgentStatus(model.AgentStatusMockSuccess)
		} else {
			meta = model.NewAgentResponseMeta(model.AgentNameRecommendation)
		}
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
		ResumeID:        resume.ID,
		GeneratedBy:     generatedBy,
		RecommendedJobs: jobPtrs,
		Result:          recResult,
		Summary:         summary,
		CreatedAt:       now.Format("2006-01-02T15:04:05Z07:00"),
		Meta:            meta,
		GenerationTier:  responseTier,
		PipelineMode:    pipelineMode,
		JudgeResult:     &model.RecommendationJudgeResult{Pass: true, Severity: "pass", Issues: nil, FixStrategy: "none"},
	}

	// K9-1-U-C: Save recommendation result snapshot with full metadata
	// K15-P0-R2: use safe background-derived context for DB persistence.
	// The original request context may have been cancelled by this point.
	if len(recommendedJobs) > 0 {
		resultJSON, err := json.Marshal(recommendedJobs)
		if err != nil {
			log.Printf("[RecommendationService] failed to marshal recommendation result for fallback: %v", err)
			// Don't fail the response — the return value is already built.
			s.fillProfileFields(resp, profileType, profileDominant, profileSource)
			return resp, nil
		}

		recSnapshot := &model.RecommendationResult{
			UserID:                 userID,
			ResumeID:               resume.ID,
			GenerationTier:         dbTier,
			GeneratedBy:            generatedBy,
			Status:                 "ready",
			ProfileType:            profileType,
			ProfileDominantProfile: profileDominant,
			ResultJSON:             resultJSON,
		}

		saveCtx, saveCancel := safeFallbackCtx()
		createErr := s.recommendationRepo.Create(saveCtx, recSnapshot)
		saveCancel()
		if createErr != nil {
			log.Printf("[RecommendationService] failed to save fallback result snapshot (non-fatal): %v", createErr)
		} else {
			resp.ResultID = recSnapshot.ID
			resp.GenerationTier = responseTier
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
// If tier is non-empty, filter by generation_tier; otherwise return most recent non-mock.
func (s *RecommendationService) GetLatestRecommendation(ctx context.Context, userID, resumeID uint64, tier string) (*model.RecommendationLatestResult, error) {
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

	latest, err := s.recommendationRepo.FindLatestByResumeID(ctx, userID, resumeID, tier)
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
		scaledScore := int(math.Min(95, math.Max(70, 70+c.TagScore*0.5)))
		rec := &model.RecommendedJob{
			JobID:                  c.JobID,
			Title:                  title,
			Department:             dept,
			City:                   city,
			Category:               c.Category,
			JobType:                "",
			RecommendationScore:    scaledScore,
			RecommendationReason:   buildFastOneSentenceReason(c),
			RecommendationLabel:    buildFastLabel(c.TagScore),
			RecommendationPriority: buildFastPriority(c.TagScore),
			MatchHighlights:        buildFastHighlights(c),
			RiskPoints:             buildFastRiskReminders(c),
			SuggestedAction:        buildFastSuggestedAction(c),
			WhyRankedHere:          buildFastRecommendReasonExplainer(c),
			GeneratedBy:            "fast_tag_engine",
			AgentStatus:            "success",
			FallbackReason:         "none",
			RepairApplied:          false,
			DataQuality:            "rule",
			ResumeEvidence:         c.EvidenceRefs,
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
		ResumeVersionID:        resume.ID,
		ResumeID:               resume.ID,
		GeneratedBy:            "fast_tag_engine",
		RecommendedJobs:        recommended,
		Summary:                summary,
		CreatedAt:              now,
		Meta:                   meta,
		GenerationTier:         "fast",
		PipelineMode:           "fast_tag_engine",
		ProfileStatus:          "ready",
		ProfileType:            profileType,
		ProfileSource:          profileSource,
		ProfileDominantProfile: profileDominant,
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)
	return resp, nil
}

// ─── Fast Recommendation via SEP ────────────────────────────────────────

// fastRecommendWithSEP generates Fast-tier recommendations using SEP semantic profiles.
// This function is strictly read-only with respect to SEP generation:
// it reads ready verified SEP profiles only and falls back quickly when SEP is missing.
// Missing profiles trigger async background generation, but the current request never waits for LLM.
func (s *RecommendationService) fastRecommendWithSEP(ctx context.Context, userID uint64, resume *model.ResumeVersion, allJobs []*model.Job, limit int, profileType, profileDominant, profileSource string) (*RecommendationResponse, error) {
	if s.semanticGate == nil || s.semanticRepo == nil {
		return nil, fmt.Errorf("SEP services not available")
	}

	resumeID := int64(resume.ID)

	// 1. Read resume SEP (read-only — no generation)
	resumeProfile, err := s.findReadyResumeSEP(ctx, resumeID)
	if err != nil {
		return nil, fmt.Errorf("SEP resume profile lookup: %w", err)
	}
	if resumeProfile == nil {
		// Schedule async background generation, then fall back immediately
		log.Printf("[SEP-Fast] resume %d: SEP not ready, scheduling async generation", resumeID)
		s.semanticGate.EnsureResumeProfileAsync(resumeID)
		return nil, ErrFastRecommendSEPMissing
	}

	// 2. Read resume SEP tags
	resumeTags, err := s.semanticRepo.ListTagsByProfileID(ctx, resumeProfile.ID)
	if err != nil {
		return nil, fmt.Errorf("SEP read resume tags: %w", err)
	}
	if len(resumeTags) < 3 {
		log.Printf("[SEP-Fast] resume %d: SEP profile %d has only %d tags (<3), triggering async regeneration",
			resumeID, resumeProfile.ID, len(resumeTags))
		s.semanticGate.EnsureResumeProfileAsync(resumeID)
		return nil, ErrFastRecommendSEPMissing
	}

	// 3. Read job SEP (read-only — no generation)
	jobMap := make(map[int64]*model.Job)
	for _, j := range allJobs {
		jobMap[int64(j.ID)] = j
	}

	var candidates []*SEPRecommendCandidate
	skipped := 0
	for _, j := range allJobs {
		jid := int64(j.ID)
		jobProfile, err := s.findReadyJobSEP(ctx, jid)
		if err != nil {
			log.Printf("[SEP-Fast] skip job %d: SEP lookup error: %v", jid, err)
			skipped++
			continue
		}
		if jobProfile == nil {
			log.Printf("[SEP-Fast] job %d: SEP not ready, scheduling async generation", jid)
			s.semanticGate.EnsureJobProfileAsync(jid)
			skipped++
			continue
		}
		jobTags, err := s.semanticRepo.ListTagsByProfileID(ctx, jobProfile.ID)
		if err != nil || len(jobTags) < 3 {
			log.Printf("[SEP-Fast] skip job %d: tags insufficient (%d, err=%v)", jid, len(jobTags), err)
			skipped++
			continue
		}
		candidates = append(candidates, &SEPRecommendCandidate{
			JobID:     jid,
			JobTitle:  j.Title,
			Direction: j.Direction,
			Category:  j.Category,
			JobTags:   jobTags,
		})
	}
	log.Printf("[SEP-Fast] candidates=%d skipped=%d resume_tags=%d", len(candidates), skipped, len(resumeTags))

	if len(candidates) == 0 {
		return nil, fmt.Errorf("SEP: all %d jobs skipped (SEP profiles not available)", len(allJobs))
	}

	// 4. SEP ranking
	scores := rankSEPCandidates(resumeTags, candidates)

	// 5. Build RecommendedJob entries via GQCL Composer
	var recommended []*model.RecommendedJob
	for i, score := range scores {
		if i >= limit {
			break
		}
		j := jobMap[score.JobID]
		title, dept, city, direction, category := "", "", "", "", ""
		if j != nil {
			title, dept, city = j.Title, j.Department, j.City
			direction, category = j.Direction, j.Category
		}

		rec := &model.RecommendedJob{
			JobID:                  uint64(score.JobID),
			Title:                  title,
			Department:             dept,
			City:                   city,
			Category:               category,
			JobType:                direction,
			RecommendationScore:    score.Score,
			RecommendationReason:   buildSEPOneSentenceReason(score),
			RecommendationLabel:    sepScoreLabel(score.Score),
			RecommendationPriority: sepScorePriority(score.Score),
			MatchHighlights:        buildSEPMatchHighlights(score),
			RiskPoints:             buildSEPRiskSummaries(score, title),
			SuggestedAction:        sepScoreAction(score.Score),
			WhyRankedHere:          score.RankReason,
			DisplayTags:            score.DisplayTags,
			GeneratedBy:            "fast_tag_engine",
			AgentStatus:            "success",
			FallbackReason:         "none",
			RepairApplied:          false,
			DataQuality:            "rule",
			ResumeEvidence:         score.EvidenceTexts,
		}
		recommended = append(recommended, rec)
	}

	// 6. Build response
	summary := fmt.Sprintf("基于 SEP 语义画像匹配 %d 个岗位，共推荐 %d 个岗位", len(allJobs)-skipped, len(recommended))
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
		ResumeVersionID:        resume.ID,
		ResumeID:               resume.ID,
		GeneratedBy:            "fast_tag_engine",
		RecommendedJobs:        recommended,
		Summary:                summary,
		CreatedAt:              now,
		Meta:                   meta,
		GenerationTier:         "fast",
		PipelineMode:           "sep_ranker",
		ProfileStatus:          "verified",
		ProfileType:            profileType,
		ProfileSource:          profileSource,
		ProfileDominantProfile: profileDominant,
	}

	s.fillProfileFields(resp, profileType, profileDominant, profileSource)
	return resp, nil
}

// ─── SEP Read-only Helpers ───────────────────────────────────────────────

// findReadyResumeSEP returns a verified resume SEP profile, or nil if not ready.
// This is a read-only lookup — it never triggers profile generation.
func (s *RecommendationService) findReadyResumeSEP(ctx context.Context, resumeID int64) (*model.SemanticProfile, error) {
	p, err := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetResume, resumeID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status != model.SemanticProfileStatusVerified {
		return nil, nil
	}
	return p, nil
}

// findReadyJobSEP returns a verified job SEP profile, or nil if not ready.
// This is a read-only lookup — it never triggers profile generation.
func (s *RecommendationService) findReadyJobSEP(ctx context.Context, jobID int64) (*model.SemanticProfile, error) {
	p, err := s.semanticRepo.FindLatestByTarget(ctx, model.SemanticTargetJob, jobID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status != model.SemanticProfileStatusVerified {
		return nil, nil
	}
	return p, nil
}

// ─── SEP Recommend Ranking — kept from original, unchanged ───────────────

func jobResultStatus(r *SemanticProfileGateResult) string {
	if r == nil {
		return "nil"
	}
	return r.Status
}

// ─── SEP Composer Helper ─────────────────────────────────────────────────

func buildSEPOneSentenceReason(s *SEPMatchScore) string {
	top := s.MatchedTags
	if len(top) > 2 {
		top = top[:2]
	}
	core := strings.Join(top, "、")
	if core == "" {
		core = "基础背景"
	}
	switch {
	case s.Score >= 85:
		return fmt.Sprintf("你的简历与该岗位在 %s 上重合度较高，适合作为优先投递岗位。", core)
	case s.Score >= 70:
		return fmt.Sprintf("你的简历与该岗位在 %s 上存在较明确的能力交集，适合作为重点关注岗位。", core)
	default:
		return fmt.Sprintf("该岗位与当前简历在 %s 上存在一定关联，建议先补强相关经历后再投递。", core)
	}
}

func buildSEPMatchHighlights(s *SEPMatchScore) []string {
	var result []string
	for _, ev := range s.EvidenceTexts {
		if len(result) >= 3 {
			break
		}
		result = append(result, ev)
	}
	if len(result) == 0 && len(s.MatchedTags) > 0 {
		result = append(result, fmt.Sprintf("简历与岗位在 %s 等方面存在语义匹配", strings.Join(s.MatchedTags[:min(3, len(s.MatchedTags))], "、")))
	}
	if len(result) == 0 {
		result = append(result, "该岗位与当前简历在基础背景上存在一定关联，建议进一步做岗位匹配分析确认。")
	}
	return result
}

func buildSEPRiskSummaries(s *SEPMatchScore, jobTitle string) []string {
	// M2: de-template — generate specific risk descriptions
	if len(s.RiskTags) == 0 {
		if s.RoleMatch == 0 {
			return []string{"当前岗位方向与简历主方向存在差异，建议优先关注方向匹配度更高的岗位。"}
		}
		return []string{"当前未发现明显硬性风险，仍建议结合岗位详情进一步确认。"}
	}
	var result []string
	for _, r := range s.RiskTags {
		if len(result) >= 2 {
			break
		}
		risk := fmt.Sprintf("岗位要求关注「%s」，当前简历在该方面的体现较弱，可能会降低面试机会。", r)
		if s.RoleMatch == 0 {
			risk += fmt.Sprintf(" 结合岗位类型（%s），建议优先确认方向匹配后再投递。", jobTitle)
		}
		result = append(result, risk)
	}
	return result
}

func sepScoreLabel(score int) string {
	switch {
	case score >= 85:
		return "优先投递"
	case score >= 70:
		return "重点关注"
	case score >= 55:
		return "可以备选"
	default:
		return "谨慎尝试"
	}
}

func sepScorePriority(score int) string {
	switch {
	case score >= 85:
		return "strong"
	case score >= 70:
		return "high"
	case score >= 55:
		return "try"
	default:
		return "backup"
	}
}

func sepScoreAction(score int) string {
	switch {
	case score >= 85:
		return "建议进入岗位详情确认要求，再进行深度匹配分析，判断是否适合作为重点投递岗位。"
	case score >= 70:
		return "建议先做岗位匹配分析，确认主要风险后再决定是否投递。"
	case score >= 55:
		return "建议作为备选岗位关注，优先补充相关项目或实习证据。"
	default:
		return "建议暂不作为优先投递目标，先优化简历中的相关能力表达。"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			JobID:                  item.JobID,
			Title:                  title,
			Department:             dept,
			City:                   city,
			RecommendationScore:    item.RecommendationScore,
			RecommendationReason:   item.RecommendationReason,
			RecommendationLabel:    item.RecommendationLevel,
			RecommendationPriority: item.RecommendationLevel,
			MatchHighlights:        item.MatchHighlights,
			RiskPoints:             item.RiskPoints,
			ResumeEvidence:         item.ResumeEvidence,
			SuggestedAction:        item.SuggestedAction,
			ApplyReadiness:         item.OptimizationSignal,
			GeneratedBy:            result.FinalOutput.GeneratedBy,
			AgentStatus:            result.FinalOutput.AgentStatus,
			RepairApplied:          result.FinalOutput.RepairApplied,
			DataQuality:            "llm",
		}
		recommended = append(recommended, rec)
	}

	meta := model.AgentResponseMeta{
		AgentName:     "DeepRecommendationEngine",
		GeneratedBy:   result.FinalOutput.GeneratedBy,
		AgentStatus:   result.FinalOutput.AgentStatus,
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
	if score >= 15 {
		return "强匹配"
	}
	if score >= 8 {
		return "较高匹配"
	}
	return "可关注"
}

func buildFastPriority(score float64) string {
	if score >= 15 {
		return "strong"
	}
	if score >= 8 {
		return "high"
	}
	return "try"
}

func convertProfileTags(tags []model.ProfileTag) []agent.StoredTag {
	result := make([]agent.StoredTag, len(tags))
	for i, t := range tags {
		result[i] = agent.StoredTag{Tag: t.Tag, Group: t.Group, Confidence: t.Confidence, Evidence: t.Evidence}
	}
	return result
}
