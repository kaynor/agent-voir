package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	agentregistry "github.com/agentvoir/agentvoir/apps/gateway/internal/registry"
)

// Decision is the outcome of a policy evaluation.
type Decision struct {
	Allowed bool
	Reason  string
}

// Evaluator checks gateway requests against OPA policies.
type Evaluator interface {
	Allow(ctx context.Context, input Input) Decision
}

// Input is the OPA evaluation payload.
type Input struct {
	Agent       agentregistry.AgentConfig
	Environment string
	Provider    string
}

// OPAClient calls an Open Policy Agent server.
type OPAClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOPAClient(baseURL string) *OPAClient {
	return &OPAClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *OPAClient) Allow(ctx context.Context, input Input) Decision {
	if c == nil || c.baseURL == "" {
		return Decision{Allowed: true}
	}

	payload := map[string]any{
		"input": map[string]any{
			"agent": map[string]any{
				"lifecycle": input.Agent.Lifecycle,
				"policies":  input.Agent.Policies.OPAFormat(),
				"cache": map[string]any{
					"mode":                 input.Agent.CacheMode,
					"semanticCacheAllowed": input.Agent.SemanticCacheAllowed,
				},
				"dataClasses": input.Agent.DataClasses,
			},
			"request": map[string]any{
				"provider":        input.Provider,
				"contains_pii":    false,
				"contains_secret": false,
			},
			"environment": input.Environment,
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/v1/data/agentvoir/authz/allow",
		bytes.NewReader(body),
	)
	if err != nil {
		return Decision{Allowed: true}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Decision{Allowed: true}
	}
	defer resp.Body.Close()

	var result struct {
		Result bool `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Decision{Allowed: true}
	}
	if result.Result {
		return Decision{Allowed: true}
	}
	return Decision{
		Allowed: false,
		Reason:  fmt.Sprintf("request denied by policy for lifecycle=%s environment=%s", input.Agent.Lifecycle, input.Environment),
	}
}

// NopEvaluator allows all requests when OPA is disabled.
type NopEvaluator struct{}

func (NopEvaluator) Allow(context.Context, Input) Decision {
	return Decision{Allowed: true}
}
