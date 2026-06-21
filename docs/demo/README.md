# AgentVoir demo walkthrough

This walkthrough uses the **customer support agent** scenario — a realistic enterprise use case for registry, gateway, cache, and usage tracking.

## What you will see

1. Register an agent from YAML
2. Call the OpenAI-compatible gateway
3. Observe exact cache (miss → hit)
4. Inspect usage events

## Prerequisites

- Docker with Compose v2
- `curl` and `python3` (optional, for pretty JSON)

## One-command quickstart

Uses **`cache-demo-agent`** — no PII data classes, so exact cache miss → hit works reliably. Governance demos still use `customer-support-agent` (see below).

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
chmod +x scripts/quickstart.sh
./scripts/quickstart.sh
```

Expected output (abbreviated):

```text
==> Registering demo agent from manifest
    Registered cache-demo-agent (201 Created)

==> Gateway chat completion — first request (expect cache miss)
    x-cache-status: miss

==> Gateway chat completion — second request (expect cache hit)
    x-cache-status: hit
    Cache behavior: OK (miss → hit)
```

If onebox is already running:

```bash
./scripts/quickstart.sh --no-start
```

## Step-by-step

### 1. Start the stack

```bash
./scripts/onebox.sh
```

### 2. Register the demo agent

```bash
curl -X POST http://localhost:8081/v1/agents/from-manifest \
  -H "Content-Type: application/yaml" \
  --data-binary @examples/agents/customer-support-agent.yaml
```

The manifest includes model routes, budget, cache policy, dependencies, and governance fields. See [examples/agents/customer-support-agent.yaml](../../examples/agents/customer-support-agent.yaml).

### 3. Call the gateway

```bash
curl -sD - -o /tmp/response.json \
  -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-onebox-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -H "x-tenant-id: acme" \
  --data-binary @examples/demo/sample-chat-request.json
```

Check response headers:

```text
x-cache-status: miss
x-agent-id: customer-support-agent
x-model-used: gpt-4.1-mini
```

Repeat the same command — `x-cache-status` should become `hit`.

Sample response shape: [examples/demo/sample-chat-response.json](../../examples/demo/sample-chat-response.json)

### 4. View usage events

```bash
curl "http://localhost:8082/v1/usage-events?agent_id=customer-support-agent&limit=5" | python3 -m json.tool
```

### 5. Policy denial demo (live)

```bash
./scripts/demo-policy-denial.sh
```

Sends a gateway request with `x-agent-environment: production` for a **draft** agent. OPA denies it before the model is called (HTTP 403).

### 6. Budget block demo (live)

```bash
./scripts/demo-budget-block.sh
```

Registers `budget-demo-agent` with a $0.001 monthly cap, simulates spend via usage ingestion, then shows the gateway blocking the next request (HTTP 429).

### 7. Rate limit demo (live)

```bash
./scripts/demo-rate-limit.sh
```

Registers `rate-limit-demo-agent` with 5 requests/minute, sends repeated gateway calls until the limiter returns HTTP 429 with `Retry-After`.

### 8. Provider fallback demo (live)

```bash
./scripts/demo-fallback.sh
```

Registers `fallback-demo-agent` with primary provider `unavailable` and fallback `mock`. The gateway succeeds via fallback and sets `x-routing-fallback: true`.

### 9. Budget status API

```bash
./scripts/demo-budget-status.sh
```

Calls `GET /v1/agents/budget-demo-agent/budget/status` for monthly utilization fields.

### 10. Policy simulation API

```bash
./scripts/demo-policy-simulate.sh
```

Calls `POST /v1/policies/simulate` to evaluate a draft agent against production policy without sending live traffic.

### 11. Admin console

```bash
make run-web   # http://localhost:3000
```

Dashboard shows agent count, monthly spend, cache hit rate, and agent detail pages with budgets, policies, and dependencies.

### 12. Policy denial scenario (reference)

See [examples/demo/policy-denial-scenario.md](../../examples/demo/policy-denial-scenario.md) for the OPA input shape. The gateway now calls OPA on every upstream request when `OPA_URL` is set.

## Related files

| File | Purpose |
| ---- | ------- |
| `examples/agents/cache-demo-agent.yaml` | Quickstart agent (cache-friendly, no PII) |
| `examples/agents/customer-support-agent.yaml` | Full demo agent (PII, deps, governance) |
| `examples/agents/rate-limit-demo-agent.yaml` | Rate limit demo (5 req/min) |
| `examples/agents/fallback-demo-agent.yaml` | Provider fallback demo |
| `examples/demo/quickstart-chat-request.json` | Quickstart gateway request body |
| `examples/prompts/support-ticket-summary.yaml` | Sample prompt template |
| `examples/policies/no-semantic-cache-for-pii.yaml` | Policy example |
| `examples/demo/sample-chat-request.json` | Gateway request body |
| `scripts/demo-policy-denial.sh` | OPA policy denial demo (403) |
| `scripts/demo-budget-block.sh` | Budget enforcement demo (429) |
| `scripts/demo-rate-limit.sh` | Per-agent rate limit demo (429) |
| `scripts/demo-fallback.sh` | Provider fallback demo |
| `scripts/demo-budget-status.sh` | Budget utilization API demo |
| `scripts/demo-policy-simulate.sh` | Policy simulation API demo |

## Live Proxy Flow dashboard

Chrome Network–style operations console at **http://localhost:3000/live**.

```bash
make demo-live-dashboard   # seed dummy rows + API smoke checks
make run-web               # start console (separate terminal)
```

Full guide: [live-dashboard.md](live-dashboard.md)

## Troubleshooting

| Problem | Fix |
| ------- | --- |
| Port in use | Change ports in `deployments/docker/.env.onebox` |
| `docker pull` fails | Pin `AGENTVOIR_VERSION` to a GitHub Release tag |
| Cache stays `miss` | Ensure `temperature: 0` and identical request body |
| Agent 409 on register | Agent already exists — safe to continue |
| Services not ready | Wait 60s or run `./scripts/onebox-smoke.sh` |

Full install guide: [deployments/docker/INSTALL.md](../../deployments/docker/INSTALL.md)
