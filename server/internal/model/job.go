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
	ExternalSource    string
	ExternalID        *string
	ExternalInnerID   *string
	SourceURL         *string
	RawJSON           []byte // stored but not exposed in API responses
	DataQuality       string
	IsFeatured        bool
	SyncedAt          *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
