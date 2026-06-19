package usage

import (
	"time"

	"github.com/agentvoir/agentvoir/services/token-accounting/internal/pricing"
	"github.com/google/uuid"
)

// Event captures token usage and cost for a single gateway or agent request.
type Event struct {
	EventTime        time.Time `json:"event_time"`
	TraceID          string    `json:"trace_id"`
	TenantID         string    `json:"tenant_id"`
	AgentID          string    `json:"agent_id"`
	AgentVersion     string    `json:"agent_version"`
	UserID           string    `json:"user_id"`
	Provider         string    `json:"provider"`
	Model            string    `json:"model"`
	CacheStatus      string    `json:"cache_status"`
	PromptTokens     uint64    `json:"prompt_tokens"`
	CompletionTokens uint64    `json:"completion_tokens"`
	CachedTokens     uint64    `json:"cached_tokens"`
	CostUSD          float64   `json:"cost_usd"`
	LatencyMS        uint64    `json:"latency_ms"`
	StatusCode       uint16    `json:"status_code"`
	ErrorCode        string    `json:"error_code"`
}

// IngestRequest is the JSON body for POST /v1/usage-events.
type IngestRequest struct {
	EventTime        *time.Time `json:"event_time,omitempty"`
	TraceID          string     `json:"trace_id"`
	TenantID         string     `json:"tenant_id"`
	AgentID          string     `json:"agent_id"`
	AgentVersion     string     `json:"agent_version"`
	UserID           string     `json:"user_id"`
	Provider         string     `json:"provider"`
	Model            string     `json:"model"`
	CacheStatus      string     `json:"cache_status"`
	PromptTokens     uint64     `json:"prompt_tokens"`
	CompletionTokens uint64     `json:"completion_tokens"`
	CachedTokens     uint64     `json:"cached_tokens"`
	CostUSD          float64    `json:"cost_usd"`
	LatencyMS        uint64     `json:"latency_ms"`
	StatusCode       uint16     `json:"status_code"`
	ErrorCode        string     `json:"error_code"`
}

// ListFilter narrows usage event queries.
type ListFilter struct {
	TenantID string
	AgentID  string
	Limit    int
}

// Normalize converts an ingest request into a persisted event with defaults applied.
func (req IngestRequest) Normalize() (Event, string) {
	if req.AgentID == "" {
		return Event{}, "agent_id is required"
	}

	event := Event{
		TraceID:          req.TraceID,
		TenantID:         req.TenantID,
		AgentID:          req.AgentID,
		AgentVersion:     req.AgentVersion,
		UserID:           req.UserID,
		Provider:         req.Provider,
		Model:            req.Model,
		CacheStatus:      req.CacheStatus,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		CachedTokens:     req.CachedTokens,
		CostUSD:          req.CostUSD,
		LatencyMS:        req.LatencyMS,
		StatusCode:       req.StatusCode,
		ErrorCode:        req.ErrorCode,
	}
	if req.EventTime != nil {
		event.EventTime = req.EventTime.UTC()
	} else {
		event.EventTime = time.Now().UTC()
	}
	if event.TraceID == "" {
		event.TraceID = uuid.NewString()
	}
	if event.TenantID == "" {
		event.TenantID = "default"
	}
	if event.AgentVersion == "" {
		event.AgentVersion = "0.1.0"
	}
	if event.StatusCode == 0 {
		event.StatusCode = 200
	}
	if event.CostUSD == 0 && event.Model != "" {
		event.CostUSD = pricing.ComputeCostUSD(event.Model, event.PromptTokens, event.CompletionTokens)
	}
	return event, ""
}
