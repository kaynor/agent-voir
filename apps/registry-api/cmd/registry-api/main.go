package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentvoir/agentvoir/apps/registry-api/internal/postgres"
	"github.com/agentvoir/agentvoir/apps/registry-api/internal/server"
)

func main() {
	addr := env("REGISTRY_API_ADDR", ":8081")
	dsn := os.Getenv("POSTGRES_DSN")

	var stores *server.Stores
	if dsn != "" {
		pool, err := postgres.Open(context.Background(), dsn)
		if err != nil {
			log.Fatalf("postgres init failed: %v", err)
		}
		defer pool.Close()
		stores = server.NewPostgresStores(pool)
		log.Printf("AgentVoir registry API using PostgreSQL metadata store")
	} else {
		stores = server.NewMemoryStores()
		log.Printf("AgentVoir registry API using in-memory store (set POSTGRES_DSN for PostgreSQL)")
	}

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
