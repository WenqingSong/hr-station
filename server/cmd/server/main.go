package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"offer-catcher/server/internal/agent"
	"offer-catcher/server/internal/agent/pipeline"
	"offer-catcher/server/internal/agent/repair"
	"offer-catcher/server/internal/config"
	"offer-catcher/server/internal/database"
	"offer-catcher/server/internal/health"
	httpserver "offer-catcher/server/internal/http"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/repository"
	"offer-catcher/server/internal/schema"
	"offer-catcher/server/internal/service"
)

func main() {
	cfg := config.Load()

	// Initialize MySQL connection
	mysqlDB, err := database.NewMySQL(cfg.Database)
	if err != nil {
		log.Fatalf("[offer-catcher-server] Failed to initialize database: %v", err)
	}
	defer mysqlDB.Close()

	// Ping database with timeout to verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mysqlDB.Ping(ctx); err != nil {
		log.Fatalf("[offer-catcher-server] Failed to ping database: %v", err)
	}
	log.Println("[offer-catcher-server] MySQL connected")

	// Initialize health check
	healthService := health.NewService(cfg.App.Name, cfg.App.Env, mysqlDB)

	// Initialize auth module
	userRepo := repository.NewUserRepository(mysqlDB.DB)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	authHandler := httpserver.NewAuthHandler(authService)
	authMiddleware := middleware.Auth(authService)

	// Initialize resume module
	resumeRepo := repository.NewResumeRepository(mysqlDB.DB)
	resumeService := service.NewResumeService(resumeRepo)
	resumeHandler := httpserver.NewResumeHandler(resumeService)

	// Initialize job module
	jobRepo := repository.NewJobRepository(mysqlDB.DB)
	jobService := service.NewJobService(jobRepo)
	jobHandler := httpserver.NewJobHandler(jobService)

	// Initialize recommendation module
	recommendationRepo := repository.NewRecommendationRepository(mysqlDB.DB)
	mockAgent := agent.NewRecommendationMockAgent()
	// recommendationService is created after LLM init (needs pipelineRunner)

	// Initialize match module
	matchRepo := repository.NewMatchRepository(mysqlDB.DB)
	matchMockAgent := agent.NewMatchMockAgent()

	// Initialize LLM infrastructure for MatchAnalysisAgent pipeline
	var pipelineRunner *pipeline.Runner
	llmEnabled := cfg.LLM.Enable
	if llmEnabled {
		llmCfg := llm.Config{
			Enable:            cfg.LLM.Enable,
			Provider:          llm.Provider(cfg.LLM.Provider),
			BaseURL:           cfg.LLM.BaseURL,
			APIKey:            cfg.LLM.APIKey,
			Model:             cfg.LLM.Model,
			Temperature:       cfg.LLM.Temperature,
			MaxOutputTokens:   cfg.LLM.MaxOutputTokens,
			TimeoutSeconds:    cfg.LLM.TimeoutSeconds,
			RepairEnable:      cfg.LLM.RepairEnable,
			RepairMaxAttempts: cfg.LLM.RepairMaxAttempts,
			FallbackEnable:    cfg.LLM.FallbackEnable,
			LogRawOutput:      cfg.LLM.LogRawOutput,
			PromptDir:         cfg.LLM.PromptDir,
			SchemaDir:         cfg.LLM.SchemaDir,
		}

		if err := llmCfg.Validate(); err != nil {
			log.Fatalf("[offer-catcher-server] Invalid LLM config: %v", err)
		}

		llmClient, err := llm.NewHunyuanClient(llmCfg, nil)
		if err != nil {
			log.Fatalf("[offer-catcher-server] Failed to create LLM client: %v", err)
		}

		promptLoader := prompt.NewLoader(llmCfg.PromptDir)
		schemaLoader := schema.NewLoader(llmCfg.SchemaDir)
		validator := schema.NewValidator()
		repairAgent := repair.NewAgent(llmClient, promptLoader, schemaLoader, validator)

		pipelineRunner = pipeline.NewRunner(llmClient, promptLoader, schemaLoader, validator, repairAgent)
		log.Printf("[offer-catcher-server] LLM pipeline initialized (provider=%s model=%s)", llmCfg.Provider, llmCfg.Model)
	} else {
		log.Println("[offer-catcher-server] LLM disabled, using Mock Agent only")
	}

	matchService := service.NewMatchService(resumeRepo, jobRepo, matchRepo, matchMockAgent, pipelineRunner, llmEnabled)
	matchHandler := httpserver.NewMatchHandler(matchService)

	// Initialize recommendation service with Pipeline support (after LLM init)
	recommendationService := service.NewRecommendationService(resumeRepo, jobRepo, recommendationRepo, mockAgent, pipelineRunner, llmEnabled)
	recommendationHandler := httpserver.NewRecommendationHandler(recommendationService)

	// Initialize version compare module
	versionCompareRepo := repository.NewVersionCompareRepository(mysqlDB.DB)
	versionCompareMockAgent := agent.NewVersionCompareMockAgent()
	versionCompareService := service.NewVersionCompareService(resumeRepo, versionCompareRepo, versionCompareMockAgent)
	versionCompareHandler := httpserver.NewVersionCompareHandler(versionCompareService)

	// Initialize resume optimization module
	resumeOptimizationRepo := repository.NewResumeOptimizationRepository(mysqlDB.DB)
	resumeOptimizationMockAgent := agent.NewResumeOptimizationMockAgent()
	resumeOptimizationService := service.NewResumeOptimizationService(resumeOptimizationRepo, resumeOptimizationMockAgent)
	resumeOptimizationHandler := httpserver.NewResumeOptimizationHandler(resumeOptimizationService)

	// Initialize router with all dependencies
	router := httpserver.NewRouter(httpserver.RouterDeps{
		HealthService:              healthService,
		AuthHandler:                authHandler,
		AuthMiddleware:             authMiddleware,
		ResumeHandler:              resumeHandler,
		JobHandler:                 jobHandler,
		RecommendationHandler:      recommendationHandler,
		MatchHandler:               matchHandler,
		VersionCompareHandler:      versionCompareHandler,
		ResumeOptimizationHandler:  resumeOptimizationHandler,
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("[offer-catcher-server] Starting on %s (env=%s, logLevel=%s)", addr, cfg.AppEnv, cfg.LogLevel)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[offer-catcher-server] Server failed: %v", err)
		os.Exit(1)
	}
}
