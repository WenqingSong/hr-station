package model

import "time"

// ResumeVersion represents a resume version in the system.
// Maps to the `resume_versions` table in MySQL.
// User has a single main resume, managed through version iteration.
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
	CreatedAt      time.Time
}
