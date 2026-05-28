package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration.
// Top-level fields (ServerPort, AppEnv, LogLevel) are kept for backward compatibility
// with existing code (e.g., main.go). Structured sub-configs provide grouped access.
type Config struct {
	// Backward-compatible fields (used by main.go)
	ServerPort string
	AppEnv     string
	LogLevel   string

	// Structured sub-configs
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	LLM      LLMConfig
	Feature  FeatureConfig
	Log      LogConfig
}

// AppConfig holds application identity configuration.
type AppConfig struct {
	Name string
	Env  string
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds MySQL database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	DSN      string
}

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret      string
	ExpireHours int
}

// LLMConfig holds LLM API configuration.
type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// FeatureConfig holds feature flag configuration.
type FeatureConfig struct {
	EnableMockAgent bool
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	cfg := &Config{
		// Structured sub-configs
		App: AppConfig{
			Name: getEnv("APP_NAME", "offer-catcher"),
			Env:  getEnv("APP_ENV", "local"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			Name:     getEnv("DB_NAME", "offer_catcher"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
			DSN:      getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/offer_catcher?charset=utf8mb4&parseTime=True&loc=Local"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "change-me-in-local"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 168),
		},
		LLM: LLMConfig{
			BaseURL: getEnv("LLM_BASE_URL", ""),
			APIKey:  getEnv("LLM_API_KEY", ""),
			Model:   getEnv("LLM_MODEL", ""),
		},
		Feature: FeatureConfig{
			EnableMockAgent: getEnvAsBool("ENABLE_MOCK_AGENT", true),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
		},
	}

	// Backward-compatible top-level fields
	cfg.ServerPort = cfg.Server.Port
	cfg.AppEnv = cfg.App.Env
	cfg.LogLevel = cfg.Log.Level

	return cfg
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as int or returns a default value.
func getEnvAsInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return n
}

// getEnvAsBool reads an environment variable as bool or returns a default value.
// Supports "true"/"false" string values (case-insensitive).
func getEnvAsBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return b
}
