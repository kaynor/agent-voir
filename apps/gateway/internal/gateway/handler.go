package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/middleware"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/openai"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/providers"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/usage"
	"github.com/google/uuid"
)

// Handler serves OpenAI-compatible gateway endpoints.
type Handler struct {
	config   Config
	cache    cache.Store
	registry *providers.Registry
	usage    usage.Recorder
}

func NewHandler(config Config, cacheStore cache.Store, registry *providers.Registry, usageRecorder usage.Recorder) *Handler {
	if usageRecorder == nil {
		usageRecorder = usage.NopRecorder{}
	}
	return &Handler{
		config:   config,
		cache:    cacheStore,
		registry: registry,
		usage:    usageRecorder,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat/completions", h.chatCompletions)
	mux.HandleFunc("GET /v1/models", h.listModels)
}

func (h *Handler) chatCompletions(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	if !authorize(r, h.config.APIKey) {
		writeOpenAIError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
		return
	}

	agentID := strings.TrimSpace(r.Header.Get(middleware.HeaderAgentID))
	if agentID == "" {
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "x-agent-id header is required")
		return
	}
	agentVersion := strings.TrimSpace(r.Header.Get(middleware.HeaderAgentVersion))
	if agentVersion == "" {
		agentVersion = "0.1.0"
	}

	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.recordUsage(r, started, "", agentID, agentVersion, "", "", "bypass", 0, 0, 0, http.StatusBadRequest, "invalid_json")
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON body")
		return
	}
	if req.Model == "" {
		h.recordUsage(r, started, "", agentID, agentVersion, "", "", "bypass", 0, 0, 0, http.StatusBadRequest, "missing_model")
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	if len(req.Messages) == 0 {
		h.recordUsage(r, started, "", agentID, agentVersion, "", req.Model, "bypass", 0, 0, 0, http.StatusBadRequest, "missing_messages")
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "messages are required")
		return
	}
	if req.Stream {
		h.streamChatCompletions(w, r, started, agentID, agentVersion, req)
		return
	}

	traceID := uuid.NewString()
	cacheStatus := "miss"
	var raw []byte
	var providerResp *providers.ChatResponse

	if h.config.CacheReadEnabled() {
		key := cache.ExactKey(agentID, agentVersion, req)
		if entry, err := h.cache.Get(r.Context(), key); err == nil && entry != nil {
			cacheStatus = "hit"
			raw = entry.Value
			w.Header().Set(middleware.HeaderCacheStatus, cacheStatus)
			h.writeOperationalHeaders(w, agentID, agentVersion, "cache", req.Model, 0, 0, 0, traceID)
			h.recordUsage(r, started, traceID, agentID, agentVersion, "cache", req.Model, cacheStatus, 0, 0, 0, http.StatusOK, "")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(raw)
			return
		}
	}

	provider := h.registry.Resolve(req.Model)
	providerReq := providers.ChatRequest{
		Model:    req.Model,
		Messages: toProviderMessages(req.Messages),
	}

	var err error
	providerResp, err = provider.ChatCompletions(r.Context(), providerReq)
	if err != nil {
		h.recordUsage(r, started, traceID, agentID, agentVersion, "", req.Model, cacheStatus, 0, 0, 0, http.StatusBadGateway, "upstream_error")
		writeOpenAIError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	raw = providerResp.Raw

	if h.config.CacheWriteEnabled() {
		key := cache.ExactKey(agentID, agentVersion, req)
		_ = h.cache.Set(r.Context(), cache.Entry{
			Key:         key,
			Value:       raw,
			TTLSeconds:  h.config.CacheTTLSeconds,
			CacheStatus: cacheStatus,
			AgentID:     agentID,
		})
	}

	h.writeOperationalHeaders(
		w,
		agentID,
		agentVersion,
		providerResp.Provider,
		providerResp.Model,
		providerResp.PromptTokens,
		providerResp.CompletionTokens,
		providerResp.CostUSD,
		traceID,
	)
	w.Header().Set(middleware.HeaderCacheStatus, cacheStatus)
	h.recordUsage(
		r,
		started,
		traceID,
		agentID,
		agentVersion,
		providerResp.Provider,
		providerResp.Model,
		cacheStatus,
		providerResp.PromptTokens,
		providerResp.CompletionTokens,
		providerResp.CostUSD,
		http.StatusOK,
		"",
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func (h *Handler) streamChatCompletions(
	w http.ResponseWriter,
	r *http.Request,
	started time.Time,
	agentID, agentVersion string,
	req openai.ChatCompletionRequest,
) {
	provider := h.registry.Resolve(req.Model)
	providerReq := providers.ChatRequest{
		Model:    req.Model,
		Messages: toProviderMessages(req.Messages),
	}

	resp, err := provider.ChatCompletions(r.Context(), providerReq)
	if err != nil {
		h.recordUsage(r, started, "", agentID, agentVersion, "", req.Model, "bypass", 0, 0, 0, http.StatusBadGateway, "upstream_error")
		writeOpenAIError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}

	var completion openai.ChatCompletionResponse
	if err := json.Unmarshal(resp.Raw, &completion); err != nil {
		h.recordUsage(r, started, "", agentID, agentVersion, "", req.Model, "bypass", 0, 0, 0, http.StatusInternalServerError, "decode_error")
		writeOpenAIError(w, http.StatusInternalServerError, "server_error", "failed to decode provider response")
		return
	}

	content := ""
	if len(completion.Choices) > 0 {
		content = completion.Choices[0].Message.Content
	}

	traceID := uuid.NewString()
	h.writeOperationalHeaders(
		w,
		agentID,
		agentVersion,
		resp.Provider,
		resp.Model,
		resp.PromptTokens,
		resp.CompletionTokens,
		resp.CostUSD,
		traceID,
	)
	w.Header().Set(middleware.HeaderCacheStatus, "bypass")
	h.recordUsage(
		r,
		started,
		traceID,
		agentID,
		agentVersion,
		resp.Provider,
		resp.Model,
		"bypass",
		resp.PromptTokens,
		resp.CompletionTokens,
		resp.CostUSD,
		http.StatusOK,
		"",
	)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeOpenAIError(w, http.StatusInternalServerError, "server_error", "streaming not supported")
		return
	}

	chunkID := completion.ID
	created := completion.Created
	first := openai.ChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []openai.ChatChunkChoice{{
			Index: 0,
			Delta: openai.ChatMessageDelta{Role: "assistant"},
		}},
	}
	writeSSE(w, first)
	flusher.Flush()

	second := openai.ChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []openai.ChatChunkChoice{{
			Index: 0,
			Delta: openai.ChatMessageDelta{Content: content},
		}},
	}
	writeSSE(w, second)
	flusher.Flush()

	stop := "stop"
	final := openai.ChatCompletionChunk{
		ID:      chunkID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []openai.ChatChunkChoice{{
			Index:        0,
			FinishReason: &stop,
		}},
	}
	writeSSE(w, final)
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (h *Handler) listModels(w http.ResponseWriter, r *http.Request) {
	if !authorize(r, h.config.APIKey) {
		writeOpenAIError(w, http.StatusUnauthorized, "invalid_api_key", "invalid API key")
		return
	}

	payload := map[string]any{
		"object": "list",
		"data": []map[string]any{
			{"id": "gpt-4.1-mini", "object": "model", "created": time.Now().Unix(), "owned_by": "agentvoir"},
			{"id": "gpt-4.1", "object": "model", "created": time.Now().Unix(), "owned_by": "agentvoir"},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *Handler) writeOperationalHeaders(
	w http.ResponseWriter,
	agentID, agentVersion, provider, model string,
	promptTokens, completionTokens int64,
	costUSD float64,
	traceID string,
) {
	w.Header().Set(middleware.HeaderAgentID, agentID)
	w.Header().Set(middleware.HeaderAgentVersion, agentVersion)
	w.Header().Set("x-model-provider", provider)
	w.Header().Set("x-model-used", model)
	w.Header().Set("x-token-input", fmt.Sprintf("%d", promptTokens))
	w.Header().Set("x-token-output", fmt.Sprintf("%d", completionTokens))
	w.Header().Set("x-cost-usd", fmt.Sprintf("%.4f", costUSD))
	w.Header().Set(middleware.HeaderTraceID, traceID)
}

func (h *Handler) recordUsage(
	r *http.Request,
	started time.Time,
	traceID, agentID, agentVersion, provider, model, cacheStatus string,
	promptTokens, completionTokens int64,
	costUSD float64,
	statusCode int,
	errorCode string,
) {
	if agentID == "" {
		return
	}
	if traceID == "" {
		traceID = uuid.NewString()
	}

	tenantID := strings.TrimSpace(r.Header.Get(middleware.HeaderTenantID))
	if tenantID == "" {
		tenantID = "default"
	}

	h.usage.Record(usage.Event{
		EventTime:        time.Now().UTC(),
		TraceID:          traceID,
		TenantID:         tenantID,
		AgentID:          agentID,
		AgentVersion:     agentVersion,
		UserID:           strings.TrimSpace(r.Header.Get(middleware.HeaderUserID)),
		Provider:         provider,
		Model:            model,
		CacheStatus:      cacheStatus,
		PromptTokens:     uint64(maxInt64(promptTokens, 0)),
		CompletionTokens: uint64(maxInt64(completionTokens, 0)),
		CostUSD:          costUSD,
		LatencyMS:        uint64(time.Since(started).Milliseconds()),
		StatusCode:       uint16(statusCode),
		ErrorCode:        errorCode,
	})
}

func maxInt64(value int64, floor int64) int64 {
	if value < floor {
		return floor
	}
	return value
}

func writeOpenAIError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(openai.ErrorResponse{
		Error: openai.APIError{
			Message: message,
			Type:    errType,
		},
	})
}

func writeSSE(w http.ResponseWriter, payload any) {
	data, _ := json.Marshal(payload)
	_, _ = fmt.Fprintf(w, "data: %s\n\n", data)
}

func toProviderMessages(messages []openai.ChatMessage) []providers.ChatMessage {
	out := make([]providers.ChatMessage, len(messages))
	for i, msg := range messages {
		out[i] = providers.ChatMessage{Role: msg.Role, Content: msg.Content}
	}
	return out
}
