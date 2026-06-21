package proxyevents

import (
	"sort"
	"sync"
	"time"
)

// Store persists recent proxy events for the live dashboard.
type Store interface {
	Insert(event Event) (Event, error)
	InsertBatch(events []Event) ([]Event, error)
	List(filter ListFilter) (events []Event, matchedCount int)
	GetTrace(traceID string) (*TraceDetail, error)
	Reset()
	Subscribe() (<-chan Event, func())
	Metrics(since time.Time) MetricsSnapshot
	Count() int
}

// MemoryStore is an in-process ring buffer with pub/sub for WebSocket fanout.
type MemoryStore struct {
	mu       sync.RWMutex
	events   []Event
	seq      int64
	maxRows  int
	subs     map[int]chan Event
	nextSub  int
}

func NewMemoryStore(maxRows int) *MemoryStore {
	if maxRows <= 0 {
		maxRows = 5000
	}
	return &MemoryStore{
		maxRows: maxRows,
		subs:    make(map[int]chan Event),
	}
}

func (s *MemoryStore) Insert(event Event) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.insertLocked(event), nil
}

func (s *MemoryStore) InsertBatch(events []Event) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Event, 0, len(events))
	for _, event := range events {
		out = append(out, s.insertLocked(event))
	}
	return out, nil
}

func (s *MemoryStore) insertLocked(event Event) Event {
	s.seq++
	event.Seq = s.seq
	if event.EventTime.IsZero() {
		event.EventTime = time.Now().UTC()
	}
	s.events = append(s.events, event)
	if len(s.events) > s.maxRows {
		s.events = s.events[len(s.events)-s.maxRows:]
	}
	for _, ch := range s.subs {
		select {
		case ch <- event:
		default:
		}
	}
	return event
}

func (s *MemoryStore) List(filter ListFilter) ([]Event, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := filter.Limit
	if limit <= 0 {
		limit = 500
	}

	matches := make([]Event, 0)
	for i := len(s.events) - 1; i >= 0; i-- {
		event := s.events[i]
		if !filter.Since.IsZero() && event.EventTime.Before(filter.Since) {
			continue
		}
		if !filter.Until.IsZero() && event.EventTime.After(filter.Until) {
			continue
		}
		if filter.AgentID != "" && event.AgentID != filter.AgentID {
			continue
		}
		if filter.ResponseType != "" && event.ResponseType != filter.ResponseType {
			continue
		}
		if filter.StatusMin > 0 && event.StatusCode < filter.StatusMin {
			continue
		}
		if filter.TraceID != "" && event.TraceID != filter.TraceID {
			continue
		}
		if filter.Tag != "" && !hasTag(event.Tags, filter.Tag) {
			continue
		}
		matches = append(matches, event)
	}

	matchedCount := len(matches)
	if len(matches) > limit {
		matches = matches[:limit]
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].EventTime.After(matches[j].EventTime)
	})
	return matches, matchedCount
}

func (s *MemoryStore) GetTrace(traceID string) (*TraceDetail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rows []Event
	for _, event := range s.events {
		if event.TraceID == traceID {
			rows = append(rows, event)
		}
	}
	if len(rows) == 0 {
		return nil, nil
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].EventTime.Before(rows[j].EventTime)
	})

	detail := &TraceDetail{
		TraceID:   traceID,
		AgentID:   rows[0].AgentID,
		UserID:    rows[0].UserID,
		StartedAt: rows[0].EventTime,
		Status:    "complete",
		Tags:      uniqueTags(rows),
	}
	var totalCost float64
	var totalDuration int64
	for i, row := range rows {
		totalCost += row.CostUSD
		totalDuration += row.DurationMS
		kind := stepKind(row)
		detail.Steps = append(detail.Steps, FlowStep{
			Step:         i + 1,
			SpanID:       row.SpanID,
			Kind:         kind,
			ResponseType: row.ResponseType,
			Status:       "complete",
			DurationMS:   row.DurationMS,
			NextAction:   row.NextAction,
			Tool:         row.Tool,
		})
		if row.ResponseType == ResponseToolCall && detail.ToolCall == nil {
			detail.ToolCall = &ToolCall{
				Name: row.Tool,
				Arguments: map[string]any{
					"repo":  "agentvoir",
					"query": "open bug",
					"limit": 10,
				},
			}
		}
		if !row.Terminal {
			detail.Status = "in_progress"
		}
	}
	detail.DurationMS = totalDuration
	detail.CostUSD = totalCost
	if detail.Status != "in_progress" {
		detail.Status = "complete"
	}
	return detail, nil
}

func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = nil
	s.seq = 0
}

func (s *MemoryStore) Subscribe() (<-chan Event, func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextSub
	s.nextSub++
	ch := make(chan Event, 256)
	s.subs[id] = ch
	unsub := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if sub, ok := s.subs[id]; ok {
			delete(s.subs, id)
			close(sub)
		}
	}
	return ch, unsub
}

func (s *MemoryStore) Metrics(since time.Time) MetricsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := MetricsSnapshot{Window: "last_5m"}
	var latencies []int64
	traceTerminal := make(map[string]bool)

	for _, event := range s.events {
		if !since.IsZero() && event.EventTime.Before(since) {
			continue
		}
		snap.RequestsTotal++
		snap.TokensIn += event.TokensIn
		snap.TokensOut += event.TokensOut
		snap.CostUSD += event.CostUSD
		latencies = append(latencies, event.DurationMS)

		if event.StatusCode >= 400 {
			snap.Errors++
		}
		switch event.ResponseType {
		case ResponseToolCall:
			snap.ToolCalls++
		case ResponseFinalAnswer, ResponseStreamFinal:
			snap.FinalAnswers++
		case ResponseGuardrailBlock:
			snap.Blocked++
		case ResponseCacheResponse:
			snap.CacheHits++
		}
		if event.CacheStatus == "hit" {
			snap.CacheHits++
		}
		if !event.Terminal {
			traceTerminal[event.TraceID] = true
		}
	}
	for _, open := range traceTerminal {
		if open {
			snap.ActiveRequests++
		}
	}
	if snap.RequestsTotal > 0 {
		snap.CacheHitRate = float64(snap.CacheHits) / float64(snap.RequestsTotal)
	}
	snap.P50LatencyMS = percentile(latencies, 50)
	snap.P95LatencyMS = percentile(latencies, 95)
	snap.P99LatencyMS = percentile(latencies, 99)
	return snap
}

func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

func hasTag(tags []string, want string) bool {
	want = trimTag(want)
	for _, tag := range tags {
		if trimTag(tag) == want {
			return true
		}
	}
	return false
}

func trimTag(tag string) string {
	for len(tag) > 0 && tag[0] == '#' {
		tag = tag[1:]
	}
	return tag
}

func uniqueTags(rows []Event) []string {
	seen := make(map[string]bool)
	var out []string
	for _, row := range rows {
		for _, tag := range row.Tags {
			t := trimTag(tag)
			if t != "" && !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}

func stepKind(row Event) string {
	switch row.ResponseType {
	case ResponseToolCall, ResponseToolResult:
		if row.ResponseType == ResponseToolCall {
			return "LLM Call"
		}
		return "Tool Execution"
	case ResponseFinalAnswer, ResponseStreamFinal:
		return "LLM Call"
	case ResponseCacheResponse:
		return "Cache Lookup"
	case ResponseGuardrailBlock:
		return "Policy Check"
	default:
		return "Request"
	}
}

func percentile(values []int64, p int) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := (len(sorted) * p) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
