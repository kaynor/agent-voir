import Link from "next/link";
import type { ReactNode } from "react";
import "./styles.css";

export const metadata = {
  title: "AgentVoir Console",
  description: "Enterprise AI agent registry and governance console",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="topbar">
          <Link href="/" className="brand">
            AgentVoir
          </Link>
          <nav className="nav">
            <Link href="/">Dashboard</Link>
            <Link href="/agents">Agents</Link>
          </nav>
        </header>
        <main className="shell">{children}</main>
      </body>
    </html>
  );
}
