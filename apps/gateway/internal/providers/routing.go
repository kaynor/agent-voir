package providers

import (
	"context"
	"strings"
)

// RouteRequest describes a provider/model invocation attempt.
type RouteRequest struct {
	Provider string
	Model    string
	Messages []ChatMessage
}

// RouteResult is a successful provider response with routing metadata.
type RouteResult struct {
	Response *ChatResponse
	Provider string
	Model    string
	Fallback bool
}

// CallWithFallback tries the primary route, then optional fallback on failure.
func (r *Registry) CallWithFallback(
	ctx context.Context,
	primaryProvider, primaryModel, fallbackProvider, fallbackModel, routingPolicy string,
	messages []ChatMessage,
) (*RouteResult, error) {
	if routingPolicy == "" {
		routingPolicy = "primary_then_fallback"
	}

	primaryModel = strings.TrimSpace(primaryModel)
	if primaryProvider == "" {
		provider := r.Resolve(primaryModel)
		resp, err := provider.ChatCompletions(ctx, ChatRequest{Model: primaryModel, Messages: messages})
		if err != nil {
			return nil, err
		}
		return &RouteResult{Response: resp, Provider: provider.Name(), Model: primaryModel}, nil
	}

	primary := r.ResolveProvider(primaryProvider, primaryModel)
	resp, err := primary.ChatCompletions(ctx, ChatRequest{Model: primaryModel, Messages: messages})
	if err == nil {
		return &RouteResult{Response: resp, Provider: primary.Name(), Model: primaryModel}, nil
	}

	if routingPolicy != "primary_then_fallback" || fallbackProvider == "" {
		return nil, err
	}

	fbModel := strings.TrimSpace(fallbackModel)
	if fbModel == "" {
		fbModel = primaryModel
	}
	fallback := r.ResolveProvider(fallbackProvider, fbModel)
	fbResp, fbErr := fallback.ChatCompletions(ctx, ChatRequest{Model: fbModel, Messages: messages})
	if fbErr != nil {
		return nil, fbErr
	}
	return &RouteResult{
		Response: fbResp,
		Provider: fallback.Name(),
		Model:    fbModel,
		Fallback: true,
	}, nil
}

func (req RouteRequest) toChat() ChatRequest {
	return ChatRequest{Model: req.Model, Messages: req.Messages}
}
