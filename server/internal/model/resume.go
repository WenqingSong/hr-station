package model

import "time"

// Resume represents a resume in the system.
// Maps to the `resumes` table in MySQL.
//
// V1 fields: id..raw_text (schema_version defaults to 1)
// V2 fields: schema_version, target_direction, job_type, expected_cities,
//            school, degree, major, graduation_year,
//            profile_json, job_intention_json, internship_json,
//            certificates_json, awards_json, portfolio_json, extra_json
type Resume struct {
	ID             uint64
	UserID         uint64
	Name           string
	Phone          string
	Email          string
	Location       string
	TargetCity     string
	TargetRole     string
	EducationJSON  []byte
	ExperienceJSON []byte
	ProjectJSON    []byte
	SkillsJSON     []byte
	SelfEvaluation string
	RawText        string

	// V2 fields (K8-5-22-B)
	SchemaVersion int

	TargetDirection string
	JobType         string
	ExpectedCities  string
	School          string
	Degree          string
	Major           string
	GraduationYear  string

	ProfileJSON      []byte
	JobIntentionJSON []byte
	InternshipJSON   []byte
	CertificatesJSON []byte
	AwardsJSON       []byte
	PortfolioJSON    []byte
	ExtraJSON        []byte

	CreatedAt time.Time
	UpdatedAt time.Time

	// VersionNo is NOT a DB column in the `resumes` table (removed in K8-5-51-B).
	// It is kept here only as a transient field for API response backward compatibility.
	VersionNo int
}

// ResumeVersion is an alias for Resume (backward compatibility).
// Deprecated: use Resume directly.
type ResumeVersion = Resume
