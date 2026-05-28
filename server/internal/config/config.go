package config

import "os"

// Config holds application configuration.
type Config struct {
	AppEnv     string
	LogLevel   string
	ServerPort string
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		AppEnv:     envOrDefault("APP_ENV", "local"),
		LogLevel:   envOrDefault("LOG_LEVEL", "debug"),
		ServerPort: envOrDefault("SERVER_PORT", "8080"),
	}
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
