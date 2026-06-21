"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import type { LiveEventRow, ResponseType, TraceFlowStep } from "../../lib/live-types";
import { useLiveStore } from "../../lib/useLiveStore";

const RESPONSE_BADGE: Record<ResponseType, string> = {
  TOOL_CALL: "badge-tool-call",
  TOOL_RESULT: "badge-tool-result",
  FINAL_ANSWER: "badge-final",
  STREAM_FINAL: "badge-stream",
  CACHE_RESPONSE: "badge-cache",
  GUARDRAIL_BLOCK: "badge-block",
};

const TRACE_TABS = ["Flow", "Request", "Response", "Headers", "Tokens", "Tool Calls", "OTel", "Datadog", "Raw JSON"];

const TIME_RANGES = ["Live: Last 1 min", "Live: Last 5 min", "Last 15 min", "Last 1 hour", "Last 24 hours", "Custom range"];
const RECORD_LIMITS = ["Latest 100", "Latest 500", "Latest 1,000", "Latest 5,000"];
const RESPONSE_TYPES = ["All types", "Tool call", "Tool result", "Final answer", "Stream final", "Cache response", "Guardrail block"];
const PROVIDERS = ["All providers", "OpenAI", "Anthropic", "Google", "AgentVoir Tools"];
const TAGS = ["All tags", "#tool-call", "#error", "#high-cost", "#cache-hit", "#policy-blocked", "#review"];

function formatUsd(value: number): string {
  if (value === 0) return "$0";
  if (value < 0.01) return `$${value.toFixed(4)}`;
  return `$${value.toFixed(2)}`;
}

function formatBudget(value: number): string {
  if (value >= 1000) return `${(value / 1000).toFixed(0)}k`;
  return value.toLocaleString();
}

function KpiCard({ label, value, sub }: { label: string; value: string; sub?: string }) {
  return (
    <article className="kpi-card">
      <p className="kpi-label">{label}</p>
      <p className="kpi-value">{value}</p>
      {sub ? <p className="kpi-sub">{sub}</p> : null}
    </article>
  );
}

export function LiveProxyFlow() {
  const {
    rows,
    metrics,
    selectedTraceId,
    traceSteps,
    traceDetail,
    toolCallJson,
    connection,
    error,
    followTail,
    paused,
    setFollowTail,
    setPaused,
    selectTrace,
    loadInitial,
    connectWebSocket,
  } = useLiveStore();

  useEffect(() => {
    void loadInitial();
    return connectWebSocket();
  }, [loadInitial, connectWebSocket]);

  const wsConnected = connection === "live";
  const matchedCount = metrics.requestsTotal;

  return (
    <div className="live-flow">
      <header className="live-header">
        <div className="live-title-block">
          <Image src="/agentvoir-logo.svg" alt="" width={18} height={18} className="live-header-logo" />
          <h1 className="live-title">
            Live Proxy Flow
            <span className="live-badge-live">LIVE</span>
          </h1>
        </div>
        <div className="live-header-actions">
          <span className={`ws-status ${wsConnected ? "connected" : "disconnected"}`}>
            <span className="ws-dot" aria-hidden />
            {wsConnected ? "WebSocket Connected" : connection === "connecting" ? "Connecting…" : "WebSocket Disconnected"}
          </span>
          <button type="button" className="btn secondary" disabled title="Datadog link when OTel export is configured">
            Open in Datadog
          </button>
          <button type="button" className="btn secondary" disabled>
            Export ▾
          </button>
        </div>
      </header>

      {error ? (
        <div className="alert">
          <strong>{error}</strong> Run <code>make seed-live-events</code> or{" "}
          <code>make demo-live-dashboard</code> after onebox is up.
        </div>
      ) : null}

      <section className="kpi-row" aria-label="Key metrics">
        <KpiCard
          label="Requests (5m)"
          value={metrics.requestsTotal.toLocaleString()}
          sub={`Active: ${metrics.activeRequests}`}
        />
        <KpiCard
          label="Tokens (5m)"
          value={metrics.tokensIn.toLocaleString()}
          sub={`Out: ${metrics.tokensOut.toLocaleString()}`}
        />
        <KpiCard label="Cost (5m)" value={formatUsd(metrics.costUsd)} />
        <KpiCard
          label="Latency P95"
          value={`${(metrics.p95LatencyMs / 1000).toFixed(2)}s`}
          sub={`P99 ${(metrics.p99LatencyMs / 1000).toFixed(2)}s`}
        />
        <KpiCard label="Errors (5m)" value={String(metrics.errors)} />
        <KpiCard
          label="Cache hit rate"
          value={`${(metrics.cacheHitRate * 100).toFixed(1)}%`}
          sub={`${metrics.cacheHits.toLocaleString()} hits`}
        />
      </section>

      <section className="live-controls panel" aria-label="Filters">
        <div className="filter-row">
          <label>
            Time range
            <select defaultValue="5m" disabled={paused}>
              {TIME_RANGES.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </label>
          <label>
            Record limit
            <select defaultValue="500">
              {RECORD_LIMITS.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </label>
          <input
            className="search-input"
            type="search"
            placeholder="agent:research-agent response_type:tool_call status:200 tag:review model:gpt-4.1-mini"
            aria-label="Search and filter events"
          />
          <label>
            Response types
            <select defaultValue="all">
              {RESPONSE_TYPES.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </label>
          <label>
            Providers
            <select defaultValue="all">
              {PROVIDERS.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </label>
          <label>
            Tags
            <select defaultValue="all">
              {TAGS.map((opt) => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          </label>
          <button type="button" className="btn secondary" disabled title="Advanced filters coming soon">
            More filters
          </button>
          <label className="toggle-switch">
            <input type="checkbox" checked={followTail} onChange={(e) => setFollowTail(e.target.checked)} />
            <span className="toggle-track" aria-hidden />
            Follow tail
          </label>
          <button type="button" className="btn" onClick={() => setPaused(!paused)}>
            {paused ? "Resume" : "Pause"}
          </button>
        </div>

        <div className="quick-filters" aria-label="Quick filters">
          <button type="button" className="quick-filter">
            Errors <span>{metrics.errors}</span>
          </button>
          <button type="button" className="quick-filter">
            Tool calls <span>{metrics.toolCalls}</span>
          </button>
          <button type="button" className="quick-filter">
            Final answers <span>{metrics.finalAnswers}</span>
          </button>
          <button type="button" className="quick-filter">
            Blocked <span>{metrics.blocked}</span>
          </button>
          <button type="button" className="quick-filter">
            Cache hits <span>{metrics.cacheHits}</span>
          </button>
        </div>

        <p className="grid-banner">
          Showing latest {rows.length} of {matchedCount.toLocaleString()} matching events in last 5 minutes.
          <span className="grid-banner-hint"> Narrow filters or open Analytics for full aggregation.</span>
        </p>
      </section>

      <section className="live-grid panel" aria-label="Live event grid">
        <div className="live-grid-scroll">
          <table className="live-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>Trace ID</th>
                <th>Agent / User</th>
                <th>Req → Resp</th>
                <th>Status</th>
                <th>Provider / Model</th>
                <th>Response type</th>
                <th>Next action</th>
                <th>Tool</th>
                <th>Terminal</th>
                <th>Tags</th>
                <th>Tokens</th>
                <th>Budget left</th>
                <th>Duration</th>
                <th>Cost</th>
                <th>OTel</th>
                <th>Datadog</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <EventRow
                  key={`${row.traceId}:${row.spanId}`}
                  row={row}
                  selected={row.traceId === selectedTraceId}
                  onSelect={() => void selectTrace(row.traceId)}
                />
              ))}
            </tbody>
          </table>
        </div>
      </section>

      {selectedTraceId && traceDetail ? (
        <TraceDrilldown
          traceId={selectedTraceId}
          traceSteps={traceSteps}
          traceDetail={traceDetail}
          toolCallJson={toolCallJson}
        />
      ) : null}
    </div>
  );
}

function TraceDrilldown({
  traceId,
  traceSteps,
  traceDetail,
  toolCallJson,
}: {
  traceId: string;
  traceSteps: TraceFlowStep[];
  traceDetail: NonNullable<ReturnType<typeof useLiveStore.getState>["traceDetail"]>;
  toolCallJson: string | null;
}) {
  const [activeTab, setActiveTab] = useState("Flow");
  const [activeStep, setActiveStep] = useState(2);
  const step = traceSteps.find((s) => s.step === activeStep) ?? traceSteps[0];

  return (
    <section className="trace-drilldown panel" aria-label="Trace detail">
      <header className="trace-drilldown-header">
        <div className="trace-header-main">
          <strong className="trace-id">{traceId}</strong>
          <span className="trace-header-meta">
            {traceDetail.agentId} · {traceDetail.userId} · {(traceDetail.durationMs / 1000).toFixed(2)}s ·{" "}
            {formatUsd(traceDetail.costUsd)}
          </span>
        </div>
        <div className="trace-header-tags">
          {traceDetail.tags.map((t) => (
            <span key={t} className="tag">
              #{t}
            </span>
          ))}
        </div>
      </header>

      <div className="trace-tabs" role="tablist">
        {TRACE_TABS.map((tab) => (
          <button
            key={tab}
            type="button"
            role="tab"
            className={activeTab === tab ? "active" : ""}
            aria-selected={activeTab === tab}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </div>

      <div className="trace-three-col">
        <div className="trace-col trace-col-flow">
          <h3 className="trace-col-title">Call Flow</h3>
          <ol className="flow-timeline">
            {traceSteps.map((s) => (
              <li key={s.step}>
                <button
                  type="button"
                  className={`flow-step-btn${activeStep === s.step ? " active" : ""}`}
                  onClick={() => setActiveStep(s.step)}
                >
                  <span className="flow-step-circle">{s.step}</span>
                  <span className="flow-step-body">
                    <span className="flow-step-kind">{s.kind}</span>
                    {s.responseType ? (
                      <span className={`badge ${RESPONSE_BADGE[s.responseType]}`}>
                        {s.responseType.replace("_", " ")}
                      </span>
                    ) : null}
                    {s.durationMs ? (
                      <span className="flow-step-dur">{(s.durationMs / 1000).toFixed(2)}s</span>
                    ) : null}
                  </span>
                </button>
              </li>
            ))}
          </ol>
        </div>

        <div className="trace-col trace-col-detail">
          <h3 className="trace-col-title">Step Details</h3>
          {step ? (
            <>
              <p className="step-summary">
                <strong>{step.kind}</strong>
                {step.tool ? <span className="muted small"> · {step.tool}</span> : null}
              </p>
              {step.nextAction ? (
                <p className="step-next">
                  Next: <span>{step.nextAction}</span>
                </p>
              ) : null}
              {toolCallJson ? (
                <>
                  <p className="step-block-label">Tool Call (Arguments)</p>
                  <pre className="code-block">{toolCallJson}</pre>
                  <p className="step-block-label">Tool Call Summary</p>
                  <p className="step-summary-text">
                    LLM requested tool execution with structured arguments. Result forwarded back to the model for final
                    answer synthesis.
                  </p>
                </>
              ) : (
                <p className="muted small">Select a tool-call step to view argument payload.</p>
              )}
            </>
          ) : null}
        </div>

        <div className="trace-col trace-col-meta">
          <h3 className="trace-col-title">Metadata &amp; Links</h3>

          <div className="meta-widget">
            <p className="meta-widget-title">Tokens &amp; Budget</p>
            <dl className="meta-dl compact">
              <dt>In / Out</dt>
              <dd>
                {traceDetail.tokensIn.toLocaleString()} / {traceDetail.tokensOut.toLocaleString()}
              </dd>
              <dt>Request cost</dt>
              <dd>{formatUsd(traceDetail.costUsd)}</dd>
              <dt>Budget left</dt>
              <dd>{formatBudget(traceDetail.budgetRemaining)} tokens</dd>
            </dl>
            <div className="budget-bar" aria-hidden>
              <span style={{ width: "42%" }} />
            </div>
          </div>

          <div className="meta-widget">
            <p className="meta-widget-title">OTel / Datadog</p>
            <dl className="meta-dl compact">
              <dt>Trace ID</dt>
              <dd className="mono">{traceDetail.traceId}</dd>
              <dt>Span ID</dt>
              <dd className="mono">{traceDetail.spanId}</dd>
            </dl>
            <button type="button" className="linkish meta-link" disabled>
              View in Trace Viewer
            </button>
          </div>

          <div className="meta-widget">
            <p className="meta-widget-title">Datadog Links</p>
            <ul className="meta-links">
              <li>
                <button type="button" className="linkish" disabled>
                  APM Trace
                </button>
              </li>
              <li>
                <button type="button" className="linkish" disabled>
                  Related Logs
                </button>
              </li>
              <li>
                <button type="button" className="linkish" disabled>
                  Cost Dashboard
                </button>
              </li>
            </ul>
          </div>

          <div className="meta-widget">
            <p className="meta-widget-title">Export Status</p>
            <ul className="export-checklist">
              <li className={traceDetail.otelStatus === "exported" ? "ok" : ""}>
                OTel {traceDetail.otelStatus === "exported" ? "✓ exported" : "… queued"}
              </li>
              <li className={traceDetail.datadogStatus === "indexed" ? "ok" : ""}>
                Datadog {traceDetail.datadogStatus === "indexed" ? "✓ indexed" : "… queued"}
              </li>
              <li className="ok">Logs ✓ exported</li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  );
}

function EventRow({
  row,
  selected,
  onSelect,
}: {
  row: LiveEventRow;
  selected: boolean;
  onSelect: () => void;
}) {
  const budgetLeft = 500000 - row.tokensIn - row.tokensOut;

  return (
    <tr className={selected ? "selected" : ""} onClick={onSelect}>
      <td>{row.time}</td>
      <td>
        <button type="button" className="linkish" onClick={onSelect}>
          {row.traceId}
        </button>
      </td>
      <td>
        {row.agent}
        <span className="muted small"> / {row.user}</span>
      </td>
      <td className="mono">{row.reqResp}</td>
      <td>
        <span className={row.status >= 400 ? "status-err" : "status-ok"}>{row.status}</span>
      </td>
      <td>
        {row.provider} {row.model}
      </td>
      <td>
        <span className={`badge ${RESPONSE_BADGE[row.responseType]}`}>
          {row.responseType.replace("_", " ")}
        </span>
      </td>
      <td className="col-next" title={row.nextAction}>
        {row.nextAction}
      </td>
      <td>{row.tool}</td>
      <td>{row.terminal ? "✓" : "—"}</td>
      <td>
        {row.tags.map((t) => (
          <span key={t} className="tag">
            #{t}
          </span>
        ))}
      </td>
      <td>
        {row.tokensIn} / {row.tokensOut}
      </td>
      <td className="mono">{formatBudget(budgetLeft)}</td>
      <td>{(row.durationMs / 1000).toFixed(2)}s</td>
      <td>{formatUsd(row.costUsd)}</td>
      <td>{row.otelStatus === "exported" ? "✓" : "…"}</td>
      <td>{row.datadogStatus === "indexed" ? "✓" : "…"}</td>
    </tr>
  );
}
