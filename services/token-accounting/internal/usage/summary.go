package usage

import "time"

// SummaryRollup aggregates usage over a time window.
type SummaryRollup struct {
	Period           string  `json:"period"`
	AgentID          string  `json:"agent_id,omitempty"`
	TenantID         string  `json:"tenant_id,omitempty"`
	EventCount       int     `json:"event_count"`
	PromptTokens     uint64  `json:"prompt_tokens"`
	CompletionTokens uint64  `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	CacheHitRate     float64 `json:"cache_hit_rate"`
}

// SummaryFilter selects events for rollup aggregation.
type SummaryFilter struct {
	Period   string
	AgentID  string
	TenantID string
	Since    time.Time
}
