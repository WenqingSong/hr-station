package model

import "time"

// Job represents a job listing in the system.
// Maps to the `jobs` table in MySQL.
type Job struct {
	ID                uint64
	Title             string
	Department        string
	City              string
	Category          string
	JobType           string
	Description       string
	Requirements      string
	SkillKeywordsJSON []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
