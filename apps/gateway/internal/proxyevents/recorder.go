package proxyevents

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Recorder publishes proxy event rows alongside usage events.
type Recorder struct {
	store Store
}

func NewRecorder(store Store) *Recorder {
	if store == nil {
		return nil
	}
	return &Recorder{store: store}
}

// RecordFromRequest maps a gateway outcome to a dashboard row.
func (r *Recorder) RecordFromRequest(input RecordInput) {
	if r == nil || r.store == nil || input.AgentID == "" {
		return
	}
	traceID := input.TraceID
	if traceID == "" {
		traceID = uuid.NewString()
	}
	responseType, nextAction, tool := classifyOutcome(input)
	tags := buildTags(input, responseType)
	event := Event{
		EventTime:     time.Now().UTC(),
		TraceID:       traceID,
		SpanID:        "span_" + uuid.NewString()[:8],
		AgentID:       input.AgentID,
		UserID:        input.UserID,
		ReqMethod:     "POST",
		ReqPath:       "/v1/chat/completions",
		StatusCode:    input.StatusCode,
		Provider:      input.Provider,
		Model:         input.Model,
		ResponseType:  responseType,
		NextAction:    nextAction,
		Tool:          tool,
		Terminal:      true,
		Tags:          tags,
		TokensIn:      input.TokensIn,
		TokensOut:     input.TokensOut,
		DurationMS:    input.DurationMS,
		CostUSD:       input.CostUSD,
		CacheStatus:   input.CacheStatus,
		ErrorCode:     input.ErrorCode,
		OTelStatus:    "exported",
		DatadogStatus: "indexed",
	}
	if input.CacheStatus == "hit" {
		event.ResponseType = ResponseCacheResponse
		event.NextAction = "Return cached answer"
	}
	_, _ = r.store.Insert(event)
}

// RecordInput is gateway-side context for a proxy event row.
type RecordInput struct {
	TraceID     string
	AgentID     string
	UserID      string
	Provider    string
	Model       string
	CacheStatus string
	TokensIn    int64
	TokensOut   int64
	CostUSD     float64
	DurationMS  int64
	StatusCode  int
	ErrorCode   string
}

func classifyOutcome(input RecordInput) (ResponseType, string, string) {
	if input.StatusCode == http.StatusForbidden {
		return ResponseGuardrailBlock, "Policy denied", "—"
	}
	if input.StatusCode == http.StatusTooManyRequests {
		if strings.Contains(input.ErrorCode, "rate") {
			return ResponseGuardrailBlock, "Rate limit exceeded", "—"
		}
		return ResponseGuardrailBlock, "Budget exceeded", "—"
	}
	if input.StatusCode >= 400 {
		return ResponseGuardrailBlock, "Request failed", "—"
	}
	if input.CacheStatus == "hit" {
		return ResponseCacheResponse, "Return cached answer", "—"
	}
	return ResponseFinalAnswer, "Return to client", "—"
}

func buildTags(input RecordInput, responseType ResponseType) []string {
	var tags []string
	switch responseType {
	case ResponseToolCall:
		tags = append(tags, "tool-call")
	case ResponseCacheResponse:
		tags = append(tags, "cache-hit")
	case ResponseGuardrailBlock:
		tags = append(tags, "policy-blocked", "error")
	case ResponseFinalAnswer, ResponseStreamFinal:
		tags = append(tags, "final-answer")
	}
	if input.StatusCode >= 400 {
		tags = append(tags, "error")
	}
	if input.CostUSD > 0.05 {
		tags = append(tags, "high-cost")
	}
	if input.DurationMS > 3000 {
		tags = append(tags, "slow")
	}
	return tags
}
