"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { LiveTopbar } from "./LiveTopbar";
import { SidebarNav } from "./SidebarNav";

function Sidebar({ live = false }: { live?: boolean }) {
  return (
    <aside className="sidebar">
      {!live ? (
        <Link href="/live" className="sidebar-brand" title="AgentVoir">
          <span className="agentvoir-logo-mark" aria-hidden />
        </Link>
      ) : null}
      <SidebarNav />
      <div className="sidebar-footer">
        <div className="user-chip">
          <span className="user-avatar" aria-hidden>
            K
          </span>
          <div className="user-meta">
            <strong>Kailash</strong>
            <span className="user-role">Admin</span>
          </div>
        </div>
      </div>
    </aside>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const isLive = pathname === "/live" || pathname.startsWith("/live/");

  if (isLive) {
    return (
      <div className="app-shell app-shell--live">
        <LiveTopbar />
        <Sidebar live />
        <div className="app-main app-main--live">{children}</div>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <Sidebar />
      <div className="app-main">{children}</div>
    </div>
  );
}
