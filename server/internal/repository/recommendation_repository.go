package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// RecommendationRepository handles database operations for recommendation results.
type RecommendationRepository struct {
	db *sql.DB
}

// NewRecommendationRepository creates a new RecommendationRepository.
func NewRecommendationRepository(db *sql.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// Create inserts a new recommendation result with full metadata.
func (r *RecommendationRepository) Create(ctx context.Context, result *model.RecommendationResult) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO recommendation_results
		 (user_id, resume_id, generation_tier, generated_by, status, profile_type, profile_dominant_profile, result_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		result.UserID, result.ResumeID,
		result.GenerationTier, result.GeneratedBy, result.Status,
		result.ProfileType, result.ProfileDominantProfile,
		toNullJSONFromBytes(result.ResultJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create recommendation result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	result.ID = uint64(id)
	return nil
}

// GetLatestByUserID returns the most recent recommendation result for the given user.
func (r *RecommendationRepository) GetLatestByUserID(ctx context.Context, userID uint64) (*model.RecommendationResult, error) {
	var result model.RecommendationResult
	var resultJSON sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, resume_id, result_json, created_at FROM recommendation_results WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&result.ID, &result.UserID, &result.ResumeID, &resultJSON, &result.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest recommendation: %w", err)
	}

	result.ResultJSON = fromNullJSON(resultJSON)
	return &result, nil
}

// FindLatestByResumeID returns the most recent ready recommendation for the given user+resume.
func (r *RecommendationRepository) FindLatestByResumeID(ctx context.Context, userID, resumeID uint64) (*model.RecommendationResult, error) {
	var result model.RecommendationResult
	var resultJSON sql.NullString
	var profileType, profileDominant sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, resume_id, generation_tier, generated_by, status,
		        profile_type, profile_dominant_profile, result_json, created_at, updated_at
		 FROM recommendation_results
		 WHERE user_id = ? AND resume_id = ? AND status = 'ready'
		 ORDER BY created_at DESC, id DESC
		 LIMIT 1`,
		userID, resumeID,
	).Scan(&result.ID, &result.UserID, &result.ResumeID,
		&result.GenerationTier, &result.GeneratedBy, &result.Status,
		&profileType, &profileDominant, &resultJSON,
		&result.CreatedAt, &updatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find latest recommendation: %w", err)
	}

	result.ResultJSON = fromNullJSON(resultJSON)
	if profileType.Valid {
		result.ProfileType = profileType.String
	}
	if profileDominant.Valid {
		result.ProfileDominantProfile = profileDominant.String
	}
	if updatedAt.Valid {
		result.UpdatedAt = updatedAt.Time
	}
	return &result, nil
}

func toNullJSONFromBytes(data []byte) sql.NullString {
	if len(data) == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(data), Valid: true}
}

func fromNullJSON(ns sql.NullString) []byte {
if !ns.Valid || ns.String == "" { return nil }
return []byte(ns.String)
}
