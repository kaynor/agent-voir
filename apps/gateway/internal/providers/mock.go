package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/openai"
	"github.com/google/uuid"
)

// MockProvider returns a deterministic OpenAI-compatible response for local development.
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Name() string {
	return "mock"
}

func (p *MockProvider) ChatCompletions(_ context.Context, req ChatRequest) (*ChatResponse, error) {
	lastUser := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUser = req.Messages[i].Content
			break
		}
	}

	content := fmt.Sprintf(
		"AgentVoir mock response for model %s: %s",
		req.Model,
		strings.TrimSpace(lastUser),
	)
	if content == fmt.Sprintf("AgentVoir mock response for model %s:", req.Model) {
		content = fmt.Sprintf("AgentVoir mock response for model %s.", req.Model)
	}

	promptTokens := estimateTokens(req.Messages)
	completionTokens := int64(len(content) / 4)
	if completionTokens < 1 {
		completionTokens = 1
	}

	response := openai.ChatCompletionResponse{
		ID:      "chatcmpl-" + uuid.NewString(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []openai.ChatCompletionChoice{{
			Index: 0,
			Message: openai.ChatMessage{
				Role:    "assistant",
				Content: content,
			},
			FinishReason: "stop",
		}},
		Usage: openai.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}

	raw, err := json.Marshal(response)
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Provider:         p.Name(),
		Model:            req.Model,
		Raw:              raw,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CostUSD:          0,
	}, nil
}

func estimateTokens(messages []ChatMessage) int64 {
	var chars int
	for _, msg := range messages {
		chars += len(msg.Role) + len(msg.Content)
	}
	tokens := int64(chars / 4)
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}
