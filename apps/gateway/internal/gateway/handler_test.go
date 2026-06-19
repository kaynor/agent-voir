package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/middleware"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/policy"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/providers"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/usage"
	agentauth "github.com/agentvoir/agentvoir/packages/auth-go"
)

func TestChatCompletionsMockProvider(t *testing.T) {
	cfg := Config{
		APIKey:          "test-key",
		CacheMode:       "exact_only",
		CacheTTLSeconds: 60,
	}
	handler := NewHandler(cfg, cache.NewMemoryStore(), providers.NewRegistry(nil, providers.NewMockProvider()), nil, nil, policy.NopEvaluator{}, nil, usage.NopRecorder{})

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	secured := middleware.Auth(agentauth.Config{StaticAPIKeys: []string{"test-key"}})(mux)

	body := `{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer test-key")
	req.Header.Set(middleware.HeaderAgentID, "customer-support-agent")
	rec := httptest.NewRecorder()
	secured.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(middleware.HeaderAgentID) != "customer-support-agent" {
		t.Fatalf("agent header missing")
	}
	if rec.Header().Get(middleware.HeaderCacheStatus) != "miss" {
		t.Fatalf("cache status = %q", rec.Header().Get(middleware.HeaderCacheStatus))
	}

	var completion map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&completion); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if completion["object"] != "chat.completion" {
		t.Fatalf("object = %v", completion["object"])
	}

	// Second identical request should be a cache hit.
	req2 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	req2.Header.Set("Authorization", "Bearer test-key")
	req2.Header.Set(middleware.HeaderAgentID, "customer-support-agent")
	rec2 := httptest.NewRecorder()
	secured.ServeHTTP(rec2, req2)
	if rec2.Header().Get(middleware.HeaderCacheStatus) != "hit" {
		t.Fatalf("second cache status = %q", rec2.Header().Get(middleware.HeaderCacheStatus))
	}
}

func TestChatCompletionsRequiresAgentID(t *testing.T) {
	handler := NewHandler(Config{APIKey: "test-key"}, cache.NewMemoryStore(), providers.NewRegistry(nil, providers.NewMockProvider()), nil, nil, policy.NopEvaluator{}, nil, usage.NopRecorder{})
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	secured := middleware.Auth(agentauth.Config{StaticAPIKeys: []string{"test-key"}})(mux)

	body := `{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"Hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer test-key")
	rec := httptest.NewRecorder()
	secured.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
