package llm

import (
	"fmt"
	"strings"
)

// Config holds the LLM configuration. It mirrors the environment
// variables defined in .env.example (4-2.1) and provides validation.
//
// This Config is the LLM package's own representation; it is
// populated by the caller (e.g., from config.LLMConfig or directly
// from environment variables). The LLM package does not read
// environment variables itself.
type Config struct {
	// Enable controls whether LLM calls are attempted at all.
	// When false, Chat returns LLM_DISABLED.
	Enable bool `json:"enable"`
	// Provider identifies which backend to use ("hunyuan", "noop").
	Provider Provider `json:"provider"`
	// BaseURL is the OpenAI-compatible API endpoint.
	BaseURL string `json:"base_url"`
	// APIKey is the authentication key. Must only be stored in .env.
	APIKey string `json:"-"`
	// Model is the model identifier (e.g. "hunyuan-turbos-latest").
	Model string `json:"model"`
	// Temperature controls output randomness. Recommended [0, 2].
	Temperature float64 `json:"temperature"`
	// MaxOutputTokens limits response length.
	MaxOutputTokens int `json:"max_output_tokens"`
	// TimeoutSeconds is the per-request timeout.
	TimeoutSeconds int `json:"timeout_seconds"`
	// RepairEnable controls whether JSON repair is attempted.
	RepairEnable bool `json:"repair_enable"`
	// RepairMaxAttempts is the maximum number of repair attempts.
	// Must be exactly 1.
	RepairMaxAttempts int `json:"repair_max_attempts"`
	// FallbackEnable controls whether Mock fallback is allowed.
	FallbackEnable bool `json:"fallback_enable"`
	// LogRawOutput controls whether raw model output is logged.
	// Should be false in production.
	LogRawOutput bool `json:"log_raw_output"`
	// PromptDir is the directory containing Prompt template files.
	PromptDir string `json:"prompt_dir"`
	// SchemaDir is the directory containing JSON Schema files.
	SchemaDir string `json:"schema_dir"`
}

// DefaultConfig returns a Config with default values that match
// the .env.example defaults established in 4-2.1.
func DefaultConfig() Config {
	return Config{
		Enable:           false,
		Provider:         ProviderHunyuan,
		BaseURL:          "",
		APIKey:           "",
		Model:            "",
		Temperature:      0.2,
		MaxOutputTokens:  8192,
		TimeoutSeconds:   30,
		RepairEnable:     true,
		RepairMaxAttempts: 1,
		FallbackEnable:   true,
		LogRawOutput:     false,
		PromptDir:        "server/prompts",
		SchemaDir:        "server/prompts/schema",
	}
}

// Validate checks that the configuration is consistent.
// When Enable is false, only basic sanity checks are performed
// (TimeoutSeconds > 0, MaxOutputTokens > 0, Temperature in range,
// RepairMaxAttempts == 1, directories non-empty).
// When Enable is true, Provider, BaseURL, APIKey, and Model are
// additionally required.
func (c Config) Validate() error {
	var errs []string

	// Always-required fields
	if c.TimeoutSeconds <= 0 {
		errs = append(errs, "TimeoutSeconds must be > 0")
	}
	if c.MaxOutputTokens <= 0 {
		errs = append(errs, "MaxOutputTokens must be > 0")
	}
	if c.Temperature < 0 || c.Temperature > 2 {
		errs = append(errs, "Temperature must be in [0, 2]")
	}
	if c.RepairMaxAttempts != 1 {
		errs = append(errs, "RepairMaxAttempts must be 1")
	}
	if strings.TrimSpace(c.PromptDir) == "" {
		errs = append(errs, "PromptDir must not be empty")
	}
	if strings.TrimSpace(c.SchemaDir) == "" {
		errs = append(errs, "SchemaDir must not be empty")
	}

	// LLM-enabled-specific required fields
	if c.Enable {
		if strings.TrimSpace(string(c.Provider)) == "" {
			errs = append(errs, "Provider must not be empty when LLM is enabled")
		}
		if strings.TrimSpace(c.BaseURL) == "" {
			errs = append(errs, "BaseURL must not be empty when LLM is enabled")
		}
		if strings.TrimSpace(c.APIKey) == "" {
			errs = append(errs, "APIKey must not be empty when LLM is enabled")
		}
		if strings.TrimSpace(c.Model) == "" {
			errs = append(errs, "Model must not be empty when LLM is enabled")
		}
	}

	if len(errs) > 0 {
		return &LLMError{
			Code:    ErrLLMConfigMissing,
			Message: fmt.Sprintf("invalid LLM config: %s", strings.Join(errs, "; ")),
		}
	}
	return nil
}

// MaskedAPIKey returns a masked representation of the API key
// suitable for logging.
func (c Config) MaskedAPIKey() string {
	return maskAPIKey(c.APIKey)
}
