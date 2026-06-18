package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Agent struct {
	AgentID     string   `json:"agent_id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	OwnerTeam   string   `json:"owner_team"`
	Lifecycle   string   `json:"lifecycle"`
	RiskLevel   string   `json:"risk_level"`
	DataClasses []string `json:"data_classes"`
}

func main() {
	addr := env("REGISTRY_API_ADDR", ":8081")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/v1/agents", agentsPlaceholder)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir registry API listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("registry API failed: %v", err)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service":  "agentvoir-registry-api",
		"status":   "ok",
		"time_utc": time.Now().UTC().Format(time.RFC3339),
	})
}

func agentsPlaceholder(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, []Agent{})
	case http.MethodPost:
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "AgentVoir registry scaffold: create agent is not implemented yet",
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
