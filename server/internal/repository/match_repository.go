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

// Create inserts a new match analysis with full metadata.
func (r *MatchRepository) Create(ctx context.Context, analysis *model.MatchAnalysis) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO match_analyses
		 (user_id, resume_id, job_id, generation_tier, generated_by, status, profile_type, profile_dominant_profile, score, result_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		analysis.UserID, analysis.ResumeID, analysis.JobID,
		analysis.GenerationTier, analysis.GeneratedBy, analysis.Status,
		analysis.ProfileType, analysis.ProfileDominantProfile,
		analysis.Score, analysis.ResultJSON,
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
func (r *MatchRepository) GetLatestByUserIDAndJobID(ctx context.Context, userID uint64, jobID uint64) (*model.MatchAnalysis, error) {
	var analysis model.MatchAnalysis
	var resultJSON sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, resume_id, job_id, score, result_json, created_at FROM match_analyses WHERE user_id = ? AND job_id = ? ORDER BY created_at DESC LIMIT 1",
		userID, jobID,
	).Scan(&analysis.ID, &analysis.UserID, &analysis.ResumeID, &analysis.JobID, &analysis.Score, &resultJSON, &analysis.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest match analysis: %w", err)
	}

	analysis.ResultJSON = fromNullJSON(resultJSON)

	return &analysis, nil
}

// FindLatestByResumeAndJob returns the latest ready match analysis for a specific resume+job pair.
func (r *MatchRepository) FindLatestByResumeAndJob(ctx context.Context, userID, resumeID, jobID uint64) (*model.MatchAnalysis, error) {
	var a model.MatchAnalysis
	var resultJSON sql.NullString
	var profileType, profileDominant sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, resume_id, job_id, generation_tier, generated_by, status,
		        profile_type, profile_dominant_profile, score, result_json, created_at, updated_at
		 FROM match_analyses
		 WHERE user_id = ? AND resume_id = ? AND job_id = ? AND status = 'ready'
		 ORDER BY created_at DESC, id DESC
		 LIMIT 1`,
		userID, resumeID, jobID,
	).Scan(&a.ID, &a.UserID, &a.ResumeID, &a.JobID,
		&a.GenerationTier, &a.GeneratedBy, &a.Status,
		&profileType, &profileDominant, &a.Score,
		&resultJSON, &a.CreatedAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find latest match analysis: %w", err)
	}

	a.ResultJSON = fromNullJSON(resultJSON)
	if profileType.Valid {
		a.ProfileType = profileType.String
	}
	if profileDominant.Valid {
		a.ProfileDominantProfile = profileDominant.String
	}
	if updatedAt.Valid {
		a.UpdatedAt = updatedAt.Time
	}
	return &a, nil
}

