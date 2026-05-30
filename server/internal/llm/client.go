package llm

import "context"

// Client is the unified LLM calling abstraction. Business agents
// depend on this interface rather than on any concrete provider.
type Client interface {
	// Chat sends a chat request and returns the model's raw text
	// output. The caller is responsible for JSON extraction, parsing,
	// and schema validation using the llmjson and schema packages.
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Provider returns the provider identifier for this client.
	Provider() Provider
}
