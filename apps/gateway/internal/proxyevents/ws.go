package proxyevents

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsEnvelope struct {
	Type       string          `json:"type"`
	Version    int             `json:"version"`
	StreamID   string          `json:"stream_id"`
	Seq        int64           `json:"seq,omitempty"`
	ServerTime time.Time       `json:"server_time,omitempty"`
	Payload    json.RawMessage `json:"payload,omitempty"`
}

func (h *Handler) wsProxyEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	since := time.Now().UTC().Add(-5 * time.Minute)
	filter := parseListFilter(r)
	if !filter.Since.IsZero() {
		since = filter.Since
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 500
	}
	filter.Since = since
	filter.Limit = limit

	events, matched := h.store.List(filter)
	metrics := h.store.Metrics(since)
	snapshotPayload, _ := json.Marshal(map[string]any{
		"window":         "last_5m",
		"limit":          limit,
		"matched_count":  matched,
		"returned_count": len(events),
		"rows":           events,
	})
	_ = writeWSEnvelope(conn, "snapshot", snapshotPayload)
	metricsPayload, _ := json.Marshal(metrics)
	_ = writeWSEnvelope(conn, "metrics_delta", metricsPayload)

	eventsCh, unsub := h.store.Subscribe()
	defer unsub()

	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()

	var writeMu sync.Mutex
	write := func(envelope wsEnvelope) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(envelope)
	}

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case event, ok := <-eventsCh:
			if !ok {
				return
			}
			payload, _ := json.Marshal(map[string]any{"type": "row_upsert", "row": event})
			if err := write(wsEnvelope{
				Type:       "row_upsert",
				Version:    1,
				StreamID:   "live-proxy-flow",
				Seq:        event.Seq,
				ServerTime: time.Now().UTC(),
				Payload:    payload,
			}); err != nil {
				return
			}
			metricsPayload, _ := json.Marshal(h.store.Metrics(since))
			_ = write(wsEnvelope{
				Type:       "metrics_delta",
				Version:    1,
				StreamID:   "live-proxy-flow",
				ServerTime: time.Now().UTC(),
				Payload:    metricsPayload,
			})
		case <-ping.C:
			payload, _ := json.Marshal(map[string]any{
				"server_time": time.Now().UTC(),
				"connected":   true,
			})
			if err := write(wsEnvelope{
				Type:       "heartbeat",
				Version:    1,
				StreamID:   "live-proxy-flow",
				ServerTime: time.Now().UTC(),
				Payload:    payload,
			}); err != nil {
				return
			}
		}
	}
}

func writeWSEnvelope(conn *websocket.Conn, msgType string, payload json.RawMessage) error {
	return conn.WriteJSON(wsEnvelope{
		Type:       msgType,
		Version:    1,
		StreamID:   "live-proxy-flow",
		ServerTime: time.Now().UTC(),
		Payload:    payload,
	})
}
