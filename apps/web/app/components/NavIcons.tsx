type IconProps = { className?: string };

const stroke = {
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.75,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
};

export function IconLiveFlow({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M22 12h-4l-3 9L9 3l-3 9H2" />
    </svg>
  );
}

export function IconTraces({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M6 3v12M18 9v12M6 15l6-6 6 6" />
    </svg>
  );
}

export function IconAgents({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <rect {...stroke} x="3" y="8" width="18" height="12" rx="2" />
      <path {...stroke} d="M12 8V5M9 5h6M8 13h.01M12 13h.01M16 13h.01M8 17h.01M12 17h.01M16 17h.01" />
    </svg>
  );
}

export function IconModels({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M12 2l8 4.5v9L12 20l-8-4.5v-9L12 2z" />
      <path {...stroke} d="M12 12l8-4.5M12 12v8M12 12L4 7.5" />
    </svg>
  );
}

export function IconTools({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M14.7 6.3a4 4 0 0 0-5.4 5.4L3 18l3 3 6.3-6.3a4 4 0 0 0 5.4-5.4l-2.1 2.1-3.3-3.3 2.1-2.1z" />
    </svg>
  );
}

export function IconAlerts({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0" />
    </svg>
  );
}

export function IconAnalytics({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M3 3v18h18M7 16v-5M12 16V8M17 16v-3" />
    </svg>
  );
}

export function IconAudit({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6z" />
      <path {...stroke} d="M14 2v6h6M8 13h8M8 17h5" />
    </svg>
  );
}

export function IconPolicies({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <path {...stroke} d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
  );
}

export function IconSettings({ className }: IconProps) {
  return (
    <svg className={className} width="15" height="15" viewBox="0 0 24 24" aria-hidden>
      <circle {...stroke} cx="12" cy="12" r="3" />
      <path
        {...stroke}
        d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"
      />
    </svg>
  );
}

export const NAV_ICONS = {
  "/live": IconLiveFlow,
  "/traces": IconTraces,
  "/agents": IconAgents,
  "/models": IconModels,
  "/tools": IconTools,
  "/alerts": IconAlerts,
  "/analytics": IconAnalytics,
  "/audit": IconAudit,
  "/policies": IconPolicies,
  "/settings": IconSettings,
} as const;
