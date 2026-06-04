package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// ResumeRepository handles database operations for resume versions.
type ResumeRepository struct {
	db *sql.DB
}

// NewResumeRepository creates a new ResumeRepository.
func NewResumeRepository(db *sql.DB) *ResumeRepository {
	return &ResumeRepository{db: db}
}

// CreateVersion inserts a new resume version and returns the auto-generated ID.
func (r *ResumeRepository) CreateVersion(ctx context.Context, resume *model.ResumeVersion) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO resume_versions
			(user_id, version_no, name, phone, email, location, target_city, target_role,
			 education_json, experience_json, project_json, skills_json,
			 self_evaluation, raw_text,
			 schema_version,
			 target_direction, job_type, expected_cities,
			 school, degree, major, graduation_year,
			 profile_json, job_intention_json, internship_json,
			 certificates_json, awards_json, portfolio_json, extra_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
		         ?, ?, ?, ?, ?, ?, ?, ?,
		         ?, ?, ?, ?, ?, ?, ?)`,
		resume.UserID, resume.VersionNo, resume.Name, resume.Phone, resume.Email,
		resume.Location, resume.TargetCity, resume.TargetRole,
		toNullJSON(resume.EducationJSON), toNullJSON(resume.ExperienceJSON),
		toNullJSON(resume.ProjectJSON), toNullJSON(resume.SkillsJSON),
		toNullString(resume.SelfEvaluation), toNullString(resume.RawText),
		resume.SchemaVersion,
		resume.TargetDirection, resume.JobType, resume.ExpectedCities,
		resume.School, resume.Degree, resume.Major, resume.GraduationYear,
		toNullJSON(resume.ProfileJSON), toNullJSON(resume.JobIntentionJSON),
		toNullJSON(resume.InternshipJSON), toNullJSON(resume.CertificatesJSON),
		toNullJSON(resume.AwardsJSON), toNullJSON(resume.PortfolioJSON),
		toNullJSON(resume.ExtraJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to create resume version: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	resume.ID = uint64(id)
	return nil
}

// FindByID returns a resume version by its ID. Returns nil if not found.
func (r *ResumeRepository) FindByID(ctx context.Context, id uint64) (*model.ResumeVersion, error) {
	var resume model.ResumeVersion
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, version_no, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at
		 FROM resume_versions
		 WHERE id = ?`,
		id,
	).Scan(&resume.ID, &resume.UserID, &resume.VersionNo, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find resume version by id: %w", err)
	}

	resume.EducationJSON = fromNullJSON(educationJSON)
	resume.ExperienceJSON = fromNullJSON(experienceJSON)
	resume.ProjectJSON = fromNullJSON(projectJSON)
	resume.SkillsJSON = fromNullJSON(skillsJSON)
	resume.SelfEvaluation = selfEvaluation.String
	resume.RawText = rawText.String

	resume.TargetDirection = targetDirection.String
	resume.JobType = jobType.String
	resume.ExpectedCities = expectedCities.String
	resume.School = school.String
	resume.Degree = degree.String
	resume.Major = major.String
	resume.GraduationYear = graduationYear.String

	resume.ProfileJSON = fromNullJSON(profileJSON)
	resume.JobIntentionJSON = fromNullJSON(jobIntentionJSON)
	resume.InternshipJSON = fromNullJSON(internshipJSON)
	resume.CertificatesJSON = fromNullJSON(certificatesJSON)
	resume.AwardsJSON = fromNullJSON(awardsJSON)
	resume.PortfolioJSON = fromNullJSON(portfolioJSON)
	resume.ExtraJSON = fromNullJSON(extraJSON)

	return &resume, nil
}

// GetCurrentByUserID returns the latest resume version for the given user.
// Returns nil if no resume exists.
func (r *ResumeRepository) GetCurrentByUserID(ctx context.Context, userID uint64) (*model.ResumeVersion, error) {
	var resume model.ResumeVersion
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, version_no, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at
		 FROM resume_versions
		 WHERE user_id = ?
		 ORDER BY version_no DESC
		 LIMIT 1`,
		userID,
	).Scan(&resume.ID, &resume.UserID, &resume.VersionNo, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current resume: %w", err)
	}

	resume.EducationJSON = fromNullJSON(educationJSON)
	resume.ExperienceJSON = fromNullJSON(experienceJSON)
	resume.ProjectJSON = fromNullJSON(projectJSON)
	resume.SkillsJSON = fromNullJSON(skillsJSON)
	resume.SelfEvaluation = selfEvaluation.String
	resume.RawText = rawText.String

	resume.TargetDirection = targetDirection.String
	resume.JobType = jobType.String
	resume.ExpectedCities = expectedCities.String
	resume.School = school.String
	resume.Degree = degree.String
	resume.Major = major.String
	resume.GraduationYear = graduationYear.String

	resume.ProfileJSON = fromNullJSON(profileJSON)
	resume.JobIntentionJSON = fromNullJSON(jobIntentionJSON)
	resume.InternshipJSON = fromNullJSON(internshipJSON)
	resume.CertificatesJSON = fromNullJSON(certificatesJSON)
	resume.AwardsJSON = fromNullJSON(awardsJSON)
	resume.PortfolioJSON = fromNullJSON(portfolioJSON)
	resume.ExtraJSON = fromNullJSON(extraJSON)

	return &resume, nil
}

// ListVersionsByUserID returns all resume versions for the given user,
// ordered by version_no descending.
func (r *ResumeRepository) ListVersionsByUserID(ctx context.Context, userID uint64) ([]*model.ResumeVersion, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, version_no, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at
		 FROM resume_versions
		 WHERE user_id = ?
		 ORDER BY version_no DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list resume versions: %w", err)
	}
	defer rows.Close()

	var versions []*model.ResumeVersion
	for rows.Next() {
		var v model.ResumeVersion
		var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
		var selfEvaluation, rawText sql.NullString
		var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
		var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
		var targetDirection, jobType, expectedCities sql.NullString
		var school, degree, major, graduationYear sql.NullString

		err := rows.Scan(&v.ID, &v.UserID, &v.VersionNo, &v.Name, &v.Phone,
			&v.Email, &v.Location, &v.TargetCity, &v.TargetRole,
			&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
			&selfEvaluation, &rawText,
			&v.SchemaVersion,
			&targetDirection, &jobType, &expectedCities,
			&school, &degree, &major, &graduationYear,
			&profileJSON, &jobIntentionJSON, &internshipJSON,
			&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
			&v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resume version: %w", err)
		}

		v.EducationJSON = fromNullJSON(educationJSON)
		v.ExperienceJSON = fromNullJSON(experienceJSON)
		v.ProjectJSON = fromNullJSON(projectJSON)
		v.SkillsJSON = fromNullJSON(skillsJSON)
		v.SelfEvaluation = selfEvaluation.String
		v.RawText = rawText.String

		v.TargetDirection = targetDirection.String
		v.JobType = jobType.String
		v.ExpectedCities = expectedCities.String
		v.School = school.String
		v.Degree = degree.String
		v.Major = major.String
		v.GraduationYear = graduationYear.String

		v.ProfileJSON = fromNullJSON(profileJSON)
		v.JobIntentionJSON = fromNullJSON(jobIntentionJSON)
		v.InternshipJSON = fromNullJSON(internshipJSON)
		v.CertificatesJSON = fromNullJSON(certificatesJSON)
		v.AwardsJSON = fromNullJSON(awardsJSON)
		v.PortfolioJSON = fromNullJSON(portfolioJSON)
		v.ExtraJSON = fromNullJSON(extraJSON)

		versions = append(versions, &v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating resume versions: %w", err)
	}

	return versions, nil
}

// GetByUserIDAndVersionNo returns a specific resume version by user ID and version number.
// Returns nil if not found.
func (r *ResumeRepository) GetByUserIDAndVersionNo(ctx context.Context, userID uint64, versionNo int) (*model.ResumeVersion, error) {
	var resume model.ResumeVersion
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, version_no, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at
		 FROM resume_versions
		 WHERE user_id = ? AND version_no = ?`,
		userID, versionNo,
	).Scan(&resume.ID, &resume.UserID, &resume.VersionNo, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get resume version: %w", err)
	}

	resume.EducationJSON = fromNullJSON(educationJSON)
	resume.ExperienceJSON = fromNullJSON(experienceJSON)
	resume.ProjectJSON = fromNullJSON(projectJSON)
	resume.SkillsJSON = fromNullJSON(skillsJSON)
	resume.SelfEvaluation = selfEvaluation.String
	resume.RawText = rawText.String

	resume.TargetDirection = targetDirection.String
	resume.JobType = jobType.String
	resume.ExpectedCities = expectedCities.String
	resume.School = school.String
	resume.Degree = degree.String
	resume.Major = major.String
	resume.GraduationYear = graduationYear.String

	resume.ProfileJSON = fromNullJSON(profileJSON)
	resume.JobIntentionJSON = fromNullJSON(jobIntentionJSON)
	resume.InternshipJSON = fromNullJSON(internshipJSON)
	resume.CertificatesJSON = fromNullJSON(certificatesJSON)
	resume.AwardsJSON = fromNullJSON(awardsJSON)
	resume.PortfolioJSON = fromNullJSON(portfolioJSON)
	resume.ExtraJSON = fromNullJSON(extraJSON)

	return &resume, nil
}

// GetNextVersionNo returns the next version number for the given user.
// If the user has no resume versions, returns 1.
func (r *ResumeRepository) GetNextVersionNo(ctx context.Context, userID uint64) (int, error) {
	var maxVersionNo sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		"SELECT MAX(version_no) FROM resume_versions WHERE user_id = ?",
		userID,
	).Scan(&maxVersionNo)

	if err != nil {
		return 0, fmt.Errorf("failed to get next version number: %w", err)
	}

	if !maxVersionNo.Valid {
		return 1, nil
	}

	return int(maxVersionNo.Int64) + 1, nil
}

// toNullString converts a string to sql.NullString for database writes.
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// toNullJSON converts a []byte to sql.NullString for JSON database writes.
func toNullJSON(data []byte) sql.NullString {
	if len(data) == 0 {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: string(data), Valid: true}
}

// fromNullJSON converts a sql.NullString to []byte for JSON field reading.
func fromNullJSON(ns sql.NullString) []byte {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return []byte(ns.String)
}
