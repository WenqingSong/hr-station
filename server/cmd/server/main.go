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

	// K14-Hy3: Provider factory — create LLM client based on LLM_PROVIDER config
	var fastRunner, qualityRunner *pipeline.Runner
	var sepSemanticGate *service.SemanticProfileGate
	var sepSemanticRepo *repository.SemanticProfileRepository
	var qualityClient llm.Client
	var qualityModelName string
	llmEnabled := cfg.LLM.Enable
	if llmEnabled {
		provider := cfg.LLM.Provider
		maxTok := cfg.LLM.MaxOutputTokens
		if maxTok <= 0 { maxTok = 4096 }
		temp := cfg.LLM.Temperature
		if temp <= 0 { temp = 0.2 }

		// Fast client
		fastCfg := llm.Config{
			Enable:            cfg.LLM.Enable,
			Provider:          llm.Provider(provider),
			BaseURL:           cfg.LLM.ResolveBaseURL(provider),
			APIKey:            cfg.LLM.ResolveAPIKey(provider),
			Model:             cfg.LLM.FastModel,
			Temperature:       temp,
			MaxOutputTokens:   maxTok,
			TimeoutSeconds:    cfg.LLM.TimeoutSeconds,
			RepairEnable:      cfg.LLM.RepairEnable,
			RepairMaxAttempts: cfg.LLM.RepairMaxAttempts,
			FallbackEnable:    cfg.LLM.FallbackEnable,
			PromptDir:         cfg.LLM.PromptDir,
			SchemaDir:         cfg.LLM.SchemaDir,
		}
		fastClient := newLLMClient(fastCfg, provider)

		// Quality client
		qualityCfg := fastCfg
		qualityCfg.Model = cfg.LLM.QualityModel
		qualityClient = newLLMClient(qualityCfg, provider)
		qualityModelName = qualityCfg.Model

		promptLoader := prompt.NewLoader(fastCfg.PromptDir)
		schemaLoader := schema.NewLoader(fastCfg.SchemaDir)
		validator := schema.NewValidator()
		repairAgent := repair.NewAgent(qualityClient, promptLoader, schemaLoader, validator)

		repairReEffort := cfg.LLM.Hy3ReasonEffortRepair

		fastRunner = pipeline.NewRunner(fastClient, promptLoader, schemaLoader, validator, repairAgent, maxTok, temp, repairReEffort)
		qualityRunner = pipeline.NewRunner(qualityClient, promptLoader, schemaLoader, validator, repairAgent, maxTok, temp, repairReEffort)

		log.Printf("[offer-catcher-server] LLM pipeline initialized (provider=%s fast=%s quality=%s timeout_seconds=%d max_tokens=%d)",
			provider, cfg.LLM.FastModel, cfg.LLM.QualityModel, cfg.LLM.TimeoutSeconds, maxTok)

		// SEP Semantic Profile Gate for Fast recommendation
		sepSemanticRepo = repository.NewSemanticProfileRepository(mysqlDB.DB)
		sepSemanticSvc := service.NewSemanticProfileService(mysqlDB.DB, sepSemanticRepo, resumeRepo, jobRepo)
		sepProfiler := service.NewSemanticProfileProfiler(sepSemanticSvc, resumeRepo, jobRepo, qualityClient, qualityModelName)
		sepSemanticGate = service.NewSemanticProfileGate(sepSemanticSvc, sepProfiler)
		log.Printf("[offer-catcher-server] SEP semantic profile gate enabled (model=%s)", qualityModelName)
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

	// K13-Fix-A: Initialize TagEvidence module
	jobProfileRepo := repository.NewJobProfileRepository(mysqlDB.DB)
	tagEvidenceService := service.NewTagEvidenceService(resumeRepo, resumeProfileRepo, jobProfileRepo)
	tagEvidenceService.SetJobRepo(jobRepo)
	// Backfill job profiles in background
	go func() {
		if err := tagEvidenceService.EnsureAllOpenJobTagProfiles(context.Background(), 100); err != nil {
			log.Printf("[TagEvidence] backfill jobs: %v", err)
		} else {
			count, _ := jobProfileRepo.Count(context.Background())
			log.Printf("[TagEvidence] job profiles ready: %d", count)
		}
	}()

	// K13-Fix-A: Backfill resume profiles with empty evidence
	go func() {
		pairs, err := resumeRepo.ListAllIDPairs(context.Background())
		if err != nil {
			log.Printf("[TagEvidence] backfill resumes: list error: %v", err)
			return
		}
		refreshed := 0
		for _, p := range pairs {
			profile, _ := resumeProfileRepo.FindByResumeIDAndType(context.Background(), p.ID, "fast")
			if profile == nil || profile.EvidenceJSON == "" || profile.EvidenceJSON == "{}" || profile.EvidenceJSON == "null" {
				if _, err := tagEvidenceService.EnsureResumeTagProfile(context.Background(), p.ID, p.UserID); err != nil {
					log.Printf("[TagEvidence] backfill resume %d: %v", p.ID, err)
				} else {
					refreshed++
				}
			}
		}
		log.Printf("[TagEvidence] resume profiles backfilled: %d", refreshed)
	}()

	// K13-Fix-1: Read fallback/mock flags from config
	fallbackEnabled := cfg.LLM.FallbackEnable
	mockAgentEnabled := cfg.Feature.EnableMockAgent
	log.Printf("[offer-catcher-server] fallback_enable=%v enable_mock_agent=%v", fallbackEnabled, mockAgentEnabled)

	matchService := service.NewMatchService(resumeRepo, jobRepo, matchRepo, matchMockAgent, fastRunner, qualityRunner, llmEnabled, fallbackEnabled, mockAgentEnabled, tagEvidenceService, sepSemanticRepo, qualityClient)
	matchHandler := httpserver.NewMatchHandler(matchService)

	// K13-Fix-B: RecommendationService with TagEvidenceService for Fast tier
	recommendationService := service.NewRecommendationService(resumeRepo, jobRepo, recommendationRepo, mockAgent, fastRunner, qualityRunner, llmEnabled, fallbackEnabled, mockAgentEnabled, resumeProfileService, tagEvidenceService, sepSemanticGate, sepSemanticRepo, qualityClient, qualityModelName)
	recommendationHandler := httpserver.NewRecommendationHandler(recommendationService)

	// Initialize version compare module with both runners (tier selected per request)
	versionCompareRepo := repository.NewResumeCompareRepository(mysqlDB.DB)
	versionCompareMockAgent := agent.NewVersionCompareMockAgent()
	versionCompareMockAgent.SetProfileProvider(resumeProfileService)
	versionCompareService := service.NewVersionCompareService(resumeRepo, versionCompareRepo, jobRepo, versionCompareMockAgent, fastRunner, qualityRunner, llmEnabled, fallbackEnabled, mockAgentEnabled, resumeProfileService, tagEvidenceService)
	versionCompareHandler := httpserver.NewVersionCompareHandler(versionCompareService)

	// Initialize resume optimization module with both runners (tier selected per request)
	resumeOptimizationRepo := repository.NewResumeOptimizationRepository(mysqlDB.DB)
	resumeOptimizationMockAgentV2 := agent.NewResumeOptimizationMockAgentV2()
	resumeOptimizationMockAgentV2.SetProfileProvider(resumeProfileService)
	resumeOptimizationMockAgentV2.SetMatchLatestProvider(func(ctx context.Context, userID, resumeID, jobID uint64) *model.MatchAnalysis {
		ma, _ := matchRepo.FindLatestByResumeAndJob(ctx, userID, resumeID, jobID)
		return ma
	})
	resumeOptimizationService := service.NewResumeOptimizationService(resumeRepo, jobRepo, resumeOptimizationRepo, resumeOptimizationMockAgentV2, fastRunner, qualityRunner, llmEnabled, fallbackEnabled, mockAgentEnabled, tagEvidenceService)
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

// newLLMClient creates an LLM client based on provider string.
// K14-Hy3: supports hy3 (LKEAP) and deepseek/hunyuan/tokenhub (OpenAI-compatible).
func newLLMClient(cfg llm.Config, provider string) llm.Client {
	switch provider {
	case "hy3":
		c, err := llm.NewHy3Client(cfg, nil)
		if err != nil {
			log.Fatalf("[offer-catcher-server] Failed to create Hy3 client: %v", err)
		}
		return c
	default:
		// deepseek, hunyuan, tokenhub all use OpenAI-compatible path
		c, err := llm.NewHunyuanClient(cfg, nil)
		if err != nil {
			log.Fatalf("[offer-catcher-server] Failed to create LLM client: %v", err)
		}
		return c
	}
}
