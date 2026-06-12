package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"time"

	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/repository"
)

// SemanticProfileService provides profile lifecycle management for SEP.
// At DataLayer-C stage this is a skeleton — no Hy3 calls, no recommendation wiring.
type SemanticProfileService struct {
	semanticRepo *repository.SemanticProfileRepository
	resumeRepo   *repository.ResumeRepository
	jobRepo      *repository.JobRepository
	db           *sql.DB
}

// NewSemanticProfileService creates a new SemanticProfileService.
func NewSemanticProfileService(db *sql.DB, semanticRepo *repository.SemanticProfileRepository, resumeRepo *repository.ResumeRepository, jobRepo *repository.JobRepository) *SemanticProfileService {
	return &SemanticProfileService{
		semanticRepo: semanticRepo,
		resumeRepo:   resumeRepo,
		jobRepo:      jobRepo,
		db:           db,
	}
}

// ─── Profile Read ────────────────────────────────────────────────────────

// GetLatestProfile returns the latest profile and its tags for a target.
func (s *SemanticProfileService) GetLatestProfile(ctx context.Context, targetType string, targetID int64) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	profile, err := s.semanticRepo.FindLatestByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, nil, fmt.Errorf("get latest profile: %w", err)
	}
	if profile == nil {
		return nil, nil, nil
	}
	tags, err := s.semanticRepo.ListTagsByProfileID(ctx, profile.ID)
	if err != nil {
		return profile, nil, fmt.Errorf("get latest profile tags: %w", err)
	}
	return profile, tags, nil
}

// GetVerifiedProfile returns a profile matching the content hash with verified status.
func (s *SemanticProfileService) GetVerifiedProfile(ctx context.Context, targetType string, targetID int64, contentHash string) (*model.SemanticProfile, []*model.SemanticProfileTag, error) {
	profile, err := s.semanticRepo.FindVerifiedByTargetAndHash(ctx, targetType, targetID, contentHash)
	if err != nil {
		return nil, nil, fmt.Errorf("get verified profile: %w", err)
	}
	if profile == nil {
		return nil, nil, nil
	}
	tags, err := s.semanticRepo.ListTagsByProfileID(ctx, profile.ID)
	if err != nil {
		return profile, nil, fmt.Errorf("get verified profile tags: %w", err)
	}
	return profile, tags, nil
}

// ─── Profile Write ───────────────────────────────────────────────────────

// CreatePendingProfile creates a pending profile entry, marking any stale existing ones.
func (s *SemanticProfileService) CreatePendingProfile(ctx context.Context, targetType string, targetID int64, contentHash, modelName, reasoningEffort string) (*model.SemanticProfile, error) {
	// Mark old verified/pending profiles as stale
	if err := s.semanticRepo.MarkStaleByTarget(ctx, targetType, targetID); err != nil {
		return nil, fmt.Errorf("create pending: mark stale: %w", err)
	}

	now := time.Now()
	p := &model.SemanticProfile{
		TargetType:      targetType,
		TargetID:        targetID,
		ProfileVersion:  model.SemanticProfileVersionV1,
		PromptVersion:   model.SemanticPromptVersionV1,
		SchemaVersion:   model.SemanticSchemaVersionV1,
		ContentHash:     contentHash,
		ModelProvider:   "hy3",
		ModelName:       modelName,
		ReasoningEffort: reasoningEffort,
		Status:          model.SemanticProfileStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.semanticRepo.CreateProfile(ctx, p); err != nil {
		return nil, fmt.Errorf("create pending profile: %w", err)
	}
	return p, nil
}

// MarkStale marks all non-current verified/pending profiles for this target.
func (s *SemanticProfileService) MarkStale(ctx context.Context, targetType string, targetID int64) error {
	return s.semanticRepo.MarkStaleByTarget(ctx, targetType, targetID)
}

// SaveVerifiedProfile saves a verified profile and its tags with consistency guarantee.
// If tag write fails, the profile status is rolled back to failed.
func (s *SemanticProfileService) SaveVerifiedProfile(ctx context.Context, profile *model.SemanticProfile, tags []*model.SemanticProfileTag) error {
	profile.Status = model.SemanticProfileStatusVerified
	profile.UpdatedAt = time.Now()
	if profile.ID == 0 {
		if err := s.semanticRepo.CreateProfile(ctx, profile); err != nil {
			return fmt.Errorf("save verified profile: create: %w", err)
		}
	}
	if err := s.semanticRepo.UpsertTags(ctx, profile.ID, profile.TargetType, profile.TargetID, tags); err != nil {
		// Roll back: mark profile as failed since tags didn't write
		_ = s.semanticRepo.UpdateProfileStatus(ctx, profile.ID, model.SemanticProfileStatusFailed,
			fmt.Sprintf("tag upsert failed: %v", err))
		profile.Status = model.SemanticProfileStatusFailed
		return fmt.Errorf("save verified profile: upsert tags: %w", err)
	}
	// Verify tag count after write
	count, err := s.semanticRepo.CountTagsByProfileID(ctx, profile.ID)
	if err != nil {
		log.Printf("[SEP] SaveVerifiedProfile: count tags check failed for profile %d: %v", profile.ID, err)
	} else if count < 5 {
		_ = s.semanticRepo.UpdateProfileStatus(ctx, profile.ID, model.SemanticProfileStatusFailed,
			fmt.Sprintf("insufficient tags: wrote %d, need >=5", count))
		profile.Status = model.SemanticProfileStatusFailed
		return fmt.Errorf("save verified profile: only %d tags written (need >=5)", count)
	}
	return nil
}

// UpdateProfileStatus is a passthrough to mark profile status (used by gate integrity checks).
func (s *SemanticProfileService) UpdateProfileStatus(ctx context.Context, profileID int64, status, errorMessage string) error {
	return s.semanticRepo.UpdateProfileStatus(ctx, profileID, status, errorMessage)
}

// ─── Content Hash ────────────────────────────────────────────────────────

// ComputeResumeContentHash computes sha256 hash for resume content.
func (s *SemanticProfileService) ComputeResumeContentHash(ctx context.Context, resumeID int64) (string, error) {
	// Fetch resume fields for hash computation
	var rawText, name string
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(raw_text,''), COALESCE(name,''), COALESCE(updated_at,NOW()) FROM resumes WHERE id=?`,
		resumeID).Scan(&rawText, &name, &updatedAt)
	if err != nil {
		return "", fmt.Errorf("compute resume hash: %w", err)
	}
	input := name + "|" + updatedAt.Format(time.RFC3339) + "|" + rawText
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input))), nil
}

// ComputeJobContentHash computes sha256 hash for job content.
func (s *SemanticProfileService) ComputeJobContentHash(ctx context.Context, jobID int64) (string, error) {
	var title, description, requirements, direction, category string
	var updatedAt time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(title,''), COALESCE(description,''), COALESCE(requirements,''),
		        COALESCE(direction,''), COALESCE(category,''), COALESCE(updated_at,NOW())
		 FROM jobs WHERE id=?`,
		jobID).Scan(&title, &description, &requirements, &direction, &category, &updatedAt)
	if err != nil {
		return "", fmt.Errorf("compute job hash: %w", err)
	}
	input := fmt.Sprintf("%s|%s|%s|%s|%s|%s", title, description, requirements, direction, category, updatedAt.Format(time.RFC3339))
	return fmt.Sprintf("%x", sha256.Sum256([]byte(input))), nil
}
