package llm

// Provider identifies the LLM backend implementation.
type Provider string

const (
	// ProviderHunyuan is the Tencent Hunyuan model provider.
	ProviderHunyuan Provider = "hunyuan"
	// ProviderHy3 is the Tencent Cloud LKEAP hy3-preview provider.
	ProviderHy3 Provider = "hy3"
	// ProviderNoop is a non-functional provider used for testing the
	// Client interface contract without accessing external networks.
	ProviderNoop Provider = "noop"
)

// Role classifies a chat message participant.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message represents a single chat message.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the unified request structure for LLM chat calls.
// It is provider-agnostic; concrete clients translate it into
// provider-specific wire formats.
type ChatRequest struct {
	// Model may be empty; the concrete client uses its configured default.
	Model string `json:"model,omitempty"`
	// Messages must contain at least one entry.
	Messages []Message `json:"messages"`
	// Temperature controls output randomness. Recommended range [0, 2].
	Temperature float64 `json:"temperature,omitempty"`
	// TopP controls nucleus sampling. Recommended range [0, 1]. Zero = use default.
	TopP float64 `json:"top_p,omitempty"`
	// MaxTokens limits the response length. Must be > 0.
	MaxTokens int `json:"max_tokens,omitempty"`
	// ReasoningEffort is the Hy3-preview thinking effort: no_think / low / high.
	// Empty string means use provider default. Ignored by non-Hy3 providers.
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
	// ExtraBody carries provider-specific fields merged into the request body.
	// When ReasoningEffort is set on a Hy3 provider, it is automatically added
	// under extra_body.chat_template_kwargs.reasoning_effort.
	ExtraBody map[string]interface{} `json:"-"`
	// TimeoutSec overrides the client-level timeout for this request.
	// Zero means use the client default.
	TimeoutSec int `json:"-"`
	// Metadata carries non-sensitive diagnostic information such as
	// prompt_name, schema_name, agent_name. API keys must never be
	// placed here.
	Metadata map[string]string `json:"-"`
}

// ChatResponse is the unified response from an LLM chat call.
type ChatResponse struct {
	// Provider identifies which backend produced this response.
	Provider Provider `json:"provider"`
	// Model is the actual model name used (may differ from request
	// when the client applies a default).
	Model string `json:"model"`
	// RawContent is the model's raw text output. Downstream consumers
	// (llmjson + schema validator) are responsible for parsing it.
	RawContent string `json:"raw_content"`
	// FinishReason is the provider's stop reason, e.g. "stop" or "length".
	FinishReason string `json:"finish_reason,omitempty"`
	// Usage reports token consumption. Values may be 0 when the
	// provider does not return usage information.
	Usage Usage `json:"usage"`
	// RequestID is an optional provider-assigned identifier.
	RequestID string `json:"request_id,omitempty"`
	// DurationMs is the wall-clock time of the LLM call in milliseconds.
	DurationMs int64 `json:"duration_ms"`
	// RawOutputLength is the byte length of RawContent, useful for
	// logging without exposing full content.
	RawOutputLength int `json:"raw_output_length"`
}

// Usage reports token consumption for an LLM call.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
