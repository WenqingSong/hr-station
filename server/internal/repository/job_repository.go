package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"offer-catcher/server/internal/model"
)

// JobQuery holds filter parameters for job listing queries.
type JobQuery struct {
	Keyword        string
	Direction      string
	City           string
	RecruitType    string
	Category       string
	ExternalSource string
	Limit          int
	Offset         int
}

// FacetItem represents a single facet entry (value + count).
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// JobFacets holds aggregated facet counts for the job listing.
type JobFacets struct {
	Directions  []FacetItem `json:"directions"`
	Cities      []FacetItem `json:"cities"`
	RecruitTypes []FacetItem `json:"recruit_types"`
	Categories  []FacetItem `json:"categories"`
}

const jobSelectCols = "id, title, department, city, direction, category, job_type, description, requirements, skill_keywords_json, status, external_source, external_id, external_inner_id, source_url, data_quality, is_featured, synced_at, created_at, updated_at"

// JobRepository handles database operations for jobs.
type JobRepository struct {
	db *sql.DB
}

// NewJobRepository creates a new JobRepository.
func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{db: db}
}

// List returns a list of jobs matching the query, along with the total count.
func (r *JobRepository) List(ctx context.Context, query JobQuery) ([]*model.Job, int, error) {
	where, args := r.buildWhereClause(query)

	// Count total
	countSQL := "SELECT COUNT(*) FROM jobs" + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Query list
	listSQL := "SELECT " + jobSelectCols + " FROM jobs" +
		where + " ORDER BY is_featured DESC, id ASC LIMIT ? OFFSET ?"
	listArgs := append(args, query.Limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	jobs, err := scanJobs(rows)
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// FindByID returns a job by its ID. Returns nil if not found.
func (r *JobRepository) FindByID(ctx context.Context, id uint64) (*model.Job, error) {
	var j model.Job
	var skillKeywordsJSON sql.NullString
	var description, requirements sql.NullString
	var externalID, externalInnerID, sourceURL sql.NullString
	var syncedAt sql.NullTime

	err := r.db.QueryRowContext(ctx,
		"SELECT "+jobSelectCols+" FROM jobs WHERE id = ?",
		id,
	).Scan(&j.ID, &j.Title, &j.Department, &j.City, &j.Direction,
		&j.Category, &j.JobType, &description, &requirements, &skillKeywordsJSON,
		&j.Status,
		&j.ExternalSource, &externalID, &externalInnerID, &sourceURL,
		&j.DataQuality, &j.IsFeatured, &syncedAt,
		&j.CreatedAt, &j.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find job by id: %w", err)
	}

	j.Description = description.String
	j.Requirements = requirements.String
	j.SkillKeywordsJSON = fromNullJSON(skillKeywordsJSON)
	j.ExternalID = fromNullString(externalID)
	j.ExternalInnerID = fromNullString(externalInnerID)
	j.SourceURL = fromNullString(sourceURL)
	j.SyncedAt = fromNullTime(syncedAt)

	return &j, nil
}

// GetFacets returns aggregated facet counts across all jobs (directions, cities, recruit_types, categories).
func (r *JobRepository) GetFacets(ctx context.Context) (*JobFacets, error) {
	f := &JobFacets{}

	// directions
	rows, err := r.db.QueryContext(ctx, "SELECT direction, COUNT(*) FROM jobs WHERE direction != '' GROUP BY direction ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query direction facets: %w", err)
	}
	f.Directions, err = scanFacets(rows)
	if err != nil {
		return nil, err
	}

	// cities
	rows, err = r.db.QueryContext(ctx, "SELECT city, COUNT(*) FROM jobs WHERE city != '' GROUP BY city ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query city facets: %w", err)
	}
	f.Cities, err = scanFacets(rows)
	if err != nil {
		return nil, err
	}

	// recruit_types (maps to job_type)
	rows, err = r.db.QueryContext(ctx, "SELECT job_type, COUNT(*) FROM jobs WHERE job_type != '' GROUP BY job_type ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query recruit_type facets: %w", err)
	}
	f.RecruitTypes, err = scanFacets(rows)
	if err != nil {
		return nil, err
	}

	// categories
	rows, err = r.db.QueryContext(ctx, "SELECT category, COUNT(*) FROM jobs WHERE category != '' GROUP BY category ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query category facets: %w", err)
	}
	f.Categories, err = scanFacets(rows)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func scanFacets(rows *sql.Rows) ([]FacetItem, error) {
	defer rows.Close()
	var items []FacetItem
	for rows.Next() {
		var item FacetItem
		if err := rows.Scan(&item.Value, &item.Count); err != nil {
			return nil, fmt.Errorf("failed to scan facet row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating facets: %w", err)
	}
	return items, nil
}

// buildWhereClause builds the WHERE clause and args for job queries.
func (r *JobRepository) buildWhereClause(query JobQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if query.Keyword != "" {
		conditions = append(conditions, "(title LIKE ? OR department LIKE ? OR description LIKE ? OR requirements LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like, like)
	}

	if query.Direction != "" {
		conditions = append(conditions, "direction = ?")
		args = append(args, query.Direction)
	}

	if query.City != "" {
		conditions = append(conditions, "city = ?")
		args = append(args, query.City)
	}

	if query.RecruitType != "" {
		conditions = append(conditions, "job_type = ?")
		args = append(args, query.RecruitType)
	}

	if query.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, query.Category)
	}

	if query.ExternalSource != "" {
		conditions = append(conditions, "external_source = ?")
		args = append(args, query.ExternalSource)
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}

	return "", args
}

// fromNullString converts sql.NullString to *string pointer.
func fromNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// fromNullTime converts sql.NullTime to *time.Time pointer.
func fromNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// ListRecommendCandidates returns jobs suitable for recommendation candidates.
func (r *JobRepository) ListRecommendCandidates(ctx context.Context, limit int) ([]*model.Job, error) {
	if limit <= 0 {
		limit = 50
	}

	query := "SELECT " + jobSelectCols + " FROM jobs ORDER BY is_featured DESC, CASE data_quality WHEN 'full_jd' THEN 0 WHEN 'manual_enriched' THEN 1 WHEN 'summary_only' THEN 2 ELSE 3 END, id ASC LIMIT ?"

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recommend candidates: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

// scanJobs scans multiple job rows from a sql.Rows result.
func scanJobs(rows *sql.Rows) ([]*model.Job, error) {
	var jobs []*model.Job
	for rows.Next() {
		var j model.Job
		var skillKeywordsJSON sql.NullString
		var description, requirements sql.NullString
		var externalID, externalInnerID, sourceURL sql.NullString
		var syncedAt sql.NullTime

		if err := rows.Scan(&j.ID, &j.Title, &j.Department, &j.City, &j.Direction,
			&j.Category, &j.JobType, &description, &requirements, &skillKeywordsJSON,
			&j.Status,
			&j.ExternalSource, &externalID, &externalInnerID, &sourceURL,
			&j.DataQuality, &j.IsFeatured, &syncedAt,
			&j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		j.Description = description.String
		j.Requirements = requirements.String
		j.SkillKeywordsJSON = fromNullJSON(skillKeywordsJSON)
		j.ExternalID = fromNullString(externalID)
		j.ExternalInnerID = fromNullString(externalInnerID)
		j.SourceURL = fromNullString(sourceURL)
		j.SyncedAt = fromNullTime(syncedAt)

		jobs = append(jobs, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}
