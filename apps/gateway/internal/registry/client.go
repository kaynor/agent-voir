package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AgentPolicies holds governance settings from the registry.
type AgentPolicies struct {
	AllowedProviders []string
	PIIAllowed       bool
	RequireAuditLog  bool
}

func (p AgentPolicies) OPAFormat() map[string]any {
	return map[string]any{
		"allowedProviders": p.AllowedProviders,
		"piiAllowed":       p.PIIAllowed,
		"requireAuditLog":  p.RequireAuditLog,
	}
}

// AgentConfig is the runtime configuration loaded from the registry API.
type AgentConfig struct {
	AgentID              string
	Version              string
	Environment          string
	Lifecycle            string
	CacheMode            string
	CacheTTLSeconds      int64
	SemanticCacheAllowed bool
	Policies             AgentPolicies
	DataClasses          []string
	PrimaryModel         string
	PrimaryProvider      string
	FallbackProvider     string
	FallbackModel        string
	RoutingPolicy        string
}

// BudgetLimits are spend caps loaded from the registry API.
type BudgetLimits struct {
	MonthlyUSD                    float64
	MaxPromptTokensPerRequest     int64
	MaxCompletionTokensPerRequest int64
	RequestsPerMinute             int64
}

// Client loads agent runtime configuration from the registry API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]cachedConfig
	budgetCache map[string]cachedBudget
	ttl        time.Duration
}

type cachedConfig struct {
	config    AgentConfig
	expiresAt time.Time
}

type cachedBudget struct {
	limits    BudgetLimits
	expiresAt time.Time
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache:       make(map[string]cachedConfig),
		budgetCache: make(map[string]cachedBudget),
		ttl:         30 * time.Second,
	}
}

func (c *Client) GetAgentConfig(ctx context.Context, agentID, version, environment string) (AgentConfig, error) {
	if c.baseURL == "" {
		return AgentConfig{}, fmt.Errorf("registry API URL is not configured")
	}
	if environment == "" {
		environment = "dev"
	}
	key := agentID + ":" + version + ":" + environment

	c.mu.RLock()
	if entry, ok := c.cache[key]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.config, nil
	}
	c.mu.RUnlock()

	cfg, err := c.fetch(ctx, agentID, version, environment)
	if err != nil {
		for _, fallback := range []string{"staging", "dev", "production"} {
			if fallback == environment {
				continue
			}
			cfg, err = c.fetch(ctx, agentID, version, fallback)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return AgentConfig{}, err
	}

	c.mu.Lock()
	c.cache[key] = cachedConfig{config: cfg, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return cfg, nil
}

func (c *Client) GetBudget(ctx context.Context, agentID, version string) (BudgetLimits, error) {
	if c.baseURL == "" {
		return BudgetLimits{}, nil
	}
	key := agentID + ":" + version

	c.mu.RLock()
	if entry, ok := c.budgetCache[key]; ok && time.Now().Before(entry.expiresAt) {
		c.mu.RUnlock()
		return entry.limits, nil
	}
	c.mu.RUnlock()

	endpoint := c.baseURL + "/v1/agents/" + url.PathEscape(agentID) + "/budget?version=" + url.QueryEscape(version)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return BudgetLimits{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return BudgetLimits{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return BudgetLimits{}, nil
	}
	if resp.StatusCode >= 400 {
		return BudgetLimits{}, fmt.Errorf("budget lookup failed with status %d", resp.StatusCode)
	}

	var payload struct {
		MonthlyUSD                    float64 `json:"monthly_usd"`
		MaxPromptTokensPerRequest     int64   `json:"max_prompt_tokens_per_request"`
		MaxCompletionTokensPerRequest int64   `json:"max_completion_tokens_per_request"`
		RequestsPerMinute             int64   `json:"requests_per_minute"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return BudgetLimits{}, err
	}
	limits := BudgetLimits{
		MonthlyUSD:                    payload.MonthlyUSD,
		MaxPromptTokensPerRequest:     payload.MaxPromptTokensPerRequest,
		MaxCompletionTokensPerRequest: payload.MaxCompletionTokensPerRequest,
		RequestsPerMinute:             payload.RequestsPerMinute,
	}

	c.mu.Lock()
	c.budgetCache[key] = cachedBudget{limits: limits, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return limits, nil
}

func (c *Client) fetch(ctx context.Context, agentID, version, environment string) (AgentConfig, error) {
	endpoint, err := url.Parse(c.baseURL + "/v1/agents/" + url.PathEscape(agentID))
	if err != nil {
		return AgentConfig{}, err
	}
	query := endpoint.Query()
	query.Set("version", version)
	query.Set("environment", environment)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return AgentConfig{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AgentConfig{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return AgentConfig{}, fmt.Errorf("agent %s@%s (%s) not found in registry", agentID, version, environment)
	}
	if resp.StatusCode >= 400 {
		return AgentConfig{}, fmt.Errorf("registry lookup failed with status %d", resp.StatusCode)
	}

	var payload struct {
		AgentID              string   `json:"agent_id"`
		Version              string   `json:"version"`
		Environment          string   `json:"environment"`
		Lifecycle            string   `json:"lifecycle"`
		CacheMode            string   `json:"cache_mode"`
		CacheTTLSeconds      int64    `json:"cache_ttl_seconds"`
		SemanticCacheAllowed bool     `json:"semantic_cache_allowed"`
		Policies             struct {
			AllowedProviders []string `json:"allowed_providers"`
			PIIAllowed       bool     `json:"pii_allowed"`
			RequireAuditLog  bool     `json:"require_audit_log"`
		} `json:"policies"`
		DataClasses []string `json:"data_classes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return AgentConfig{}, err
	}

	cfg := AgentConfig{
		AgentID:              payload.AgentID,
		Version:              payload.Version,
		Environment:          payload.Environment,
		Lifecycle:            payload.Lifecycle,
		CacheMode:            payload.CacheMode,
		CacheTTLSeconds:      payload.CacheTTLSeconds,
		SemanticCacheAllowed: payload.SemanticCacheAllowed,
		Policies: AgentPolicies{
			AllowedProviders: payload.Policies.AllowedProviders,
			PIIAllowed:       payload.Policies.PIIAllowed,
			RequireAuditLog:  payload.Policies.RequireAuditLog,
		},
		DataClasses: payload.DataClasses,
	}

	routeEndpoint := c.baseURL + "/v1/agents/" + url.PathEscape(agentID) + "/model-route?version=" + url.QueryEscape(version)
	routeReq, err := http.NewRequestWithContext(ctx, http.MethodGet, routeEndpoint, nil)
	if err == nil {
		routeResp, err := c.httpClient.Do(routeReq)
		if err == nil {
			defer routeResp.Body.Close()
			if routeResp.StatusCode == http.StatusOK {
				var route struct {
					PrimaryProvider  string `json:"primary_provider"`
					PrimaryModel     string `json:"primary_model"`
					FallbackProvider string `json:"fallback_provider"`
					FallbackModel    string `json:"fallback_model"`
					RoutingPolicy    string `json:"routing_policy"`
				}
				if json.NewDecoder(routeResp.Body).Decode(&route) == nil {
					cfg.PrimaryProvider = route.PrimaryProvider
					cfg.PrimaryModel = route.PrimaryModel
					cfg.FallbackProvider = route.FallbackProvider
					cfg.FallbackModel = route.FallbackModel
					cfg.RoutingPolicy = route.RoutingPolicy
				}
			}
		}
	}

	if cfg.CacheMode == "" {
		cfg.CacheMode = "exact_only"
	}
	if cfg.CacheTTLSeconds == 0 {
		cfg.CacheTTLSeconds = 86400
	}
	if cfg.Policies.AllowedProviders == nil {
		cfg.Policies.AllowedProviders = []string{}
	}
	return cfg, nil
}
