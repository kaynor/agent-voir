package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/server"
)

func TestRegistryAPIEndpoints(t *testing.T) {
	stores := server.NewStores()
	mux := http.NewServeMux()
	server.RegisterRoutes(mux, stores)

	agentBody := `{
		"agent_id": "customer-support-agent",
		"name": "Customer Support Agent",
		"version": "0.1.0",
		"owner_team": "support-platform",
		"environment": "staging"
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewBufferString(agentBody))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create agent: %d %s", rec.Code, rec.Body.String())
	}

	depBody := `{"dependency_type":"tool","dependency_name":"zendesk"}`
	req = httptest.NewRequest(http.MethodPost, "/v1/agents/customer-support-agent/dependencies?version=0.1.0", bytes.NewBufferString(depBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create dependency: %d %s", rec.Code, rec.Body.String())
	}

	budgetBody := `{"monthly_usd":1000,"max_prompt_tokens_per_request":12000}`
	req = httptest.NewRequest(http.MethodPut, "/v1/agents/customer-support-agent/budget?version=0.1.0", bytes.NewBufferString(budgetBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert budget: %d %s", rec.Code, rec.Body.String())
	}

	routeBody := `{"primary_provider":"openai","primary_model":"gpt-4.1-mini"}`
	req = httptest.NewRequest(http.MethodPut, "/v1/agents/customer-support-agent/model-route?version=0.1.0", bytes.NewBufferString(routeBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("upsert model route: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agents/customer-support-agent/dependency-graph?version=0.1.0", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dependency graph: %d %s", rec.Code, rec.Body.String())
	}

	promptBody := `{
		"prompt_id": "support-greeting",
		"name": "Support Greeting",
		"version": "1.0.0",
		"owner_team": "support-platform",
		"template": "Hello {{name}}"
	}`
	req = httptest.NewRequest(http.MethodPost, "/v1/prompts", bytes.NewBufferString(promptBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create prompt: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/prompts", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list prompts: %d %s", rec.Code, rec.Body.String())
	}

	var prompts []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&prompts); err != nil {
		t.Fatalf("decode prompts: %v", err)
	}
	if len(prompts) != 1 {
		t.Fatalf("prompt count = %d, want 1", len(prompts))
	}
}
