package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	TimeUTC string `json:"time_utc"`
}

func main() {
	addr := env("GATEWAY_ADDR", ":8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/v1/chat/completions", chatCompletionsPlaceholder)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir gateway listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway failed: %v", err)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Service: "agentvoir-gateway",
		Status:  "ok",
		TimeUTC: time.Now().UTC().Format(time.RFC3339),
	})
}

func chatCompletionsPlaceholder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	// TODO: Implement OpenAI-compatible request handling:
	// 1. Authenticate request.
	// 2. Resolve x-agent-id from registry.
	// 3. Run policy checks.
	// 4. Canonicalize request and compute cache key.
	// 5. Serve exact/semantic cache hit or route to provider.
	// 6. Stream response when requested.
	// 7. Emit usage, cost, cache, and trace events.
	writeJSON(w, http.StatusNotImplemented, map[string]string{
		"error": "AgentVoir gateway scaffold: /v1/chat/completions is not implemented yet",
	})
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
