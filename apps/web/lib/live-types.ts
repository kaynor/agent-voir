/** Live grid row — small summary packet (see docs/architecture/ui-dashboard.md). */
export type ResponseType =
  | "TOOL_CALL"
  | "TOOL_RESULT"
  | "FINAL_ANSWER"
  | "STREAM_FINAL"
  | "CACHE_RESPONSE"
  | "GUARDRAIL_BLOCK";

export type LiveEventRow = {
  traceId: string;
  spanId: string;
  time: string;
  agent: string;
  user: string;
  reqResp: string;
  status: number;
  provider: string;
  model: string;
  responseType: ResponseType;
  nextAction: string;
  tool: string;
  terminal: boolean;
  tags: string[];
  tokensIn: number;
  tokensOut: number;
  durationMs: number;
  costUsd: number;
  otelStatus: "exported" | "queued" | "failed";
  datadogStatus: "indexed" | "queued" | "failed";
};

export type LiveMetrics = {
  requestsTotal: number;
  errors: number;
  toolCalls: number;
  finalAnswers: number;
  blocked: number;
  cacheHits: number;
  tokensIn: number;
  tokensOut: number;
  activeRequests: number;
  costUsd: number;
  cacheHitRate: number;
  p50LatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
};

export type TraceFlowStep = {
  step: number;
  kind: string;
  responseType?: ResponseType;
  status: "complete" | "in_progress" | "pending";
  durationMs?: number;
  nextAction?: string;
  tool?: string;
};

export type TraceDetailView = {
  traceId: string;
  agentId: string;
  userId: string;
  status: string;
  startedAt: string;
  durationMs: number;
  costUsd: number;
  tags: string[];
  tokensIn: number;
  tokensOut: number;
  budgetRemaining: number;
  spanId: string;
  otelStatus: string;
  datadogStatus: string;
};
