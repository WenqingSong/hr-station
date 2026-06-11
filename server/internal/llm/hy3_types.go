package llm

// hy3ChatCompletionRequest is the request body for LKEAP hy3-preview.
type hy3ChatCompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Temperature *float64               `json:"temperature,omitempty"`
	TopP        *float64               `json:"top_p,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	ExtraBody   map[string]interface{} `json:"extra_body,omitempty"`
}

// hy3ChatCompletionResponse is the response from LKEAP hy3-preview.
type hy3ChatCompletionResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object,omitempty"`
	Model   string          `json:"model,omitempty"`
	Choices []hy3ChatChoice `json:"choices"`
	Usage   hy3Usage        `json:"usage,omitempty"`
}

type hy3ChatChoice struct {
	Index        int         `json:"index"`
	Message      hy3Message  `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type hy3Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type hy3Usage struct {
	PromptTokens     int                   `json:"prompt_tokens"`
	CompletionTokens int                   `json:"completion_tokens"`
	TotalTokens      int                   `json:"total_tokens"`
	PromptTokensDetails     hy3TokensDetail `json:"prompt_tokens_details,omitempty"`
	CompletionTokensDetails hy3TokensDetail `json:"completion_tokens_details,omitempty"`
}

type hy3TokensDetail struct {
	CachedTokens    int `json:"cached_tokens,omitempty"`
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}
