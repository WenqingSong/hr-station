package model

import "time"

// Job represents a job listing in the system (J3 expanded).
// Maps to the `jobs` table in MySQL.
type Job struct {
	ID                     uint64
	Title                  string
	Department             string
	City                   string
	Direction              string
	Category               string
	JobType                string
	ApplicationType        string
	ApplicationProject     string
	ApplicationSubtype     string
	Slogan                 string
	Summary                string
	Description            string
	Requirements           string
	ResponsibilitiesJSON   []byte
	MustHaveJSON           []byte
	NiceToHaveJSON         []byte
	BonusPointsJSON        []byte
	NotesJSON              []byte
	SkillKeywordsJSON      []byte
	SkillsJSON             []byte
	OrganizationsJSON      []byte
	LocationsJSON          []byte
	InterviewLocationsJSON []byte
	FacetsJSON             []byte
	Status                 string
	ExternalSource         string
	ExternalID             *string
	ExternalInnerID        *string
	SourceURL              *string
	SourceName             string
	RawJSON                []byte
	DataQuality            string
	CollectedAt            *time.Time
	IsFeatured             bool
	SyncedAt               *time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
