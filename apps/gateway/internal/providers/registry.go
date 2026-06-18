package providers

import "strings"

// Registry selects a provider for a requested model.
type Registry struct {
	openai *OpenAIProvider
	mock   *MockProvider
}

func NewRegistry(openai *OpenAIProvider, mock *MockProvider) *Registry {
	return &Registry{openai: openai, mock: mock}
}

func (r *Registry) Resolve(model string) Provider {
	if r.openai != nil && looksLikeOpenAIModel(model) {
		return r.openai
	}
	return r.mock
}

func looksLikeOpenAIModel(model string) bool {
	model = strings.ToLower(model)
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}
