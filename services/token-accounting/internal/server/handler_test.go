package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentvoir/agentvoir/services/token-accounting/internal/server"
	"github.com/agentvoir/agentvoir/services/token-accounting/internal/usage"
)

func TestUsageEventIngestionAndList(t *testing.T) {
	mux := http.NewServeMux()
	server.NewHandler(usage.NewMemoryStore()).RegisterRoutes(mux)

	body := `{
		"trace_id": "trace-123",
		"tenant_id": "acme",
		"agent_id": "customer-support-agent",
		"agent_version": "0.1.0",
		"user_id": "user-42",
		"provider": "openai",
		"model": "gpt-4.1-mini",
		"cache_status": "miss",
		"prompt_tokens": 120,
		"completion_tokens": 45,
		"cost_usd": 0.0021,
		"latency_ms": 812,
		"status_code": 200
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/usage-events", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("ingest status = %d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/usage-events?agent_id=customer-support-agent", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rec.Code, rec.Body.String())
	}

	var events []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if events[0]["trace_id"] != "trace-123" {
		t.Fatalf("trace_id = %v", events[0]["trace_id"])
	}
}

func TestUsageEventRequiresAgentID(t *testing.T) {
	mux := http.NewServeMux()
	server.NewHandler(usage.NewMemoryStore()).RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/usage-events", bytes.NewBufferString(`{"model":"gpt-4.1-mini"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
