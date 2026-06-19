package providers

import (
	"strings"
)

// Registry selects a provider for a requested model.
type Registry struct {
	openai      *OpenAIProvider
	mock        *MockProvider
	unavailable *UnavailableProvider
}

func NewRegistry(openai *OpenAIProvider, mock *MockProvider) *Registry {
	return &Registry{
		openai:      openai,
		mock:        mock,
		unavailable: NewUnavailableProvider(),
	}
}

func (r *Registry) Resolve(model string) Provider {
	if r.openai != nil && looksLikeOpenAIModel(model) {
		return r.openai
	}
	return r.mock
}

func (r *Registry) ResolveProvider(providerName, model string) Provider {
	switch strings.ToLower(strings.TrimSpace(providerName)) {
	case "openai":
		if r.openai != nil {
			return r.openai
		}
	case "mock":
		if r.mock != nil {
			return r.mock
		}
	case "unavailable", "mock-unavailable":
		return r.unavailable
	}
	return r.Resolve(model)
}

func looksLikeOpenAIModel(model string) bool {
	model = strings.ToLower(model)
	return strings.HasPrefix(model, "gpt-") ||
		strings.HasPrefix(model, "o1") ||
		strings.HasPrefix(model, "o3") ||
		strings.HasPrefix(model, "o4")
}
