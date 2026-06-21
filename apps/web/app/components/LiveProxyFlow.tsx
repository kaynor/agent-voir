"use client";

import { useEffect, useState } from "react";
import { formatJsonDisplay } from "../../lib/live-api";
import type { LiveEventRow, ResponseType, TraceFlowStep } from "../../lib/live-types";
import { useLiveStore } from "../../lib/useLiveStore";

const RESPONSE_BADGE: Record<ResponseType, string> = {
  TOOL_CALL: "badge-toolcall",
  TOOL_RESULT: "badge-toolresult",
  FINAL_ANSWER: "badge-final",
  STREAM_FINAL: "badge-stream",
  CACHE_RESPONSE: "badge-cache",
  GUARDRAIL_BLOCK: "badge-block",
};

const TRACE_TABS = [
  "Flow",
  "Request",
  "Response",
  "Headers",
  "Tool Calls",
  "Tokens",
  "Attributes",
  "Logs",
  "OTel",
  "Datadog",
  "Raw JSON",
];

function formatUsd(value: number): string {
  if (value === 0) return "$0.0000";
  if (value < 0.01) return `$${value.toFixed(4)}`;
  return `$${value.toFixed(2)}`;
}

function formatBudget(value: number): string {
  return value.toLocaleString();
}

function Sparkline({ points, color = "#22d36b" }: { points: string; color?: string }) {
  return (
    <span className="spark" aria-hidden>
      <svg viewBox="0 0 100 35">
        <polyline points={points} fill="none" stroke={color} strokeWidth="3" />
      </svg>
    </span>
  );
}

function ProgressBar({ pct, variant }: { pct: number; variant?: "blue" }) {
  return (
    <div className={`bar${variant === "blue" ? " blue" : ""}`}>
      <i style={{ width: `${Math.min(100, pct)}%` }} />
    </div>
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

  const tokensTotal = metrics.tokensIn + metrics.tokensOut;
  const budgetUsed = 184220;
  const budgetLimit = 500000;
  const budgetPct = (budgetUsed / budgetLimit) * 100;
  const estHourly = metrics.costUsd * 12;
  const hasDrilldown = Boolean(selectedTraceId && traceDetail);

  return (
    <div className={`live-flow${hasDrilldown ? " has-drilldown" : ""}`}>
      {error ? (
        <div className="live-alert">
          <strong>{error}</strong> Run <code>make seed-live-events</code> or{" "}
          <code>make demo-live-dashboard</code> after onebox is up.
        </div>
      ) : null}

      <section className="live-cards" aria-label="Key metrics">
        <article className="live-card">
          <div className="card-title">Requests (5m)</div>
          <div className="metric">{metrics.requestsTotal.toLocaleString()}</div>
          <div className="mini-row">
            <span className="delta">↗ 12.6%</span>
            <Sparkline points="0,30 10,24 20,27 30,14 42,20 54,9 66,17 78,7 88,12 100,4" />
          </div>
          <div className="mini-row">
            <span>Active Now</span>
            <b>{metrics.activeRequests}</b>
          </div>
        </article>

        <article className="live-card">
          <div className="card-title">Tokens (5m)</div>
          <div className="metric">
            {tokensTotal.toLocaleString()} <span className="delta inline-delta">↗ 8.7%</span>
          </div>
          <div className="donut-wrap">
            <div className="donut" aria-hidden />
            <div>
              <div className="small-label">Input</div>
              <b>{metrics.tokensIn.toLocaleString()}</b>
            </div>
            <div>
              <div className="small-label">Output</div>
              <b>{metrics.tokensOut.toLocaleString()}</b>
            </div>
          </div>
        </article>

        <article className="live-card">
          <div className="card-title">Cost (5m)</div>
          <div className="metric">
            {formatUsd(metrics.costUsd)} <span className="delta inline-delta">↗ 9.4%</span>
          </div>
          <div className="mini-row">
            <span>Est. (1h)</span>
            <b>{formatUsd(estHourly)}</b>
          </div>
          <Sparkline points="0,26 10,29 18,20 30,25 42,15 52,20 65,8 75,13 84,5 100,22" />
        </article>

        <article className="live-card">
          <div className="card-title">Budget (Daily)</div>
          <div className="bar-label">
            <span>
              Used <b>{budgetUsed.toLocaleString()}</b>
            </span>
            <span>
              Limits <b>{budgetLimit.toLocaleString()}</b>
            </span>
          </div>
          <ProgressBar pct={budgetPct} />
          <div className="mini-row">
            <span>Left</span>
            <b className="delta">{(budgetLimit - budgetUsed).toLocaleString()}</b>
          </div>
        </article>

        <article className="live-card">
          <div className="card-title">
            Provider Limits <span className="small-label">(OpenAI)</span>
          </div>
          <div className="bar-label">
            <span>Requests Left</span>
            <b>24,320 / 30K</b>
          </div>
          <ProgressBar pct={81} />
          <div className="bar-label">
            <span>Tokens Left</span>
            <b>1.2M / 2M</b>
          </div>
          <ProgressBar pct={60} variant="blue" />
        </article>

        <article className="live-card live-card-compact">
          <div className="card-title">Latency (5m)</div>
          <div className="latency-lines">
            <b>P50</b>
            <span>{metrics.p50LatencyMs}ms</span>
            <b>P95</b>
            <span className="bad">{(metrics.p95LatencyMs / 1000).toFixed(2)}s</span>
            <b>P99</b>
            <span className="bad">{(metrics.p99LatencyMs / 1000).toFixed(2)}s</span>
          </div>
        </article>

        <article className="live-card live-card-compact">
          <div className="card-title">Errors (5m)</div>
          <div className="metric">
            {metrics.errors} <span className="delta red inline-delta">↗ 8.3%</span>
          </div>
          <div className="mini-row">
            <span>Rate</span>
            <b>
              {metrics.requestsTotal
                ? ((metrics.errors / metrics.requestsTotal) * 100).toFixed(2)
                : "0.00"}
              %
            </b>
          </div>
        </article>

        <article className="live-card live-card-budget">
          <div className="big-donut" aria-hidden>
            <span>{Math.round(budgetPct)}%</span>
          </div>
          <div>
            <div className="card-title">Token Budget (Today)</div>
            <div className="mini-row">
              <span>Used</span>
              <b>{budgetUsed.toLocaleString()}</b>
            </div>
            <div className="mini-row">
              <span>Limit</span>
              <b>{budgetLimit.toLocaleString()}</b>
            </div>
            <ProgressBar pct={budgetPct} />
          </div>
        </article>
      </section>

      <section className="live-controls" aria-label="Filters">
        <label className="live-select">
          <span className="sr-only">Time range</span>
          <select defaultValue="5m" disabled={paused}>
            <option value="5m">◷ Live: Last 5 minutes</option>
            <option value="1m">Live: Last 1 minute</option>
            <option value="15m">Last 15 minutes</option>
          </select>
        </label>
        <label className="live-select">
          <span className="sr-only">Record limit</span>
          <select defaultValue="500">
            <option value="500">Limit: Latest 500</option>
            <option value="1000">Limit: Latest 1,000</option>
          </select>
        </label>
        <input
          className="live-search"
          type="search"
          placeholder="agent:research-agent response_type:tool_call"
          aria-label="Search and filter events"
        />
        <label className="live-select">
          <span className="sr-only">Providers</span>
          <select defaultValue="all">
            <option value="all">Providers: All</option>
          </select>
        </label>
        <label className="live-select">
          <span className="sr-only">Models</span>
          <select defaultValue="all">
            <option value="all">Models: All</option>
          </select>
        </label>
        <label className="live-select">
          <span className="sr-only">Tags</span>
          <select defaultValue="all">
            <option value="all">Tags: All Tags</option>
          </select>
        </label>
        <button type="button" className="live-select live-select-btn" disabled>
          ☷ More Filters
        </button>
        <label className="live-toggle">
          <span>Follow Tail</span>
          <input type="checkbox" checked={followTail} onChange={(e) => setFollowTail(e.target.checked)} />
          <span className="live-switch" aria-hidden />
        </label>
      </section>

      <section className="live-summary" aria-label="Grid summary">
        <div className="summary-left">
          Showing latest <b>{rows.length}</b> of <b>{metrics.requestsTotal.toLocaleString()}</b> matching events in
          last 5 minutes (Auto-refresh){" "}
          <button type="button" className="summary-link">
            Auto-scroll ON
          </button>
        </div>
        <div className="summary-right">
          <span className="summary-stat error">
            <span className="dot" aria-hidden />
            Errors <b>{metrics.errors}</b>
          </span>
          <span className="summary-stat tool">
            ⌁ Tool Calls <b>{metrics.toolCalls.toLocaleString()}</b>
          </span>
          <span className="summary-stat final">
            ♧ Final Answers <b>{metrics.finalAnswers.toLocaleString()}</b>
          </span>
          <span className="summary-stat block">
            ▱ Blocked <b>{metrics.blocked}</b>
          </span>
          <span className="summary-stat cache">
            ϟ Cache Hits <b>{metrics.cacheHits.toLocaleString()}</b>
          </span>
          <button type="button" className="summary-link" disabled>
            Save View
          </button>
          <button type="button" className="live-pause-btn" onClick={() => setPaused(!paused)}>
            {paused ? "▶ Resume" : "Ⅱ Pause"}
          </button>
        </div>
      </section>

      <section className="live-grid-panel" aria-label="Live event grid">
        <div className="live-grid-scroll">
          <table className="live-table grid-table">
            <thead>
              <tr>
                <th className="idx" />
                <th>Time</th>
                <th>Trace ID</th>
                <th>Agent / User</th>
                <th>Req → Resp</th>
                <th>Status</th>
                <th>Provider / Model</th>
                <th>Response Type</th>
                <th>Next Action</th>
                <th>Tool</th>
                <th>Terminal</th>
                <th>Tags</th>
                <th>Tokens (In/Out)</th>
                <th>Duration</th>
                <th>Cost</th>
                <th>OTel</th>
                <th>Datadog</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row, i) => (
                <EventRow
                  key={`${row.traceId}:${row.spanId}`}
                  index={i + 1}
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
        <>
          <section className="live-selected" aria-label="Selected trace">
            <div className="sel-label">
              Selected Trace: <b>{selectedTraceId}</b>
            </div>
            <span className="in-progress">{traceDetail.status === "complete" ? "Complete" : "In Progress"}</span>
            <div>
              Agent: <b>{traceDetail.agentId}</b>
            </div>
            <div>
              User: <b>{traceDetail.userId}</b>
            </div>
            <div>
              Duration: <b>{(traceDetail.durationMs / 1000).toFixed(2)}s</b>
            </div>
            <div>
              Total Cost: <b>{formatUsd(traceDetail.costUsd)}</b>
            </div>
            <div>
              Tags:{" "}
              <b className="sel-tags">
                {traceDetail.tags.map((t) => `#${t}`).join(", ")}
              </b>
            </div>
            <button type="button" className="small-btn" disabled>
              ↗ View Full Trace
            </button>
            <button type="button" className="small-btn" disabled>
              ◇ Add Tag
            </button>
          </section>

          <TraceDrilldown
            traceId={selectedTraceId}
            traceSteps={traceSteps}
            traceDetail={traceDetail}
            toolCallJson={toolCallJson}
          />
        </>
      ) : null}

      <footer className="live-footer">
        <span>
          System Health <b className="health">● Healthy</b>
        </span>
        <span className="sep" aria-hidden />
        <span>
          Events/sec <b>1,820</b>
        </span>
        <span className="sep" aria-hidden />
        <span>
          WS <b>{connection === "live" ? "Connected" : "Mock"}</b>
        </span>
        <span className="sep" aria-hidden />
        <span>
          Build <b>v0.4.0</b>
        </span>
        <span className="live-footer-right">Time Zone · UTC</span>
      </footer>
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
  const stepTotal = traceSteps.length || 1;

  const circleClass = (s: TraceFlowStep) => {
    if (s.responseType === "TOOL_CALL" || s.responseType === "FINAL_ANSWER") return "circle purple";
    if (s.kind.toLowerCase().includes("response")) return "circle green";
    return "circle";
  };

  return (
    <section className="live-bottom-panel" aria-label="Trace detail">
      <div className="live-tabs" role="tablist">
        {TRACE_TABS.map((tab) => (
          <button
            key={tab}
            type="button"
            role="tab"
            className={`live-tab${activeTab === tab ? " active" : ""}`}
            aria-selected={activeTab === tab}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </div>

      <div className="live-detail">
        <div className="detail-card">
          <h3>
            Call Flow (Trace View)
            <button type="button" className="graph-btn" disabled>
              ↔ View as Graph
            </button>
          </h3>
          <div className="flow-list">
            {traceSteps.map((s) => (
              <div key={s.step} className="flow-step">
                <button
                  type="button"
                  className={`flow-step-inner${activeStep === s.step ? " active-step" : ""}`}
                  onClick={() => setActiveStep(s.step)}
                >
                  <span className={circleClass(s)}>{s.step}</span>
                  <div className={`flow-box${activeStep === s.step ? " active" : ""}`}>
                    <div className="flow-title">
                      {s.kind}
                      {s.responseType ? (
                        <span className={`badge ${RESPONSE_BADGE[s.responseType]}`}>
                          {s.responseType.replace("_", " ")}
                        </span>
                      ) : null}
                    </div>
                    <div className="flow-meta">
                      <span>{s.tool ?? s.nextAction ?? traceDetail.agentId}</span>
                      {s.durationMs ? <span>{(s.durationMs / 1000).toFixed(2)}s</span> : null}
                    </div>
                  </div>
                </button>
              </div>
            ))}
          </div>
        </div>

        <div className="detail-card middle-detail">
          <div className="step-header">
            <span className="step-chip">
              Step {step?.step ?? 1} of {stepTotal}
            </span>
            <span className="step-title">{step?.kind ?? "—"}</span>
            {step?.responseType ? (
              <span className="mini-pill">
                Response Type: <b>{step.responseType.replace("_", " ")}</b>
              </span>
            ) : null}
            {step?.nextAction ? (
              <span className="mini-pill">
                Next Action: <b>{step.nextAction}</b>
              </span>
            ) : null}
          </div>
          <div className="detail-body">
            <div className="tool-summary">
              <div className="kv-card">
                <div className="kv-row">
                  <span>Tool Name</span>
                  <b>{step?.tool ?? "—"}</b>
                </div>
                <div className="kv-row">
                  <span>Span ID</span>
                  <b>{traceDetail.spanId}</b>
                </div>
              </div>
              {toolCallJson ? (
                <pre className="args-box">{toolCallJson}</pre>
              ) : null}
            </div>
            <div className="code-card">
              <div className="code-title">LLM Response (Tool Call Request)</div>
              <pre className="code-pre">{toolCallJson ?? formatJsonDisplay({ trace_id: traceId })}</pre>
            </div>
          </div>
        </div>

        <div className="detail-card side-detail">
          <div className="side-tabs">
            <div className="side-tab active">OTelemetry</div>
            <div className="side-tab">Datadog</div>
          </div>
          <div className="side-content">
            <div className="side-section">
              <div className="side-kv">
                <span>Trace ID</span>
                <b>{traceDetail.traceId}</b>
              </div>
              <div className="side-kv">
                <span>Span ID</span>
                <b>{traceDetail.spanId}</b>
              </div>
              <div className="side-kv">
                <span>Status</span>
                <b className="ok-text">OK</b>
              </div>
              <button type="button" className="link-btn" disabled>
                ↗ Open in Trace Viewer
              </button>
            </div>
            <div className="side-section">
              <h4>Export Status</h4>
              <div className="side-kv">
                <span>OTel Export</span>
                <b className="ok-text">◎ Exported</b>
              </div>
              <div className="side-kv">
                <span>Datadog Export</span>
                <b className="ok-text">◎ Indexed</b>
              </div>
            </div>
          </div>
        </div>

        <div className="detail-card side-detail">
          <div className="side-tabs">
            <div className="side-tab active">Tokens &amp; Budget</div>
          </div>
          <div className="side-content">
            <div className="side-section">
              <h4>This Call</h4>
              <div className="side-kv">
                <span>Input Tokens</span>
                <b>{traceDetail.tokensIn.toLocaleString()}</b>
              </div>
              <div className="side-kv">
                <span>Output Tokens</span>
                <b>{traceDetail.tokensOut.toLocaleString()}</b>
              </div>
              <div className="side-kv">
                <span>Cost</span>
                <b>{formatUsd(traceDetail.costUsd)}</b>
              </div>
            </div>
            <div className="side-section">
              <h4>Today (User)</h4>
              <div className="budget-ring-row">
                <div className="small-donut">
                  <span>36%</span>
                </div>
                <div>
                  <div>
                    Used <b>184,220</b>
                  </div>
                  <div>
                    Limit <b>500,000</b>
                  </div>
                  <div className="delta">Left {formatBudget(traceDetail.budgetRemaining)}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function tagClass(tag: string): string {
  if (tag.includes("tool")) return "tag purple";
  if (tag.includes("error") || tag.includes("block") || tag.includes("policy")) return "tag red";
  if (tag.includes("final") || tag.includes("cache-hit")) return "tag green";
  return "tag";
}

function EventRow({
  index,
  row,
  selected,
  onSelect,
}: {
  index: number;
  row: LiveEventRow;
  selected: boolean;
  onSelect: () => void;
}) {
  const durLabel = row.durationMs >= 1000 ? `${(row.durationMs / 1000).toFixed(2)}s` : `${row.durationMs}ms`;

  return (
    <tr className={selected ? "selected" : ""} onClick={onSelect}>
      <td className="idx">{index}</td>
      <td className="cell-mono">{row.time}</td>
      <td>
        <button type="button" className="trace linkish" onClick={onSelect}>
          {row.traceId}
        </button>
      </td>
      <td>
        <div className="agent">{row.agent}</div>
        <div className="sub">{row.user}</div>
      </td>
      <td className="req">
        <div className="req-line">{row.reqResp}</div>
        <div className="req-dur">{durLabel}</div>
      </td>
      <td>
        <span className={`status ${row.status >= 400 ? "err" : "ok"}`}>{row.status}</span>
      </td>
      <td>
        {row.provider}
        <div className="sub">{row.model}</div>
      </td>
      <td>
        <span className={`badge ${RESPONSE_BADGE[row.responseType]}`}>
          {row.responseType.replace("_", " ")}
        </span>
      </td>
      <td className="col-next" title={row.nextAction}>
        <div className="next-action">{row.nextAction}</div>
        {row.tool && row.tool !== "—" ? <div className="next-tool">{row.tool}</div> : null}
      </td>
      <td className="tool-cell">{row.tool !== "—" ? row.tool : "—"}</td>
      <td>
        <span className={`term ${row.terminal ? "yes" : "no"}`}>{row.terminal ? "Yes" : "No"}</span>
      </td>
      <td>
        {row.tags.map((t) => (
          <span key={t} className={tagClass(t)}>
            #{t}
          </span>
        ))}
      </td>
      <td className="cell-mono">
        {row.tokensIn.toLocaleString()} / {row.tokensOut.toLocaleString()}
      </td>
      <td className="cell-mono">{durLabel}</td>
      <td className="cell-mono">{formatUsd(row.costUsd)}</td>
      <td>
        <span className="check">{row.otelStatus === "exported" ? "◎" : "…"}</span>
      </td>
      <td>
        <span className="wave">{row.datadogStatus === "indexed" ? "⌁" : "…"}</span>
      </td>
    </tr>
  );
}
