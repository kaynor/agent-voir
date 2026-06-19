package agents_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/agents"
)

func TestRegisterAndListAgents(t *testing.T) {
	store := agents.NewMemoryStore()
	handler := agents.NewHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := map[string]any{
		"agent_id":     "customer-support-agent",
		"name":         "Customer Support Agent",
		"version":      "0.1.0",
		"owner_team":   "support-platform",
		"environment":  "staging",
		"risk_level":   "medium",
		"data_classes": []string{"customer_pii"},
	}
	payload, _ := json.Marshal(body)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewReader(payload))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/agents", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	var listed agents.ListResult
	if err := json.NewDecoder(listRec.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("listed %d agents, want 1", len(listed.Items))
	}
	if listed.Items[0].AgentID != "customer-support-agent" {
		t.Fatalf("agent_id = %q", listed.Items[0].AgentID)
	}
}

func TestRegisterAgentConflict(t *testing.T) {
	store := agents.NewMemoryStore()
	handler := agents.NewHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	payload := []byte(`{
		"agent_id": "demo-agent",
		"name": "Demo",
		"version": "1.0.0",
		"owner_team": "platform"
	}`)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/agents", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if i == 0 && rec.Code != http.StatusCreated {
			t.Fatalf("first create status = %d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusConflict {
			t.Fatalf("second create status = %d, want %d", rec.Code, http.StatusConflict)
		}
	}
}
