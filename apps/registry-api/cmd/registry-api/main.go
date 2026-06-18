package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/server"
)

func main() {
	addr := env("REGISTRY_API_ADDR", ":8081")

	stores := server.NewStores()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	server.RegisterRoutes(mux, stores)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir registry API listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
