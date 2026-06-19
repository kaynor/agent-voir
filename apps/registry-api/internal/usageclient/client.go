package usageclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client reads usage summaries from token-accounting.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type MonthlySummary struct {
	CostUSD    float64 `json:"cost_usd"`
	EventCount int     `json:"event_count"`
}

func (c *Client) GetMonthlySummary(ctx context.Context, tenantID, agentID string) (MonthlySummary, error) {
	if c == nil || c.baseURL == "" {
		return MonthlySummary{}, nil
	}
	endpoint, err := url.Parse(c.baseURL + "/v1/usage-events/summary")
	if err != nil {
		return MonthlySummary{}, err
	}
	query := endpoint.Query()
	query.Set("period", "monthly")
	if agentID != "" {
		query.Set("agent_id", agentID)
	}
	if tenantID != "" {
		query.Set("tenant_id", tenantID)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return MonthlySummary{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MonthlySummary{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return MonthlySummary{}, fmt.Errorf("usage summary failed with status %d", resp.StatusCode)
	}
	var summary MonthlySummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return MonthlySummary{}, err
	}
	return summary, nil
}
