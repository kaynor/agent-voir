# AgentVoir

**AgentVoir** is an open-source control plane and low-latency LLM gateway for enterprise AI agents.

---

## A project of the Agents, by the Agents and for the Agents !!

---

**[Kailash Aynor] This project uses Agent Bots to design and code features for this project.



It helps enterprises **register agents**, **govern model/tool access**, **track token usage and cost**, **map dependencies**, **enforce policy-as-code**, and **cache repeated LLM requests** through an OpenAI-compatible proxy.

> Status: early scaffold. This repository is structured for an enterprise-grade implementation, but many modules are intentionally placeholders until the first implementation milestone lands.

---

## Why AgentVoir?

Modern enterprises are quickly accumulating dozens or hundreds of AI agents across teams. Without a central registry and gateway, it becomes difficult to answer basic operational questions:

- Which agents exist in production?
- Who owns each agent?
- Which models, tools, APIs, vector stores, and other agents does each agent depend on?
- Which agents can access sensitive data?
- What is the token usage and cost by team, agent, user, workflow, and model?
- Which agents are approved for production?
- Which requests can be cached safely?
- Which model/provider should each agent use?
- What broke after an agent, prompt, model, or dependency changed?

AgentVoir is designed to become the **agent registry + LLM gateway + governance ledger** for enterprise agentic systems.

---

## Core capabilities

### Agent Registry

Register and manage enterprise agents with metadata such as:

- Unique agent identity and version
- Owner team, business unit, cost center
- Runtime framework: LangGraph, CrewAI, AutoGen, custom, etc.
- Capabilities and input/output schemas
- Model routes and fallback models
- Tool, API, vector DB, MCP server, and agent dependencies
- Risk level, data classification, and compliance attributes
- Lifecycle state: draft, review, staging, production, deprecated, retired
- Budgets, token limits, rate limits, and SLOs

### LLM Gateway / Proxy

A low-latency, OpenAI-compatible middle layer for model requests:

- Exact request cache
- Optional semantic cache
- Provider routing and fallback
- Per-agent budgets and rate limits
- Token/cost accounting
- Streaming support
- Audit logging
- OpenTelemetry tracing
- Policy enforcement before model/tool access

### Governance and Observability

- Policy-as-code with OPA/Rego
- Agent dependency graph
- Prompt registry and versioning
- Agent scorecards
- Cost dashboards
- Cache hit-rate metrics
- Eval hooks and regression checks
- PII/secret detection hooks

---

## Repository layout

```text
agentvoir/
  apps/
    gateway/                # Go LLM gateway/proxy: cache, routing, model providers
    registry-api/           # Go registry API: agents, prompts, dependencies, budgets
    web/                    # Next.js admin console
  packages/
    sdk-python/             # Python SDK
    sdk-typescript/         # TypeScript SDK
    sdk-go/                 # Go SDK/client package
  services/
    evaluator/              # Agent eval runner and regression checks
    worker/                 # Async jobs: rollups, cache warming, policy sync
    token-accounting/       # Usage/cost aggregation service
    pii-redactor/           # PII and secret redaction service/plugin
  db/
    migrations/postgres/    # Registry metadata schema
    clickhouse/             # Usage, trace, and analytics schema
  policies/
    opa/                    # Rego policies
  config/                   # Example config files
  deployments/
    docker/                 # Docker Compose and local deployment files
    helm/                   # Kubernetes Helm chart
  infra/
    terraform/              # Terraform examples
    kubernetes/crds/        # Future Agent/Prompt/ModelRoute CRDs
  observability/
    prometheus/             # Prometheus config
    grafana/dashboards/     # Grafana dashboards
    otel/                   # OpenTelemetry Collector config
  examples/
    agents/                 # Example agent manifests
    prompts/                # Example prompt manifests
    policies/               # Example policy manifests
    apps/                   # Example client apps
  docs/
    architecture/           # Design docs and architecture diagrams
    api/                    # API specs
    guides/                 # How-to guides
  tests/
    e2e/                    # End-to-end tests
    contracts/              # API and provider contract tests
    load/                   # Gateway load tests
```

---

## Tech stack


| Layer          | Stack                                   |
| -------------- | --------------------------------------- |
| Gateway        | Go, net/http, Redis, OpenTelemetry      |
| Registry API   | Go, PostgreSQL, sqlc or pgx             |
| Admin UI       | Next.js, React, TypeScript              |
| Hot cache      | Redis                                   |
| Semantic cache | RedisVL, Qdrant, or pgvector            |
| Metadata DB    | PostgreSQL                              |
| Analytics      | ClickHouse                              |
| Policy engine  | OPA/Rego                                |
| Observability  | OpenTelemetry, Prometheus, Grafana      |
| Events/jobs    | NATS, Kafka, Redpanda, or Redis Streams |
| Deployment     | Docker Compose, Helm, Kubernetes        |
| SDKs           | Python, TypeScript, Go                  |


---

## Local development

### Prerequisites

- Go 1.22+
- Node.js 20+
- Python 3.11+
- Docker and Docker Compose
- Make

### Start local infrastructure

```bash
cp .env.example .env
make dev-up
```

This starts local dependencies such as PostgreSQL, Redis, ClickHouse, OPA, Prometheus, Grafana, and the OpenTelemetry Collector.

### Run the registry API

```bash
make run-api
```

Default local URL:

```text
http://localhost:8081
```

### Run the gateway

```bash
make run-gateway
```

Default local URL:

```text
http://localhost:8080
```

### Run the web console

```bash
make run-web
```

Default local URL:

```text
http://localhost:3000
```

---

## Example agent manifest

```yaml
apiVersion: agentvoir.dev/v1alpha1
kind: Agent
metadata:
  name: customer-support-agent
  version: 0.1.0
spec:
  ownerTeam: support-platform
  costCenter: support-ai
  environment: staging
  framework: langgraph
  riskLevel: medium
  dataClasses:
    - customer_pii
  lifecycle: draft
  models:
    primary:
      provider: openai
      model: gpt-4.1-mini
    fallback:
      provider: anthropic
      model: claude-sonnet
  cache:
    mode: exact_only
    ttlSeconds: 86400
    semanticCacheAllowed: false
  budget:
    monthlyUsd: 1000
    maxPromptTokensPerRequest: 12000
    maxCompletionTokensPerRequest: 2000
  dependencies:
    tools:
      - zendesk
      - salesforce
    vectorStores:
      - support-kb-qdrant
    agents:
      - policy-lookup-agent
  policies:
    piiAllowed: true
    requireAuditLog: true
    allowedProviders:
      - openai
      - anthropic
```

---

## Gateway cache modes


| Mode                  | Description                                                                 |
| --------------------- | --------------------------------------------------------------------------- |
| `off`                 | No cache read or write                                                      |
| `exact_only`          | Cache only exact normalized requests                                        |
| `semantic_safe`       | Semantic cache for approved non-sensitive use cases                         |
| `semantic_aggressive` | Higher hit rate, not default for enterprise workloads                       |
| `write_only`          | Write entries but never serve from cache; useful for testing                |
| `shadow`              | Compare cache answer with live model answer without serving cached response |


---

## OpenAI-compatible gateway usage

Apps and agents should be able to point their OpenAI-compatible client to AgentVoir:

```bash
export OPENAI_BASE_URL="http://localhost:8080/v1"
export OPENAI_API_KEY="agentvoir-local-dev-key"
```

Example future endpoint:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-local-dev-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [{"role": "user", "content": "Summarize this support ticket."}],
    "temperature": 0
  }'
```

Gateway responses should eventually include operational headers:

```text
x-agent-id: customer-support-agent
x-agent-version: 0.1.0
x-cache-status: hit | miss | bypass | semantic-hit
x-model-provider: openai
x-model-used: gpt-4.1-mini
x-token-input: 1241
x-token-output: 317
x-cost-usd: 0.0042
x-trace-id: trace-id
```

---

## Development roadmap

### Phase 1: Registry and exact cache

- Agent registration API
- Agent YAML manifest parser
- OpenAI-compatible gateway endpoint
- Redis exact cache
- PostgreSQL metadata schema
- Usage event ingestion
- Docker Compose environment
- Python and TypeScript SDK skeletons

### Phase 2: Enterprise controls

- OIDC authentication
- RBAC and service accounts
- Per-agent budgets
- Per-agent and per-tenant rate limits
- Audit logging
- Provider routing and fallback
- Dependency graph API
- OpenTelemetry traces and Prometheus metrics

### Phase 3: Semantic cache and evals

- RedisVL/Qdrant semantic cache
- Cache shadow mode
- Prompt registry
- Eval datasets and regression runner
- Agent scorecards
- PII/secret detection hooks

### Phase 4: Kubernetes-native control plane

- Helm chart
- Kubernetes CRDs: `Agent`, `Prompt`, `ModelRoute`, `AgentPolicy`
- Admission controller
- GitOps examples
- Multi-region routing examples

---

## Contributing

Contributions are welcome. For early development, start with:

```bash
make fmt
make test
make lint
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

---

## Security

AgentVoir is expected to handle sensitive enterprise AI metadata and traffic. Please do not open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

---

## License

Apache License 2.0. See [LICENSE](LICENSE).
