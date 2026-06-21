"use client";

import Link from "next/link";
import { useLiveStore } from "../../lib/useLiveStore";

export function LiveTopbar() {
  const connection = useLiveStore((s) => s.connection);
  const wsConnected = connection === "live";

  return (
    <header className="live-topbar">
      <Link href="/live" className="live-topbar-brand">
        <span className="live-logo-mark" aria-hidden />
        <span className="live-topbar-brand-name">AgentVoir</span>
      </Link>
      <span className="live-topbar-divider" aria-hidden />
      <h1 className="live-topbar-title">Live Proxy Flow</h1>
      <span className="live-pill">LIVE</span>
      <span className="live-topbar-spacer" />
      <span className="live-topbar-status">
        <span className={`live-status-dot${wsConnected ? " on" : ""}`} aria-hidden />
        WebSocket
      </span>
      <span className="live-topbar-status">
        <span className={`live-status-dot${wsConnected ? " on" : ""}`} aria-hidden />
        {wsConnected ? "Connected" : connection === "connecting" ? "Connecting…" : "Disconnected"}
      </span>
      <button type="button" className="live-btn live-btn-purple" disabled title="Datadog link when OTel export is configured">
        ↗ Open in Datadog ↗
      </button>
      <button type="button" className="live-btn" disabled>
        ⇧ Export ⌄
      </button>
      <button type="button" className="live-btn live-btn-icon" disabled aria-label="Settings">
        ⚙
      </button>
      <span className="live-avatar-top" aria-hidden>
        K
      </span>
    </header>
  );
}
