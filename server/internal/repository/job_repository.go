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
	// K8-4-3: 筛选维表参数（多选）
	Cities          []string
	BusinessGroups  []string
	Skills          []string
	JobCategories   []string
	JobTypes        []string // R2
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

// ─── K8-4-3: 筛选维表类型 ─────────────────────────────────────────────────────

// FilterGroupRow 对应 job_filter_groups 表
type FilterGroupRow struct {
	ID                int
	Code              string
	Name              string
	DisplayOrder      int
	DefaultExpanded   bool
	MaxVisibleOptions int
	Visible           bool
}

// FilterOptionRow 对应 job_filter_options 表
type FilterOptionRow struct {
	ID             int
	GroupID        int
	Code           string
	Label          string
	Value          string
	DisplayOrder   int
	Visible        bool
	ParentOptionID *int
}

// FilterOptionWithCount 带岗位数量的筛选项
type FilterOptionWithCount struct {
	Option FilterOptionRow
	Count  int
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
	where, args, hasJoin := r.buildWhereClause(ctx, query)
	fromClause := "FROM jobs"
	if hasJoin {
		fromClause = "FROM jobs" + r.buildJoinClause(query)
	}

	countSQL := "SELECT COUNT(DISTINCT jobs.id) " + fromClause + where
	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

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

// FindByID returns a job by its ID.
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

// GetFacets returns aggregated facet counts across all jobs (旧兼容版本).
func (r *JobRepository) GetFacets(ctx context.Context) (*JobFacets, error) {
	f := &JobFacets{}
	if rows, err := r.db.QueryContext(ctx, "SELECT direction, COUNT(*) FROM jobs WHERE direction != '' GROUP BY direction ORDER BY COUNT(*) DESC"); err == nil {
		f.Directions, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT jl.city, COUNT(DISTINCT jl.job_id) FROM job_locations jl INNER JOIN jobs j ON j.id=jl.job_id GROUP BY jl.city ORDER BY COUNT(DISTINCT jl.job_id) DESC"); err == nil {
		f.Cities, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT job_type, COUNT(*) FROM jobs WHERE job_type != '' GROUP BY job_type ORDER BY COUNT(*) DESC"); err == nil {
		f.RecruitTypes, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT category, COUNT(*) FROM jobs WHERE category != '' GROUP BY category ORDER BY COUNT(*) DESC"); err == nil {
		f.Categories, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT application_project, COUNT(*) FROM jobs WHERE application_project != '' GROUP BY application_project ORDER BY COUNT(*) DESC"); err == nil {
		f.ApplicationProjects, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT application_subtype, COUNT(*) FROM jobs WHERE application_subtype != '' GROUP BY application_subtype ORDER BY COUNT(*) DESC"); err == nil {
		f.ApplicationSubtypes, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT job_type, COUNT(*) FROM jobs WHERE job_type != '' GROUP BY job_type ORDER BY COUNT(*) DESC"); err == nil {
		f.JobTypes, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT jbg.code, COUNT(DISTINCT jbg.job_id) FROM job_business_groups jbg INNER JOIN jobs j ON j.id=jbg.job_id GROUP BY jbg.code ORDER BY COUNT(DISTINCT jbg.job_id) DESC"); err == nil {
		f.BusinessGroups, _ = scanFacets(rows)
	}
	if rows, err := r.db.QueryContext(ctx, "SELECT js.skill, COUNT(DISTINCT js.job_id) FROM job_skills js INNER JOIN jobs j ON j.id=js.job_id GROUP BY js.skill ORDER BY COUNT(DISTINCT js.job_id) DESC LIMIT 50"); err == nil {
		f.Skills, _ = scanFacets(rows)
	}
	return f, nil
}

// ─── K8-4-3: 筛选维表查询 ─────────────────────────────────────────────────────

// GetFilterGroups returns all visible filter groups ordered by display_order.
func (r *JobRepository) GetFilterGroups(ctx context.Context) ([]FilterGroupRow, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, code, name, display_order, default_expanded, max_visible_options, visible FROM job_filter_groups WHERE visible = 1 ORDER BY display_order ASC")
	if err != nil {
		return nil, fmt.Errorf("get filter groups: %w", err)
	}
	defer rows.Close()
	var groups []FilterGroupRow
	for rows.Next() {
		var g FilterGroupRow
		if err := rows.Scan(&g.ID, &g.Code, &g.Name, &g.DisplayOrder, &g.DefaultExpanded, &g.MaxVisibleOptions, &g.Visible); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

// GetFilterOptions returns visible options for a group, ordered by display_order.
func (r *JobRepository) GetFilterOptions(ctx context.Context, groupID int) ([]FilterOptionRow, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, group_id, code, label, value, display_order, visible, parent_option_id FROM job_filter_options WHERE group_id = ? AND visible = 1 ORDER BY display_order ASC", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var opts []FilterOptionRow
	for rows.Next() {
		var o FilterOptionRow
		if err := rows.Scan(&o.ID, &o.GroupID, &o.Code, &o.Label, &o.Value, &o.DisplayOrder, &o.Visible, &o.ParentOptionID); err != nil {
			return nil, err
		}
		opts = append(opts, o)
	}
	return opts, rows.Err()
}

// GetOptionCounts returns a map: option_id → number of associated jobs.
func (r *JobRepository) GetOptionCounts(ctx context.Context) (map[int]int, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT r.option_id, COUNT(DISTINCT r.job_id) FROM job_filter_option_relations r INNER JOIN jobs j ON j.id = r.job_id WHERE j.status = 'active' OR j.status = 'open' GROUP BY r.option_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[int]int{}
	for rows.Next() {
		var oid, cnt int
		if err := rows.Scan(&oid, &cnt); err != nil {
			return nil, err
		}
		counts[oid] = cnt
	}
	return counts, rows.Err()
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
	if query.BusinessGroup != "" || len(query.BusinessGroups) > 0 {
		joins = append(joins, " INNER JOIN job_business_groups jbg ON jobs.id = jbg.job_id")
	}
	if query.Skill != "" || len(query.Skills) > 0 {
		joins = append(joins, " INNER JOIN job_skills js ON jobs.id = js.job_id")
	}
	return strings.Join(joins, " ")
}

func (r *JobRepository) buildWhereClause(ctx context.Context, query JobQuery) (string, []interface{}, bool) {
	var conditions []string
	var args []interface{}
	hasJoin := false

	if query.Keyword != "" {
		conditions = append(conditions, "(jobs.title LIKE ? OR jobs.department LIKE ? OR jobs.summary LIKE ? OR jobs.requirements LIKE ?)")
		like := "%" + query.Keyword + "%"
		args = append(args, like, like, like, like)
	}
	if query.Direction != "" {
		conditions = append(conditions, "jobs.direction = ?")
		args = append(args, query.Direction)
	}
	if query.RecruitType != "" {
		conditions = append(conditions, "jobs.job_type = ?")
		args = append(args, query.RecruitType)
	}
	if query.Category != "" {
		conditions = append(conditions, "jobs.category = ?")
		args = append(args, query.Category)
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

	// 多选筛选参数（使用 ? 占位符）
	if len(query.Cities) > 0 {
		phs := make([]string, len(query.Cities))
		for i, c := range query.Cities { phs[i] = "?"; args = append(args, c) }
		conditions = append(conditions, "jobs.city IN ("+strings.Join(phs, ",")+")")
	}
	if len(query.BusinessGroups) > 0 {
		phs := make([]string, len(query.BusinessGroups))
		for i, bg := range query.BusinessGroups { phs[i] = "?"; args = append(args, bg) }
		conditions = append(conditions, "jbg.code IN ("+strings.Join(phs, ",")+")")
		hasJoin = true
	}
	if len(query.Skills) > 0 {
		phs := make([]string, len(query.Skills))
		for i, sk := range query.Skills { phs[i] = "?"; args = append(args, sk) }
		conditions = append(conditions, "js.skill IN ("+strings.Join(phs, ",")+")")
		hasJoin = true
	}
	if len(query.JobCategories) > 0 {
		phs := make([]string, len(query.JobCategories))
		for i, cat := range query.JobCategories { phs[i] = "?"; args = append(args, cat) }
		conditions = append(conditions, "jobs.category IN ("+strings.Join(phs, ",")+")")
	}
	if len(query.JobTypes) > 0 {
		phs := make([]string, len(query.JobTypes))
		for i, jt := range query.JobTypes { phs[i] = "?"; args = append(args, jt) }
		conditions = append(conditions, "jobs.job_type IN ("+strings.Join(phs, ",")+")")
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
