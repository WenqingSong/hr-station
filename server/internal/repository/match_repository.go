package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// MatchRepository handles database operations for match analyses.
type MatchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new MatchRepository.
func NewMatchRepository(db *sql.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// Create inserts a new match analysis and updates the ID.
func (r *MatchRepository) Create(ctx context.Context, analysis *model.MatchAnalysis) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO match_analyses (user_id, resume_version_id, job_id, score, result_json) VALUES (?, ?, ?, ?, ?)",
		analysis.UserID, analysis.ResumeVersionID, analysis.JobID, analysis.Score, analysis.ResultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create match analysis: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	analysis.ID = uint64(id)
	return nil
}

// GetLatestByUserIDAndJobID returns the latest match analysis for a user and job.
// Returns nil if not found.
func (r *MatchRepository) GetLatestByUserIDAndJobID(ctx context.Context, userID uint64, jobID uint64) (*model.MatchAnalysis, error) {
	var analysis model.MatchAnalysis
	var resultJSON sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, resume_version_id, job_id, score, result_json, created_at FROM match_analyses WHERE user_id = ? AND job_id = ? ORDER BY created_at DESC LIMIT 1",
		userID, jobID,
	).Scan(&analysis.ID, &analysis.UserID, &analysis.ResumeVersionID, &analysis.JobID, &analysis.Score, &resultJSON, &analysis.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest match analysis: %w", err)
	}

	analysis.ResultJSON = fromNullJSON(resultJSON)

	return &analysis, nil
}
