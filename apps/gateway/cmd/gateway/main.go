package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/gateway"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/providers"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/usage"
)

func main() {
	config := gateway.LoadConfig()

	cacheStore, err := cache.NewStore(context.Background(), config.RedisAddr)
	if err != nil {
		log.Fatalf("cache init failed: %v", err)
	}
	if config.RedisAddr != "" {
		log.Printf("AgentVoir gateway using Redis exact cache at %s", config.RedisAddr)
	} else {
		log.Printf("AgentVoir gateway using in-memory exact cache (set REDIS_ADDR for Redis)")
	}

	var openaiProvider *providers.OpenAIProvider
	if config.OpenAIAPIKey != "" {
		openaiProvider = providers.NewOpenAIProvider(config.OpenAIAPIKey, config.OpenAIBaseURL)
	}
	registry := providers.NewRegistry(openaiProvider, providers.NewMockProvider())
	handler := gateway.NewHandler(config, cacheStore, registry, usage.NewRecorder(config.TokenAccountingURL))

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir gateway listening on %s (cache_mode=%s)", config.Addr, config.CacheMode)
	if config.TokenAccountingURL != "" {
		log.Printf("AgentVoir gateway emitting usage events to %s", config.TokenAccountingURL)
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway failed: %v", err)
	}
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"service":  "agentvoir-gateway",
		"status":   "ok",
		"time_utc": time.Now().UTC().Format(time.RFC3339),
	})
}
