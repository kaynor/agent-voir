# AgentVoir web console

Next.js admin and **Live Proxy Flow** operations dashboard.

## Routes

| Path | Purpose |
| ---- | ------- |
| `/` | Registry overview (agents, spend, cache) |
| `/live` | Live Proxy Flow UI (mock data until WebSocket backend ships) |
| `/agents` | Agent list and detail |
| `/traces`, `/models`, … | Stubs — see roadmap Milestone H |

## Design

- Architecture: [docs/architecture/ui-dashboard.md](../../docs/architecture/ui-dashboard.md)
- Mockup: [docs/architecture/agent-voir-dashboard-01.png](../../docs/architecture/agent-voir-dashboard-01.png)
- Roadmap: [docs/development-roadmap.md](../../docs/development-roadmap.md) → Operations dashboard

## Stack

- Next.js 14 App Router, React, TypeScript
- TanStack Table + Virtual (grid — wire in Milestone B)
- Zustand (live stream state — Milestone D)
- WebSocket client (Milestone D)

## Dev

```bash
make run-web
# or
cd apps/web && npm install && REGISTRY_API_URL=http://localhost:8081 TOKEN_ACCOUNTING_URL=http://localhost:8082 npm run dev
```

Open http://localhost:3000/live for the Live Proxy Flow console.

**Verify with dummy data:**

```bash
make onebox-up-build
make demo-live-dashboard
make run-web
```
