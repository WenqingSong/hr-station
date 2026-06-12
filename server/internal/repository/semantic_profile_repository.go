package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"offer-catcher/server/internal/model"
)

// SemanticProfileRepository manages SEP semantic profile and tag storage.
type SemanticProfileRepository struct {
	db *sql.DB
}

func NewSemanticProfileRepository(db *sql.DB) *SemanticProfileRepository {
	return &SemanticProfileRepository{db: db}
}

// ─── Profile Queries ─────────────────────────────────────────────────────

// FindLatestByTarget returns the most recent semantic profile for a target.
func (r *SemanticProfileRepository) FindLatestByTarget(ctx context.Context, targetType string, targetID int64) (*model.SemanticProfile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, target_type, target_id, profile_version, prompt_version, schema_version,
		        content_hash, model_provider, model_name, reasoning_effort,
		        COALESCE(summary_sentence,''), COALESCE(strengths_json,''), COALESCE(risks_json,''),
		        COALESCE(fit_directions_json,''), COALESCE(profile_json,''),
		        status, COALESCE(error_message,''), created_at, updated_at
		 FROM semantic_profiles
		 WHERE target_type=? AND target_id=?
		 ORDER BY updated_at DESC, id DESC
		 LIMIT 1`, targetType, targetID)
	return scanProfile(row)
}

// FindVerifiedByTargetAndHash finds a verified profile matching the content hash.
func (r *SemanticProfileRepository) FindVerifiedByTargetAndHash(ctx context.Context, targetType string, targetID int64, contentHash string) (*model.SemanticProfile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, target_type, target_id, profile_version, prompt_version, schema_version,
		        content_hash, model_provider, model_name, reasoning_effort,
		        COALESCE(summary_sentence,''), COALESCE(strengths_json,''), COALESCE(risks_json,''),
		        COALESCE(fit_directions_json,''), COALESCE(profile_json,''),
		        status, COALESCE(error_message,''), created_at, updated_at
		 FROM semantic_profiles
		 WHERE target_type=? AND target_id=? AND profile_version=? AND schema_version=? AND content_hash=? AND status=?
		 ORDER BY updated_at DESC, id DESC
		 LIMIT 1`,
		targetType, targetID, model.SemanticProfileVersionV1, model.SemanticSchemaVersionV1, contentHash, model.SemanticProfileStatusVerified)
	return scanProfile(row)
}

// CreateProfile inserts a new semantic profile, or reuses an existing non-verified record with matching unique key.
func (r *SemanticProfileRepository) CreateProfile(ctx context.Context, p *model.SemanticProfile) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO semantic_profiles
		 (target_type, target_id, profile_version, prompt_version, schema_version,
		  content_hash, model_provider, model_name, reasoning_effort,
		  summary_sentence, strengths_json, risks_json, fit_directions_json, profile_json,
		  status, error_message, created_at, updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p.TargetType, p.TargetID, p.ProfileVersion, p.PromptVersion, p.SchemaVersion,
		p.ContentHash, p.ModelProvider, p.ModelName, p.ReasoningEffort,
		p.SummarySentence, p.StrengthsJSON, p.RisksJSON, p.FitDirectionsJSON, p.ProfileJSON,
		p.Status, p.ErrorMessage, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		// Check for duplicate key — may be an old failed/stale record that we can update
		if isDuplicateKeyError(err) {
			existing, err2 := r.FindVerifiedByTargetAndHash(ctx, p.TargetType, p.TargetID, p.ContentHash)
			if err2 != nil {
				return fmt.Errorf("create semantic profile: duplicate key, lookup existing: %w", err2)
			}
			if existing != nil && existing.Status == model.SemanticProfileStatusVerified {
				// Already verified — skip creation, reuse ID
				p.ID = existing.ID
				return nil
			}
			// Find the non-verified record and update it
			existing, err2 = r.findAnyByTargetAndHash(ctx, p.TargetType, p.TargetID, p.ContentHash)
			if err2 != nil || existing == nil {
				return fmt.Errorf("create semantic profile: duplicate key, existing lookup: %w", err)
			}
			// Update the existing non-verified record
			if err3 := r.updateProfileAll(ctx, existing.ID, p); err3 != nil {
				return fmt.Errorf("create semantic profile: duplicate key, update existing: %w", err3)
			}
			p.ID = existing.ID
			return nil
		}
		return fmt.Errorf("create semantic profile: %w", err)
	}
	id, _ := result.LastInsertId()
	p.ID = id
	return nil
}

// UpdateProfileStatus updates the status and error message of a profile.
func (r *SemanticProfileRepository) UpdateProfileStatus(ctx context.Context, profileID int64, status, errorMessage string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE semantic_profiles SET status=?, error_message=?, updated_at=? WHERE id=?`,
		status, errorMessage, time.Now(), profileID)
	if err != nil {
		return fmt.Errorf("update semantic profile status: %w", err)
	}
	return nil
}

// MarkStaleByTarget marks non-matching-hash profiles as stale for a target.
func (r *SemanticProfileRepository) MarkStaleByTarget(ctx context.Context, targetType string, targetID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE semantic_profiles SET status=?, updated_at=? WHERE target_type=? AND target_id=? AND status IN (?,?)`,
		model.SemanticProfileStatusStale, time.Now(), targetType, targetID,
		model.SemanticProfileStatusVerified, model.SemanticProfileStatusPending)
	if err != nil {
		return fmt.Errorf("mark semantic profile stale: %w", err)
	}
	return nil
}

// ListProfilesByStatus returns profiles with the given status, up to limit.
func (r *SemanticProfileRepository) ListProfilesByStatus(ctx context.Context, status string, limit int) ([]*model.SemanticProfile, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, target_type, target_id, profile_version, prompt_version, schema_version,
		        content_hash, model_provider, model_name, reasoning_effort,
		        COALESCE(summary_sentence,''), COALESCE(strengths_json,''), COALESCE(risks_json,''),
		        COALESCE(fit_directions_json,''), COALESCE(profile_json,''),
		        status, COALESCE(error_message,''), created_at, updated_at
		 FROM semantic_profiles WHERE status=? ORDER BY updated_at ASC LIMIT ?`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list profiles by status: %w", err)
	}
	defer rows.Close()
	var result []*model.SemanticProfile
	for rows.Next() {
		p, err := scanProfileRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// ─── Tag Queries ─────────────────────────────────────────────────────────

// UpsertTags replaces all tags for a profile in a transaction.
func (r *SemanticProfileRepository) UpsertTags(ctx context.Context, profileID int64, targetType string, targetID int64, tags []*model.SemanticProfileTag) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("upsert tags: begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing tags
	if _, err := tx.ExecContext(ctx, `DELETE FROM semantic_profile_tags WHERE profile_id=?`, profileID); err != nil {
		return fmt.Errorf("upsert tags: delete old: %w", err)
	}

	// Insert new tags
	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO semantic_profile_tags
		 (profile_id, target_type, target_id, tag_code, tag_label, category,
		  confidence, evidence_text, explanation, source_section,
		  is_user_visible, weight, created_at, updated_at)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("upsert tags: prepare insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, t := range tags {
		if _, err := stmt.ExecContext(ctx,
			profileID, targetType, targetID, t.TagCode, t.TagLabel, t.Category,
			t.Confidence, t.EvidenceText, t.Explanation, t.SourceSection,
			t.IsUserVisible, t.Weight, now, now,
		); err != nil {
			return fmt.Errorf("upsert tags: insert: %w", err)
		}
	}

	return tx.Commit()
}

// CountTagsByProfileID returns the number of tags for a profile.
func (r *SemanticProfileRepository) CountTagsByProfileID(ctx context.Context, profileID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM semantic_profile_tags WHERE profile_id=?`, profileID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tags by profile id: %w", err)
	}
	return count, nil
}

// ListTagsByProfileID returns all tags for a profile.
func (r *SemanticProfileRepository) ListTagsByProfileID(ctx context.Context, profileID int64) ([]*model.SemanticProfileTag, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, profile_id, target_type, target_id, tag_code, tag_label, category,
		        confidence, COALESCE(evidence_text,''), COALESCE(explanation,''), COALESCE(source_section,''),
		        is_user_visible, weight, created_at, updated_at
		 FROM semantic_profile_tags WHERE profile_id=? ORDER BY weight DESC, category, tag_code`, profileID)
	if err != nil {
		return nil, fmt.Errorf("list tags by profile id: %w", err)
	}
	defer rows.Close()
	var result []*model.SemanticProfileTag
	for rows.Next() {
		var t model.SemanticProfileTag
		if err := rows.Scan(&t.ID, &t.ProfileID, &t.TargetType, &t.TargetID, &t.TagCode, &t.TagLabel, &t.Category,
			&t.Confidence, &t.EvidenceText, &t.Explanation, &t.SourceSection,
			&t.IsUserVisible, &t.Weight, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, &t)
	}
	return result, rows.Err()
}

// ─── Scan Helpers ────────────────────────────────────────────────────────

func scanProfile(row *sql.Row) (*model.SemanticProfile, error) {
	var p model.SemanticProfile
	err := row.Scan(&p.ID, &p.TargetType, &p.TargetID, &p.ProfileVersion, &p.PromptVersion, &p.SchemaVersion,
		&p.ContentHash, &p.ModelProvider, &p.ModelName, &p.ReasoningEffort,
		&p.SummarySentence, &p.StrengthsJSON, &p.RisksJSON, &p.FitDirectionsJSON, &p.ProfileJSON,
		&p.Status, &p.ErrorMessage, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan semantic profile: %w", err)
	}
	return &p, nil
}

func scanProfileRow(rows *sql.Rows) (*model.SemanticProfile, error) {
	var p model.SemanticProfile
	err := rows.Scan(&p.ID, &p.TargetType, &p.TargetID, &p.ProfileVersion, &p.PromptVersion, &p.SchemaVersion,
		&p.ContentHash, &p.ModelProvider, &p.ModelName, &p.ReasoningEffort,
		&p.SummarySentence, &p.StrengthsJSON, &p.RisksJSON, &p.FitDirectionsJSON, &p.ProfileJSON,
		&p.Status, &p.ErrorMessage, &p.CreatedAt, &p.UpdatedAt)
	return &p, err
}

// ─── Internal Helpers ────────────────────────────────────────────────────

// isDuplicateKeyError checks if the error is a MySQL 1062 duplicate key error.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Error 1062") || strings.Contains(err.Error(), "Duplicate entry")
}

// findAnyByTargetAndHash finds any profile matching target and hash, regardless of status.
func (r *SemanticProfileRepository) findAnyByTargetAndHash(ctx context.Context, targetType string, targetID int64, contentHash string) (*model.SemanticProfile, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, target_type, target_id, profile_version, prompt_version, schema_version,
		        content_hash, model_provider, model_name, reasoning_effort,
		        COALESCE(summary_sentence,''), COALESCE(strengths_json,''), COALESCE(risks_json,''),
		        COALESCE(fit_directions_json,''), COALESCE(profile_json,''),
		        status, COALESCE(error_message,''), created_at, updated_at
		 FROM semantic_profiles
		 WHERE target_type=? AND target_id=? AND content_hash=?
		 ORDER BY updated_at DESC, id DESC LIMIT 1`, targetType, targetID, contentHash)
	return scanProfile(row)
}

// updateProfileAll replaces all fields of an existing profile by ID.
func (r *SemanticProfileRepository) updateProfileAll(ctx context.Context, profileID int64, p *model.SemanticProfile) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE semantic_profiles SET
		 profile_version=?, prompt_version=?, schema_version=?,
		 content_hash=?, model_provider=?, model_name=?, reasoning_effort=?,
		 summary_sentence=?, strengths_json=?, risks_json=?, fit_directions_json=?, profile_json=?,
		 status=?, error_message=?, updated_at=?
		 WHERE id=?`,
		p.ProfileVersion, p.PromptVersion, p.SchemaVersion,
		p.ContentHash, p.ModelProvider, p.ModelName, p.ReasoningEffort,
		p.SummarySentence, p.StrengthsJSON, p.RisksJSON, p.FitDirectionsJSON, p.ProfileJSON,
		p.Status, p.ErrorMessage, p.UpdatedAt,
		profileID)
	if err != nil {
		return fmt.Errorf("update profile all: %w", err)
	}
	return nil
}
