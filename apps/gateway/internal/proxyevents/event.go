package proxyevents

import "time"

// ResponseType classifies a proxy event row for the Live Proxy Flow grid.
type ResponseType string

const (
	ResponseToolCall      ResponseType = "TOOL_CALL"
	ResponseToolResult    ResponseType = "TOOL_RESULT"
	ResponseFinalAnswer   ResponseType = "FINAL_ANSWER"
	ResponseStreamFinal   ResponseType = "STREAM_FINAL"
	ResponseCacheResponse ResponseType = "CACHE_RESPONSE"
	ResponseGuardrailBlock ResponseType = "GUARDRAIL_BLOCK"
)

// Event is a lightweight summary row for the operations dashboard grid.
type Event struct {
	Seq           int64          `json:"seq"`
	EventTime     time.Time      `json:"event_time"`
	TraceID       string         `json:"trace_id"`
	SpanID        string         `json:"span_id"`
	AgentID       string         `json:"agent_id"`
	UserID        string         `json:"user_id"`
	ReqMethod     string         `json:"req_method"`
	ReqPath       string         `json:"req_path"`
	StatusCode    int            `json:"status_code"`
	Provider      string         `json:"provider"`
	Model         string         `json:"model"`
	ResponseType  ResponseType   `json:"response_type"`
	NextAction    string         `json:"next_action"`
	Tool          string         `json:"tool"`
	Terminal      bool           `json:"terminal"`
	Tags          []string       `json:"tags"`
	TokensIn      int64          `json:"tokens_in"`
	TokensOut     int64          `json:"tokens_out"`
	DurationMS    int64          `json:"duration_ms"`
	CostUSD       float64        `json:"cost_usd"`
	OTelStatus    string         `json:"otel_status"`
	DatadogStatus string         `json:"datadog_status"`
	CacheStatus   string         `json:"cache_status,omitempty"`
	ErrorCode     string         `json:"error_code,omitempty"`
}

// FlowStep is one step in a trace call-flow timeline.
type FlowStep struct {
	Step         int          `json:"step"`
	SpanID       string       `json:"span_id"`
	Kind         string       `json:"kind"`
	ResponseType ResponseType `json:"response_type,omitempty"`
	Status       string       `json:"status"`
	DurationMS   int64        `json:"duration_ms,omitempty"`
	NextAction   string       `json:"next_action,omitempty"`
	Tool         string       `json:"tool,omitempty"`
}

// TraceDetail is returned when a grid row is selected.
type TraceDetail struct {
	TraceID    string     `json:"trace_id"`
	AgentID    string     `json:"agent_id"`
	UserID     string     `json:"user_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	DurationMS int64      `json:"duration_ms"`
	CostUSD    float64    `json:"cost_usd"`
	Tags       []string   `json:"tags"`
	Steps      []FlowStep `json:"steps"`
	ToolCall   *ToolCall  `json:"tool_call,omitempty"`
}

// ToolCall holds tool invocation details for the drilldown panel.
type ToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ListFilter query parameters for GET /v1/proxy-events.
type ListFilter struct {
	Since        time.Time
	Until        time.Time
	Limit        int
	AgentID      string
	ResponseType ResponseType
	StatusMin    int
	Tag          string
	TraceID      string
}

// MetricsSnapshot aggregates KPI cards for a time window.
type MetricsSnapshot struct {
	Window         string  `json:"window"`
	RequestsTotal  int     `json:"requests_total"`
	Errors         int     `json:"errors"`
	ToolCalls      int     `json:"tool_calls"`
	FinalAnswers   int     `json:"final_answers"`
	Blocked        int     `json:"blocked"`
	CacheHits      int     `json:"cache_hits"`
	TokensIn       int64   `json:"tokens_in"`
	TokensOut      int64   `json:"tokens_out"`
	ActiveRequests int     `json:"active_requests"`
	CostUSD        float64 `json:"cost_usd"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	P50LatencyMS   int64   `json:"p50_latency_ms"`
	P95LatencyMS   int64   `json:"p95_latency_ms"`
	P99LatencyMS   int64   `json:"p99_latency_ms"`
}

func (e Event) ReqResp() string {
	status := "OK"
	if e.StatusCode >= 400 {
		status = httpStatusText(e.StatusCode)
	}
	return e.ReqMethod + " " + e.ReqPath + " → " + status
}

func httpStatusText(code int) string {
	switch code {
	case 403:
		return "403 Forbidden"
	case 429:
		return "429 Too Many Requests"
	case 200:
		return "200 OK"
	default:
		if code >= 400 {
			return "error"
		}
		return "200 OK"
	}
}
