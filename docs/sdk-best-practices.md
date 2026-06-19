# SDK client best practices

Guidance for Python and TypeScript AgentVoir SDK clients.

## Timeouts

Set explicit timeouts to avoid hung requests when the gateway or registry is unavailable.

**Python**

```python
client = AgentVoirClient("http://localhost:8081", timeout=30.0)
gateway = GatewayClient(..., timeout=120.0)
usage = UsageClient("http://localhost:8082", timeout=30.0)
```

**TypeScript**

```typescript
const client = new AgentVoirClient({ baseUrl: "http://localhost:8081", timeoutMs: 30_000 });
const gateway = new GatewayClient({ baseUrl: "http://localhost:8080", timeoutMs: 120_000 });
const usage = new UsageClient({ baseUrl: "http://localhost:8082", timeoutMs: 30_000 });
```

## Retries

Retry only **idempotent** operations (GET, list) on transient failures (HTTP 502/503/504, connection reset).

- Do **not** blindly retry `POST /v1/chat/completions` — use idempotency keys upstream if your provider supports them.
- `POST /v1/usage-events` may duplicate events if retried; prefer at-most-once delivery or dedupe on `trace_id`.

Example retry loop (Python):

```python
import time
import httpx
from agentvoir import AgentVoirError

def list_agents_with_retry(client, attempts=3):
    delay = 0.5
    for attempt in range(attempts):
        try:
            return client.list_agents()
        except AgentVoirError as exc:
            if exc.status_code not in (502, 503, 504) or attempt == attempts - 1:
                raise
            time.sleep(delay)
            delay *= 2
```

## Error handling

Both SDKs raise `AgentVoirError` / reject with response text and HTTP status. Parse gateway OpenAI-shaped errors from `GatewayClient` responses when status ≥ 400.

## Usage / analytics client

Both SDKs include `UsageClient` for `POST /v1/usage-events` and `GET /v1/usage-events`. Rollup summaries are available at `GET /v1/usage-events/summary?period=daily|monthly`.

## Publishing

Packages are published to PyPI (`agentvoir`) and npm (`@agentvoir/sdk`) on release. See [docs/RELEASE.md](./RELEASE.md).

OpenAPI-driven codegen is optional; hand-written clients track Phase 1 API surface closely.
