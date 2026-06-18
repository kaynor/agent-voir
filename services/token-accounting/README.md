# AgentVoir token-accounting

The token-accounting service ingests LLM usage events from the gateway and other AgentVoir components. Events capture token counts, cost, cache status, latency, and request metadata for downstream cost dashboards and budget enforcement.

## Storage

- **Local dev (default):** in-memory store when `CLICKHOUSE_DSN` is unset
- **Production path:** ClickHouse via the HTTP interface when `CLICKHOUSE_DSN` is set (for example `http://localhost:8123`)

The ClickHouse schema matches `db/clickhouse/001_usage_events.sql` and is created automatically on startup.

## Run locally

```bash
export TOKEN_ACCOUNTING_ADDR=:8082
# optional for durable analytics storage
export CLICKHOUSE_DSN=http://localhost:8123
make run-token-accounting
```

## API

### Ingest a usage event

```bash
curl -X POST http://localhost:8082/v1/usage-events \
  -H "Content-Type: application/json" \
  -d '{
    "trace_id": "trace-123",
    "tenant_id": "acme",
    "agent_id": "customer-support-agent",
    "agent_version": "0.1.0",
    "provider": "openai",
    "model": "gpt-4.1-mini",
    "cache_status": "miss",
    "prompt_tokens": 120,
    "completion_tokens": 45,
    "cost_usd": 0.0021,
    "latency_ms": 812,
    "status_code": 200
  }'
```

### List recent events

```bash
curl "http://localhost:8082/v1/usage-events?agent_id=customer-support-agent&limit=20"
```

The gateway emits events automatically when `TOKEN_ACCOUNTING_URL` is configured.
