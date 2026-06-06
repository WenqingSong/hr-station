package repository

import (
	"context"
	"database/sql"
	"fmt"

	"offer-catcher/server/internal/model"
)

// ResumeRepository handles database operations for resumes.
type ResumeRepository struct {
	db *sql.DB
}

// NewResumeRepository creates a new ResumeRepository.
func NewResumeRepository(db *sql.DB) *ResumeRepository {
	return &ResumeRepository{db: db}
}

// Create inserts a new resume and returns the auto-generated ID.
func (r *ResumeRepository) Create(ctx context.Context, resume *model.Resume) error {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO resumes
			(user_id, name, phone, email, location, target_city, target_role,
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
		resume.UserID, resume.Name, resume.Phone, resume.Email,
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
		return fmt.Errorf("failed to create resume: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	resume.ID = uint64(id)
	return nil
}

// FindByID returns a resume by its ID. Returns nil if not found.
func (r *ResumeRepository) FindByID(ctx context.Context, id uint64) (*model.Resume, error) {
	var resume model.Resume
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at, updated_at
		 FROM resumes
		 WHERE id = ?`,
		id,
	).Scan(&resume.ID, &resume.UserID, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt, &resume.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find resume by id: %w", err)
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

// GetCurrentByUserID returns the latest resume for the given user.
// Returns nil if no resume exists.
func (r *ResumeRepository) GetCurrentByUserID(ctx context.Context, userID uint64) (*model.Resume, error) {
	var resume model.Resume
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at, updated_at
		 FROM resumes
		 WHERE user_id = ?
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&resume.ID, &resume.UserID, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt, &resume.UpdatedAt)

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

// ListByUserID returns all resumes for the given user, ordered by created_at descending.
func (r *ResumeRepository) ListByUserID(ctx context.Context, userID uint64) ([]*model.Resume, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at, updated_at
		 FROM resumes
		 WHERE user_id = ?
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list resumes: %w", err)
	}
	defer rows.Close()

	var versions []*model.Resume
	for rows.Next() {
		var v model.Resume
		var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
		var selfEvaluation, rawText sql.NullString
		var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
		var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
		var targetDirection, jobType, expectedCities sql.NullString
		var school, degree, major, graduationYear sql.NullString

		err := rows.Scan(&v.ID, &v.UserID, &v.Name, &v.Phone,
			&v.Email, &v.Location, &v.TargetCity, &v.TargetRole,
			&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
			&selfEvaluation, &rawText,
			&v.SchemaVersion,
			&targetDirection, &jobType, &expectedCities,
			&school, &degree, &major, &graduationYear,
			&profileJSON, &jobIntentionJSON, &internshipJSON,
			&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
			&v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resume: %w", err)
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
		return nil, fmt.Errorf("error iterating resumes: %w", err)
	}

	return versions, nil
}

// FindByUserIDAndID returns a specific resume by user ID and resume ID.
func (r *ResumeRepository) FindByUserIDAndID(ctx context.Context, userID uint64, id uint64) (*model.Resume, error) {
	var resume model.Resume
	var educationJSON, experienceJSON, projectJSON, skillsJSON sql.NullString
	var selfEvaluation, rawText sql.NullString
	var profileJSON, jobIntentionJSON, internshipJSON sql.NullString
	var certificatesJSON, awardsJSON, portfolioJSON, extraJSON sql.NullString
	var targetDirection, jobType, expectedCities sql.NullString
	var school, degree, major, graduationYear sql.NullString

	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, phone, email, location, target_city, target_role,
		        education_json, experience_json, project_json, skills_json,
		        self_evaluation, raw_text,
		        schema_version,
		        target_direction, job_type, expected_cities,
		        school, degree, major, graduation_year,
		        profile_json, job_intention_json, internship_json,
		        certificates_json, awards_json, portfolio_json, extra_json,
		        created_at, updated_at
		 FROM resumes
		 WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&resume.ID, &resume.UserID, &resume.Name, &resume.Phone,
		&resume.Email, &resume.Location, &resume.TargetCity, &resume.TargetRole,
		&educationJSON, &experienceJSON, &projectJSON, &skillsJSON,
		&selfEvaluation, &rawText,
		&resume.SchemaVersion,
		&targetDirection, &jobType, &expectedCities,
		&school, &degree, &major, &graduationYear,
		&profileJSON, &jobIntentionJSON, &internshipJSON,
		&certificatesJSON, &awardsJSON, &portfolioJSON, &extraJSON,
		&resume.CreatedAt, &resume.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find resume: %w", err)
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
