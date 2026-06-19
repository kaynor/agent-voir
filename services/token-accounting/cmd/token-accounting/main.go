package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentvoir/agentvoir/services/token-accounting/internal/httputil"
	"github.com/agentvoir/agentvoir/services/token-accounting/internal/server"
)

func main() {
	addr := env("TOKEN_ACCOUNTING_ADDR", ":8082")
	clickhouseDSN := os.Getenv("CLICKHOUSE_DSN")

	store, err := server.OpenStore(context.Background(), clickhouseDSN)
	if err != nil {
		log.Fatalf("usage store init failed: %v", err)
	}
	if clickhouseDSN != "" {
		log.Printf("AgentVoir token-accounting using ClickHouse at %s", clickhouseDSN)
	} else {
		log.Printf("AgentVoir token-accounting using in-memory store (set CLICKHOUSE_DSN for ClickHouse)")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	server.NewHandler(store).RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           httputil.WrapDevCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir token-accounting listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("token-accounting failed: %v", err)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service":  "agentvoir-token-accounting",
		"status":   "ok",
		"time_utc": time.Now().UTC().Format(time.RFC3339),
	})
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
