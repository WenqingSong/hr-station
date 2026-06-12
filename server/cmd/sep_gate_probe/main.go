// sep_gate_probe is a local development tool that triggers SEP semantic profile
// generation through the SemanticProfileGate. It connects only to the development DB
// and does NOT expose an HTTP endpoint, modify old profile tables, or print API keys.
//
// Usage:
//
//	go run ./server/cmd/sep_gate_probe --resume-id=8 --job-id=10001
//	go run ./server/cmd/sep_gate_probe --resume-id=8
//	go run ./server/cmd/sep_gate_probe --job-id=10001
//
// Safety:
//   - Never prints API key.
//   - Never prints raw resume text or full JD.
//   - Only writes to semantic_profiles / semantic_profile_tags.
//   - Does not modify old resume_profiles / job_profiles.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"offer-catcher/server/internal/config"
	"offer-catcher/server/internal/database"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/repository"
	"offer-catcher/server/internal/service"
)

func main() {
	resumeID := flag.Int64("resume-id", 0, "Resume ID to profile")
	jobID := flag.Int64("job-id", 0, "Job ID to profile")
	skipResume := flag.Bool("skip-resume", false, "Skip resume profiling")
	skipJob := flag.Bool("skip-job", false, "Skip job profiling")
	flag.Parse()

	if *resumeID == 0 && *jobID == 0 {
		fmt.Println("Usage: sep_gate_probe --resume-id=N [--job-id=N] [--skip-resume] [--skip-job]")
		os.Exit(1)
	}

	fmt.Println("=== SEP Gate Probe ===")
	fmt.Printf("resume-id=%d job-id=%d skip-resume=%v skip-job=%v\n\n", *resumeID, *jobID, *skipResume, *skipJob)

	// Load config from .env
	cfg := config.Load()
	cfg.Validate()
	fmt.Printf("config: provider=%s quality_model=%s timeout=%ds\n\n",
		string(cfg.LLM.Provider), cfg.LLM.QualityModel, cfg.LLM.TimeoutSeconds)

	// Connect to dev DB
	mysqlDB, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("DB connect: %v", err)
	}
	defer mysqlDB.Close()
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := mysqlDB.Ping(pingCtx); err != nil {
		log.Fatalf("DB ping: %v", err)
	}
	fmt.Println("DB connected OK")

	// Long-lived context for LLM calls
	ctx := context.Background()

	// Initialize LLM client
	provider := string(cfg.LLM.Provider)
	qualityCfg := llm.Config{
		Enable:            true,
		Provider:          llm.Provider(provider),
		BaseURL:           cfg.LLM.ResolveBaseURL(provider),
		APIKey:            cfg.LLM.ResolveAPIKey(provider),
		Model:             cfg.LLM.QualityModel,
		Temperature:       0.3,
		MaxOutputTokens:   4096,
		TimeoutSeconds:    180,
		RepairEnable:      true,
		RepairMaxAttempts: 1,
		FallbackEnable:    false,
		PromptDir:         cfg.LLM.PromptDir,
		SchemaDir:         cfg.LLM.SchemaDir,
	}
	qualityClient, err := newLLMClient(qualityCfg, provider)
	if err != nil {
		log.Fatalf("LLM client: %v", err)
	}
	fmt.Printf("LLM client: provider=%s model=%s\n\n", provider, qualityCfg.Model)

	// Build service stack
	db := mysqlDB.DB
	resumeRepo := repository.NewResumeRepository(db)
	jobRepo := repository.NewJobRepository(db)
	semanticRepo := repository.NewSemanticProfileRepository(db)
	semanticService := service.NewSemanticProfileService(db, semanticRepo, resumeRepo, jobRepo)
	profiler := service.NewSemanticProfileProfiler(semanticService, resumeRepo, jobRepo, qualityClient, qualityCfg.Model)
	gate := service.NewSemanticProfileGate(semanticService, profiler)

	var allResults []*service.SemanticProfileGateResult

	if !*skipResume && *resumeID > 0 {
		fmt.Printf("--- Resume %d ---\n", *resumeID)
		start := time.Now()
		r, err := gate.EnsureResumeProfile(ctx, *resumeID)
		elapsed := time.Since(start).Milliseconds()
		r.DurationMs = elapsed
		allResults = append(allResults, r)
		if err != nil {
			fmt.Printf("  ERROR: %v (elapsed=%dms)\n\n", err, elapsed)
		} else {
			printResult(r)
		}

		// Second call — verify reuse
		fmt.Printf("--- Resume %d (verify reuse) ---\n", *resumeID)
		start = time.Now()
		r2, err2 := gate.EnsureResumeProfile(ctx, *resumeID)
		elapsed2 := time.Since(start).Milliseconds()
		r2.DurationMs = elapsed2
		if err2 != nil {
			fmt.Printf("  ERROR: %v (elapsed=%dms)\n\n", err2, elapsed2)
		} else {
			fmt.Printf("  reused=%v generated=%v profile_id=%d status=%s elapsed=%dms\n\n",
				r2.Reused, r2.Generated, r2.ProfileID, r2.Status, elapsed2)
		}
	}

	if !*skipJob && *jobID > 0 {
		fmt.Printf("--- Job %d ---\n", *jobID)
		start := time.Now()
		r, err := gate.EnsureJobProfile(ctx, *jobID)
		elapsed := time.Since(start).Milliseconds()
		r.DurationMs = elapsed
		allResults = append(allResults, r)
		if err != nil {
			fmt.Printf("  ERROR: %v (elapsed=%dms)\n\n", err, elapsed)
		} else {
			printResult(r)
		}

		// Second call — verify reuse
		fmt.Printf("--- Job %d (verify reuse) ---\n", *jobID)
		start = time.Now()
		r2, err2 := gate.EnsureJobProfile(ctx, *jobID)
		elapsed2 := time.Since(start).Milliseconds()
		r2.DurationMs = elapsed2
		if err2 != nil {
			fmt.Printf("  ERROR: %v (elapsed=%dms)\n\n", err2, elapsed2)
		} else {
			fmt.Printf("  reused=%v generated=%v profile_id=%d status=%s elapsed=%dms\n\n",
				r2.Reused, r2.Generated, r2.ProfileID, r2.Status, elapsed2)
		}
	}

	// Summary
	fmt.Println("=== Summary ===")
	for _, r := range allResults {
		fmt.Printf("  %s %d: profile_id=%d status=%s tags=%d reused=%v generated=%v duration=%dms\n",
			r.TargetType, r.TargetID, r.ProfileID, r.Status, r.TagsCount, r.Reused, r.Generated, r.DurationMs)
		if r.ErrorMessage != "" {
			fmt.Printf("    error: %s\n", r.ErrorMessage)
		}
	}
}

func printResult(r *service.SemanticProfileGateResult) {
	fmt.Printf("  target_type=%s target_id=%d\n", r.TargetType, r.TargetID)
	fmt.Printf("  content_hash=%s...\n", truncate(r.ContentHash, 16))
	fmt.Printf("  profile_id=%d status=%s\n", r.ProfileID, r.Status)
	fmt.Printf("  tags_count=%d reused=%v generated=%v elapsed=%dms\n", r.TagsCount, r.Reused, r.Generated, r.DurationMs)
	if r.Summary != "" {
		fmt.Printf("  summary=%s\n", r.Summary)
	}
	if r.ErrorMessage != "" {
		fmt.Printf("  error=%s\n", r.ErrorMessage)
	}
	fmt.Println()
}

func newLLMClient(cfg llm.Config, provider string) (llm.Client, error) {
	switch llm.Provider(provider) {
	case llm.ProviderHy3:
		return llm.NewHy3Client(cfg, nil)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
