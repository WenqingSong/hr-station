package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// VersionCompareRepository handles database operations for version comparison results.
type VersionCompareRepository struct {
	db *sql.DB
}

// NewVersionCompareRepository creates a new VersionCompareRepository.
func NewVersionCompareRepository(db *sql.DB) *VersionCompareRepository {
	return &VersionCompareRepository{db: db}
}

// Create inserts a new version compare result and updates the ID.
func (r *VersionCompareRepository) Create(ctx context.Context, result *model.VersionCompareResult) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO version_compare_results (user_id, from_version_id, to_version_id, result_json) VALUES (?, ?, ?, ?)",
		result.UserID, result.FromVersionID, result.ToVersionID, result.ResultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create version compare result: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	result.ID = uint64(id)
	return nil
}

// GetLatestByUserID returns the latest version compare result for a user.
// Returns nil if not found.
func (r *VersionCompareRepository) GetLatestByUserID(ctx context.Context, userID uint64) (*model.VersionCompareResult, error) {
	var result model.VersionCompareResult
	var resultJSON sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, user_id, from_version_id, to_version_id, result_json, created_at FROM version_compare_results WHERE user_id = ? ORDER BY created_at DESC LIMIT 1",
		userID,
	).Scan(&result.ID, &result.UserID, &result.FromVersionID, &result.ToVersionID, &resultJSON, &result.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version compare result: %w", err)
	}

	result.ResultJSON = fromNullJSON(resultJSON)

	return &result, nil
}
