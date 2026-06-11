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

// VersionCompareRequest represents a request for version comparison.
type VersionCompareRequest struct {
	FromResumeID  uint64
	ToResumeID    uint64
	GenerationTier model.GenerationTier
}

// VersionCompareResponse represents the response for version comparison.
type VersionCompareResponse struct {
	CompareID   uint64                        `json:"compare_id"`
	GeneratedBy string                        `json:"generated_by"`
	Result      *model.VersionCompareAnalysis `json:"result"`
	CreatedAt   string                        `json:"created_at"`
}

// VersionCompareService handles version comparison business logic.
// Supports generation_tier: "fast" uses fastRunner, "quality" uses qualityRunner.
type VersionCompareService struct {
	resumeRepo    *repository.ResumeRepository
	compareRepo   *repository.ResumeCompareRepository
	jobRepo       *repository.JobRepository
	mockAgent     *agent.VersionCompareMockAgent
	fastRunner    agent.PipelineRunner
	qualityRunner agent.PipelineRunner
	llmEnabled    bool
	fallbackEnable bool
	enableMockAgent bool
	profileSvc    *ResumeProfileService
}

// NewVersionCompareService creates a new VersionCompareService.
func NewVersionCompareService(
	resumeRepo *repository.ResumeRepository,
	compareRepo *repository.ResumeCompareRepository,
	jobRepo *repository.JobRepository,
	mockAgent *agent.VersionCompareMockAgent,
	fastRunner agent.PipelineRunner,
	qualityRunner agent.PipelineRunner,
	llmEnabled bool,
	fallbackEnable bool,
	enableMockAgent bool,
	profileSvc *ResumeProfileService,
) *VersionCompareService {
	return &VersionCompareService{
		resumeRepo:    resumeRepo,
		compareRepo:   compareRepo,
		jobRepo:       jobRepo,
		mockAgent:     mockAgent,
		fastRunner:    fastRunner,
		qualityRunner: qualityRunner,
		llmEnabled:    llmEnabled,
		fallbackEnable: fallbackEnable,
		enableMockAgent: enableMockAgent,
		profileSvc:    profileSvc,
	}
}

// selectRunner returns the appropriate pipeline runner for the given tier.
func (s *VersionCompareService) selectRunner(tier model.GenerationTier) agent.PipelineRunner {
	// tier validated by handler
	if tier == model.GenerationTierQuality && s.qualityRunner != nil {
		return s.qualityRunner
	}
	return s.fastRunner
}

// Compare generates a version comparison for the given user's two resume versions.
func (s *VersionCompareService) Compare(ctx context.Context, userID uint64, req VersionCompareRequest) (*VersionCompareResponse, error) {
	runner := s.selectRunner(req.GenerationTier)
	log.Printf("[VersionCompareService] generation_tier=%s", string(req.GenerationTier))

	// Validate from/to
	if req.FromResumeID <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid from version number", 400)
	}
	if req.ToResumeID <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid to version number", 400)
	}
	if req.FromResumeID == req.ToResumeID {
		return nil, apperror.New(apperror.CodeInvalidParam, "from and to version numbers must be different", 400)
	}

	// Get from version
	fromVersion, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, req.FromResumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}
	if fromVersion == nil {
		return nil, apperror.ErrResumeFromVerNotFound
	}

	// Get to version
	toVersion, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, req.ToResumeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}
	if toVersion == nil {
		return nil, apperror.ErrResumeToVerNotFound
	}

	// Try LLM Pipeline first if enabled
	if s.llmEnabled && runner != nil {
		log.Printf("[VersionCompareService] use_llm=true tier=%s", string(req.GenerationTier))
		llmAgent := agent.NewVersionCompareLLMAgent(runner)
		result, fallbackErr := llmAgent.Compare(fromVersion, toVersion)
		if fallbackErr == nil {
			resp := &VersionCompareResponse{
				CompareID:   0, // will be set after snapshot save
				GeneratedBy: result.GeneratedBy,
				Result:      result,
				CreatedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
			}
			return resp, nil
		}
		log.Printf("[VersionCompareService] LLM pipeline fallback: %v", fallbackErr)

		// K13-Fix-1: Check fallback enable flag
		if !s.fallbackEnable {
			reason := "llm_failed"
			if fbErr, ok := fallbackErr.(*agent.VersionCompareFallbackError); ok {
				reason = fbErr.FallbackReason
			}
			return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed and fallback disabled: %s", reason), 500)
		}
	} else if !s.enableMockAgent {
		// LLM disabled — check if mock is allowed
		return nil, apperror.New(apperror.CodeInternal, "LLM is disabled and ENABLE_MOCK_AGENT=false, no generation path available", 500)
	}

	// Fall back to Mock Agent
	log.Printf("[VersionCompareService] use_llm=false use_mock=true tier=%s", string(req.GenerationTier))
	return s.compareWithMock(ctx, userID, fromVersion, toVersion)
}

func (s *VersionCompareService) compareWithMock(ctx context.Context, userID uint64, fromVer *model.ResumeVersion, toVer *model.ResumeVersion) (*VersionCompareResponse, error) {
	result, err := s.mockAgent.Compare(fromVer, toVer)
	if err != nil {
		return nil, fmt.Errorf("failed to generate version comparison: %w", err)
	}

	now := time.Now()
	resp := &VersionCompareResponse{
		GeneratedBy: model.GeneratedByMockAgent,
		Result:      result,
		CreatedAt:   now.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Save comparison snapshot
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comparison result: %w", err)
	}

	compareResult := &model.VersionCompareResult{
		UserID:        userID,
		FromResumeID: fromVer.ID,
		ToResumeID:   toVer.ID,
		ResultJSON:    resultJSON,
	}

	if err := s.compareRepo.Create(ctx, compareResult); err != nil {
		// Log but don't fail the response if snapshot save fails
		_ = err
	}

	resp.CompareID = compareResult.ID

	return resp, nil
}

// ── K9-4-C: CompareV2 ─────────────────────────────────────────────────────

// CompareV2 generates a profile-aware resume comparison for the V2 API.
func (s *VersionCompareService) CompareV2(ctx context.Context, userID uint64, req model.CompareRequestV2) (*model.CompareResponseV2, error) {
	// Validate
	if req.FromResumeID <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "from_resume_id is required", 400)
	}
	if req.ToResumeID <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "to_resume_id is required", 400)
	}
	if req.FromResumeID == req.ToResumeID {
		return nil, apperror.New(apperror.CodeInvalidParam, "from_resume_id and to_resume_id must be different", 400)
	}

	// Normalize tier
	tier, err := model.NormalizeGenerationTier(req.GenerationTier)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid generation_tier", 400)
	}
	qualityMode := tier == model.GenerationTierQuality

	// Validate mode
	mode := req.ComparisonMode
	if mode == "" { mode = "general" }
	if mode != "general" && mode != "target_direction" && mode != "target_job" {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid comparison_mode", 400)
	}
	if mode == "target_direction" && req.TargetDirection == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "target_direction is required for target_direction mode", 400)
	}
	if mode == "target_job" && req.JobID == 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "job_id is required for target_job mode", 400)
	}

	// Validate job exists for target_job mode
	if mode == "target_job" && s.jobRepo != nil {
		job, err := s.jobRepo.FindByID(ctx, req.JobID)
		if err != nil || job == nil {
			return nil, apperror.New(apperror.CodeInvalidParam, "job not found", 404)
		}
	}

	// Load resumes
	fromVer, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, req.FromResumeID)
	if err != nil || fromVer == nil {
		return nil, apperror.ErrResumeNotFound
	}
	toVer, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, req.ToResumeID)
	if err != nil || toVer == nil {
		return nil, apperror.ErrResumeToVerNotFound
	}

	// Get profiles
	fromProf, toProf := "", ""
	if s.profileSvc != nil {
		fromP, _, _ := s.profileSvc.EnsureProfile(ctx, userID, req.FromResumeID, "fast")
		if fromP != nil { fromProf = fromP.DominantProfile }
		toP, _, _ := s.profileSvc.EnsureProfile(ctx, userID, req.ToResumeID, "fast")
		if toP != nil { toProf = toP.DominantProfile }
	}

	// K13-Fix-1: Try LLM pipeline first for quality tier
	genTier := string(tier)
	runner := s.selectRunner(tier)
	if s.llmEnabled && runner != nil && qualityMode {
		log.Printf("[VersionCompareService] CompareV2 use_llm=true tier=%s", genTier)
		llmAgent := agent.NewVersionCompareLLMAgent(runner)
		llmResult, fbErr := llmAgent.Compare(fromVer, toVer)
		if fbErr == nil {
			// Map LLM result to V2 format
			result := mapLLMResultToV2(llmResult, req.FromResumeID, req.ToResumeID, mode, req.TargetDirection, req.JobID, qualityMode, fromProf, toProf, genTier)
			if result != nil {
				// Save to DB
				if err := s.saveCompareV2(ctx, userID, result, mode, string(tier), req.TargetDirection, req.JobID); err != nil {
					log.Printf("[VersionCompareService] failed to save LLM V2 compare: %v", err)
				}
				return result, nil
			}
			log.Printf("[VersionCompareService] LLM mapping to V2 format failed, falling back to mock")
		} else {
			log.Printf("[VersionCompareService] CompareV2 LLM fallback: %v", fbErr)
			if !s.fallbackEnable {
				return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed and fallback disabled: %v", fbErr), 500)
			}
		}
	}

	// Run mock agent
	log.Printf("[VersionCompareService] CompareV2 use_llm=false use_mock=true tier=%s", genTier)
	result := s.mockAgent.CompareV2(fromVer, toVer, mode, req.TargetDirection, req.JobID, qualityMode, fromProf, toProf)
	result.FromResumeID = req.FromResumeID
	result.ToResumeID = req.ToResumeID
	result.ComparisonMode = mode
	if mode == "target_direction" { result.TargetDirection = req.TargetDirection }
	if mode == "target_job" { var jid = req.JobID; result.JobID = &jid }

	// Save to DB
	if err := s.saveCompareV2(ctx, userID, result, mode, string(tier), req.TargetDirection, req.JobID); err != nil {
		log.Printf("[VersionCompareService] failed to save V2 compare: %v", err)
	}

	return result, nil
}

// saveCompareV2 persists the compare result.
func (s *VersionCompareService) saveCompareV2(ctx context.Context, userID uint64, result *model.CompareResponseV2, mode, tier, targetDir string, jobID uint64) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal compare result: %w", err)
	}

	var jobIDPtr *uint64
	if mode == "target_job" { jobIDPtr = &jobID }
	targetDirection := ""
	if mode == "target_direction" { targetDirection = targetDir }

	dbResult := &model.ResumeCompareResultV2{
		UserID:                     userID,
		FromResumeID:               result.FromResumeID,
		ToResumeID:                 result.ToResumeID,
		ComparisonMode:             mode,
		GenerationTier:             tier,
		GeneratedBy:                result.GeneratedBy,
		Status:                     "ready",
		TargetDirection:            targetDirection,
		JobID:                      jobIDPtr,
		FromProfileType:            result.FromProfileType,
		FromProfileDominantProfile: result.FromProfileDominantProfile,
		ToProfileType:              result.ToProfileType,
		ToProfileDominantProfile:   result.ToProfileDominantProfile,
		ResultJSON:                 resultJSON,
	}

	if err := s.compareRepo.CreateV2(ctx, dbResult); err != nil {
		return err
	}
	result.CompareID = dbResult.ID
	return nil
}

// mapLLMResultToV2 maps a V1 VersionCompareAnalysis to V2 CompareResponseV2.
func mapLLMResultToV2(llm *model.VersionCompareAnalysis, fromID, toID uint64, mode, targetDir string, jobID uint64, qualityMode bool, fromProf, toProf, genTier string) *model.CompareResponseV2 {
	result := &model.CompareResponseV2{
		FromResumeID:               fromID,
		ToResumeID:                 toID,
		ComparisonMode:             mode,
		GenerationTier:             genTier,
		GeneratedBy:                llm.GeneratedBy,
		Status:                     "ready",
		FromProfileType:            genTier,
		ToProfileType:              genTier,
		FromProfileDominantProfile: fromProf,
		ToProfileDominantProfile:   toProf,
		CompareScore:               llm.ScoreAfter,
		CompareLevel:               llm.OverallChange,
		WinnerResumeID:             toID, // default to "to" version
		WinnerLabel:                llm.OverallChange,
		OneSentenceSummary:         llm.Summary,
		FromStrengths:              nil,
		ToStrengths:                nil,
		FromRisks:                  llm.RemainingRisks,
		ToRisks:                    nil,
		KeyDifferences:             llm.KeyImprovements,
		MergeSuggestions:           nil,
		DeliveryAdvice:             llm.Recommendation,
	}
	if mode == "target_direction" { result.TargetDirection = targetDir }
	if mode == "target_job" { var jid = jobID; result.JobID = &jid }
	return result
}

// ── K9-4-D: GetLatestCompareV2 ─────────────────────────────────────────────

// GetLatestCompareV2 returns the latest ready compare result for the given criteria.
func (s *VersionCompareService) GetLatestCompareV2(ctx context.Context, userID, fromResumeID, toResumeID uint64, mode, targetDirection string, jobID uint64) (*model.CompareLatestResponseV2, error) {
	if fromResumeID <= 0 || toResumeID <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "from_resume_id and to_resume_id are required", 400)
	}
	if fromResumeID == toResumeID {
		return nil, apperror.New(apperror.CodeInvalidParam, "from_resume_id and to_resume_id must be different", 400)
	}
	if mode == "" { mode = "general" }
	if mode != "general" && mode != "target_direction" && mode != "target_job" {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid comparison_mode", 400)
	}
	if mode == "target_direction" && targetDirection == "" {
		return nil, apperror.New(apperror.CodeInvalidParam, "target_direction is required", 400)
	}
	if mode == "target_job" && jobID == 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "job_id is required", 400)
	}
	// Validate job exists for target_job mode
	if mode == "target_job" && s.jobRepo != nil {
		job, err := s.jobRepo.FindByID(ctx, jobID)
		if err != nil || job == nil {
			return nil, apperror.New(apperror.CodeInvalidParam, "job not found", 404)
		}
	}

	// Verify resumes belong to user
	fromVer, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, fromResumeID)
	if err != nil || fromVer == nil {
		return nil, apperror.ErrResumeNotFound
	}
	toVer, err := s.resumeRepo.FindByUserIDAndID(ctx, userID, toResumeID)
	if err != nil || toVer == nil {
		return nil, apperror.ErrResumeToVerNotFound
	}

	latest, err := s.compareRepo.FindLatestV2(ctx, userID, fromResumeID, toResumeID, mode, targetDirection, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to find latest compare: %w", err)
	}
	if latest == nil {
		return &model.CompareLatestResponseV2{HasResult: false}, nil
	}

	var resultData interface{}
	if len(latest.ResultJSON) > 0 {
		if err := json.Unmarshal(latest.ResultJSON, &resultData); err != nil {
			return nil, fmt.Errorf("failed to parse saved compare result: %w", err)
		}
	}

	return &model.CompareLatestResponseV2{
		HasResult:                  true,
		CompareID:                  latest.ID,
		FromResumeID:               latest.FromResumeID,
		ToResumeID:                 latest.ToResumeID,
		ComparisonMode:             latest.ComparisonMode,
		TargetDirection:            latest.TargetDirection,
		JobID:                      latest.JobID,
		GenerationTier:             latest.GenerationTier,
		GeneratedBy:                latest.GeneratedBy,
		Status:                     latest.Status,
		FromProfileType:            latest.FromProfileType,
		FromProfileDominantProfile: latest.FromProfileDominantProfile,
		ToProfileType:              latest.ToProfileType,
		ToProfileDominantProfile:   latest.ToProfileDominantProfile,
		CreatedAt:                  latest.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Result:                     resultData,
	}, nil
}
