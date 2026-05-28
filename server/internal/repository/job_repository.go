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
	Keyword  string
	City     string
	Category string
	Limit    int
	Offset   int
}

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
	listSQL := "SELECT id, title, department, city, category, job_type, description, requirements, skill_keywords_json, created_at, updated_at FROM jobs" +
		where + " ORDER BY id ASC LIMIT ? OFFSET ?"
	listArgs := append(args, query.Limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*model.Job
	for rows.Next() {
		var j model.Job
		var skillKeywordsJSON sql.NullString
		var description, requirements sql.NullString

		if err := rows.Scan(&j.ID, &j.Title, &j.Department, &j.City, &j.Category,
			&j.JobType, &description, &requirements, &skillKeywordsJSON,
			&j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}

		j.Description = description.String
		j.Requirements = requirements.String
		j.SkillKeywordsJSON = fromNullJSON(skillKeywordsJSON)

		jobs = append(jobs, &j)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, total, nil
}

// FindByID returns a job by its ID. Returns nil if not found.
func (r *JobRepository) FindByID(ctx context.Context, id uint64) (*model.Job, error) {
	var j model.Job
	var skillKeywordsJSON sql.NullString
	var description, requirements sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, title, department, city, category, job_type, description, requirements, skill_keywords_json, created_at, updated_at FROM jobs WHERE id = ?",
		id,
	).Scan(&j.ID, &j.Title, &j.Department, &j.City, &j.Category,
		&j.JobType, &description, &requirements, &skillKeywordsJSON,
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

	return &j, nil
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

	if query.City != "" {
		conditions = append(conditions, "city = ?")
		args = append(args, query.City)
	}

	if query.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, query.Category)
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}

	return "", args
}
