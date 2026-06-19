package providers

import (
	"context"
	"errors"
)

// ErrUnavailable marks a deliberate upstream failure (used for fallback demos).
var ErrUnavailable = errors.New("provider unavailable")

// UnavailableProvider always fails so the gateway can exercise fallback routing.
type UnavailableProvider struct{}

func NewUnavailableProvider() *UnavailableProvider {
	return &UnavailableProvider{}
}

func (p *UnavailableProvider) Name() string {
	return "unavailable"
}

func (p *UnavailableProvider) ChatCompletions(context.Context, ChatRequest) (*ChatResponse, error) {
	return nil, ErrUnavailable
}
