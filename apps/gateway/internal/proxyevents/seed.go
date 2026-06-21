package proxyevents

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// SeedOptions controls dummy data generation for the live dashboard.
type SeedOptions struct {
	Count      int
	Spread     time.Duration
	BaseTime   time.Time
	Agents     []string
	Users      []string
}

// GenerateSeed builds realistic multi-step traces for dashboard verification.
func GenerateSeed(opts SeedOptions) []Event {
	if opts.Count <= 0 {
		opts.Count = 50
	}
	if opts.Spread <= 0 {
		opts.Spread = 5 * time.Minute
	}
	if opts.BaseTime.IsZero() {
		opts.BaseTime = time.Now().UTC()
	}
	if len(opts.Agents) == 0 {
		opts.Agents = []string{
			"research-agent",
			"customer-support-agent",
			"cache-demo-agent",
			"rate-limit-demo-agent",
			"fallback-demo-agent",
		}
	}
	if len(opts.Users) == 0 {
		opts.Users = []string{"user_42", "user_17", "user_03", "user_88", "user_12"}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var events []Event
	traces := (opts.Count / 3)
	if traces < 1 {
		traces = 1
	}

	for i := 0; i < traces; i++ {
		traceID := "trace_" + uuid.NewString()[:6]
		agent := opts.Agents[rng.Intn(len(opts.Agents))]
		user := opts.Users[rng.Intn(len(opts.Users))]
		offset := time.Duration(rng.Int63n(int64(opts.Spread)))
		base := opts.BaseTime.Add(-offset)

		scenario := rng.Intn(5)
		switch scenario {
		case 0:
			events = append(events, toolCallTrace(rng, traceID, agent, user, base)...)
		case 1:
			events = append(events, cacheHitTrace(rng, traceID, agent, user, base)...)
		case 2:
			events = append(events, policyBlockTrace(rng, traceID, agent, user, base)...)
		case 3:
			events = append(events, streamFinalTrace(rng, traceID, agent, user, base)...)
		default:
			events = append(events, simpleAnswerTrace(rng, traceID, agent, user, base)...)
		}
	}
	return events
}

func toolCallTrace(rng *rand.Rand, traceID, agent, user string, base time.Time) []Event {
	tool := "github.search_issues"
	t0 := base
	return []Event{
		{
			EventTime:     t0,
			TraceID:       traceID,
			SpanID:        "span_llm_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    200,
			Provider:      "OpenAI",
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseToolCall,
			NextAction:    "Execute " + tool,
			Tool:          tool,
			Terminal:      false,
			Tags:          []string{"tool-call", "review"},
			TokensIn:      int64(800 + rng.Intn(1200)),
			TokensOut:     0,
			DurationMS:    int64(600 + rng.Intn(900)),
			CostUSD:       0.002 + rng.Float64()*0.004,
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
		{
			EventTime:     t0.Add(450 * time.Millisecond),
			TraceID:       traceID,
			SpanID:        "span_tool_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "TOOL",
			ReqPath:       tool,
			StatusCode:    200,
			Provider:      "AgentVoir Tools",
			Model:         tool,
			ResponseType:  ResponseToolResult,
			NextAction:    "Send result to LLM",
			Tool:          tool,
			Terminal:      false,
			Tags:          []string{"tool-result"},
			TokensIn:      0,
			TokensOut:     0,
			DurationMS:    int64(200 + rng.Intn(400)),
			CostUSD:       0,
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
		{
			EventTime:     t0.Add(1200 * time.Millisecond),
			TraceID:       traceID,
			SpanID:        "span_llm_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    200,
			Provider:      "OpenAI",
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseFinalAnswer,
			NextAction:    "Return to client",
			Tool:          "—",
			Terminal:      true,
			Tags:          []string{"final-answer"},
			TokensIn:      int64(1500 + rng.Intn(1000)),
			TokensOut:     int64(300 + rng.Intn(800)),
			DurationMS:    int64(900 + rng.Intn(1200)),
			CostUSD:       0.004 + rng.Float64()*0.008,
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
	}
}

func cacheHitTrace(rng *rand.Rand, traceID, agent, user string, base time.Time) []Event {
	return []Event{
		{
			EventTime:     base,
			TraceID:       traceID,
			SpanID:        "span_cache_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    200,
			Provider:      "OpenAI",
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseCacheResponse,
			NextAction:    "Return cached answer",
			Tool:          "—",
			Terminal:      true,
			Tags:          []string{"cache-hit"},
			TokensIn:      int64(200 + rng.Intn(400)),
			TokensOut:     int64(100 + rng.Intn(200)),
			DurationMS:    int64(5 + rng.Intn(20)),
			CostUSD:       0,
			CacheStatus:   "hit",
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
	}
}

func policyBlockTrace(rng *rand.Rand, traceID, agent, user string, base time.Time) []Event {
	return []Event{
		{
			EventTime:     base,
			TraceID:       traceID,
			SpanID:        "span_policy_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    403,
			Provider:      "OpenAI",
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseGuardrailBlock,
			NextAction:    "Policy denied",
			Tool:          "—",
			Terminal:      true,
			Tags:          []string{"policy-blocked", "error"},
			TokensIn:      int64(400 + rng.Intn(600)),
			TokensOut:     0,
			DurationMS:    int64(20 + rng.Intn(80)),
			CostUSD:       0,
			ErrorCode:     "policy_denied",
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
	}
}

func streamFinalTrace(rng *rand.Rand, traceID, agent, user string, base time.Time) []Event {
	return []Event{
		{
			EventTime:     base,
			TraceID:       traceID,
			SpanID:        "span_stream_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    200,
			Provider:      "OpenAI",
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseStreamFinal,
			NextAction:    "Return to client",
			Tool:          "—",
			Terminal:      true,
			Tags:          []string{"streaming", "final-answer"},
			TokensIn:      int64(600 + rng.Intn(800)),
			TokensOut:     int64(2000 + rng.Intn(2000)),
			DurationMS:    int64(3000 + rng.Intn(4000)),
			CostUSD:       0.008 + rng.Float64()*0.015,
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
	}
}

func simpleAnswerTrace(rng *rand.Rand, traceID, agent, user string, base time.Time) []Event {
	cost := 0.001 + rng.Float64()*0.012
	tags := []string{"final-answer"}
	if cost > 0.05 {
		tags = append(tags, "high-cost")
	}
	return []Event{
		{
			EventTime:     base,
			TraceID:       traceID,
			SpanID:        "span_llm_" + uuid.NewString()[:4],
			AgentID:       agent,
			UserID:        user,
			ReqMethod:     "POST",
			ReqPath:       "/v1/chat/completions",
			StatusCode:    200,
			Provider:      pickProvider(rng),
			Model:         "gpt-4.1-mini",
			ResponseType:  ResponseFinalAnswer,
			NextAction:    "Return to client",
			Tool:          "—",
			Terminal:      true,
			Tags:          tags,
			TokensIn:      int64(300 + rng.Intn(1500)),
			TokensOut:     int64(100 + rng.Intn(900)),
			DurationMS:    int64(400 + rng.Intn(2500)),
			CostUSD:       cost,
			OTelStatus:    "exported",
			DatadogStatus: "indexed",
		},
	}
}

func pickProvider(rng *rand.Rand) string {
	providers := []string{"OpenAI", "Anthropic", "Mock"}
	return providers[rng.Intn(len(providers))]
}

// FormatEventTime returns HH:MM:SS.mmm for grid display.
func FormatEventTime(t time.Time) string {
	return fmt.Sprintf("%s.%03d", t.Format("15:04:05"), t.Nanosecond()/1e6)
}
