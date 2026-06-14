// sep-backfill scans and generates missing SEP semantic profiles for
// resumes and jobs.  It uses the same DB + LLM stack as the main server
// (config.Load) but runs as a standalone operational tool.
//
// Usage:
//
//	# Dry-run — show what profiles are missing/stale without generating anything.
//	go run ./server/cmd/sep-backfill --target=all --dry-run
//
//	# Generate missing resume profiles only (concurrency=1, safe).
//	go run ./server/cmd/sep-backfill --target=resume --concurrency=2
//
//	# Generate all missing profiles (resume + job).
//	go run ./server/cmd/sep-backfill --target=all --concurrency=1
//
// Safety:
//   - Never prints API key.
//   - Never prints raw resume text or full JD.
//   - Only writes to semantic_profiles / semantic_profile_tags.
//   - Does not modify old resume_profiles / job_profiles.
//   - Idempotent: ready profiles are always skipped.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"offer-catcher/server/internal/config"
	"offer-catcher/server/internal/database"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/repository"
	"offer-catcher/server/internal/service"
)

func main() {
	target := flag.String("target", "all", "Target type: resume, job, all")
	targetID := flag.Int64("target-id", 0, "Only process a specific resume/job ID (0=all)")
	dryRun := flag.Bool("dry-run", false, "Only scan, do not generate")
	limit := flag.Int("limit", 0, "Max profiles to generate per type (0=unlimited)")
	concurrency := flag.Int("concurrency", 1, "Max concurrent generations")
	rebuildStale := flag.Bool("rebuild-stale", false, "Also rebuild stale profiles")
	flag.Parse()

	fmt.Println("=== SEP Backfill ===")
	fmt.Printf("target=%s target-id=%d dry-run=%v limit=%d concurrency=%d rebuild-stale=%v\n\n", *target, *targetID, *dryRun, *limit, *concurrency, *rebuildStale)

	// Load config from .env
	cfg := config.Load()
	cfg.Validate()
	fmt.Printf("config: provider=%s model=%s\n\n", string(cfg.LLM.Provider), cfg.LLM.QualityModel)

	// Connect to dev DB
	mysqlDB, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("DB connect: %v", err)
	}
	defer mysqlDB.Close()
	pgCtx, pgCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pgCancel()
	if err := mysqlDB.Ping(pgCtx); err != nil {
		log.Fatalf("DB ping: %v", err)
	}
	fmt.Println("DB connected OK")

	db := mysqlDB.DB
	resumeRepo := repository.NewResumeRepository(db)
	jobRepo := repository.NewJobRepository(db)
	semanticRepo := repository.NewSemanticProfileRepository(db)
	semanticService := service.NewSemanticProfileService(db, semanticRepo, resumeRepo, jobRepo)

	// Build LLM client (only needed for execute mode)
	provider := string(cfg.LLM.Provider)
	llmCfg := llm.Config{
		Enable:            true,
		Provider:          llm.Provider(provider),
		BaseURL:           cfg.LLM.ResolveBaseURL(provider),
		APIKey:            cfg.LLM.ResolveAPIKey(provider),
		Model:             cfg.LLM.QualityModel,
		Temperature:       0.3,
		MaxOutputTokens:   4096,
		TimeoutSeconds:    300,
		RepairEnable:      true,
		RepairMaxAttempts: 1,
		FallbackEnable:    false,
		PromptDir:         cfg.LLM.PromptDir,
		SchemaDir:         cfg.LLM.SchemaDir,
	}
	qualityClient, err := newLLMClient(llmCfg, provider)
	if err != nil {
		log.Fatalf("LLM client: %v", err)
	}
	fmt.Printf("LLM client: provider=%s model=%s\n\n", provider, llmCfg.Model)

	profiler := service.NewSemanticProfileProfiler(semanticService, resumeRepo, jobRepo, qualityClient, llmCfg.Model)
	gate := service.NewSemanticProfileGate(semanticService, profiler)
	backfill := service.NewSemanticProfileBackfillService(gate, semanticRepo, resumeRepo, jobRepo)

	ctx := context.Background()

	if *dryRun {
		fmt.Println("=== DRY-RUN: scanning only ===")
		runScan(ctx, backfill, *target, *targetID)
		fmt.Println("\nNo profiles were generated.")
		return
	}

	if *target == "all" || *target == "job" {
		fmt.Println("=== Scanning jobs ===")
		jobItems, err := backfill.Scan(ctx, "job")
		if err != nil {
			log.Fatalf("scan jobs: %v", err)
		}
		jobItems = filterByTargetID(jobItems, *targetID)
		printScanSummary("job", jobItems)
	}

	if *target == "all" || *target == "resume" {
		fmt.Println("=== Scanning resumes ===")
		resumeItems, err := backfill.Scan(ctx, "resume")
		if err != nil {
			log.Fatalf("scan resumes: %v", err)
		}
		resumeItems = filterByTargetID(resumeItems, *targetID)
		printScanSummary("resume", resumeItems)
	}

	fmt.Println("\n=== Generating missing profiles ===")
	opts := service.BackfillOptions{
		TargetType:   *target,
		TargetID:     *targetID,
		DryRun:       false,
		Limit:        *limit,
		Concurrency:  *concurrency,
		RebuildStale: *rebuildStale,
	}
	result, err := backfill.BackfillMissing(ctx, opts)
	if err != nil {
		log.Fatalf("backfill: %v", err)
	}

	fmt.Println()
	printResult(result)
}

func runScan(ctx context.Context, bf *service.SemanticProfileBackfillService, target string, targetID int64) {
	targets := []string{target}
	if target == "all" {
		targets = []string{"resume", "job"}
	}
	for _, t := range targets {
		items, err := bf.Scan(ctx, t)
		if err != nil {
			log.Printf("scan %s: %v", t, err)
			continue
		}
		items = filterByTargetID(items, targetID)
		printScanSummary(t, items)
	}
}

func filterByTargetID(items []service.BackfillScanItem, targetID int64) []service.BackfillScanItem {
	if targetID <= 0 {
		return items
	}
	var filtered []service.BackfillScanItem
	for _, it := range items {
		if it.TargetID == targetID {
			filtered = append(filtered, it)
		}
	}
	return filtered
}

func printScanSummary(targetType string, items []service.BackfillScanItem) {
	ready, missing, stale, failed := 0, 0, 0, 0
	for _, it := range items {
		switch it.Status {
		case "ready":
			ready++
		case "missing":
			missing++
		case "stale":
			stale++
		case "failed":
			failed++
		}
	}
	fmt.Printf("  %s: total=%d ready=%d missing=%d stale=%d failed=%d\n",
		targetType, len(items), ready, missing, stale, failed)
	for _, it := range items {
		if it.Status != "ready" {
			fmt.Printf("    %s %d (%s): %s\n", it.TargetType, it.TargetID, it.TargetName, it.Status)
		}
	}
}

func printResult(r *service.BackfillResult) {
	fmt.Println("=== Backfill Result ===")
	fmt.Printf("  target=%s total=%d\n", r.TargetType, r.Total)
	fmt.Printf("  ready=%d missing=%d stale=%d failed=%d\n", r.Ready, r.Missing, r.Stale, r.Failed)
	fmt.Printf("  scheduled=%d generated=%d gen_failed=%d skipped=%d\n", r.Scheduled, r.Generated, r.GenFailed, r.Skipped)
	if r.Duration != "" {
		fmt.Printf("  duration=%s\n", r.Duration)
	}
}

func newLLMClient(cfg llm.Config, provider string) (llm.Client, error) {
	switch llm.Provider(provider) {
	case llm.ProviderHy3:
		return llm.NewHy3Client(cfg, nil)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
