import type { LiveEventRow, LiveMetrics, TraceFlowStep } from "./live-types";

export type GatewayProxyEvent = {
  seq: number;
  event_time: string;
  trace_id: string;
  span_id: string;
  agent_id: string;
  user_id: string;
  req_method: string;
  req_path: string;
  status_code: number;
  provider: string;
  model: string;
  response_type: string;
  next_action: string;
  tool: string;
  terminal: boolean;
  tags: string[];
  tokens_in: number;
  tokens_out: number;
  duration_ms: number;
  cost_usd: number;
  otel_status: string;
  datadog_status: string;
};

export type ApiMetrics = {
  window: string;
  requests_total: number;
  errors: number;
  tool_calls: number;
  final_answers: number;
  blocked: number;
  cache_hits: number;
  tokens_in: number;
  tokens_out: number;
  active_requests: number;
  cost_usd: number;
  cache_hit_rate: number;
  p50_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
};

export type ProxyEventsResponse = {
  window: string;
  limit: number;
  matched_count: number;
  returned_count: number;
  metrics: ApiMetrics;
  events: GatewayProxyEvent[];
};

export type TraceDetailResponse = {
  trace_id: string;
  agent_id: string;
  user_id: string;
  status: string;
  started_at: string;
  duration_ms: number;
  cost_usd: number;
  tags: string[];
  steps: Array<{
    step: number;
    span_id: string;
    kind: string;
    response_type?: string;
    status: string;
    duration_ms?: number;
    next_action?: string;
    tool?: string;
  }>;
  tool_call?: {
    name: string;
    arguments: Record<string, unknown>;
  };
};

export function gatewayUrl(): string {
  return process.env.NEXT_PUBLIC_GATEWAY_URL ?? "http://localhost:8080";
}

export function gatewayWsUrl(): string {
  return process.env.NEXT_PUBLIC_GATEWAY_WS_URL ?? "ws://localhost:8080/ws/proxy-events";
}

export function mapGatewayEvent(event: GatewayProxyEvent): LiveEventRow {
  const time = new Date(event.event_time);
  const timeLabel = time.toLocaleTimeString("en-US", { hour12: false }) + "." + String(time.getMilliseconds()).padStart(3, "0");
  const statusLabel = event.status_code >= 400 ? `${event.status_code}` : "200 OK";
  return {
    traceId: event.trace_id,
    spanId: event.span_id,
    time: timeLabel,
    agent: event.agent_id,
    user: event.user_id,
    reqResp: `${event.req_method} ${event.req_path} → ${statusLabel}`,
    status: event.status_code,
    provider: event.provider,
    model: event.model,
    responseType: event.response_type as LiveEventRow["responseType"],
    nextAction: event.next_action,
    tool: event.tool,
    terminal: event.terminal,
    tags: event.tags,
    tokensIn: event.tokens_in,
    tokensOut: event.tokens_out,
    durationMs: event.duration_ms,
    costUsd: event.cost_usd,
    otelStatus: event.otel_status === "exported" ? "exported" : "queued",
    datadogStatus: event.datadog_status === "indexed" ? "indexed" : "queued",
  };
}

export function mapApiMetrics(raw: ApiMetrics): LiveMetrics {
  return {
    requestsTotal: raw.requests_total,
    errors: raw.errors,
    toolCalls: raw.tool_calls,
    finalAnswers: raw.final_answers,
    blocked: raw.blocked,
    cacheHits: raw.cache_hits,
    tokensIn: raw.tokens_in,
    tokensOut: raw.tokens_out,
    activeRequests: raw.active_requests,
    costUsd: raw.cost_usd,
    cacheHitRate: raw.cache_hit_rate,
    p50LatencyMs: raw.p50_latency_ms,
    p95LatencyMs: raw.p95_latency_ms,
    p99LatencyMs: raw.p99_latency_ms,
  };
}

export async function fetchProxyEvents(limit = 500): Promise<ProxyEventsResponse> {
  const res = await fetch(`${gatewayUrl()}/v1/proxy-events?limit=${limit}`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`proxy-events failed (${res.status})`);
  }
  return res.json() as Promise<ProxyEventsResponse>;
}

export async function fetchTraceDetail(traceId: string): Promise<TraceDetailResponse> {
  const res = await fetch(`${gatewayUrl()}/v1/traces/${encodeURIComponent(traceId)}`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`trace detail failed (${res.status})`);
  }
  return res.json() as Promise<TraceDetailResponse>;
}

export function mapTraceSteps(detail: TraceDetailResponse): TraceFlowStep[] {
  return detail.steps.map((step) => ({
    step: step.step,
    kind: step.kind,
    responseType: step.response_type as TraceFlowStep["responseType"],
    status: step.status as TraceFlowStep["status"],
    durationMs: step.duration_ms,
    nextAction: step.next_action,
    tool: step.tool,
  }));
}

export function mapTraceDetail(detail: TraceDetailResponse, row?: LiveEventRow): import("./live-types").TraceDetailView {
  return {
    traceId: detail.trace_id,
    agentId: detail.agent_id,
    userId: detail.user_id,
    status: detail.status,
    startedAt: detail.started_at,
    durationMs: detail.duration_ms,
    costUsd: detail.cost_usd,
    tags: detail.tags,
    tokensIn: row?.tokensIn ?? 0,
    tokensOut: row?.tokensOut ?? 0,
    budgetRemaining: 315780,
    spanId: detail.steps.find((s) => s.response_type === "TOOL_CALL")?.span_id ?? detail.steps[0]?.span_id ?? row?.spanId ?? "—",
    otelStatus: row?.otelStatus ?? "exported",
    datadogStatus: row?.datadogStatus ?? "indexed",
  };
}

const JSON_DISPLAY_INDENT = 4;

/** Pretty-print JSON for drilldown panels with visible nesting. */
export function formatJsonDisplay(value: unknown, indent = JSON_DISPLAY_INDENT): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "string") {
    const trimmed = value.trim();
    if (!trimmed) return "";
    try {
      return JSON.stringify(JSON.parse(trimmed), null, indent);
    } catch {
      return trimmed;
    }
  }
  try {
    return JSON.stringify(value, null, indent);
  } catch {
    return String(value);
  }
}

export function formatToolCallJson(toolCall: { name: string; arguments: unknown }): string {
  let args: unknown = toolCall.arguments;
  if (typeof args === "string") {
    try {
      args = JSON.parse(args);
    } catch {
      /* keep raw string */
    }
  }
  return formatJsonDisplay({ function: toolCall.name, arguments: args });
}
