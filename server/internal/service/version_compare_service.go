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
	FromVersionNo int
	ToVersionNo   int
}

// VersionCompareResponse represents the response for version comparison.
type VersionCompareResponse struct {
	CompareID   uint64                        `json:"compare_id"`
	GeneratedBy string                        `json:"generated_by"`
	Result      *model.VersionCompareAnalysis `json:"result"`
	CreatedAt   string                        `json:"created_at"`
}

// VersionCompareService handles version comparison business logic.
type VersionCompareService struct {
	resumeRepo     *repository.ResumeRepository
	compareRepo    *repository.VersionCompareRepository
	mockAgent      *agent.VersionCompareMockAgent
	pipelineRunner agent.PipelineRunner
	llmEnabled     bool
}

// NewVersionCompareService creates a new VersionCompareService.
func NewVersionCompareService(
	resumeRepo *repository.ResumeRepository,
	compareRepo *repository.VersionCompareRepository,
	mockAgent *agent.VersionCompareMockAgent,
	pipelineRunner agent.PipelineRunner,
	llmEnabled bool,
) *VersionCompareService {
	return &VersionCompareService{
		resumeRepo:     resumeRepo,
		compareRepo:    compareRepo,
		mockAgent:      mockAgent,
		pipelineRunner: pipelineRunner,
		llmEnabled:     llmEnabled,
	}
}

// Compare generates a version comparison for the given user's two resume versions.
func (s *VersionCompareService) Compare(ctx context.Context, userID uint64, req VersionCompareRequest) (*VersionCompareResponse, error) {
	// Validate from/to
	if req.FromVersionNo <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid from version number", 400)
	}
	if req.ToVersionNo <= 0 {
		return nil, apperror.New(apperror.CodeInvalidParam, "invalid to version number", 400)
	}
	if req.FromVersionNo == req.ToVersionNo {
		return nil, apperror.New(apperror.CodeInvalidParam, "from and to version numbers must be different", 400)
	}

	// Get from version
	fromVersion, err := s.resumeRepo.GetByUserIDAndVersionNo(ctx, userID, req.FromVersionNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}
	if fromVersion == nil {
		return nil, apperror.ErrResumeFromVerNotFound
	}

	// Get to version
	toVersion, err := s.resumeRepo.GetByUserIDAndVersionNo(ctx, userID, req.ToVersionNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}
	if toVersion == nil {
		return nil, apperror.ErrResumeToVerNotFound
	}

	// Try LLM Pipeline first if enabled
	if s.llmEnabled && s.pipelineRunner != nil {
		llmAgent := agent.NewVersionCompareLLMAgent(s.pipelineRunner)
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
	}

	// Fall back to Mock Agent
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
		FromVersionID: fromVer.ID,
		ToVersionID:   toVer.ID,
		ResultJSON:    resultJSON,
	}

	if err := s.compareRepo.Create(ctx, compareResult); err != nil {
		// Log but don't fail the response if snapshot save fails
		_ = err
	}

	resp.CompareID = compareResult.ID

	return resp, nil
}
