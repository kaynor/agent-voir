import Image from "next/image";
import Link from "next/link";
import type { ReactNode } from "react";
import { SidebarNav } from "./SidebarNav";

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <Link href="/live" className="sidebar-brand" title="AgentVoir">
          <Image src="/agentvoir-logo.svg" alt="AgentVoir" width={22} height={22} className="sidebar-logo" priority />
        </Link>
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
      <div className="app-main">{children}</div>
    </div>
  );
}
