# Live Proxy Flow dashboard demo

Verify the **Live Proxy Flow** operations console with dummy data and optional real gateway traffic.

Design: [architecture/ui-dashboard.md](../architecture/ui-dashboard.md)

## Quick path (recommended)

```bash
# Terminal 1 — stack (if not already running)
make onebox-up-build

# Terminal 2 — seed dummy rows + API smoke checks
make demo-live-dashboard

# Terminal 3 — web console
make run-web
```

Open **http://localhost:3000/live**

You should see KPI cards, a populated event grid, and a trace drilldown when you click a row.

## Seed commands

| Command | Purpose |
| ------- | ------- |
| `make seed-live-events` | POST ~80 dummy multi-step traces to the gateway |
| `make demo-live-dashboard` | wait for onebox + seed + curl metrics/trace APIs |
| `go run ./apps/gateway/cmd/seed-proxy-events -count 120` | Go CLI (same as seed script) |

Environment variables:

```bash
COUNT=150 RESET=false make seed-live-events   # append without clearing
GATEWAY_URL=http://localhost:8080 GATEWAY_API_KEY=agentvoir-onebox-key make seed-live-events
```

## API verification (manual)

```bash
# List recent rows + metrics
curl -s http://localhost:8080/v1/proxy-events?limit=10 | python3 -m json.tool

# KPI snapshot
curl -s http://localhost:8080/v1/proxy-events/metrics | python3 -m json.tool

# Trace drilldown (replace TRACE_ID)
curl -s http://localhost:8080/v1/traces/TRACE_ID | python3 -m json.tool
```

Seed endpoint (requires API key):

```bash
curl -X POST http://localhost:8080/v1/proxy-events/seed \
  -H "Authorization: Bearer agentvoir-onebox-key" \
  -H "Content-Type: application/json" \
  -d '{"count":80,"reset":true}'
```

## Real traffic (optional)

After seeding, generate live rows from the gateway recorder:

```bash
make quickstart          # cache miss → hit
make demo-policy         # policy block row (403)
make demo-rate-limit     # rate limit row (429)
```

New rows appear on `/live` via WebSocket when **Follow tail** is ON.

## Troubleshooting

| Problem | Fix |
| ------- | --- |
| Empty grid / mock data banner | Run `make seed-live-events` |
| `connection refused` on :8080 | `make onebox-up-build` or `make run-gateway` |
| Web console cannot reach gateway | Ensure `NEXT_PUBLIC_GATEWAY_URL=http://localhost:8080` (set by `make run-web`) |
| CORS errors from browser | Set `CORS_ALLOWED_ORIGINS=http://localhost:3000` in onebox env and restart gateway |
