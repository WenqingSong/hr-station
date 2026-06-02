package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

// JobQuery holds filter parameters for job listing queries.
type JobQuery struct {
	Keyword            string
	Direction          string
	City               string
	RecruitType        string
	Category           string
	BusinessGroup      string
	Skill              string
	ApplicationProject string
	ApplicationSubtype string
	ApplicationType    string
	ExternalSource     string
	Limit              int
	Offset             int
}

// FacetItem represents a single facet entry (value + count).
type FacetItem struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// JobFacets holds aggregated facet counts.
type JobFacets struct {
	Directions           []FacetItem `json:"directions"`
	Cities               []FacetItem `json:"cities"`
	RecruitTypes         []FacetItem `json:"recruit_types"`
	Categories           []FacetItem `json:"categories"`
	ApplicationProjects  []FacetItem `json:"application_projects"`
	ApplicationSubtypes  []FacetItem `json:"application_subtypes"`
	JobTypes             []FacetItem `json:"job_types"`
	BusinessGroups       []FacetItem `json:"business_groups"`
	Skills               []FacetItem `json:"skills"`
}

const jobSelectCols = `id, title, department, city, direction, category, job_type,
	application_type, application_project, application_subtype,
	slogan, summary, description, requirements,
	responsibilities_json, must_have_json, nice_to_have_json, bonus_points_json, notes_json,
	skill_keywords_json, skills_json,
	organizations_json, locations_json, interview_locations_json, facets_json,
	status, external_source, external_id, external_inner_id, source_url, source_name,
	data_quality, collected_at, is_featured, synced_at, created_at, updated_at`

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
	where, args, hasJoin := r.buildWhereClause(query)

	fromClause := "FROM jobs"
	if hasJoin {
		fromClause = "FROM jobs" + r.buildJoinClause(query)
	}

	// Count total
	countSQL := "SELECT COUNT(DISTINCT jobs.id) " + fromClause + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Query list
	listSQL := "SELECT DISTINCT jobs." + strings.ReplaceAll(jobSelectCols, ", ", ", jobs.") + " " +
		fromClause + where + " ORDER BY jobs.is_featured DESC, jobs.id ASC LIMIT ? OFFSET ?"
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
	err := scanJobRow(r.db.QueryRowContext(ctx, "SELECT "+jobSelectCols+" FROM jobs WHERE id = ?", id), &j)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find job by id: %w", err)
	}
	return &j, nil
}

func scanJobRow(row interface{ Scan(...interface{}) error }, j *model.Job) error {
	var summary sql.NullString
	if err := row.Scan(
		&j.ID, &j.Title, &j.Department, &j.City, &j.Direction, &j.Category, &j.JobType,
		&j.ApplicationType, &j.ApplicationProject, &j.ApplicationSubtype,
		&j.Slogan, &summary, &j.Description, &j.Requirements,
		&j.ResponsibilitiesJSON, &j.MustHaveJSON, &j.NiceToHaveJSON, &j.BonusPointsJSON, &j.NotesJSON,
		&j.SkillKeywordsJSON, &j.SkillsJSON,
		&j.OrganizationsJSON, &j.LocationsJSON, &j.InterviewLocationsJSON, &j.FacetsJSON,
		&j.Status, &j.ExternalSource, &j.ExternalID, &j.ExternalInnerID, &j.SourceURL, &j.SourceName,
		&j.DataQuality, &j.CollectedAt, &j.IsFeatured, &j.SyncedAt,
		&j.CreatedAt, &j.UpdatedAt,
	); err != nil {
		return err
	}
	j.Summary = summary.String
	return nil
}

// GetFacets returns aggregated facet counts across all jobs.
func (r *JobRepository) GetFacets(ctx context.Context) (*JobFacets, error) {
	f := &JobFacets{}

	// directions — from jobs table
	if rows, err := r.db.QueryContext(ctx, "SELECT direction, COUNT(*) FROM jobs WHERE direction != '' GROUP BY direction ORDER BY COUNT(*) DESC"); err == nil {
		f.Directions, _ = scanFacets(rows)
	}

	// cities — from job_locations
	if rows, err := r.db.QueryContext(ctx, "SELECT jl.city, COUNT(DISTINCT jl.job_id) FROM job_locations jl INNER JOIN jobs j ON j.id=jl.job_id GROUP BY jl.city ORDER BY COUNT(DISTINCT jl.job_id) DESC"); err == nil {
		f.Cities, _ = scanFacets(rows)
	}

	// recruit_types — from jobs.job_type (backward compat)
	if rows, err := r.db.QueryContext(ctx, "SELECT job_type, COUNT(*) FROM jobs WHERE job_type != '' GROUP BY job_type ORDER BY COUNT(*) DESC"); err == nil {
		f.RecruitTypes, _ = scanFacets(rows)
	}

	// categories — from jobs table
	if rows, err := r.db.QueryContext(ctx, "SELECT category, COUNT(*) FROM jobs WHERE category != '' GROUP BY category ORDER BY COUNT(*) DESC"); err == nil {
		f.Categories, _ = scanFacets(rows)
	}

	// application_projects
	if rows, err := r.db.QueryContext(ctx, "SELECT application_project, COUNT(*) FROM jobs WHERE application_project != '' GROUP BY application_project ORDER BY COUNT(*) DESC"); err == nil {
		f.ApplicationProjects, _ = scanFacets(rows)
	}

	// application_subtypes
	if rows, err := r.db.QueryContext(ctx, "SELECT application_subtype, COUNT(*) FROM jobs WHERE application_subtype != '' GROUP BY application_subtype ORDER BY COUNT(*) DESC"); err == nil {
		f.ApplicationSubtypes, _ = scanFacets(rows)
	}

	// job_types
	if rows, err := r.db.QueryContext(ctx, "SELECT job_type, COUNT(*) FROM jobs WHERE job_type != '' GROUP BY job_type ORDER BY COUNT(*) DESC"); err == nil {
		f.JobTypes, _ = scanFacets(rows)
	}

	// business_groups
	if rows, err := r.db.QueryContext(ctx, "SELECT jbg.code, COUNT(DISTINCT jbg.job_id) FROM job_business_groups jbg INNER JOIN jobs j ON j.id=jbg.job_id GROUP BY jbg.code ORDER BY COUNT(DISTINCT jbg.job_id) DESC"); err == nil {
		f.BusinessGroups, _ = scanFacets(rows)
	}

	// skills
	if rows, err := r.db.QueryContext(ctx, "SELECT js.skill, COUNT(DISTINCT js.job_id) FROM job_skills js INNER JOIN jobs j ON j.id=js.job_id GROUP BY js.skill ORDER BY COUNT(DISTINCT js.job_id) DESC LIMIT 50"); err == nil {
		f.Skills, _ = scanFacets(rows)
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

func (r *JobRepository) buildJoinClause(query JobQuery) string {
	var joins []string
	if query.City != "" {
		joins = append(joins, " INNER JOIN job_locations jl ON jobs.id = jl.job_id")
	}
	if query.BusinessGroup != "" {
		joins = append(joins, " INNER JOIN job_business_groups jbg ON jobs.id = jbg.job_id")
	}
	if query.Skill != "" {
		joins = append(joins, " INNER JOIN job_skills js ON jobs.id = js.job_id")
	}
	return strings.Join(joins, " ")
}

func (r *JobRepository) buildWhereClause(query JobQuery) (string, []interface{}, bool) {
	var conditions []string
	var args []interface{}
	hasJoin := false

	if query.Keyword != "" {
		conditions = append(conditions, "(jobs.title LIKE ? OR jobs.department LIKE ? OR jobs.description LIKE ? OR jobs.requirements LIKE ? OR jobs.skills_json LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like, like, like)
	}

	if query.Direction != "" {
		conditions = append(conditions, "jobs.direction = ?")
		args = append(args, query.Direction)
	}

	if query.City != "" {
		conditions = append(conditions, "jl.city = ?")
		args = append(args, query.City)
		hasJoin = true
	}

	if query.RecruitType != "" {
		conditions = append(conditions, "jobs.job_type = ?")
		args = append(args, query.RecruitType)
	}

	if query.Category != "" {
		conditions = append(conditions, "jobs.category = ?")
		args = append(args, query.Category)
	}

	if query.BusinessGroup != "" {
		conditions = append(conditions, "jbg.code = ?")
		args = append(args, query.BusinessGroup)
		hasJoin = true
	}

	if query.Skill != "" {
		conditions = append(conditions, "js.skill = ?")
		args = append(args, query.Skill)
		hasJoin = true
	}

	if query.ApplicationProject != "" {
		conditions = append(conditions, "jobs.application_project = ?")
		args = append(args, query.ApplicationProject)
	}

	if query.ApplicationSubtype != "" {
		conditions = append(conditions, "jobs.application_subtype = ?")
		args = append(args, query.ApplicationSubtype)
	}

	if query.ApplicationType != "" {
		conditions = append(conditions, "jobs.application_type = ?")
		args = append(args, query.ApplicationType)
	}

	if query.ExternalSource != "" {
		conditions = append(conditions, "jobs.external_source = ?")
		args = append(args, query.ExternalSource)
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args, hasJoin
	}
	return "", args, hasJoin
}

func (r *JobRepository) ListRecommendCandidates(ctx context.Context, limit int) ([]*model.Job, error) {
	if limit <= 0 {
		limit = 50
	}
	query := "SELECT jobs." + strings.ReplaceAll(jobSelectCols, ", ", ", jobs.") + " FROM jobs ORDER BY jobs.is_featured DESC, CASE jobs.data_quality WHEN 'manual_verified' THEN 0 WHEN 'manual_enriched' THEN 1 WHEN 'summary_only' THEN 2 ELSE 3 END, jobs.id ASC LIMIT ?"
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recommend candidates: %w", err)
	}
	defer rows.Close()
	return scanJobs(rows)
}

func scanJobs(rows *sql.Rows) ([]*model.Job, error) {
	var jobs []*model.Job
	for rows.Next() {
		var j model.Job
		if err := scanJobRow(rows, &j); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &j)
	}
	return jobs, rows.Err()
}
