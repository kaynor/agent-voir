package usage_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/usage"
)

func TestHTTPRecorderPostsUsageEvent(t *testing.T) {
	var (
		mu    sync.Mutex
		event usage.Event
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/usage-events" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		mu.Lock()
		defer mu.Unlock()
		if err := json.Unmarshal(body, &event); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	recorder := usage.NewHTTPRecorder(server.URL)
	recorder.Record(usage.Event{
		TraceID:          "trace-abc",
		AgentID:          "customer-support-agent",
		AgentVersion:     "0.1.0",
		Provider:         "mock",
		Model:            "gpt-4.1-mini",
		CacheStatus:      "miss",
		PromptTokens:     10,
		CompletionTokens: 5,
		LatencyMS:        120,
		StatusCode:       200,
	})

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		got := event.AgentID
		mu.Unlock()
		if got == "customer-support-agent" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for usage event")
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if event.TraceID != "trace-abc" {
		t.Fatalf("trace_id = %q", event.TraceID)
	}
}
