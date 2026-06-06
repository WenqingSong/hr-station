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
