![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)
![Stars](https://badgen.net/github/stars/kaynor/agent-voir)
![Issues](https://img.shields.io/github/issues/kaynor/agent-voir)
![Last Commit](https://img.shields.io/github/last-commit/kaynor/agent-voir)

# AgentVoir

**AgentVoir** is an open-source enterprise **AI agent registry**, **low latency LLM gateway**, **model proxy**, **semantic cache**, and **governance control plane** for **agentic AI systems**.

---

### A project of the Agents, by the Agents and for the Agents !!

---

**[Kailash Aynor] This project uses Agent Bots to design and code features for this project. See below Architecture for AI-assisted development flow.

It helps enterprises **register agents**, **govern model/tool access**, **track token usage and cost**, **map dependencies**, **enforce policy-as-code**, and **cache repeated LLM requests** through an OpenAI-compatible proxy.

> Status: early scaffold. This repository is structured for an enterprise-grade implementation, but many modules are intentionally placeholders until the first implementation milestone lands.

![AgentVoir tech architecture](docs/architecture/tech-architecture.png)


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

- Go 1.25+
- Node.js 20+ (only for the web console)
- Docker and Docker Compose
- Make (optional — contributors only; end users do not need it)

### Quick try (recommended for end users)

See **[deployments/docker/INSTALL.md](deployments/docker/INSTALL.md)** for the full Docker install guide.

The **onebox** stack uses **pre-built Docker images** — no Make, no Go, and no local compile. Download a [GitHub Release](https://github.com/kaynor/agent-voir/releases) zip (or clone the repo), then:

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
./scripts/onebox.sh
./scripts/onebox-smoke.sh
```

Or with plain Docker Compose:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml pull
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml up -d
```

Default exposed ports (change in `deployments/docker/.env.onebox` if needed):


| Service          | URL                                            | Notes                 |
| ---------------- | ---------------------------------------------- | --------------------- |
| Gateway          | [http://localhost:8080](http://localhost:8080) | OpenAI-compatible API |
| Registry API     | [http://localhost:8081](http://localhost:8081) | Agent registry        |
| Token accounting | [http://localhost:8082](http://localhost:8082) | Usage events          |


Default gateway API key: `agentvoir-onebox-key`

```bash
export OPENAI_BASE_URL="http://localhost:8080/v1"
export OPENAI_API_KEY="agentvoir-onebox-key"
```

Try a chat completion:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-onebox-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [{"role": "user", "content": "Hello from AgentVoir onebox"}]
  }'
```

Useful onebox commands:

```bash
./scripts/onebox-smoke.sh   # health checks
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml logs -f
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml down
```

**Onebox vs developer setup:** use `./scripts/onebox.sh` to try AgentVoir with pre-built images. Use `make dev-up` / `make dev-up-all` when building from source locally. See [deployments/docker/README.md](deployments/docker/README.md).

### Start local infrastructure (developers)

```bash
cp .env.example .env
make dev-up
```

This starts PostgreSQL, Redis, ClickHouse, OPA, Prometheus, Grafana, and the OpenTelemetry Collector.

### Run the full stack in Docker

To run the AgentVoir application services in containers (registry API, token accounting, and gateway) on top of the same infrastructure:

```bash
cp .env.example .env
make dev-up-all
```

This builds and starts:


| Service          | URL                                            | Storage / deps                        |
| ---------------- | ---------------------------------------------- | ------------------------------------- |
| Gateway          | [http://localhost:8080](http://localhost:8080) | Redis, registry API, token accounting |
| Registry API     | [http://localhost:8081](http://localhost:8081) | PostgreSQL                            |
| Token accounting | [http://localhost:8082](http://localhost:8082) | ClickHouse                            |
| PostgreSQL       | localhost:5432                                 | —                                     |
| Redis            | localhost:6379                                 | —                                     |
| ClickHouse       | [http://localhost:8123](http://localhost:8123) | —                                     |
| Grafana          | [http://localhost:3001](http://localhost:3001) | admin / `agentvoir`                   |
| Prometheus       | [http://localhost:9090](http://localhost:9090) | —                                     |
| OPA              | [http://localhost:8181](http://localhost:8181) | —                                     |


Useful commands:

```bash
make dev-logs     # follow container logs
make dev-down     # stop infrastructure and app containers
```

Optional: pass a real provider key to the gateway container:

```bash
OPENAI_API_KEY=sk-... make dev-up-all
```

The web console is not containerized yet; run it on the host with `make run-web` after `make dev-up-all`.

### Run apps on the host (hybrid mode)

Use `make dev-up` for infrastructure only, then run Go services locally against `localhost` URLs from `.env`.

### Run the registry API

```bash
make run-api
```

Default local URL:

```text
http://localhost:8081
```

### Run token accounting (usage ingestion)

```bash
make run-token-accounting
```

Default local URL:

```text
http://localhost:8082
```

When `CLICKHOUSE_DSN` is set, usage events are stored in ClickHouse. Otherwise the service keeps events in memory for local development.

The gateway emits usage events automatically when `TOKEN_ACCOUNTING_URL` is configured (default: `http://localhost:8082`).

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

### Client SDKs

AgentVoir ships lightweight SDK skeletons for registering agents, calling the gateway, and querying usage events.

#### Python

```bash
cd packages/sdk-python
pip install -e ".[dev]"
pytest
```

```python
from agentvoir import AgentVoirClient, GatewayClient, RegisterAgentRequest, ChatCompletionRequest, ChatMessage

registry = AgentVoirClient("http://localhost:8081")
print(registry.list_agents())

gateway = GatewayClient(
    base_url="http://localhost:8080",
    api_key="agentvoir-local-dev-key",
    agent_id="customer-support-agent",
)
print(gateway.chat_completions(
    ChatCompletionRequest(model="gpt-4.1-mini", messages=[ChatMessage(role="user", content="Hello")])
))
```

See [packages/sdk-python/README.md](packages/sdk-python/README.md).

#### TypeScript

```bash
cd packages/sdk-typescript
npm install
npm run build
npm test
```

```typescript
import { AgentVoirClient, GatewayClient } from "@agentvoir/sdk";

const registry = new AgentVoirClient({ baseUrl: "http://localhost:8081" });
console.log(await registry.listAgents());

const gateway = new GatewayClient({
  baseUrl: "http://localhost:8080",
  apiKey: "agentvoir-local-dev-key",
  agentId: "customer-support-agent",
});
console.log(await gateway.chatCompletions({
  model: "gpt-4.1-mini",
  messages: [{ role: "user", content: "Hello" }],
}));
```

See [packages/sdk-typescript/README.md](packages/sdk-typescript/README.md).

#### Go

```bash
cd packages/sdk-go
go test ./...
```

See [packages/sdk-go/agentvoir/client.go](packages/sdk-go/agentvoir/client.go).

### Usage event ingestion

Every gateway chat completion emits a usage event with token counts, cost, cache status, latency, and trace metadata. Events are sent asynchronously to the token-accounting service so request latency is not affected.

Start the ingestion service, then point the gateway at it. With Docker Compose, `make dev-up-all` starts token accounting and the gateway together:

```bash
make dev-up-all
```

Or run services on the host:

```bash
make run-token-accounting
export TOKEN_ACCOUNTING_URL=http://localhost:8082
make run-gateway
```

Send a gateway request:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-local-dev-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -H "x-tenant-id: acme" \
  -H "x-user-id: user-42" \
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [{"role": "user", "content": "Summarize this support ticket."}],
    "temperature": 0
  }'
```

Query ingested events:

```bash
curl "http://localhost:8082/v1/usage-events?agent_id=customer-support-agent&limit=10"
```

You can also ingest events directly (for custom agents or batch replays):

```bash
curl -X POST http://localhost:8082/v1/usage-events \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "customer-support-agent",
    "agent_version": "0.1.0",
    "provider": "openai",
    "model": "gpt-4.1-mini",
    "cache_status": "miss",
    "prompt_tokens": 120,
    "completion_tokens": 45,
    "cost_usd": 0.0021,
    "latency_ms": 812
  }'
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

Example endpoint:

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

See **[docs/development-roadmap.md](docs/development-roadmap.md)** for plain-language explanations and concrete TODO tasks for each item below.

Status: ✅ done · 🟡 partial · ⬜ not started · 🔒 blocked

### Phase 0: Developer experience and project trust

- ⬜ Quickstart smoke test
- ⬜ Public demo scenario
- ⬜ Contribution-ready issue backlog
- ⬜ API documentation portal

### Phase 1: Registry and exact cache

- ✅ Agent registration API
- ✅ Agent YAML manifest parser
- ✅ OpenAI-compatible gateway endpoint
- ✅ Redis exact cache
- ✅ PostgreSQL metadata schema
- ✅ Usage event ingestion
- ✅ Docker Compose environment
- ✅ Python and TypeScript SDK skeletons
- ⬜ Release security and software supply chain

### Phase 2: Enterprise controls

- ⬜ OIDC authentication
- ⬜ RBAC and service accounts
- 🟡 Per-agent budgets
- ⬜ Per-agent and per-tenant rate limits
- ⬜ Audit logging
- ⬜ Policy-as-code engine
- 🟡 Provider routing and fallback
- ⬜ Provider adapter conformance suite
- 🟡 Dependency graph API
- ⬜ Tool and MCP server registry
- 🟡 OpenTelemetry traces and Prometheus metrics
- ⬜ Pre-flight token and cost estimation
- ⬜ Human-in-the-loop approval gates
- ⬜ Prompt injection and tool-call security
- ⬜ Admin web console

### Phase 3: Semantic cache and evals

- ⬜ Cache correctness and safety framework
- ⬜ RedisVL / Qdrant semantic cache
- ⬜ Cache shadow mode
- 🟡 Prompt registry
- ⬜ Eval datasets and regression runner
- ⬜ Agent scorecards
- ⬜ PII / secret detection hooks

### Phase 4: Kubernetes-native control plane

- 🟡 Helm chart
- 🟡 Kubernetes CRDs: `Agent`, `Prompt`, `ModelRoute`, `AgentPolicy`
- ⬜ Admission controller
- ⬜ GitOps examples
- ⬜ Multi-region routing examples

### Phase 5: Ecosystem and integrations

- ⬜ Framework integrations
- ⬜ CI/CD integrations
- ⬜ Data platform and notification integrations
- ⬜ AgentVoir CLI

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