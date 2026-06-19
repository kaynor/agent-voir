# API usage examples

Copy-paste examples against a running onebox stack (default ports).

## Authentication

### Gateway (Bearer API key)

```bash
export GATEWAY_URL=http://localhost:8080
export REGISTRY_URL=http://localhost:8081
export USAGE_URL=http://localhost:8082
export API_KEY=agentvoir-onebox-key
```

Registry API is open in local onebox (no auth in Phase 1). Production deployments will add OIDC (Phase 2).

---

## Agent registration

### Register via JSON

```bash
curl -X POST "$REGISTRY_URL/v1/agents" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "customer-support-agent",
    "name": "Customer Support Agent",
    "version": "0.1.0",
    "owner_team": "support-platform",
    "environment": "staging",
    "lifecycle": "draft"
  }'
```

### Register from YAML manifest

```bash
curl -X POST "$REGISTRY_URL/v1/agents/from-manifest" \
  -H "Content-Type: application/yaml" \
  --data-binary @examples/agents/customer-support-agent.yaml
```

### List agents

```bash
curl "$REGISTRY_URL/v1/agents"
```

---

## Gateway calls

### Chat completion

```bash
curl -sD - -o /tmp/out.json \
  -X POST "$GATEWAY_URL/v1/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -H "x-tenant-id: acme" \
  --data-binary @examples/demo/sample-chat-request.json
grep -i x-cache-status /tmp/out.json 2>/dev/null || grep -i x-cache-status - /tmp/../headers 2>/dev/null
```

Inspect operational headers: `x-cache-status`, `x-agent-id`, `x-model-used`, `x-cost-usd`, `x-trace-id`.

### List models

```bash
curl "$GATEWAY_URL/v1/models" -H "Authorization: Bearer $API_KEY"
```

---

## Usage events

### Ingest (manual)

```bash
curl -X POST "$USAGE_URL/v1/usage-events" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "customer-support-agent",
    "model": "gpt-4.1-mini",
    "cache_status": "miss",
    "prompt_tokens": 120,
    "completion_tokens": 45,
    "cost_usd": 0.002
  }'
```

### Query recent events

```bash
curl "$USAGE_URL/v1/usage-events?agent_id=customer-support-agent&limit=10"
```

The gateway emits events automatically when `TOKEN_ACCOUNTING_URL` is configured.

---

## Policy simulation (OPA)

With OPA running in onebox (internal) or dev stack on port 8181:

```bash
curl -s http://localhost:8181/v1/data/agentvoir/authz/allow \
  -H "Content-Type: application/json" \
  -d @examples/demo/policy-denial-input.json
```

See [examples/demo/policy-denial-scenario.md](../../examples/demo/policy-denial-scenario.md).

---

## Python SDK

```python
from agentvoir import AgentVoirClient, GatewayClient, RegisterAgentRequest

registry = AgentVoirClient("http://localhost:8081")
print(registry.list_agents())

gateway = GatewayClient("http://localhost:8080", api_key="agentvoir-onebox-key")
print(gateway.list_models())
```

---

## TypeScript SDK

```typescript
import { AgentVoirClient, GatewayClient } from "@agentvoir/sdk";

const registry = new AgentVoirClient({ baseUrl: "http://localhost:8081" });
console.log(await registry.listAgents());

const gateway = new GatewayClient({
  baseUrl: "http://localhost:8080",
  apiKey: "agentvoir-onebox-key",
});
console.log(await gateway.listModels());
```

---

## Automated demo

```bash
./scripts/quickstart.sh
```
