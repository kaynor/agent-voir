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

// AgentConfig is the runtime configuration loaded from the registry API.
type AgentConfig struct {
	AgentID              string
	Version              string
	Environment          string
	Lifecycle            string
	CacheMode            string
	CacheTTLSeconds      int64
	SemanticCacheAllowed bool
	DataClasses          []string
	PrimaryModel         string
	PrimaryProvider      string
}

// Client loads agent runtime configuration from the registry API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]cachedConfig
	ttl        time.Duration
}

type cachedConfig struct {
	config    AgentConfig
	expiresAt time.Time
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache: make(map[string]cachedConfig),
		ttl:   30 * time.Second,
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
		return AgentConfig{}, err
	}

	c.mu.Lock()
	c.cache[key] = cachedConfig{config: cfg, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return cfg, nil
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
		DataClasses          []string `json:"data_classes"`
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
		DataClasses:          payload.DataClasses,
	}

	routeEndpoint := c.baseURL + "/v1/agents/" + url.PathEscape(agentID) + "/model-route?version=" + url.QueryEscape(version)
	routeReq, err := http.NewRequestWithContext(ctx, http.MethodGet, routeEndpoint, nil)
	if err == nil {
		routeResp, err := c.httpClient.Do(routeReq)
		if err == nil {
			defer routeResp.Body.Close()
			if routeResp.StatusCode == http.StatusOK {
				var route struct {
					PrimaryProvider string `json:"primary_provider"`
					PrimaryModel    string `json:"primary_model"`
				}
				if json.NewDecoder(routeResp.Body).Decode(&route) == nil {
					cfg.PrimaryProvider = route.PrimaryProvider
					cfg.PrimaryModel = route.PrimaryModel
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
	return cfg, nil
}
