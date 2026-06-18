# AgentVoir Tech Stack Usage

This document explains the purpose of key infrastructure components in AgentVoir — **ClickHouse**, **OPA**, **Prometheus**, and **Grafana** — and how far each is wired in the current codebase.

For the high-level system layout, see [Architecture Overview](overview.md).

---

## The three planes (context)

AgentVoir is organized around three planes:

1. **Control plane** — registry API, agent metadata, prompts, policies, lifecycle, budgets (PostgreSQL).
2. **Data plane** — low-latency LLM gateway/proxy, cache, routing, provider adapters (Redis).
3. **Observability / governance plane** — usage analytics, policy enforcement, metrics, dashboards.

```text
Agents / Apps
   |
   v
AgentVoir Gateway
   |-- AuthN/AuthZ
   |-- Policy checks (OPA — planned full integration)
   |-- Cache lookup (Redis)
   |-- Provider routing
   |-- Usage/event emission
   v
Model providers / local models

Registry API  <-> PostgreSQL
Gateway       <-> Redis / semantic cache
Analytics     <-> ClickHouse
Policy        <-> OPA
Telemetry     <-> OpenTelemetry / Prometheus / Grafana
```

End-to-end data flow for usage and governance:

```text
Agents → Gateway → model providers
              ↓
         usage events → token-accounting → ClickHouse
         policy checks → OPA (planned)
         metrics/traces → Prometheus / Grafana / OTel (partial)
```

---

## ClickHouse — usage and cost analytics

### Purpose

Store high-volume **LLM usage events** for cost, token, and performance analytics.

Every gateway request can emit an event containing:

- Agent, tenant, model, provider
- Prompt / completion / cached token counts
- Cost, latency, cache hit/miss status
- HTTP status codes and errors

### How it fits in AgentVoir

```text
Gateway → token-accounting service → ClickHouse
```

- **Schema:** `db/clickhouse/001_usage_events.sql`
- **Service:** `services/token-accounting/` — ingests events via `POST /v1/usage-events`
- **Storage driver:** `services/token-accounting/internal/usage/clickhouse.go`
- **Gateway emitter:** `apps/gateway/internal/usage/recorder.go` (posts to token-accounting when `TOKEN_ACCOUNTING_URL` is set)

When `CLICKHOUSE_DSN` is set (e.g. `http://clickhouse:8123`), events are persisted in ClickHouse. Without it, token-accounting falls back to an in-memory store for local development.

### Why ClickHouse (not PostgreSQL)?

PostgreSQL holds **registry metadata** — agents, prompts, budgets, dependencies. ClickHouse is optimized for **append-only, time-series analytics** at scale: cost by team/agent/model over time, cache analytics, latency trends, and rollups for dashboards and budget enforcement.

### Docker / deployment

| Stack | ClickHouse |
| ----- | ---------- |
| Onebox (`make onebox-up`) | Internal only — not exposed on host |
| Dev stack (`make dev-up`) | Exposed on `:8123` and `:9000` |

Compose definitions:

- `deployments/docker/docker-compose.onebox.yml`
- `deployments/docker/docker-compose.yml`

### Status

**Implemented (Phase 1).** Usage event ingestion to ClickHouse works in onebox and dev stacks.

---

## OPA — policy-as-code / governance

### Purpose

Enforce **who can do what** before model calls — governance as code in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/), not hardcoded rules in the gateway.

### Example policies

Policies live in `policies/opa/agentvoir.rego`. Examples include:

- Only **production** agents may call approved model providers
- **Staging** agents are restricted to the staging environment
- Block requests containing **PII** when the agent is not approved for PII
- Control whether **caching** (exact or semantic) is allowed for a given request

Tests: `policies/opa/agentvoir_test.rego`

### How it fits in AgentVoir

OPA runs as a standalone policy server. AgentVoir services are intended to query OPA with structured input (agent metadata + request context) and act on `allow` / `deny` results before forwarding to a model provider or serving from cache.

Configuration references:

- `config/agentvoir.local.yaml` — `opaUrl: "http://localhost:8181"`
- `deployments/helm/agentvoir/values.yaml` — Helm OPA URL
- `services/worker/README.md` — future policy sync jobs with OPA

### Docker / deployment

| Stack | OPA |
| ----- | --- |
| Onebox | Internal only |
| Dev stack | Exposed on `:8181` |

Policies are mounted read-only from `policies/opa/` into the OPA container.

### Status

**Partially implemented.** Policies are written and tested; OPA runs in Docker. The **gateway does not yet call OPA on every request** — that integration is planned. The worker service is also intended to sync governance state with OPA asynchronously.

---

## Prometheus — metrics collection

### Purpose

Scrape and store **operational metrics** from AgentVoir services for monitoring and alerting:

- Request counts and error rates
- Model and provider latency
- Cache hit / miss / bypass rates
- Token usage and cost counters
- Policy denial counts

### How it fits in AgentVoir

Prometheus is part of the **observability plane**. Services expose a `/metrics` endpoint (planned); Prometheus scrapes those endpoints on an interval and stores time-series data for querying and alerting.

Configuration:

- `observability/prometheus/prometheus.yml` — scrape targets for gateway (`:8080`) and registry API (`:8081`)
- `apps/gateway/internal/observability/metrics.go` — planned metric definitions (TODO)

OpenTelemetry Collector (`observability/otel/collector.yml`) is also included in the dev stack for traces and future metric pipelines.

### Docker / deployment

| Stack | Prometheus |
| ----- | ---------- |
| Onebox | Not included |
| Dev stack | Exposed on `:9090` |

Prometheus and Grafana are **developer observability tools**. They are not required for the onebox try-out experience.

### Status

**Scaffold only.** Prometheus runs in the dev Docker stack, but application-level metrics emission from the gateway and registry API is still TODO.

---

## Grafana — dashboards and visualization

### Purpose

Turn Prometheus (and eventually ClickHouse) data into **operator dashboards**:

- Token usage and cost trends
- Cache hit rates by agent and model
- Gateway latency and error rates
- Per-agent, per-team, and per-tenant views

### How it fits in AgentVoir

Grafana connects to Prometheus as a data source and renders dashboards for the observability plane. Longer term, Grafana can also query ClickHouse directly for cost and usage analytics that are not exposed as Prometheus metrics.

Configuration:

- `observability/grafana/dashboards/agentvoir-overview.json` — placeholder overview dashboard (panels not yet populated)
- Dev stack default credentials: `admin` / `agentvoir` on `:3001`

### Docker / deployment

| Stack | Grafana |
| ----- | ------- |
| Onebox | Not included |
| Dev stack | Exposed on `:3001` (maps to container port 3000) |

### Status

**Scaffold only.** Grafana container runs in the dev stack; dashboards are placeholders awaiting metrics and ClickHouse datasource wiring.

---

## Onebox vs developer stack

| Component | Onebox (`make onebox-up`) | Dev stack (`make dev-up` / `make dev-up-all`) |
| --------- | ------------------------- | --------------------------------------------- |
| ClickHouse | Internal only | `:8123` exposed |
| OPA | Internal only | `:8181` exposed |
| Prometheus | Not included | `:9090` |
| Grafana | Not included | `:3001` |
| OTel Collector | Not included | `:4317` / `:4318` |

Onebox includes ClickHouse and OPA because they support core product behavior (usage analytics and future policy enforcement). Prometheus, Grafana, and OTel are included in the developer stack for local monitoring and debugging.

See also:

- [deployments/docker/INSTALL.md](../../deployments/docker/INSTALL.md) — end-user onebox install guide
- [deployments/docker/README.md](../../deployments/docker/README.md) — onebox vs dev stack comparison

---

## Implementation status summary

| Component | Role in AgentVoir | Phase 1 status |
| --------- | ----------------- | -------------- |
| **ClickHouse** | Analytics database for token, cost, latency, and cache events | Implemented — ingestion works |
| **OPA** | Policy engine: provider access, PII rules, cache permissions | Policies written; gateway integration pending |
| **Prometheus** | Metrics scraping for operational monitoring | Infra ready; app metrics TODO |
| **Grafana** | Dashboards for cost, cache, latency, and health | Infra ready; dashboards TODO |

---

## Related files

| Area | Path |
| ---- | ---- |
| ClickHouse schema | `db/clickhouse/001_usage_events.sql` |
| Token accounting service | `services/token-accounting/` |
| Gateway usage recorder | `apps/gateway/internal/usage/` |
| OPA policies | `policies/opa/` |
| Prometheus config | `observability/prometheus/prometheus.yml` |
| Grafana dashboards | `observability/grafana/dashboards/` |
| OTel collector | `observability/otel/collector.yml` |
| Docker onebox compose | `deployments/docker/docker-compose.onebox.yml` |
| Docker dev compose | `deployments/docker/docker-compose.yml` |
