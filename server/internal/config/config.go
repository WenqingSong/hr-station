package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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
	CORS     CORSConfig
	LLM      LLMConfig
	Redis    RedisConfig
	Session  SessionConfig
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
	Port        string
	FrontendURL string
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

// CORSConfig holds CORS (Cross-Origin Resource Sharing) configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   string
	AllowedHeaders   string
	AllowCredentials bool
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Addr     string
}

// SessionConfig holds session TTL configuration.
type SessionConfig struct {
	TTLSeconds int
}

// LLMConfig holds LLM API configuration.
type LLMConfig struct {
	Enable            bool
	Provider          string
	BaseURL           string
	APIKey            string
	Model             string // deprecated: kept for backward compatibility; prefer FastModel / QualityModel
	FastModel         string // LLM_FAST_MODEL
	QualityModel      string // LLM_QUALITY_MODEL
	DefaultTier       string // LLM_DEFAULT_TIER
	// K14-Hy3: Hy3-preview provider specific
	Hy3BaseURL        string // LLM_HY3_BASE_URL
	Hy3APIKey         string // LLM_HY3_API_KEY
	Hy3Model          string // LLM_HY3_MODEL
	Temperature       float64
	TopP              float64 // LLM_TOP_P
	MaxOutputTokens   int
	TimeoutSeconds    int
	RepairEnable      bool
	RepairMaxAttempts int
	FallbackEnable    bool
	LogRawOutput      bool
	PromptDir         string
	SchemaDir         string
	// K14-Hy3: Reasoning effort per stage
	Hy3ReasonEffortFast          string
	Hy3ReasonEffortTag           string
	Hy3ReasonEffortDeepContext   string
	Hy3ReasonEffortDeepDecision  string
	Hy3ReasonEffortDeepVerifier  string
	Hy3ReasonEffortRepair        string
}

// ResolveBaseURL returns provider-specific BaseURL.
func (c LLMConfig) ResolveBaseURL(provider string) string {
	if provider == "hy3" && c.Hy3BaseURL != "" { return c.Hy3BaseURL }
	return c.BaseURL
}

// ResolveAPIKey returns provider-specific API key.
func (c LLMConfig) ResolveAPIKey(provider string) string {
	if provider == "hy3" && c.Hy3APIKey != "" { return c.Hy3APIKey }
	return c.APIKey
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
// K13-Fix-1: Auto-load .env file from project root if it exists.
// This ensures LLM_ENABLE and other .env vars are available when started
// outside of the Makefile/scripts pipeline (e.g., direct go run).
func Load() *Config {
	// K13-Fix-1: Auto-load .env file, trying multiple relative paths.
	// The compiled binary may run from various working directories.
	// The server source is at <project>/server/ so .env is at <project>/.env.
	envPaths := []string{".env", "../.env", "../../.env", "../../../.env"}
	loaded := false
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				loaded = true
				break
			}
		}
	}
	if !loaded {
		log.Println("[config] WARNING: .env file not found, using OS environment only")
	}

	cfg := &Config{
		// Structured sub-configs
		App: AppConfig{
			Name: getEnv("APP_NAME", "offer-catcher"),
			Env:  getEnv("APP_ENV", "local"),
		},
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
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
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
			AllowedMethods:   getEnv("CORS_ALLOWED_METHODS", "GET, POST, PUT, PATCH, DELETE, OPTIONS"),
			AllowedHeaders:   getEnv("CORS_ALLOWED_HEADERS", "Content-Type, Authorization, X-Requested-With"),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "127.0.0.1"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Addr:     getEnv("REDIS_HOST", "127.0.0.1") + ":" + getEnv("REDIS_PORT", "6379"),
		},
		Session: SessionConfig{
			TTLSeconds: getEnvAsInt("REDIS_SESSION_TTL", 604800),
		},
		LLM: LLMConfig{
			Enable:            getEnvAsBool("LLM_ENABLE", false),
			Provider:          getEnv("LLM_PROVIDER", "hunyuan"),
			BaseURL:           getEnv("LLM_BASE_URL", ""),
			APIKey:            getEnv("LLM_API_KEY", ""),
			Model:             getEnv("LLM_MODEL", ""),
			FastModel:         coalesceModel(getEnv("LLM_FAST_MODEL", ""), getEnv("LLM_MODEL", "")),
			QualityModel:      coalesceModel(getEnv("LLM_QUALITY_MODEL", ""), getEnv("LLM_MODEL", "")),
			DefaultTier:       getEnv("LLM_DEFAULT_TIER", "fast"),
			// K14-Hy3
			Hy3BaseURL:        getEnv("LLM_HY3_BASE_URL", "https://api.lkeap.cloud.tencent.com/plan/v3/chat/completions"),
			Hy3APIKey:         getEnv("LLM_HY3_API_KEY", ""),
			Hy3Model:          getEnv("LLM_HY3_MODEL", "hy3-preview"),
			Temperature:       getEnvAsFloat64("LLM_TEMPERATURE", 0.2),
			TopP:              getEnvAsFloat64("LLM_TOP_P", 1.0),
			MaxOutputTokens:   getEnvAsInt("LLM_MAX_OUTPUT_TOKENS", 4096),
			TimeoutSeconds:    getEnvAsInt("LLM_TIMEOUT_SECONDS", 120),
			RepairEnable:      getEnvAsBool("LLM_REPAIR_ENABLE", true),
			RepairMaxAttempts: getEnvAsInt("LLM_REPAIR_MAX_ATTEMPTS", 1),
			FallbackEnable:    getEnvAsBool("LLM_FALLBACK_ENABLE", true),
			LogRawOutput:      getEnvAsBool("LLM_LOG_RAW_OUTPUT", false),
			PromptDir:         getEnv("LLM_PROMPT_DIR", "server/prompts"),
			SchemaDir:         getEnv("LLM_SCHEMA_DIR", "server/prompts/schema"),
			// K14-Hy3 reasoning effort defaults
			Hy3ReasonEffortFast:         getEnv("HY3_REASONING_EFFORT_FAST", "no_think"),
			Hy3ReasonEffortTag:          getEnv("HY3_REASONING_EFFORT_TAG", "no_think"),
			Hy3ReasonEffortDeepContext:  getEnv("HY3_REASONING_EFFORT_DEEP_CONTEXT", "low"),
			Hy3ReasonEffortDeepDecision: getEnv("HY3_REASONING_EFFORT_DEEP_DECISION", "high"),
			Hy3ReasonEffortDeepVerifier: getEnv("HY3_REASONING_EFFORT_DEEP_VERIFIER", "low"),
			Hy3ReasonEffortRepair:       getEnv("HY3_REASONING_EFFORT_REPAIR", "no_think"),
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

// Validate checks for configuration issues and warns on insecure defaults.
func (c *Config) Validate() {
	weakJWTSecrets := []string{"change-me-in-local", "change-me", "secret", "your-secret", ""}
	for _, weak := range weakJWTSecrets {
		if c.JWT.Secret == weak {
			log.Printf("[warn] JWT_SECRET is set to a weak default (%q). Please set a strong secret in .env for non-local environments.", c.JWT.Secret)
			break
		}
	}
	if c.Database.Password == "password" {
		log.Printf("[warn] DB_PASSWORD is set to the default 'password'. Consider changing it for security.")
	}
	if c.JWT.Secret == "" {
		fmt.Fprintln(os.Stderr, "WARNING: JWT_SECRET is empty. Authentication will be insecure.")
	}
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

// getEnvAsFloat64 reads an environment variable as float64 or returns a default value.
func getEnvAsFloat64(key string, defaultValue float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

// getEnvAsCSV reads an environment variable as comma-separated values or returns a default slice.
// It trims whitespace and ignores empty items.
func getEnvAsCSV(key string, defaultValues []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValues
	}
	var result []string
	for _, item := range strings.Split(val, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) == 0 {
		return defaultValues
	}
	return result
}

// coalesceModel returns the tier-specific model if set, otherwise falls back to the legacy model.
func coalesceModel(tier, legacy string) string {
	if tier != "" {
		return tier
	}
	return legacy
}
