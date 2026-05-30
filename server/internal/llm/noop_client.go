package llm

import "context"

// NoopClient is a Client implementation that never accesses external
// networks. It is used to test the Client interface contract and
// error-handling paths without requiring a real LLM provider.
//
// When config.Enable is false, Chat returns LLM_DISABLED.
// When config.Enable is true, Chat returns a fixed response with
// ProviderNoop metadata.
type NoopClient struct {
	config Config
}

// NewNoopClient creates a NoopClient with the given configuration.
func NewNoopClient(config Config) *NoopClient {
	return &NoopClient{config: config}
}

// Provider returns ProviderNoop.
func (c *NoopClient) Provider() Provider {
	return ProviderNoop
}

// Chat implements the Client interface. If LLM is disabled it
// returns LLM_DISABLED; otherwise it returns a synthetic
// ChatResponse with a fixed RawContent.
func (c *NoopClient) Chat(_ context.Context, req ChatRequest) (*ChatResponse, error) {
	if !c.config.Enable {
		return nil, NewLLMError(ErrLLMDisabled, "LLM is disabled (LLM_ENABLE=false)")
	}

	rawContent := `{"noop":true,"message":"noop client response"}`
	return &ChatResponse{
		Provider:        ProviderNoop,
		Model:           coalesce(req.Model, c.config.Model, "noop-model"),
		RawContent:      rawContent,
		FinishReason:    "stop",
		Usage:           Usage{},
		RequestID:       "noop-request",
		DurationMs:      0,
		RawOutputLength: len(rawContent),
	}, nil
}

// coalesce returns the first non-empty string from the provided
// candidates.
func coalesce(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
