package providers

import (
	"context"
	"testing"
)

func TestCallWithFallback_PrimarySuccess(t *testing.T) {
	reg := NewRegistry(nil, NewMockProvider())
	result, err := reg.CallWithFallback(
		context.Background(),
		"mock", "gpt-4.1-mini", "openai", "gpt-4.1-mini", "primary_then_fallback",
		[]ChatMessage{{Role: "user", Content: "hello"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Fallback {
		t.Fatal("expected primary route, got fallback")
	}
	if result.Provider != "mock" {
		t.Fatalf("expected mock provider, got %q", result.Provider)
	}
}

func TestCallWithFallback_UsesFallbackOnPrimaryFailure(t *testing.T) {
	reg := NewRegistry(nil, NewMockProvider())
	result, err := reg.CallWithFallback(
		context.Background(),
		"unavailable", "gpt-4.1-mini", "mock", "gpt-4.1-mini", "primary_then_fallback",
		[]ChatMessage{{Role: "user", Content: "hello"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Fallback {
		t.Fatal("expected fallback route")
	}
	if result.Provider != "mock" {
		t.Fatalf("expected mock fallback provider, got %q", result.Provider)
	}
}

func TestCallWithFallback_NoFallbackWhenPolicyPrimaryOnly(t *testing.T) {
	reg := NewRegistry(nil, NewMockProvider())
	_, err := reg.CallWithFallback(
		context.Background(),
		"unavailable", "gpt-4.1-mini", "mock", "gpt-4.1-mini", "primary_only",
		[]ChatMessage{{Role: "user", Content: "hello"}},
	)
	if err == nil {
		t.Fatal("expected primary failure error")
	}
}
