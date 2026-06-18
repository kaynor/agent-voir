package usage

import "time"

// Event is emitted by the gateway after each chat completion request.
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

// Recorder persists usage events asynchronously.
type Recorder interface {
	Record(event Event)
}
