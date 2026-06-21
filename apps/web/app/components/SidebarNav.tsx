"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { NAV_ICONS } from "./NavIcons";

const NAV_ITEMS = [
  { href: "/live", label: "Live Flow" },
  { href: "/traces", label: "Traces" },
  { href: "/agents", label: "Agents" },
  { href: "/models", label: "Models" },
  { href: "/tools", label: "Tools" },
  { href: "/alerts", label: "Alerts", count: 12 },
  { href: "/analytics", label: "Analytics" },
  { href: "/audit", label: "Audit Logs" },
  { href: "/policies", label: "Policies" },
  { href: "/settings", label: "Settings" },
];

export function SidebarNav() {
  const pathname = usePathname();

  return (
    <nav className="sidebar-nav" aria-label="Main">
      {NAV_ITEMS.map((item) => {
        const active =
          pathname === item.href || (item.href !== "/live" && pathname.startsWith(`${item.href}/`));
        const Icon = NAV_ICONS[item.href as keyof typeof NAV_ICONS];
        return (
          <Link
            key={item.href}
            href={item.href}
            className={`sidebar-link${active ? " active" : ""}`}
            aria-current={active ? "page" : undefined}
          >
            <span className="sidebar-link-icon" aria-hidden>
              {Icon ? <Icon /> : null}
            </span>
            <span className="sidebar-link-label">{item.label}</span>
            {item.count ? <span className="nav-count-badge">{item.count}</span> : null}
          </Link>
        );
      })}
    </nav>
  );
}
