package llm

// chatCompletionRequest is the OpenAI-compatible chat completions request
// body sent to Hunyuan / TokenHub.
type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []Message     `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// chatCompletionResponse is the OpenAI-compatible chat completions response
// body returned by Hunyuan / TokenHub.
type chatCompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object,omitempty"`
	Model   string             `json:"model,omitempty"`
	Choices []chatChoice       `json:"choices"`
	Usage   chatCompletionUsage `json:"usage,omitempty"`
}

// chatChoice represents a single completion choice.
type chatChoice struct {
	Index        int          `json:"index"`
	Message      chatMessage  `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

// chatMessage is the message structure within a choice.
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatCompletionUsage reports token consumption.
type chatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
