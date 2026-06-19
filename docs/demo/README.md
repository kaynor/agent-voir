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

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
chmod +x scripts/quickstart.sh
./scripts/quickstart.sh
```

Expected output (abbreviated):

```text
==> Starting AgentVoir onebox
...
==> Registering demo agent from manifest
    Registered customer-support-agent (201 Created)

==> Gateway chat completion — first request (expect cache miss)
    x-cache-status: miss

==> Gateway chat completion — second request (expect cache hit)
    x-cache-status: hit
    Cache behavior: OK (miss → hit)

==> Recent usage events for customer-support-agent
[
  { "agent_id": "customer-support-agent", "cache_status": "hit", ... },
  { "agent_id": "customer-support-agent", "cache_status": "miss", ... }
]
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

### 5. Policy denial scenario (preview)

See [examples/demo/policy-denial-scenario.md](../../examples/demo/policy-denial-scenario.md) for how OPA policies will gate requests in Phase 2.

## Related files

| File | Purpose |
| ---- | ------- |
| `examples/agents/customer-support-agent.yaml` | Agent manifest (cache, budget, deps) |
| `examples/prompts/support-ticket-summary.yaml` | Sample prompt template |
| `examples/policies/no-semantic-cache-for-pii.yaml` | Policy example |
| `examples/demo/sample-chat-request.json` | Gateway request body |
| `scripts/quickstart.sh` | Automated demo |

## Troubleshooting

| Problem | Fix |
| ------- | --- |
| Port in use | Change ports in `deployments/docker/.env.onebox` |
| `docker pull` fails | Pin `AGENTVOIR_VERSION` to a GitHub Release tag |
| Cache stays `miss` | Ensure `temperature: 0` and identical request body |
| Agent 409 on register | Agent already exists — safe to continue |
| Services not ready | Wait 60s or run `./scripts/onebox-smoke.sh` |

Full install guide: [deployments/docker/INSTALL.md](../../deployments/docker/INSTALL.md)
