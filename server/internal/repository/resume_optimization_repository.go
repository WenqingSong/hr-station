package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// ResumeOptimizationRepository handles database operations for resume optimization results.
type ResumeOptimizationRepository struct {
	db *sql.DB
}

// NewResumeOptimizationRepository creates a new ResumeOptimizationRepository.
func NewResumeOptimizationRepository(db *sql.DB) *ResumeOptimizationRepository {
	return &ResumeOptimizationRepository{db: db}
}

// Create inserts a new resume optimization result and updates the ID.
func (r *ResumeOptimizationRepository) Create(ctx context.Context, result *model.ResumeOptimizationSnapshot) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO resume_optimization_results (user_id, job_id, result_json) VALUES (?, ?, ?)",
		result.UserID, result.JobID, result.ResultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create resume optimization result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	result.ID = uint64(id)
	return nil
}

// CreateV2 inserts with full K9-3-C metadata.
func (r *ResumeOptimizationRepository) CreateV2(ctx context.Context, result *model.ResumeOptimizationSnapshot) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO resume_optimization_results
		 (user_id, resume_id, job_id, optimization_mode, target_direction,
		  generation_tier, generated_by, status, profile_type, profile_dominant_profile, result_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.UserID, result.ResumeID, result.JobID,
		result.OptimizationMode, result.TargetDirection,
		result.GenerationTier, result.GeneratedBy, result.Status,
		result.ProfileType, result.ProfileDominantProfile,
		result.ResultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create v2 optimization result: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	result.ID = uint64(id)
	return nil
}

// FindLatestByMode returns the latest ready optimization for a given user+resume+mode.
func (r *ResumeOptimizationRepository) FindLatestByMode(ctx context.Context, userID, resumeID uint64, mode string, targetDirection string, jobID uint64) (*model.ResumeOptimizationSnapshot, error) {
	var s model.ResumeOptimizationSnapshot
	var resultJSON sql.NullString
	var profileType, profileDominant, targetDir sql.NullString
	var genTier, genBy, status sql.NullString
	var jobIDVal sql.NullInt64
	var resumeIDVal sql.NullInt64
	var updatedAt sql.NullTime

	var query string
	var args []interface{}

	switch mode {
	case "general":
		query = `SELECT id, user_id, resume_id, COALESCE(job_id,0), optimization_mode, target_direction,
		         generation_tier, generated_by, status, profile_type, profile_dominant_profile, result_json, created_at, updated_at
		         FROM resume_optimization_results
		         WHERE user_id=? AND resume_id=? AND optimization_mode='general' AND status='ready'
		         ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, resumeID}
	case "target_direction":
		query = `SELECT id, user_id, resume_id, COALESCE(job_id,0), optimization_mode, target_direction,
		         generation_tier, generated_by, status, profile_type, profile_dominant_profile, result_json, created_at, updated_at
		         FROM resume_optimization_results
		         WHERE user_id=? AND resume_id=? AND optimization_mode='target_direction' AND target_direction=? AND status='ready'
		         ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, resumeID, targetDirection}
	case "target_job":
		query = `SELECT id, user_id, resume_id, COALESCE(job_id,0), optimization_mode, target_direction,
		         generation_tier, generated_by, status, profile_type, profile_dominant_profile, result_json, created_at, updated_at
		         FROM resume_optimization_results
		         WHERE user_id=? AND resume_id=? AND optimization_mode='target_job' AND job_id=? AND status='ready'
		         ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, resumeID, jobID}
	default:
		return nil, nil
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&s.ID, &s.UserID, &resumeIDVal, &jobIDVal,
		&s.OptimizationMode, &targetDir,
		&genTier, &genBy, &status,
		&profileType, &profileDominant,
		&resultJSON, &s.CreatedAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest optimization: %w", err)
	}

	s.ResumeID = uint64(resumeIDVal.Int64)
	if jobIDVal.Valid { s.JobID = uint64(jobIDVal.Int64) }
	if genTier.Valid { s.GenerationTier = genTier.String }
	if genBy.Valid { s.GeneratedBy = genBy.String }
	if status.Valid { s.Status = status.String }
	if profileType.Valid { s.ProfileType = profileType.String }
	if profileDominant.Valid { s.ProfileDominantProfile = profileDominant.String }
	if targetDir.Valid { s.TargetDirection = targetDir.String }
	s.ResultJSON = fromNullJSONOpt(resultJSON)
	if updatedAt.Valid { s.UpdatedAt = updatedAt.Time }
	return &s, nil
}

func fromNullJSONOpt(ns sql.NullString) []byte {
	if ns.Valid { return []byte(ns.String) }
	return nil
}

// GetLatestByUserID returns the most recent optimization result for a user.
func (r *ResumeOptimizationRepository) GetLatestByUserID(ctx context.Context, userID uint64) (*model.ResumeOptimizationSnapshot, error) {
	var s model.ResumeOptimizationSnapshot
	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, job_id, result_json, created_at FROM resume_optimization_results WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&s.ID, &s.UserID, &s.JobID, &s.ResultJSON, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest optimization result: %w", err)
	}
	return &s, nil
}
