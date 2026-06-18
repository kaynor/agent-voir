package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const createUsageEventsTableSQL = `
CREATE TABLE IF NOT EXISTS usage_events
(
  event_time DateTime64(3),
  trace_id String,
  tenant_id String,
  agent_id String,
  agent_version String,
  user_id String,
  provider String,
  model String,
  cache_status LowCardinality(String),
  prompt_tokens UInt64,
  completion_tokens UInt64,
  cached_tokens UInt64,
  cost_usd Float64,
  latency_ms UInt64,
  status_code UInt16,
  error_code String
)
ENGINE = MergeTree
PARTITION BY toDate(event_time)
ORDER BY (tenant_id, agent_id, event_time)
`

// ClickHouseStore persists usage events through the ClickHouse HTTP interface.
type ClickHouseStore struct {
	baseURL    string
	database   string
	httpClient *http.Client
}

func NewClickHouseStore(dsn string) (*ClickHouseStore, error) {
	baseURL, database, err := parseClickHouseDSN(dsn)
	if err != nil {
		return nil, err
	}

	store := &ClickHouseStore{
		baseURL:  strings.TrimRight(baseURL, "/"),
		database: database,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
	if err := store.EnsureSchema(context.Background()); err != nil {
		return nil, err
	}
	return store, nil
}

func parseClickHouseDSN(dsn string) (string, string, error) {
	if dsn == "" {
		return "", "", fmt.Errorf("clickhouse dsn is required")
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", "", fmt.Errorf("parse clickhouse dsn: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("clickhouse dsn must include scheme and host")
	}
	database := strings.TrimPrefix(parsed.Path, "/")
	if database == "" {
		database = "default"
	}
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), database, nil
}

func (s *ClickHouseStore) EnsureSchema(ctx context.Context) error {
	return s.exec(ctx, createUsageEventsTableSQL)
}

func (s *ClickHouseStore) Insert(ctx context.Context, event Event) error {
	row := map[string]any{
		"event_time":        event.EventTime.UTC().Format("2006-01-02 15:04:05.000"),
		"trace_id":          event.TraceID,
		"tenant_id":         event.TenantID,
		"agent_id":          event.AgentID,
		"agent_version":     event.AgentVersion,
		"user_id":           event.UserID,
		"provider":          event.Provider,
		"model":             event.Model,
		"cache_status":      event.CacheStatus,
		"prompt_tokens":     event.PromptTokens,
		"completion_tokens": event.CompletionTokens,
		"cached_tokens":     event.CachedTokens,
		"cost_usd":          event.CostUSD,
		"latency_ms":        event.LatencyMS,
		"status_code":       event.StatusCode,
		"error_code":        event.ErrorCode,
	}
	payload, err := json.Marshal(row)
	if err != nil {
		return fmt.Errorf("marshal usage event: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s.usage_events FORMAT JSONEachRow", s.database)
	return s.post(ctx, query, payload)
}

func (s *ClickHouseStore) List(ctx context.Context, filter ListFilter) ([]Event, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	var clauses []string
	if filter.TenantID != "" {
		clauses = append(clauses, fmt.Sprintf("tenant_id = '%s'", escapeSQL(filter.TenantID)))
	}
	if filter.AgentID != "" {
		clauses = append(clauses, fmt.Sprintf("agent_id = '%s'", escapeSQL(filter.AgentID)))
	}

	where := ""
	if len(clauses) > 0 {
		where = "WHERE " + strings.Join(clauses, " AND ")
	}

	query := fmt.Sprintf(
		"SELECT event_time, trace_id, tenant_id, agent_id, agent_version, user_id, provider, model, cache_status, prompt_tokens, completion_tokens, cached_tokens, cost_usd, latency_ms, status_code, error_code FROM %s.usage_events %s ORDER BY event_time DESC LIMIT %d FORMAT JSON",
		s.database,
		where,
		limit,
	)

	body, err := s.query(ctx, query)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []struct {
			EventTime        string  `json:"event_time"`
			TraceID          string  `json:"trace_id"`
			TenantID         string  `json:"tenant_id"`
			AgentID          string  `json:"agent_id"`
			AgentVersion     string  `json:"agent_version"`
			UserID           string  `json:"user_id"`
			Provider         string  `json:"provider"`
			Model            string  `json:"model"`
			CacheStatus      string  `json:"cache_status"`
			PromptTokens     uint64  `json:"prompt_tokens,string"`
			CompletionTokens uint64  `json:"completion_tokens,string"`
			CachedTokens     uint64  `json:"cached_tokens,string"`
			CostUSD          float64 `json:"cost_usd"`
			LatencyMS        uint64  `json:"latency_ms,string"`
			StatusCode       uint16  `json:"status_code"`
			ErrorCode        string  `json:"error_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("decode clickhouse response: %w", err)
	}

	events := make([]Event, 0, len(response.Data))
	for _, row := range response.Data {
		eventTime, err := time.Parse("2006-01-02 15:04:05.000", row.EventTime)
		if err != nil {
			eventTime, err = time.Parse(time.RFC3339Nano, row.EventTime)
			if err != nil {
				return nil, fmt.Errorf("parse event_time %q: %w", row.EventTime, err)
			}
		}
		events = append(events, Event{
			EventTime:        eventTime.UTC(),
			TraceID:          row.TraceID,
			TenantID:         row.TenantID,
			AgentID:          row.AgentID,
			AgentVersion:     row.AgentVersion,
			UserID:           row.UserID,
			Provider:         row.Provider,
			Model:            row.Model,
			CacheStatus:      row.CacheStatus,
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			CachedTokens:     row.CachedTokens,
			CostUSD:          row.CostUSD,
			LatencyMS:        row.LatencyMS,
			StatusCode:       row.StatusCode,
			ErrorCode:        row.ErrorCode,
		})
	}
	return events, nil
}

func (s *ClickHouseStore) exec(ctx context.Context, query string) error {
	endpoint := fmt.Sprintf("%s/?query=%s", s.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(""))
	if err != nil {
		return fmt.Errorf("create clickhouse request: %w", err)
	}
	req.ContentLength = 0

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse exec: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read clickhouse response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("clickhouse exec failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *ClickHouseStore) post(ctx context.Context, query string, body []byte) error {
	endpoint := fmt.Sprintf("%s/?query=%s", s.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create clickhouse request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("clickhouse insert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clickhouse insert failed: %s", strings.TrimSpace(string(responseBody)))
	}
	return nil
}

func (s *ClickHouseStore) query(ctx context.Context, query string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/?query=%s", s.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create clickhouse request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("clickhouse query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read clickhouse response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("clickhouse query failed: %s", strings.TrimSpace(string(body)))
	}
	return body, nil
}

func escapeSQL(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
