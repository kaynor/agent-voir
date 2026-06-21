"use client";

import { create } from "zustand";
import type { LiveEventRow, LiveMetrics, TraceDetailView, TraceFlowStep } from "../lib/live-types";
import {
  fetchProxyEvents,
  fetchTraceDetail,
  gatewayWsUrl,
  mapApiMetrics,
  mapGatewayEvent,
  mapTraceDetail,
  mapTraceSteps,
  type GatewayProxyEvent,
} from "../lib/live-api";
import {
  mockLiveMetrics,
  mockLiveRows,
  mockToolCallJson,
  mockTraceDetail,
  mockTraceFlow,
} from "../lib/mock-live-events";

type ConnectionState = "connecting" | "live" | "mock" | "error";

type LiveStore = {
  rows: LiveEventRow[];
  metrics: LiveMetrics;
  selectedTraceId: string | null;
  traceSteps: TraceFlowStep[];
  traceDetail: TraceDetailView | null;
  toolCallJson: string | null;
  connection: ConnectionState;
  error: string;
  followTail: boolean;
  paused: boolean;
  setFollowTail: (value: boolean) => void;
  setPaused: (value: boolean) => void;
  selectTrace: (traceId: string) => Promise<void>;
  loadInitial: () => Promise<void>;
  connectWebSocket: () => () => void;
};

const emptyMetrics: LiveMetrics = {
  requestsTotal: 0,
  errors: 0,
  toolCalls: 0,
  finalAnswers: 0,
  blocked: 0,
  cacheHits: 0,
  tokensIn: 0,
  tokensOut: 0,
  activeRequests: 0,
  costUsd: 0,
  cacheHitRate: 0,
  p50LatencyMs: 0,
  p95LatencyMs: 0,
  p99LatencyMs: 0,
};

export const useLiveStore = create<LiveStore>((set, get) => ({
  rows: [],
  metrics: emptyMetrics,
  selectedTraceId: null,
  traceSteps: [],
  traceDetail: null,
  toolCallJson: null,
  connection: "connecting",
  error: "",
  followTail: true,
  paused: false,
  setFollowTail: (value) => set({ followTail: value }),
  setPaused: (value) => set({ paused: value }),
  selectTrace: async (traceId) => {
    set({ selectedTraceId: traceId });
    const row = get().rows.find((r) => r.traceId === traceId);
    try {
      const detail = await fetchTraceDetail(traceId);
      set({
        traceSteps: mapTraceSteps(detail),
        traceDetail: mapTraceDetail(detail, row),
        toolCallJson: detail.tool_call
          ? JSON.stringify(
              { function: detail.tool_call.name, arguments: detail.tool_call.arguments },
              null,
              2,
            )
          : null,
      });
    } catch {
      set({
        traceSteps: mockTraceFlow,
        traceDetail: row
          ? {
              ...mockTraceDetail,
              traceId: row.traceId,
              agentId: row.agent,
              userId: row.user,
              tokensIn: row.tokensIn,
              tokensOut: row.tokensOut,
              costUsd: row.costUsd,
              spanId: row.spanId,
              otelStatus: row.otelStatus,
              datadogStatus: row.datadogStatus,
            }
          : mockTraceDetail,
        toolCallJson: mockToolCallJson,
      });
    }
  },
  loadInitial: async () => {
    try {
      const data = await fetchProxyEvents(500);
      const rows = data.events.map(mapGatewayEvent);
      set({
        rows: rows.length ? rows : mockLiveRows,
        metrics: data.metrics.requests_total ? mapApiMetrics(data.metrics) : mockLiveMetrics,
        connection: rows.length ? "live" : "mock",
        error: rows.length ? "" : "No proxy events yet — run: make seed-live-events",
        selectedTraceId: rows[0]?.traceId ?? mockLiveRows[0]?.traceId ?? null,
      });
      const traceId = get().selectedTraceId;
      if (traceId) {
        await get().selectTrace(traceId);
      }
    } catch (err) {
      set({
        rows: mockLiveRows,
        metrics: mockLiveMetrics,
        connection: "mock",
        error: err instanceof Error ? err.message : "Could not reach gateway — showing mock data",
        selectedTraceId: mockLiveRows[0]?.traceId ?? null,
      });
    }
  },
  connectWebSocket: () => {
    let ws: WebSocket | null = null;
    let closed = false;

    const connect = () => {
      if (closed) return;
      ws = new WebSocket(`${gatewayWsUrl()}?limit=500`);
      ws.onopen = () => set({ connection: "live", error: "" });
      ws.onmessage = (msg) => {
        if (get().paused) return;
        try {
          const envelope = JSON.parse(msg.data as string) as {
            type: string;
            payload?: string | Record<string, unknown>;
          };
          if (envelope.type === "snapshot" && envelope.payload) {
            const payload =
              typeof envelope.payload === "string" ? JSON.parse(envelope.payload) : envelope.payload;
            const rows = ((payload as { rows?: GatewayProxyEvent[] }).rows ?? []).map(mapGatewayEvent);
            if (rows.length && get().followTail) {
              set({ rows: rows.slice(0, 500) });
            }
          }
          if (envelope.type === "metrics_delta" && envelope.payload) {
            const raw =
              typeof envelope.payload === "string"
                ? JSON.parse(envelope.payload)
                : envelope.payload;
            set({ metrics: mapApiMetrics(raw as import("./live-api").ApiMetrics) });
          }
          if (envelope.type === "row_upsert" && envelope.payload) {
            const payload =
              typeof envelope.payload === "string" ? JSON.parse(envelope.payload) : envelope.payload;
            const row = mapGatewayEvent((payload as { row: GatewayProxyEvent }).row);
            if (!get().followTail) return;
            set((state) => {
              const next = [row, ...state.rows.filter((r) => r.spanId !== row.spanId)];
              return { rows: next.slice(0, 500) };
            });
          }
        } catch {
          // ignore malformed packets
        }
      };
      ws.onclose = () => {
        if (!closed) {
          set((state) => ({
            connection: state.rows.length ? "live" : "mock",
            error: state.error || "WebSocket disconnected — retrying…",
          }));
          setTimeout(connect, 2000);
        }
      };
      ws.onerror = () => ws?.close();
    };

    connect();
    return () => {
      closed = true;
      ws?.close();
    };
  },
}));
