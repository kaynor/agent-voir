package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/agentvoir/agentvoir/apps/gateway/internal/budget"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/cache"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/metrics"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/middleware"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/openai"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/policy"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/pricing"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/providers"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/ratelimit"
	agentregistry "github.com/agentvoir/agentvoir/apps/gateway/internal/registry"
	"github.com/agentvoir/agentvoir/apps/gateway/internal/usage"
	"github.com/google/uuid"
)

// Handler serves OpenAI-compatible gateway endpoints.
type Handler struct {
	config        Config
	cache         cache.Store
	providers     *providers.Registry
	agentRegistry *agentregistry.Client
	budget        *budget.Checker
	policy        policy.Evaluator
	rateLimit     *ratelimit.Limiter
	usage         usage.Recorder
}

func NewHandler(
	config Config,
	cacheStore cache.Store,
	providerRegistry *providers.Registry,
	agentRegistry *agentregistry.Client,
	budgetChecker *budget.Checker,
	policyEvaluator policy.Evaluator,
	rateLimiter *ratelimit.Limiter,
	usageRecorder usage.Recorder,
) *Handler {
	if usageRecorder == nil {
		usageRecorder = usage.NopRecorder{}
	}
	if policyEvaluator == nil {
		policyEvaluator = policy.NopEvaluator{}
	}
	return &Handler{
		config:        config,
		cache:         cacheStore,
		providers:     providerRegistry,
		agentRegistry: agentRegistry,
		budget:        budgetChecker,
		policy:        policyEvaluator,
		rateLimit:     rateLimiter,
		usage:         usageRecorder,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/chat/completions", h.chatCompletions)
	mux.HandleFunc("GET /v1/models", h.listModels)
}

func (h *Handler) chatCompletions(w http.ResponseWriter, r *http.Request) {
	started := time.Now()

	agentID := strings.TrimSpace(r.Header.Get(middleware.HeaderAgentID))
	if agentID == "" {
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "x-agent-id header is required")
		return
	}
	agentVersion := strings.TrimSpace(r.Header.Get(middleware.HeaderAgentVersion))
	if agentVersion == "" {
		agentVersion = "0.1.0"
	}
	environment := strings.TrimSpace(r.Header.Get(middleware.HeaderAgentEnvironment))
	if environment == "" {
		environment = "dev"
	}

	var req openai.ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.recordUsage(r, started, "", agentID, agentVersion, "", "", "bypass", 0, 0, 0, http.StatusBadRequest, "invalid_json")
		writeOpenAIError(w, http.StatusBadRequest, "invalid_request_error", "invalid JSON body")
		return
	}

	agentCfg := h.loadAgentConfig(r, agentID, agentVersion, environment)
	if agentCfg.PrimaryModel != "" && req.Model == "" {
		req.Model = agentCfg.PrimaryModel
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
		h.streamChatCompletions(w, r, started, agentID, agentVersion, environment, agentCfg, req)
		return
	}

	traceID := uuid.NewString()

	if h.enforceGovernance(w, r, started, traceID, agentID, agentVersion, environment, agentCfg, req) {
		return
	}

	cacheStatus := "miss"
	bypass := cache.ShouldBypass(r, agentCfg, req)
	if bypass {
		cacheStatus = "bypass"
		metrics.RecordCacheBypass()
	}

	var raw []byte
	var providerResp *providers.ChatResponse

	if !bypass && cache.CacheReadEnabled(agentCfg, h.config.CacheMode) {
		key := cache.ExactKey(agentID, agentVersion, req)
		if entry, err := h.cache.Get(r.Context(), key); err == nil && entry != nil {
			cacheStatus = "hit"
			metrics.RecordCacheHit()
			raw = entry.Value
			h.writeOperationalHeaders(w, agentID, agentVersion, "cache", req.Model, 0, 0, 0, traceID)
			w.Header().Set(middleware.HeaderCacheStatus, cacheStatus)
			h.recordUsage(r, started, traceID, agentID, agentVersion, "cache", req.Model, cacheStatus, 0, 0, 0, http.StatusOK, "")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(raw)
			return
		}
		metrics.RecordCacheMiss()
	} else if !bypass {
		metrics.RecordCacheMiss()
	}

	provider := h.providers.Resolve(req.Model)
	primaryProvider := agentCfg.PrimaryProvider
	if primaryProvider == "" {
		primaryProvider = provider.Name()
	}
	routeResult, err := h.providers.CallWithFallback(
		r.Context(),
		agentCfg.PrimaryProvider,
		req.Model,
		agentCfg.FallbackProvider,
		agentCfg.FallbackModel,
		agentCfg.RoutingPolicy,
		toProviderMessages(req.Messages),
	)
	if err != nil {
		h.recordUsage(r, started, traceID, agentID, agentVersion, primaryProvider, req.Model, cacheStatus, 0, 0, 0, http.StatusBadGateway, "upstream_error")
		writeOpenAIError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	providerResp = routeResult.Response
	raw = providerResp.Raw
	costUSD := providerResp.CostUSD
	if costUSD == 0 {
		costUSD = pricing.ComputeCostUSD(providerResp.Model, providerResp.PromptTokens, providerResp.CompletionTokens)
	}

	cacheTTL := h.config.CacheTTLSeconds
	if agentCfg.CacheTTLSeconds > 0 {
		cacheTTL = agentCfg.CacheTTLSeconds
	}
	if !bypass && cache.CacheWriteEnabled(agentCfg, h.config.CacheMode) {
		key := cache.ExactKey(agentID, agentVersion, req)
		_ = h.cache.Set(r.Context(), cache.Entry{
			Key:         key,
			Value:       raw,
			TTLSeconds:  cacheTTL,
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
		costUSD,
		traceID,
	)
	if routeResult.Fallback {
		w.Header().Set("x-routing-fallback", "true")
	}
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
		costUSD,
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
	agentID, agentVersion, environment string,
	agentCfg agentregistry.AgentConfig,
	req openai.ChatCompletionRequest,
) {
	metrics.RecordCacheBypass()
	if h.enforceGovernance(w, r, started, "", agentID, agentVersion, environment, agentCfg, req) {
		return
	}
	provider := h.providers.Resolve(req.Model)
	routeResult, err := h.providers.CallWithFallback(
		r.Context(),
		agentCfg.PrimaryProvider,
		req.Model,
		agentCfg.FallbackProvider,
		agentCfg.FallbackModel,
		agentCfg.RoutingPolicy,
		toProviderMessages(req.Messages),
	)
	if err != nil {
		h.recordUsage(r, started, "", agentID, agentVersion, provider.Name(), req.Model, "bypass", 0, 0, 0, http.StatusBadGateway, "upstream_error")
		writeOpenAIError(w, http.StatusBadGateway, "upstream_error", err.Error())
		return
	}
	resp := routeResult.Response

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

	costUSD := resp.CostUSD
	if costUSD == 0 {
		costUSD = pricing.ComputeCostUSD(resp.Model, resp.PromptTokens, resp.CompletionTokens)
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
		costUSD,
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
		costUSD,
		http.StatusOK,
		"",
	)
	_ = agentCfg // registry config influences bypass via caller; stream always bypasses cache
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

func (h *Handler) loadAgentConfig(r *http.Request, agentID, version, environment string) agentregistry.AgentConfig {
	if h.agentRegistry == nil {
		return agentregistry.AgentConfig{
			CacheMode:       h.config.CacheMode,
			CacheTTLSeconds: h.config.CacheTTLSeconds,
		}
	}
	cfg, err := h.agentRegistry.GetAgentConfig(r.Context(), agentID, version, environment)
	if err != nil {
		return agentregistry.AgentConfig{
			CacheMode:       h.config.CacheMode,
			CacheTTLSeconds: h.config.CacheTTLSeconds,
		}
	}
	return cfg
}

func (h *Handler) enforceGovernance(
	w http.ResponseWriter,
	r *http.Request,
	started time.Time,
	traceID, agentID, agentVersion, environment string,
	agentCfg agentregistry.AgentConfig,
	req openai.ChatCompletionRequest,
) bool {
	if environment == "" {
		environment = "dev"
	}
	if traceID == "" {
		traceID = uuid.NewString()
	}

	provider := h.providers.Resolve(req.Model)
	decision := h.policy.Allow(r.Context(), policy.Input{
		Agent:       agentCfg,
		Environment: environment,
		Provider:    provider.Name(),
	})
	if !decision.Allowed {
		metrics.RecordPolicyDenied()
		h.recordUsage(r, started, traceID, agentID, agentVersion, provider.Name(), req.Model, "bypass", 0, 0, 0, http.StatusForbidden, "policy_denied")
		writeOpenAIError(w, http.StatusForbidden, "policy_denied", decision.Reason)
		return true
	}

	tenantID := budget.NormalizeTenant(r.Header.Get(middleware.HeaderTenantID))
	if h.budget != nil && h.rateLimit != nil {
		rpm, _ := h.budget.RateLimit(r.Context(), tenantID, agentID, agentVersion)
		if rpm > 0 {
			allowed, retryAfter := h.rateLimit.Allow(r.Context(), ratelimit.AgentKey(tenantID, agentID), rpm)
			if !allowed {
				metrics.RecordRateLimited()
				w.Header().Set("Retry-After", ratelimit.RetryAfterHeader(retryAfter))
				h.recordUsage(r, started, traceID, agentID, agentVersion, provider.Name(), req.Model, "bypass", 0, 0, 0, http.StatusTooManyRequests, "rate_limit_exceeded")
				writeOpenAIError(w, http.StatusTooManyRequests, "rate_limit_exceeded", fmt.Sprintf("rate limit exceeded (%d requests/minute)", rpm))
				return true
			}
		}
	}

	if h.budget != nil {
		messageTexts := make([]string, len(req.Messages))
		for i, msg := range req.Messages {
			messageTexts[i] = msg.Content
		}
		if violation := h.budget.Check(r.Context(), tenantID, agentID, agentVersion, budget.EstimatePromptTokens(messageTexts)); violation != nil {
			metrics.RecordBudgetExceeded()
			h.recordUsage(r, started, traceID, agentID, agentVersion, provider.Name(), req.Model, "bypass", 0, 0, 0, http.StatusTooManyRequests, violation.Code)
			writeOpenAIError(w, http.StatusTooManyRequests, violation.Code, violation.Message)
			return true
		}
	}
	return false
}
