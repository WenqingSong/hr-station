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
