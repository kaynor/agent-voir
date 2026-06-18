package providers

import "context"

type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
	Options  map[string]any `json:"-"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Provider         string
	Model            string
	Raw              []byte
	PromptTokens     int64
	CompletionTokens int64
	CostUSD          float64
}

type Provider interface {
	Name() string
	ChatCompletions(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}
