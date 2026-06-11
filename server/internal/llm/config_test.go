package llm

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify defaults match .env.example
	if cfg.Enable != false {
		t.Errorf("Enable = %v, want false", cfg.Enable)
	}
	if cfg.Provider != ProviderHunyuan {
		t.Errorf("Provider = %q, want %q", cfg.Provider, ProviderHunyuan)
	}
	if cfg.Temperature != 0.2 {
		t.Errorf("Temperature = %v, want 0.2", cfg.Temperature)
	}
	if cfg.MaxOutputTokens != 8192 {
		t.Errorf("MaxOutputTokens = %d, want 8192", cfg.MaxOutputTokens)
	}
	if cfg.TimeoutSeconds != 120 {
		t.Errorf("TimeoutSeconds = %d, want 120", cfg.TimeoutSeconds)
	}
	if cfg.RepairEnable != true {
		t.Errorf("RepairEnable = %v, want true", cfg.RepairEnable)
	}
	if cfg.RepairMaxAttempts != 1 {
		t.Errorf("RepairMaxAttempts = %d, want 1", cfg.RepairMaxAttempts)
	}
	if cfg.FallbackEnable != true {
		t.Errorf("FallbackEnable = %v, want true", cfg.FallbackEnable)
	}
	if cfg.LogRawOutput != false {
		t.Errorf("LogRawOutput = %v, want false", cfg.LogRawOutput)
	}
	if cfg.PromptDir != "server/prompts" {
		t.Errorf("PromptDir = %q, want %q", cfg.PromptDir, "server/prompts")
	}
	if cfg.SchemaDir != "server/prompts/schema" {
		t.Errorf("SchemaDir = %q, want %q", cfg.SchemaDir, "server/prompts/schema")
	}
}

func TestConfig_Validate_Disabled_NoAPIKey(t *testing.T) {
	cfg := DefaultConfig() // Enable=false
	// When disabled, APIKey/BaseURL/Model are not required
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should pass when disabled without API key, got: %v", err)
	}
}

func TestConfig_Validate_Enabled_MissingAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://api.example.com/v1"
	cfg.Model = "test-model"
	// APIKey is empty
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when APIKey is missing and LLM is enabled")
	}
	if !IsErrorCode(err, ErrLLMConfigMissing) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMConfigMissing)
	}
}

func TestConfig_Validate_Enabled_MissingBaseURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.APIKey = "test-key"
	cfg.Model = "test-model"
	// BaseURL is empty
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when BaseURL is missing and LLM is enabled")
	}
	if !IsErrorCode(err, ErrLLMConfigMissing) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMConfigMissing)
	}
}

func TestConfig_Validate_Enabled_MissingModel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://api.example.com/v1"
	cfg.APIKey = "test-key"
	// Model is empty
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when Model is missing and LLM is enabled")
	}
	if !IsErrorCode(err, ErrLLMConfigMissing) {
		t.Errorf("error code = %q, want %q", err.(*LLMError).Code, ErrLLMConfigMissing)
	}
}

func TestConfig_Validate_Enabled_AllFieldsPresent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.BaseURL = "https://api.example.com/v1"
	cfg.APIKey = "test-key"
	cfg.Model = "test-model"
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() should pass with all required fields, got: %v", err)
	}
}

func TestConfig_Validate_RepairMaxAttemptsNotOne(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RepairMaxAttempts = 2
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when RepairMaxAttempts != 1")
	}
}

func TestConfig_Validate_TimeoutZero(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TimeoutSeconds = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when TimeoutSeconds <= 0")
	}
}

func TestConfig_Validate_TimeoutNegative(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TimeoutSeconds = -1
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when TimeoutSeconds <= 0")
	}
}

func TestConfig_Validate_MaxOutputTokensZero(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxOutputTokens = 0
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when MaxOutputTokens <= 0")
	}
}

func TestConfig_Validate_TemperatureTooHigh(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Temperature = 2.5
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when Temperature > 2")
	}
}

func TestConfig_Validate_TemperatureNegative(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Temperature = -0.1
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when Temperature < 0")
	}
}

func TestConfig_Validate_TemperatureZero(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Temperature = 0
	if err := cfg.Validate(); err != nil {
		t.Errorf("Temperature=0 should be valid, got: %v", err)
	}
}

func TestConfig_Validate_TemperatureTwo(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Temperature = 2.0
	if err := cfg.Validate(); err != nil {
		t.Errorf("Temperature=2.0 should be valid, got: %v", err)
	}
}

func TestConfig_Validate_EmptyPromptDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PromptDir = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when PromptDir is empty")
	}
}

func TestConfig_Validate_EmptySchemaDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SchemaDir = "  "
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when SchemaDir is whitespace-only")
	}
}

func TestConfig_MaskedAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "sk-1234567890abcdef"
	masked := cfg.MaskedAPIKey()
	if masked == cfg.APIKey {
		t.Error("MaskedAPIKey should not expose the full key")
	}
	if masked == "" {
		t.Error("MaskedAPIKey should not be empty")
	}
}

func TestConfig_Validate_Enabled_MissingProvider(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enable = true
	cfg.Provider = ""
	cfg.BaseURL = "https://api.example.com/v1"
	cfg.APIKey = "test-key"
	cfg.Model = "test-model"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error when Provider is empty and LLM is enabled")
	}
}
