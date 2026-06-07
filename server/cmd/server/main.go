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
	"offer-catcher/server/internal/model"
	httpserver "offer-catcher/server/internal/http"
	"offer-catcher/server/internal/llm"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/prompt"
	"offer-catcher/server/internal/repository"
	"offer-catcher/server/internal/schema"
	"offer-catcher/server/internal/service"
	"offer-catcher/server/internal/session"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	cfg.Validate()

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

	// Initialize Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[warn] Redis connection failed: %v. Session validation will not work.", err)
	} else {
		log.Println("[offer-catcher-server] Redis connected")
	}
	defer redisClient.Close()

	// Initialize session manager
	sessionMgr := session.NewManager(redisClient, cfg.Session.TTLSeconds)

	// Initialize auth module
	userRepo := repository.NewUserRepository(mysqlDB.DB)
	authService := service.NewAuthService(userRepo, sessionMgr, cfg.JWT.Secret, cfg.JWT.ExpireHours)
	authHandler := httpserver.NewAuthHandler(authService)
	authMiddleware := middleware.Auth(authService)

	// Initialize resume module
	resumeRepo := repository.NewResumeRepository(mysqlDB.DB)
	resumeService := service.NewResumeService(resumeRepo)

	// Initialize resume profile module (K9-1-H4-B)
	resumeProfileRepo := repository.NewResumeProfileRepository(mysqlDB.DB)
	resumeProfileService := service.NewResumeProfileService(resumeRepo, resumeProfileRepo)
	resumeService.SetProfileService(resumeProfileService)
	resumeHandler := httpserver.NewResumeHandler(resumeService, resumeProfileService)

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
	matchMockAgent.SetProfileProvider(resumeProfileService)

	// Initialize LLM infrastructure for MatchAnalysisAgent pipeline
	var fastRunner, qualityRunner *pipeline.Runner
	llmEnabled := cfg.LLM.Enable
	if llmEnabled {
		baseCfg := llm.Config{
			Enable:            cfg.LLM.Enable,
			Provider:          llm.Provider(cfg.LLM.Provider),
			BaseURL:           cfg.LLM.BaseURL,
			APIKey:            cfg.LLM.APIKey,
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

		// Fast client (deepseek-v4-flash)
		fastCfg := baseCfg
		fastCfg.Model = cfg.LLM.FastModel
		if err := fastCfg.Validate(); err != nil {
			log.Fatalf("[offer-catcher-server] Invalid Fast LLM config: %v", err)
		}
		fastClient, err := llm.NewHunyuanClient(fastCfg, nil)
		if err != nil {
			log.Fatalf("[offer-catcher-server] Failed to create Fast LLM client: %v", err)
		}

		// Quality client (deepseek-v4-pro-202606)
		qualityCfg := baseCfg
		qualityCfg.Model = cfg.LLM.QualityModel
		if err := qualityCfg.Validate(); err != nil {
			log.Fatalf("[offer-catcher-server] Invalid Quality LLM config: %v", err)
		}
		qualityClient, err := llm.NewHunyuanClient(qualityCfg, nil)
		if err != nil {
			log.Fatalf("[offer-catcher-server] Failed to create Quality LLM client: %v", err)
		}

		promptLoader := prompt.NewLoader(baseCfg.PromptDir)
		schemaLoader := schema.NewLoader(baseCfg.SchemaDir)
		validator := schema.NewValidator()

		// Repair agent for Quality tier (repair is quality-sensitive; fast tier uses Mock fallback on failure)
		repairAgent := repair.NewAgent(qualityClient, promptLoader, schemaLoader, validator)

		fastRunner = pipeline.NewRunner(fastClient, promptLoader, schemaLoader, validator, repairAgent)
		qualityRunner = pipeline.NewRunner(qualityClient, promptLoader, schemaLoader, validator, repairAgent)

		log.Printf("[offer-catcher-server] LLM pipeline initialized (fast=%s quality=%s)",
			cfg.LLM.FastModel, cfg.LLM.QualityModel)
	} else {
		log.Println("[offer-catcher-server] LLM disabled, using Mock Agent only")
	}

	// K9-1-H5: wire ProfileAgent for LLM-based quality profile extraction
	if llmEnabled && qualityRunner != nil {
		profileAgent := agent.NewProfileAgent(qualityRunner)
		resumeProfileService.SetProfileAgent(profileAgent)
		log.Println("[offer-catcher-server] ProfileAgent wired for LLM quality profile extraction")
	} else {
		log.Println("[offer-catcher-server] ProfileAgent disabled (LLM off), quality profiles use rule fallback")
	}

	matchService := service.NewMatchService(resumeRepo, jobRepo, matchRepo, matchMockAgent, fastRunner, qualityRunner, llmEnabled)
	matchHandler := httpserver.NewMatchHandler(matchService)

	// Initialize recommendation service with both runners (tier selected per request)
	recommendationService := service.NewRecommendationService(resumeRepo, jobRepo, recommendationRepo, mockAgent, fastRunner, qualityRunner, llmEnabled, resumeProfileService)
	recommendationHandler := httpserver.NewRecommendationHandler(recommendationService)

	// Initialize version compare module with both runners (tier selected per request)
	versionCompareRepo := repository.NewResumeCompareRepository(mysqlDB.DB)
	versionCompareMockAgent := agent.NewVersionCompareMockAgent()
	versionCompareService := service.NewVersionCompareService(resumeRepo, versionCompareRepo, versionCompareMockAgent, fastRunner, qualityRunner, llmEnabled)
	versionCompareHandler := httpserver.NewVersionCompareHandler(versionCompareService)

	// Initialize resume optimization module with both runners (tier selected per request)
	resumeOptimizationRepo := repository.NewResumeOptimizationRepository(mysqlDB.DB)
	resumeOptimizationMockAgentV2 := agent.NewResumeOptimizationMockAgentV2()
	resumeOptimizationMockAgentV2.SetProfileProvider(resumeProfileService)
	resumeOptimizationMockAgentV2.SetMatchLatestProvider(func(ctx context.Context, userID, resumeID, jobID uint64) *model.MatchAnalysis {
		ma, _ := matchRepo.FindLatestByResumeAndJob(ctx, userID, resumeID, jobID)
		return ma
	})
	resumeOptimizationService := service.NewResumeOptimizationService(resumeRepo, jobRepo, resumeOptimizationRepo, resumeOptimizationMockAgentV2, fastRunner, qualityRunner, llmEnabled)
	resumeOptimizationHandler := httpserver.NewResumeOptimizationHandler(resumeOptimizationService)
	resumeHandler.SetOptimizationV2Handler(http.HandlerFunc(resumeOptimizationHandler.OptimizeV2))
	resumeHandler.SetOptimizationLatestHandler(http.HandlerFunc(resumeOptimizationHandler.LatestV2))

	// Initialize CORS middleware with config
	middleware.InitCORS(cfg.CORS)

	// Initialize router with all dependencies
	router := httpserver.NewRouter(httpserver.RouterDeps{
		HealthService:              healthService,
		AuthHandler:                authHandler,
		AuthMiddleware:             authMiddleware,
		AuthService:                authService,
		ResumeHandler:              resumeHandler,
		JobHandler:                 jobHandler,
		RecommendationHandler:      recommendationHandler,
		MatchHandler:               matchHandler,
		VersionCompareHandler:      versionCompareHandler,
		ResumeOptimizationHandler:  resumeOptimizationHandler,
		FrontendURL:                cfg.Server.FrontendURL,
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("[offer-catcher-server] Starting on %s (env=%s, logLevel=%s)", addr, cfg.AppEnv, cfg.LogLevel)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[offer-catcher-server] Server failed: %v", err)
		os.Exit(1)
	}
}
