package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/accounting"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/budget"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/gateway"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/metrics"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/middleware"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/policy"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/providers"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/ratelimit"
	agentregistry "github.com/agentvoir/agentvoir/apps/gateway/internal/registry"
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
	providerRegistry := providers.NewRegistry(openaiProvider, providers.NewMockProvider())
	agentRegistry := agentregistry.NewClient(config.RegistryAPIURL, config.RegistryAPIKey)
	accountingClient := accounting.NewClient(config.TokenAccountingURL)
	budgetChecker := budget.NewChecker(&budget.RegistryAdapter{Client: agentRegistry}, accounting.NewSpendAdapter(accountingClient))

	var policyEvaluator policy.Evaluator = policy.NopEvaluator{}
	if config.OPAURL != "" {
		policyEvaluator = policy.NewOPAClient(config.OPAURL)
		log.Printf("AgentVoir gateway enforcing OPA policies at %s", config.OPAURL)
	}

	var rateLimiter *ratelimit.Limiter
	if config.RedisAddr != "" {
		rateLimiter, err = ratelimit.NewLimiter(config.RedisAddr)
		if err != nil {
			log.Fatalf("rate limiter init failed: %v", err)
		}
		log.Printf("AgentVoir gateway enforcing per-agent rate limits via Redis at %s", config.RedisAddr)
	}

	handler := gateway.NewHandler(
		config,
		cacheStore,
		providerRegistry,
		agentRegistry,
		budgetChecker,
		policyEvaluator,
		rateLimiter,
		usage.NewRecorder(config.TokenAccountingURL),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.Handle("/metrics", metrics.Handler())
	handler.RegisterRoutes(mux)

	authCfg := middleware.LoadAuthConfig()
	if !authCfg.Enabled() && config.APIKey != "" {
		authCfg.StaticAPIKeys = []string{config.APIKey}
	}
	if authCfg.Enabled() {
		if authCfg.IssuerURL != "" {
			log.Printf("AgentVoir gateway accepting OIDC tokens from %s", authCfg.IssuerURL)
		}
		if len(authCfg.StaticAPIKeys) > 0 {
			log.Printf("AgentVoir gateway accepting bootstrap API key (GATEWAY_API_KEY)")
		}
	}

	server := &http.Server{
		Addr:              config.Addr,
		Handler:           middleware.DevCORS(middleware.Auth(authCfg)(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("AgentVoir gateway listening on %s (cache_mode=%s)", config.Addr, config.CacheMode)
	if config.RegistryAPIURL != "" {
		log.Printf("AgentVoir gateway loading agent config from %s", config.RegistryAPIURL)
	}
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
