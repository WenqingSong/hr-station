package model

import "time"

// ResumeVersion represents a resume version in the system.
// Maps to the `resume_versions` table in MySQL.
// User has a single main resume, managed through version iteration.
//
// V1 fields: id..raw_text (schema_version defaults to 1)
// V2 fields: schema_version, target_direction, job_type, expected_cities,
//            school, degree, major, graduation_year,
//            profile_json, job_intention_json, internship_json,
//            certificates_json, awards_json, portfolio_json, extra_json
type ResumeVersion struct {
	ID             uint64
	UserID         uint64
	VersionNo      int
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
}
