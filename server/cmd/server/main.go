package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"offer-catcher/server/internal/config"
	"offer-catcher/server/internal/http"
)

func main() {
	cfg := config.Load()

	router := httpserver.NewRouter()

	addr := fmt.Sprintf(":%s", cfg.ServerPort)

	log.Printf("[offer-catcher-server] Starting on %s (env=%s, logLevel=%s)", addr, cfg.AppEnv, cfg.LogLevel)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("[offer-catcher-server] Server failed: %v", err)
		os.Exit(1)
	}
}
