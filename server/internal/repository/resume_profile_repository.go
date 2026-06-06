package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"offer-catcher/server/internal/model"
)

// ResumeProfileRepository handles resume_profiles table operations.
type ResumeProfileRepository struct {
	db *sql.DB
}

// NewResumeProfileRepository creates a new ResumeProfileRepository.
func NewResumeProfileRepository(db *sql.DB) *ResumeProfileRepository {
	return &ResumeProfileRepository{db: db}
}

// FindByResumeIDAndType returns the profile for a given resume and type.
func (r *ResumeProfileRepository) FindByResumeIDAndType(ctx context.Context, resumeID uint64, profileType string) (*model.ResumeProfile, error) {
	const q = `SELECT id,resume_id,user_id,profile_type,profile_version,status,
	dominant_profile,primary_tags_json,candidate_tags_json,risk_tags_json,evidence_json,
	profile_summary,source,generator,error_message,resume_updated_at,generated_at,created_at,updated_at
	FROM resume_profiles WHERE resume_id=? AND profile_type=?`
	row := r.db.QueryRowContext(ctx, q, resumeID, profileType)
	var p model.ResumeProfile
	var resumeUpdatedAt, generatedAt sql.NullTime
	var errMsg sql.NullString
	var profileSummary sql.NullString
	err := row.Scan(&p.ID, &p.ResumeID, &p.UserID, &p.ProfileType, &p.ProfileVersion, &p.Status,
		&p.DominantProfile, &p.PrimaryTagsJSON, &p.CandidateTagsJSON, &p.RiskTagsJSON, &p.EvidenceJSON,
		&profileSummary, &p.Source, &p.Generator, &errMsg, &resumeUpdatedAt, &generatedAt, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find profile: %w", err)
	}
	if resumeUpdatedAt.Valid {
		p.ResumeUpdatedAt = &resumeUpdatedAt.Time
	}
	if generatedAt.Valid {
		p.GeneratedAt = &generatedAt.Time
	}
	if profileSummary.Valid {
		p.ProfileSummary = profileSummary.String
	}
	if errMsg.Valid {
		p.ErrorMessage = errMsg.String
	}
	return &p, nil
}

// UpsertProfile inserts or updates a profile row.
func (r *ResumeProfileRepository) UpsertProfile(ctx context.Context, p *model.ResumeProfile) error {
	const q = `INSERT INTO resume_profiles (resume_id,user_id,profile_type,profile_version,status,
	dominant_profile,primary_tags_json,candidate_tags_json,risk_tags_json,evidence_json,
	profile_summary,source,generator,error_message,resume_updated_at,generated_at)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,NOW())
	ON DUPLICATE KEY UPDATE
	status=VALUES(status), dominant_profile=VALUES(dominant_profile),
	primary_tags_json=VALUES(primary_tags_json), candidate_tags_json=VALUES(candidate_tags_json),
	risk_tags_json=VALUES(risk_tags_json), evidence_json=VALUES(evidence_json),
	profile_summary=VALUES(profile_summary), source=VALUES(source), generator=VALUES(generator),
	error_message=VALUES(error_message), resume_updated_at=VALUES(resume_updated_at),
	generated_at=VALUES(generated_at), user_id=VALUES(user_id)`
	_, err := r.db.ExecContext(ctx, q,
		p.ResumeID, p.UserID, p.ProfileType, p.ProfileVersion, p.Status,
		p.DominantProfile, p.PrimaryTagsJSON, p.CandidateTagsJSON, p.RiskTagsJSON, p.EvidenceJSON,
		toNStr(p.ProfileSummary), p.Source, p.Generator, toNStr(p.ErrorMessage),
		toNTime(p.ResumeUpdatedAt))
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	return nil
}

// MarkStaleByResumeID sets all profiles for a resume to stale.
func (r *ResumeProfileRepository) MarkStaleByResumeID(ctx context.Context, resumeID uint64) error {
	const q = `UPDATE resume_profiles SET status=? WHERE resume_id=? AND status=?`
	_, err := r.db.ExecContext(ctx, q, model.ProfileStatusStale, resumeID, model.ProfileStatusReady)
	if err != nil {
		return fmt.Errorf("mark stale: %w", err)
	}
	return nil
}

// CreatePendingIfMissing creates a pending profile if none exists.
// Returns (profile, created, error).
func (r *ResumeProfileRepository) CreatePendingIfMissing(ctx context.Context, userID, resumeID uint64, profileType string) (*model.ResumeProfile, bool, error) {
	existing, err := r.FindByResumeIDAndType(ctx, resumeID, profileType)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}
	now := time.Now()
	p := &model.ResumeProfile{
		ResumeID:          resumeID,
		UserID:            userID,
		ProfileType:       profileType,
		ProfileVersion:    model.ProfileVersionV1,
		Status:            model.ProfileStatusPending,
		DominantProfile:   "",
		PrimaryTagsJSON:   "[]",
		CandidateTagsJSON: "[]",
		RiskTagsJSON:      "[]",
		EvidenceJSON:      "{}",
		Source:            "",
		Generator:         "",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := r.UpsertProfile(ctx, p); err != nil {
		return nil, false, err
	}
	return p, true, nil
}

// UpdateReady updates a profile to ready status with generated data.
func (r *ResumeProfileRepository) UpdateReady(ctx context.Context, p *model.ResumeProfile) error {
	p.Status = model.ProfileStatusReady
	return r.UpsertProfile(ctx, p)
}

// UpdateFailed updates a profile to failed status with error message.
func (r *ResumeProfileRepository) UpdateFailed(ctx context.Context, resumeID uint64, profileType string, errMsg string) error {
	const q = `UPDATE resume_profiles SET status=?, error_message=?, generated_at=NOW() WHERE resume_id=? AND profile_type=?`
	_, err := r.db.ExecContext(ctx, q, model.ProfileStatusFailed, errMsg, resumeID, profileType)
	return err
}

// ListMissingFastProfiles returns resume IDs without a fast profile for a given user.
func (r *ResumeProfileRepository) ListMissingFastProfiles(ctx context.Context, userID uint64) ([]uint64, error) {
	const q = `SELECT r.id FROM resumes r
	WHERE r.user_id=?
	AND NOT EXISTS (SELECT 1 FROM resume_profiles rp WHERE rp.resume_id=r.id AND rp.profile_type=?)
	ORDER BY r.id`
	rows, err := r.db.QueryContext(ctx, q, userID, model.ProfileTypeFast)
	if err != nil {
		return nil, fmt.Errorf("list missing fast: %w", err)
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

// ListByResumeID returns all profiles for a given resume.
func (r *ResumeProfileRepository) ListByResumeID(ctx context.Context, resumeID uint64) ([]*model.ResumeProfile, error) {
	const q = `SELECT id,resume_id,user_id,profile_type,profile_version,status,
	dominant_profile,primary_tags_json,candidate_tags_json,risk_tags_json,evidence_json,
	COALESCE(profile_summary,''),source,generator,COALESCE(error_message,''),resume_updated_at,generated_at,created_at,updated_at
	FROM resume_profiles WHERE resume_id=? ORDER BY profile_type`
	rows, err := r.db.QueryContext(ctx, q, resumeID)
	if err != nil {
		return nil, fmt.Errorf("list by resume: %w", err)
	}
	defer rows.Close()
	var results []*model.ResumeProfile
	for rows.Next() {
		var p model.ResumeProfile
		var resumeUpdatedAt, generatedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.ResumeID, &p.UserID, &p.ProfileType, &p.ProfileVersion, &p.Status,
			&p.DominantProfile, &p.PrimaryTagsJSON, &p.CandidateTagsJSON, &p.RiskTagsJSON, &p.EvidenceJSON,
			&p.ProfileSummary, &p.Source, &p.Generator, &p.ErrorMessage, &resumeUpdatedAt, &generatedAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if resumeUpdatedAt.Valid {
			p.ResumeUpdatedAt = &resumeUpdatedAt.Time
		}
		if generatedAt.Valid {
			p.GeneratedAt = &generatedAt.Time
		}
		results = append(results, &p)
	}
	return results, nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func toNStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func toNTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
