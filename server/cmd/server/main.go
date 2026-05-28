package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"offer-catcher/server/internal/config"
	"offer-catcher/server/internal/database"
	"offer-catcher/server/internal/health"
	httpserver "offer-catcher/server/internal/http"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/repository"
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

	// Initialize router with all dependencies
	router := httpserver.NewRouter(httpserver.RouterDeps{
		HealthService:  healthService,
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
		ResumeHandler:  resumeHandler,
		JobHandler:     jobHandler,
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("[offer-catcher-server] Starting on %s (env=%s, logLevel=%s)", addr, cfg.AppEnv, cfg.LogLevel)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[offer-catcher-server] Server failed: %v", err)
		os.Exit(1)
	}
}
