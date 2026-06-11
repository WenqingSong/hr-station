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
	tagEvidence   *TagEvidenceService // K13-Fix-C4
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
	tagEvidence *TagEvidenceService,
) *VersionCompareService {
	return &VersionCompareService{
		resumeRepo:    resumeRepo, compareRepo: compareRepo, jobRepo: jobRepo,
		mockAgent: mockAgent, fastRunner: fastRunner, qualityRunner: qualityRunner,
		llmEnabled: llmEnabled, fallbackEnable: fallbackEnable, enableMockAgent: enableMockAgent,
		profileSvc: profileSvc, tagEvidence: tagEvidence,
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

	// K13-Fix-C4: Fast tier → TagEvidence engine (no LLM)
	genTier := string(tier)
	runner := s.selectRunner(tier)

	if !qualityMode && s.tagEvidence != nil {
		log.Printf("[VersionCompareService] CompareV2 use_fast_tag_engine tier=%s from=%d to=%d", genTier, req.FromResumeID, req.ToResumeID)
		result, err := s.fastCompareWithTags(ctx, userID, req.FromResumeID, req.ToResumeID, mode, req.TargetDirection, req.JobID, fromProf, toProf)
		if err == nil { return result, nil }
		log.Printf("[VersionCompareService] fast tag compare failed: %v, falling back to mock", err)
	}

	// K13-Fix-C4: Quality tier → 3-agent pipeline
	if qualityMode && s.llmEnabled && runner != nil {
		if s.tagEvidence != nil {
			log.Printf("[VersionCompareService] CompareV2 use_deep_3agent tier=%s", genTier)
			result, err3a := s.compareWith3Agents(ctx, userID, req.FromResumeID, req.ToResumeID, mode, req.TargetDirection, req.JobID, fromProf, toProf, runner)
			if err3a == nil { return result, nil }
			log.Printf("[VersionCompareService] 3-agent compare failed: %v, falling back to old LLM", err3a)
		}
		llmAgent := agent.NewVersionCompareLLMAgent(runner)
		llmResult, fbErr := llmAgent.Compare(fromVer, toVer)
		if fbErr == nil {
			result := mapLLMResultToV2(llmResult, req.FromResumeID, req.ToResumeID, mode, req.TargetDirection, req.JobID, qualityMode, fromProf, toProf, genTier)
			if result != nil { s.saveCompareV2(ctx, userID, result, mode, string(tier), req.TargetDirection, req.JobID); return result, nil }
		}
		log.Printf("[VersionCompareService] LLM fallback: %v", fbErr)
		if !s.fallbackEnable { return nil, apperror.New(apperror.CodeInternal, fmt.Sprintf("LLM failed: %v", fbErr), 500) }
	}

	// Fallback to mock
	log.Printf("[VersionCompareService] CompareV2 use_mock_fallback tier=%s", genTier)
	result := s.mockAgent.CompareV2(fromVer, toVer, mode, req.TargetDirection, req.JobID, qualityMode, fromProf, toProf)
	result.FromResumeID = req.FromResumeID; result.ToResumeID = req.ToResumeID; result.ComparisonMode = mode
	if mode == "target_direction" { result.TargetDirection = req.TargetDirection }
	if mode == "target_job" { var jid = req.JobID; result.JobID = &jid }
	s.saveCompareV2(ctx, userID, result, mode, string(tier), req.TargetDirection, req.JobID)
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

// ─── K13-Fix-C4: fastCompareWithTags ──────────────────────────────────

func (s *VersionCompareService) fastCompareWithTags(ctx context.Context, userID uint64, fromID, toID uint64, mode, targetDir string, jobID uint64, fromProf, toProf string) (*model.CompareResponseV2, error) {
	fp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, fromID, userID)
	if err != nil { return nil, fmt.Errorf("fast compare: from profile: %w", err) }
	tp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, toID, userID)
	if err != nil { return nil, fmt.Errorf("fast compare: to profile: %w", err) }

	diff := buildCompareDiffSummary(fp, tp)
	score := 70
	winner := toID
	if len(diff.AddedTags) > 2 || tp.Evidence.ProjectCount > fp.Evidence.ProjectCount { score = 78 }
	if tp.Evidence.ProjectCount <= fp.Evidence.ProjectCount && len(diff.AddedTags) == 0 { score = 60; winner = fromID }

	summary := "目标版本在项目经历和标签上有明显提升"
	if winner == fromID { summary = "原版本表述更完整" }

	keyDiffs := diff.AddedTags
	if tp.Evidence.ProjectCount > fp.Evidence.ProjectCount { keyDiffs = append(keyDiffs, fmt.Sprintf("项目经历从%d增加到%d", fp.Evidence.ProjectCount, tp.Evidence.ProjectCount)) }
	if tp.Evidence.InternshipCount > fp.Evidence.InternshipCount { keyDiffs = append(keyDiffs, fmt.Sprintf("实习经历从%d增加到%d", fp.Evidence.InternshipCount, tp.Evidence.InternshipCount)) }

	result := &model.CompareResponseV2{
		FromResumeID: fromID, ToResumeID: toID, ComparisonMode: mode, TargetDirection: targetDir,
		GenerationTier: "fast", GeneratedBy: "fast_tag_engine", Status: "ready",
		CompareScore: score, CompareLevel: "基础对比", WinnerResumeID: uint64(winner), WinnerLabel: "推荐版本",
		OneSentenceSummary: summary, KeyDifferences: keyDiffs,
		FromStrengths: diff.FromStrengths, ToStrengths: diff.ToStrengths,
		FromRisks: diff.FromRisks, ToRisks: diff.ToRisks,
		MergeSuggestions: []string{"保留目标版本的项目描述", "合并原版本的简洁表达"},
		FromProfileType: "fast", ToProfileType: "fast",
		FromProfileDominantProfile: diff.FromDominant, ToProfileDominantProfile: diff.ToDominant,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
	s.saveCompareV2(ctx, userID, result, mode, "fast", targetDir, jobID)
	return result, nil
}

// ─── K13-Fix-C4: compareWith3Agents ───────────────────────────────────

func (s *VersionCompareService) compareWith3Agents(ctx context.Context, userID uint64, fromID, toID uint64, mode, targetDir string, jobID uint64, fromProf, toProf string, runner agent.PipelineRunner) (*model.CompareResponseV2, error) {
	fp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, fromID, userID)
	if err != nil { return nil, fmt.Errorf("3agent compare: from: %w", err) }
	tp, err := s.tagEvidence.EnsureResumeTagProfile(ctx, toID, userID)
	if err != nil { return nil, fmt.Errorf("3agent compare: to: %w", err) }

	final, _, err := agent.RunDeepCompare3Agents(ctx, fp, tp, fromID, toID, runner, runner)
	if err != nil { return nil, fmt.Errorf("3agent compare: %w", err) }

	result := &model.CompareResponseV2{
		FromResumeID: fromID, ToResumeID: toID, ComparisonMode: mode, TargetDirection: targetDir,
		GenerationTier: "quality", GeneratedBy: final.GeneratedBy, Status: "ready",
		CompareScore: final.Score, CompareLevel: final.CompareLevel,
		WinnerResumeID: final.WinnerResumeID, WinnerLabel: "推荐版本",
		OneSentenceSummary: final.OneSentenceSummary, KeyDifferences: final.KeyDifferences,
		FromStrengths: final.FromStrengths, ToStrengths: final.ToStrengths,
		FromRisks: final.FromRisks, ToRisks: final.ToRisks,
		MergeSuggestions: final.MergeSuggestions,
		FromProfileDominantProfile: fp.DominantProfile, ToProfileDominantProfile: tp.DominantProfile,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
	s.saveCompareV2(ctx, userID, result, mode, "quality", targetDir, jobID)
	return result, nil
}

type compareDiffResult struct {
	FromDominant  string
	ToDominant    string
	AddedTags     []string
	FromStrengths []string
	ToStrengths   []string
	FromRisks     []string
	ToRisks       []string
}

func buildCompareDiffSummary(fp, tp *model.ResumeTagProfile) compareDiffResult {
	d := compareDiffResult{FromDominant: fp.DominantProfile, ToDominant: tp.DominantProfile}
	fs := allOptTagSet(fp); ts := allOptTagSet(tp)
	for t := range ts { if !fs[t] { d.AddedTags = append(d.AddedTags, t) } }
	if fp.Evidence.ProjectCount > 0 { d.FromStrengths = append(d.FromStrengths, fmt.Sprintf("%d个项目经历", fp.Evidence.ProjectCount)) }
	if tp.Evidence.ProjectCount > 0 { d.ToStrengths = append(d.ToStrengths, fmt.Sprintf("%d个项目经历", tp.Evidence.ProjectCount)) }
	if fp.Evidence.ProjectCount == 0 { d.FromRisks = append(d.FromRisks, "缺少项目经历") }
	if tp.Evidence.InternshipCount == 0 { d.ToRisks = append(d.ToRisks, "缺少实习经历") }
	return d
}

func allOptTagSet(rp *model.ResumeTagProfile) map[string]bool {
	s := make(map[string]bool)
	for _, t := range rp.PrimaryTags { s[t.Tag] = true }
	for _, t := range rp.SkillTags { s[t.Tag] = true }
	return s
}
