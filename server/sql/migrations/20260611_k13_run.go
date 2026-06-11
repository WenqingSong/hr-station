//go:build ignore
// +build ignore

// K13-Fix-D: Non-destructive migration runner.
// Run: go run sql/migrations/20260611_k13_run.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(127.0.0.1:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil { log.Fatal(err) }
	defer db.Close()

	migrations := []struct{ table, col, def string }{
		// Common agent fields for recommendation/match/optimization
		{"recommendation_results", "agent_status", "VARCHAR(50) DEFAULT ''"},
		{"recommendation_results", "fallback", "TINYINT(1) DEFAULT 0"},
		{"recommendation_results", "repair_applied", "TINYINT(1) DEFAULT 0"},
		{"recommendation_results", "error_code", "VARCHAR(50) DEFAULT ''"},
		{"recommendation_results", "duration_ms", "BIGINT DEFAULT 0"},
		{"recommendation_results", "model_name", "VARCHAR(50) DEFAULT ''"},
		{"recommendation_results", "result_version", "VARCHAR(20) DEFAULT 'v1'"},

		{"match_analyses", "agent_status", "VARCHAR(50) DEFAULT ''"},
		{"match_analyses", "fallback", "TINYINT(1) DEFAULT 0"},
		{"match_analyses", "repair_applied", "TINYINT(1) DEFAULT 0"},
		{"match_analyses", "error_code", "VARCHAR(50) DEFAULT ''"},
		{"match_analyses", "duration_ms", "BIGINT DEFAULT 0"},
		{"match_analyses", "model_name", "VARCHAR(50) DEFAULT ''"},
		{"match_analyses", "result_version", "VARCHAR(20) DEFAULT 'v1'"},

		{"resume_optimization_results", "agent_status", "VARCHAR(50) DEFAULT ''"},
		{"resume_optimization_results", "fallback", "TINYINT(1) DEFAULT 0"},
		{"resume_optimization_results", "repair_applied", "TINYINT(1) DEFAULT 0"},
		{"resume_optimization_results", "error_code", "VARCHAR(50) DEFAULT ''"},
		{"resume_optimization_results", "duration_ms", "BIGINT DEFAULT 0"},
		{"resume_optimization_results", "model_name", "VARCHAR(50) DEFAULT ''"},
		{"resume_optimization_results", "result_version", "VARCHAR(20) DEFAULT 'v1'"},

		// resume_compare_results needs V2 fields
		{"resume_compare_results", "generated_by", "VARCHAR(50) DEFAULT ''"},
		{"resume_compare_results", "generation_tier", "VARCHAR(20) DEFAULT 'fast'"},
		{"resume_compare_results", "agent_status", "VARCHAR(50) DEFAULT ''"},
		{"resume_compare_results", "fallback", "TINYINT(1) DEFAULT 0"},
		{"resume_compare_results", "repair_applied", "TINYINT(1) DEFAULT 0"},
		{"resume_compare_results", "error_code", "VARCHAR(50) DEFAULT ''"},
		{"resume_compare_results", "duration_ms", "BIGINT DEFAULT 0"},
		{"resume_compare_results", "model_name", "VARCHAR(50) DEFAULT ''"},
		{"resume_compare_results", "result_version", "VARCHAR(20) DEFAULT 'v1'"},
		{"resume_compare_results", "winner_resume_id", "BIGINT UNSIGNED DEFAULT 0"},
		{"resume_compare_results", "compare_score", "INT DEFAULT 0"},
		{"resume_compare_results", "compare_level", "VARCHAR(50) DEFAULT ''"},
	}

	added := 0
	for _, m := range migrations {
		var exists int
		q := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='offer_catcher' AND TABLE_NAME='%s' AND COLUMN_NAME='%s'", m.table, m.col)
		if err := db.QueryRow(q).Scan(&exists); err != nil { log.Printf("SKIP %s.%s: %v", m.table, m.col, err); continue }
		if exists > 0 { continue }

		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", m.table, m.col, m.def)
		if _, err := db.Exec(sql); err != nil {
			log.Printf("SKIP %s.%s: %v", m.table, m.col, err)
			continue
		}
		log.Printf("ADDED %s.%s", m.table, m.col)
		added++
	}
	log.Printf("Migration complete: %d columns added", added)
}
