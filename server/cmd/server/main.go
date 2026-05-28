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
	"offer-catcher/server/internal/http"
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

	router := httpserver.NewRouter()

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("[offer-catcher-server] Starting on %s (env=%s, logLevel=%s)", addr, cfg.AppEnv, cfg.LogLevel)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[offer-catcher-server] Server failed: %v", err)
		os.Exit(1)
	}
}
