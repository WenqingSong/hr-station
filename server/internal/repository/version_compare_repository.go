package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// ResumeCompareRepository handles database operations for resume comparison results.
type ResumeCompareRepository struct {
	db *sql.DB
}

// NewResumeCompareRepository creates a new ResumeCompareRepository.
func NewResumeCompareRepository(db *sql.DB) *ResumeCompareRepository {
	return &ResumeCompareRepository{db: db}
}

// Create inserts a new resume compare result and updates the ID.
func (r *ResumeCompareRepository) Create(ctx context.Context, result *model.ResumeCompareResult) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO resume_compare_results (user_id, from_resume_id, to_resume_id, result_json) VALUES (?, ?, ?, ?)",
		result.UserID, result.FromResumeID, result.ToResumeID, result.ResultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create resume compare result: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	result.ID = uint64(id)
	return nil
}

// FindLatestV2 returns the latest ready compare result matching the given criteria.
func (r *ResumeCompareRepository) FindLatestV2(ctx context.Context, userID, fromResumeID, toResumeID uint64, mode, targetDirection string, jobID uint64) (*model.ResumeCompareResultV2, error) {
	var result model.ResumeCompareResultV2
	var resultJSON sql.NullString
	var tDir, fpt, fdp, tpt, tdp sql.NullString
	var jid sql.NullInt64
	var updatedAt sql.NullTime

	var query string
	var args []interface{}

	switch mode {
	case "target_direction":
		query = `SELECT id, user_id, from_resume_id, to_resume_id, comparison_mode, generation_tier, generated_by, status,
			target_direction, job_id, from_profile_type, from_profile_dominant_profile,
			to_profile_type, to_profile_dominant_profile, result_json, created_at, updated_at
		FROM resume_compare_results
		WHERE user_id = ? AND from_resume_id = ? AND to_resume_id = ?
		  AND comparison_mode = ? AND target_direction = ? AND status = 'ready'
		ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, fromResumeID, toResumeID, mode, targetDirection}
	case "target_job":
		query = `SELECT id, user_id, from_resume_id, to_resume_id, comparison_mode, generation_tier, generated_by, status,
			target_direction, job_id, from_profile_type, from_profile_dominant_profile,
			to_profile_type, to_profile_dominant_profile, result_json, created_at, updated_at
		FROM resume_compare_results
		WHERE user_id = ? AND from_resume_id = ? AND to_resume_id = ?
		  AND comparison_mode = ? AND job_id = ? AND status = 'ready'
		ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, fromResumeID, toResumeID, mode, jobID}
	default: // general
		query = `SELECT id, user_id, from_resume_id, to_resume_id, comparison_mode, generation_tier, generated_by, status,
			target_direction, job_id, from_profile_type, from_profile_dominant_profile,
			to_profile_type, to_profile_dominant_profile, result_json, created_at, updated_at
		FROM resume_compare_results
		WHERE user_id = ? AND from_resume_id = ? AND to_resume_id = ?
		  AND comparison_mode = ? AND status = 'ready'
		ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []interface{}{userID, fromResumeID, toResumeID, mode}
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&result.ID, &result.UserID, &result.FromResumeID, &result.ToResumeID,
		&result.ComparisonMode, &result.GenerationTier, &result.GeneratedBy, &result.Status,
		&tDir, &jid, &fpt, &fdp, &tpt, &tdp,
		&resultJSON, &result.CreatedAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find latest compare result: %w", err)
	}

	result.ResultJSON = fromNullJSON(resultJSON)
	if tDir.Valid { result.TargetDirection = tDir.String }
	if jid.Valid { v := uint64(jid.Int64); result.JobID = &v }
	if fpt.Valid { result.FromProfileType = fpt.String }
	if fdp.Valid { result.FromProfileDominantProfile = fdp.String }
	if tpt.Valid { result.ToProfileType = tpt.String }
	if tdp.Valid { result.ToProfileDominantProfile = tdp.String }
	if updatedAt.Valid { result.UpdatedAt = updatedAt.Time }

	return &result, nil
}

// GetLatestByUserID returns the latest resume compare result for a user.
func (r *ResumeCompareRepository) GetLatestByUserID(ctx context.Context, userID uint64) (*model.ResumeCompareResult, error) {
	var result model.ResumeCompareResult
	var resultJSON sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, from_resume_id, to_resume_id, result_json, created_at FROM resume_compare_results WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&result.ID, &result.UserID, &result.FromResumeID, &result.ToResumeID, &resultJSON, &result.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest resume compare result: %w", err)
	}

	result.ResultJSON = fromNullJSON(resultJSON)

	return &result, nil
}

// CreateV2 inserts a new resume compare result with full metadata.
func (r *ResumeCompareRepository) CreateV2(ctx context.Context, result *model.ResumeCompareResultV2) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO resume_compare_results
		 (user_id, from_resume_id, to_resume_id, comparison_mode, generation_tier, generated_by, status,
		  target_direction, job_id, from_profile_type, from_profile_dominant_profile,
		  to_profile_type, to_profile_dominant_profile, result_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.UserID, result.FromResumeID, result.ToResumeID,
		result.ComparisonMode, result.GenerationTier, result.GeneratedBy, result.Status,
		toNullStr(result.TargetDirection), toNullUI64(result.JobID),
		toNullStr(result.FromProfileType), toNullStr(result.FromProfileDominantProfile),
		toNullStr(result.ToProfileType), toNullStr(result.ToProfileDominantProfile),
		toNullJSONFromBytes(result.ResultJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create resume compare result: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	result.ID = uint64(id)
	return nil
}

func toNullStr(s string) sql.NullString {
	if s == "" { return sql.NullString{Valid: false} }
	return sql.NullString{String: s, Valid: true}
}

func toNullUI64(v *uint64) sql.NullInt64 {
	if v == nil { return sql.NullInt64{Valid: false} }
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}
