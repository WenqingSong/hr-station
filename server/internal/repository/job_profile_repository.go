// ─── K13-Fix-A: JobProfile Repository ──────────────────────────────────────

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"offer-catcher/server/internal/model"
)

// JobProfileRepository handles job_profiles table operations.
type JobProfileRepository struct {
	db *sql.DB
}

// NewJobProfileRepository creates a new JobProfileRepository.
func NewJobProfileRepository(db *sql.DB) *JobProfileRepository {
	return &JobProfileRepository{db: db}
}

// FindByJobID returns the profile for a given job.
func (r *JobProfileRepository) FindByJobID(ctx context.Context, jobID uint64) (*model.JobProfile, error) {
	const q = `SELECT id, job_id, direction, category,
		primary_tags_json, skill_tags_json, requirement_tags_json,
		responsibility_tags_json, risk_tags_json, evidence_json,
		source, version, created_at, updated_at
		FROM job_profiles WHERE job_id=?`

	row := r.db.QueryRowContext(ctx, q, jobID)
	var p model.JobProfile
	var ptags, stags, rtags, restags, risktags sql.NullString
	err := row.Scan(&p.ID, &p.JobID, &p.Direction, &p.Category,
		&ptags, &stags, &rtags, &restags, &risktags, &p.EvidenceJSON,
		&p.Source, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find job profile: %w", err)
	}
	p.PrimaryTagsJSON = ptags.String
	p.SkillTagsJSON = stags.String
	p.RequirementTagsJSON = rtags.String
	p.ResponsibilityTagsJSON = restags.String
	p.RiskTagsJSON = risktags.String
	return &p, nil
}

// Upsert inserts or updates a job profile row.
func (r *JobProfileRepository) Upsert(ctx context.Context, p *model.JobProfile) error {
	const q = `INSERT INTO job_profiles
		(job_id, direction, category, primary_tags_json, skill_tags_json,
		requirement_tags_json, responsibility_tags_json, risk_tags_json,
		evidence_json, source, version)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
		direction=VALUES(direction), category=VALUES(category),
		primary_tags_json=VALUES(primary_tags_json),
		skill_tags_json=VALUES(skill_tags_json),
		requirement_tags_json=VALUES(requirement_tags_json),
		responsibility_tags_json=VALUES(responsibility_tags_json),
		risk_tags_json=VALUES(risk_tags_json),
		evidence_json=VALUES(evidence_json),
		source=VALUES(source), version=VALUES(version),
		updated_at=NOW()`

	_, err := r.db.ExecContext(ctx, q,
		p.JobID, p.Direction, p.Category,
		toNStrNull(p.PrimaryTagsJSON), toNStrNull(p.SkillTagsJSON),
		toNStrNull(p.RequirementTagsJSON), toNStrNull(p.ResponsibilityTagsJSON),
		toNStrNull(p.RiskTagsJSON), toNStrNull(p.EvidenceJSON),
		p.Source, p.Version)
	if err != nil {
		return fmt.Errorf("upsert job profile: %w", err)
	}
	return nil
}

// FindAll returns all job profiles (for recall).
func (r *JobProfileRepository) FindAll(ctx context.Context) ([]*model.JobProfile, error) {
	const q = `SELECT id, job_id, direction, category,
		COALESCE(primary_tags_json,'[]'), COALESCE(skill_tags_json,'[]'),
		COALESCE(requirement_tags_json,'[]'), COALESCE(responsibility_tags_json,'[]'),
		COALESCE(risk_tags_json,'[]'), COALESCE(evidence_json,'{}'),
		COALESCE(source,''), COALESCE(version,''), created_at, updated_at
		FROM job_profiles ORDER BY job_id`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("find all job profiles: %w", err)
	}
	defer rows.Close()

	var results []*model.JobProfile
	for rows.Next() {
		var p model.JobProfile
		if err := rows.Scan(&p.ID, &p.JobID, &p.Direction, &p.Category,
			&p.PrimaryTagsJSON, &p.SkillTagsJSON, &p.RequirementTagsJSON,
			&p.ResponsibilityTagsJSON, &p.RiskTagsJSON, &p.EvidenceJSON,
			&p.Source, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, nil
}

// Count returns the total number of job profiles.
func (r *JobProfileRepository) Count(ctx context.Context) (int, error) {
	const q = `SELECT COUNT(*) FROM job_profiles`
	var count int
	err := r.db.QueryRowContext(ctx, q).Scan(&count)
	return count, err
}

// ListPendingJobIDs returns job IDs that exist in jobs table but not in job_profiles.
func (r *JobProfileRepository) ListPendingJobIDs(ctx context.Context, limit int) ([]uint64, error) {
	const q = `SELECT j.id FROM jobs j
		WHERE NOT EXISTS (SELECT 1 FROM job_profiles jp WHERE jp.job_id = j.id)
		ORDER BY j.id LIMIT ?`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending job IDs: %w", err)
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ── helpers ──────────────────────────────────────────────────────────────

func toNStrNull(s string) sql.NullString {
	if s == "" || s == "null" || s == "[]" || s == "{}" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

var _ = time.Now // unused in file but used elsewhere in package
