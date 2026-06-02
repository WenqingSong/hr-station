package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type JobRecord struct {
	Source            SourceInfo `json:"source"`
	Job               JobInfo    `json:"job"`
	Organizations     []OrgInfo  `json:"organizations"`
	Locations         []string   `json:"locations"`
	InterviewLocations []string  `json:"interview_locations"`
	Description       DescInfo   `json:"description"`
	Requirements      ReqInfo    `json:"requirements"`
	Skills            []string   `json:"skills"`
	Facets            FacetsInfo `json:"facets"`
}

type SourceInfo struct {
	SourceType  string `json:"source_type"`
	SourceName  string `json:"source_name"`
	SourceURL   string `json:"source_url"`
	DataQuality string `json:"data_quality"`
	CollectedAt string `json:"collected_at"`
}

type JobInfo struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	JobType         string `json:"job_type"`
	ApplicationType string `json:"application_type"`
	Direction       string `json:"direction"`
	Category        string `json:"category"`
	Status          string `json:"status"`
	Slogan          string `json:"slogan"`
}

type OrgInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type DescInfo struct {
	Summary         string   `json:"summary"`
	Responsibilities []string `json:"responsibilities"`
}

type ReqInfo struct {
	MustHave    []string `json:"must_have"`
	NiceToHave  []string `json:"nice_to_have"`
	BonusPoints []string `json:"bonus_points"`
	Notes       []string `json:"notes,omitempty"`
}

type FacetsInfo struct {
	ApplicationProject string   `json:"application_project"`
	ApplicationSubtype string   `json:"application_subtype"`
	JobCategory       string   `json:"job_category"`
	JobType           string   `json:"job_type"`
	Direction         string   `json:"direction"`
	Cities            []string `json:"cities"`
	BusinessGroups    []string `json:"business_groups"`
}

func main() {
	filePath := flag.String("file", "", "Path to JSON file")
	dryRun := flag.Bool("dry-run", false, "Validate only, do not write to DB")
	replace := flag.Bool("replace", false, "Drop old curated jobs and import new ones")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("--file is required")
	}
	if !*dryRun && !*replace {
		log.Fatal("either --dry-run or --replace is required")
	}

	jobs, err := loadAndValidate(*filePath)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	fmt.Printf("job_count = %d\n", len(jobs))
	fmt.Printf("validation = pass\n")

	counts := collectStats(jobs)
	fmt.Printf("job_type =\n")
	for k, v := range counts.jobTypes {
		fmt.Printf("  %s: %d\n", k, v)
	}
	fmt.Printf("direction =\n")
	for k, v := range counts.directions {
		fmt.Printf("  %s: %d\n", k, v)
	}
	fmt.Printf("cities =\n")
	for k, v := range counts.cities {
		fmt.Printf("  %s: %d\n", k, v)
	}
	fmt.Printf("business_groups =\n")
	for k, v := range counts.bgs {
		fmt.Printf("  %s: %d\n", k, v)
	}
	locCount, bgCount, skillCount := estimateAssociations(jobs)
	fmt.Printf("will_write_jobs = %d\n", len(jobs))
	fmt.Printf("will_write_locations = %d\n", locCount)
	fmt.Printf("will_write_business_groups = %d\n", bgCount)
	fmt.Printf("will_write_skills = %d\n", skillCount)

	if *dryRun {
		fmt.Println("\n=== dry-run complete, no data written ===")
		return
	}

	dsn := getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}


	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Begin tx error: %v", err)
	}
	defer tx.Rollback()

	// 1. Delete old curated jobs — ordered by FK dependency
	sourcesToReplace := []string{"local_mock", "manual_curated", "demo_curated"}
	placeholders := make([]string, len(sourcesToReplace))
	args := make([]interface{}, len(sourcesToReplace))
	for i, s := range sourcesToReplace {
		placeholders[i] = "?"
		args[i] = s
	}
	inClause := strings.Join(placeholders, ",")

	delSQL := "DELETE FROM match_analyses WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN (" + inClause + "))"
	tx.Exec(delSQL, args...)

	delSQL = "DELETE FROM recommendation_results"
	tx.Exec(delSQL)

	for _, table := range []string{"job_skills", "job_locations", "job_business_groups"} {
		_, err = tx.Exec("DELETE FROM "+table+" WHERE job_id IN (SELECT id FROM jobs WHERE external_source IN ("+inClause+"))", args...)
		if err != nil {
			log.Printf("[warn] delete %s: %v", table, err)
		}
	}

	result, err := tx.Exec("DELETE FROM jobs WHERE external_source IN ("+inClause+")", args...)
	if err != nil {
		log.Fatalf("Delete jobs error: %v", err)
	}
	deleted, _ := result.RowsAffected()
	fmt.Printf("deleted_old_jobs = %d\n", deleted)

	// 2. Insert 46 jobs
	insertSQL := `INSERT INTO jobs (title, department, city, direction, category, job_type, application_type, application_project, application_subtype, slogan, summary, description, requirements, responsibilities_json, must_have_json, nice_to_have_json, bonus_points_json, notes_json, skill_keywords_json, skills_json, organizations_json, locations_json, interview_locations_json, facets_json, status, external_source, external_id, source_url, source_name, raw_json, data_quality, collected_at, is_featured) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`

	insertStmt, err := tx.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("Prepare: %v", err)
	}
	defer insertStmt.Close()

	locStmt, _ := tx.Prepare("INSERT INTO job_locations (job_id, city) VALUES (?, ?)")
	bgStmt, _ := tx.Prepare("INSERT INTO job_business_groups (job_id, code, name) VALUES (?, ?, ?)")
	skStmt, _ := tx.Prepare("INSERT INTO job_skills (job_id, skill) VALUES (?, ?)")
	defer locStmt.Close()
	defer bgStmt.Close()
	defer skStmt.Close()

	for _, job := range jobs {
		respJSON := jsonStr(job.Description.Responsibilities)
		mustJSON := jsonStr(job.Requirements.MustHave)
		niceJSON := jsonStr(job.Requirements.NiceToHave)
		bonusJSON := jsonStr(job.Requirements.BonusPoints)
		notesJSON := jsonStr(job.Requirements.Notes)
		skillsJSON := jsonStr(job.Skills)
		orgsJSON := jsonStr(job.Organizations)
		locsJSON := jsonStr(job.Locations)
		ivLocsJSON := jsonStr(job.InterviewLocations)
		facetsJSON := jsonStr(job.Facets)
		rawJSON := jsonStr(job)

		reqText := buildRequirementsText(job.Requirements)
		department := ""
		if len(job.Organizations) > 0 {
			department = job.Organizations[0].Name
		}
		city := ""
		if len(job.Locations) > 0 {
			city = job.Locations[0]
		}
		direction := job.Job.Direction
		category := job.Job.Category

		var collectedAt interface{}
		if job.Source.CollectedAt != "" {
			if t, err := time.Parse("2006-01-02", job.Source.CollectedAt); err == nil {
				collectedAt = t
			}
		}

		res, err := insertStmt.Exec(
			job.Job.Title, department, city, direction, category, job.Job.JobType,
			job.Job.ApplicationType, job.Facets.ApplicationProject, job.Facets.ApplicationSubtype,
			job.Job.Slogan, job.Description.Summary,
			job.Description.Summary, reqText,
			respJSON, mustJSON, niceJSON, bonusJSON, notesJSON,
			skillsJSON, skillsJSON, orgsJSON, locsJSON, ivLocsJSON, facetsJSON,
			job.Job.Status,
			"manual_curated", job.Job.ID,
			job.Source.SourceURL, job.Source.SourceName,
			rawJSON, "manual_verified", collectedAt,
		)
		if err != nil {
			log.Fatalf("Insert %s (%s): %v", job.Job.ID, job.Job.Title, err)
		}
		jobID, _ := res.LastInsertId()

		for _, loc := range job.Locations {
			locStmt.Exec(jobID, loc)
		}
		for _, org := range job.Organizations {
			bgStmt.Exec(jobID, org.Code, org.Name)
		}
		for _, skill := range job.Skills {
			skStmt.Exec(jobID, skill)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Commit: %v", err)
	}
	fmt.Println("\n=== import complete ===")
	verify(db)
}

func verify(db *sql.DB) {
	var cnt int
	db.QueryRow("SELECT COUNT(*) FROM jobs WHERE external_source = 'manual_curated'").Scan(&cnt)
	fmt.Printf("jobs (manual_curated) = %d\n", cnt)
	db.QueryRow("SELECT COUNT(*) FROM job_locations").Scan(&cnt)
	fmt.Printf("job_locations = %d\n", cnt)
	db.QueryRow("SELECT COUNT(*) FROM job_business_groups").Scan(&cnt)
	fmt.Printf("job_business_groups = %d\n", cnt)
	db.QueryRow("SELECT COUNT(*) FROM job_skills").Scan(&cnt)
	fmt.Printf("job_skills = %d\n", cnt)

	fmt.Println("\n=== source/quality ===")
	rows, _ := db.Query("SELECT external_source, data_quality, COUNT(*) FROM jobs GROUP BY external_source, data_quality")
	for rows.Next() {
		var s, q string
		var c int
		rows.Scan(&s, &q, &c)
		fmt.Printf("  %s / %s: %d\n", s, q, c)
	}
	rows.Close()

	fmt.Println("\n=== orphan check ===")
	tables := []struct{ name, sql string }{
		{"match_analyses", "SELECT COUNT(*) FROM match_analyses ma LEFT JOIN jobs j ON ma.job_id = j.id WHERE ma.job_id IS NOT NULL AND j.id IS NULL"},
		{"job_locations", "SELECT COUNT(*) FROM job_locations jl LEFT JOIN jobs j ON jl.job_id = j.id WHERE j.id IS NULL"},
		{"job_business_groups", "SELECT COUNT(*) FROM job_business_groups bg LEFT JOIN jobs j ON bg.job_id = j.id WHERE j.id IS NULL"},
		{"job_skills", "SELECT COUNT(*) FROM job_skills js LEFT JOIN jobs j ON js.job_id = j.id WHERE j.id IS NULL"},
		{"resume_optimization_results", "SELECT COUNT(*) FROM resume_optimization_results ror LEFT JOIN jobs j ON ror.job_id = j.id WHERE ror.job_id IS NOT NULL AND ror.job_id <> 0 AND j.id IS NULL"},
	}
	for _, t := range tables {
		db.QueryRow(t.sql).Scan(&cnt)
		fmt.Printf("  %s orphans = %d\n", t.name, cnt)
		if cnt > 0 {
			fmt.Printf("  ⚠ WARNING: %d orphan records in %s!\n", cnt, t.name)
		}
	}
}

func loadAndValidate(filePath string) ([]JobRecord, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	var jobs []JobRecord
	if err := json.Unmarshal(data, &jobs); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("empty array")
	}

	seenIDs := map[string]bool{}
	var errs []string
	for i, j := range jobs {
		label := j.Job.Title
		if label == "" {
			label = fmt.Sprintf("index=%d", i)
		}
		if j.Job.ID == "" {
			errs = append(errs, fmt.Sprintf("%s: job.id missing", label))
		}
		if seenIDs[j.Job.ID] {
			errs = append(errs, fmt.Sprintf("%s: duplicate id=%s", label, j.Job.ID))
		}
		seenIDs[j.Job.ID] = true

		checks := map[string]string{
			"title":            j.Job.Title,
			"direction":        j.Job.Direction,
			"category":         j.Job.Category,
			"job_type":         j.Job.JobType,
			"application_type": j.Job.ApplicationType,
			"summary":          j.Description.Summary,
		}
		for field, val := range checks {
			if val == "" {
				errs = append(errs, fmt.Sprintf("%s: %s empty", label, field))
			}
		}
		if len(j.Locations) == 0 {
			errs = append(errs, fmt.Sprintf("%s: locations empty", label))
		}
		if len(j.Organizations) == 0 {
			errs = append(errs, fmt.Sprintf("%s: organizations empty", label))
		}
		if len(j.Description.Responsibilities) == 0 {
			errs = append(errs, fmt.Sprintf("%s: responsibilities empty", label))
		}
		if len(j.Requirements.MustHave) == 0 {
			errs = append(errs, fmt.Sprintf("%s: must_have empty", label))
		}
		if len(j.Skills) == 0 {
			errs = append(errs, fmt.Sprintf("%s: skills empty", label))
		}

		f := j.Facets
		if j.Job.Direction != f.Direction {
			errs = append(errs, fmt.Sprintf("%s: direction mismatch job=%s facets=%s", label, j.Job.Direction, f.Direction))
		}
		if j.Job.Category != f.JobCategory {
			errs = append(errs, fmt.Sprintf("%s: category mismatch", label))
		}
		if j.Job.JobType != f.JobType {
			errs = append(errs, fmt.Sprintf("%s: job_type mismatch", label))
		}
		if j.Job.ApplicationType != f.ApplicationSubtype {
			errs = append(errs, fmt.Sprintf("%s: application_type mismatch", label))
		}
		if join(sorted(j.Locations)) != join(sorted(f.Cities)) {
			errs = append(errs, fmt.Sprintf("%s: locations!=cities", label))
		}
		codes := make([]string, len(j.Organizations))
		for ii, o := range j.Organizations {
			codes[ii] = o.Code
		}
		if join(sorted(codes)) != join(sorted(f.BusinessGroups)) {
			errs = append(errs, fmt.Sprintf("%s: orgs!=business_groups", label))
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("validation errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return jobs, nil
}

type stats struct {
	jobTypes, directions, cities, bgs map[string]int
}

func collectStats(jobs []JobRecord) stats {
	s := stats{
		jobTypes:   map[string]int{},
		directions: map[string]int{},
		cities:     map[string]int{},
		bgs:        map[string]int{},
	}
	for _, j := range jobs {
		s.jobTypes[j.Job.JobType]++
		s.directions[j.Job.Direction]++
		for _, c := range j.Locations {
			s.cities[c]++
		}
		for _, o := range j.Organizations {
			s.bgs[o.Code]++
		}
	}
	return s
}

func estimateAssociations(jobs []JobRecord) (int, int, int) {
	l, b, s := 0, 0, 0
	for _, j := range jobs {
		l += len(j.Locations)
		b += len(j.Organizations)
		s += len(j.Skills)
	}
	return l, b, s
}

func jsonStr(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

func buildRequirementsText(r ReqInfo) string {
	var p []string
	if len(r.MustHave) > 0 {
		p = append(p, "【必备要求】\n"+strings.Join(r.MustHave, "\n"))
	}
	if len(r.NiceToHave) > 0 {
		p = append(p, "【了解项】\n"+strings.Join(r.NiceToHave, "\n"))
	}
	if len(r.BonusPoints) > 0 {
		p = append(p, "【加分项】\n"+strings.Join(r.BonusPoints, "\n"))
	}
	return strings.Join(p, "\n\n")
}

func sorted(s []string) []string {
	r := make([]string, len(s))
	copy(r, s)
	for i := 0; i < len(r); i++ {
		for j := i + 1; j < len(r); j++ {
			if r[i] > r[j] {
				r[i], r[j] = r[j], r[i]
			}
		}
	}
	return r
}

func join(s []string) string { return strings.Join(s, ",") }
func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
