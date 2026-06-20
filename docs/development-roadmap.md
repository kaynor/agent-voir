# AgentVoir Development Roadmap (Detailed)

This document expands the high-level roadmap in [README.md](../README.md) into concrete work items. Each section includes:

- **What it means** — a plain-language explanation for readers who are not deeply familiar with the underlying technology.
- **TODO items** — specific, actionable tasks for implementers.

Status legend:


| Symbol | Meaning                                              |
| ------ | ---------------------------------------------------- |
| ✅      | Done                                                 |
| 🟡     | Partially done — foundation exists, more work needed |
| ⬜      | Not started                                          |
| 🔒     | Blocked or waiting on an external prerequisite       |


For infrastructure component context (ClickHouse, OPA, Prometheus, Grafana), see [Tech Stack Usage](architecture/tech-stack-usage.md).

**Strategy and metadata discussions** (product direction — inform phases below):


| Document                                           | Topic                                                 |
| -------------------------------------------------- | ----------------------------------------------------- |
| [meta-data.md](meta-data.md)                       | Enterprise metadata model for governed runtime assets |
| [agent-voir-home.md](agent-voir-home.md)           | Personal / home use, permissions, privacy             |
| [mobile-version.md](mobile-version.md)             | Mobile app, device permissions, activity timeline     |
| [future-of-agents.md](future-of-agents.md)         | AI asset hierarchy beyond chat agents                 |
| [data-analytics.md](data-analytics.md)             | Org intelligence, conversation analytics              |
| [agent-quality-review.md](agent-quality-review.md) | Quality scores, feedback loops                        |
| [voice-agents.md](voice-agents.md)                 | Operational / voice / incident responder agents       |
| [multilingual-agents.md](multilingual-agents.md)   | Language governance and localized evals               |
| [non-llm-models.md](non-llm-models.md)             | Embeddings, classifiers, multimodal dependencies      |
| [model-performance.md](model-performance.md)       | Model SLOs and dependency health                      |
| [agents-sunsets.md](agents-sunsets.md)             | Graceful degradation, liquidation readiness           |
| [china-and-robots.md](china-and-robots.md)         | Provider residency, embodied / robot governance       |


These phase items now also include automatic discovery, provenance, sandboxing, secret governance, marketplace security scanning, backup/DR, consent tracking, model catalog drift, browser/desktop monitoring, memory governance, red-team testing, and agent contract validation.

---

## Phase 0: Developer experience and project trust

**Goal:** Make AgentVoir easy to understand, run, evaluate, and contribute to. This phase improves first impressions for open-source users, future contributors, and enterprise evaluators.

---

### ✅ Quickstart smoke test

**What it means:** A single command should prove that the core AgentVoir flow works end-to-end: start the stack, register a demo agent, call the gateway, observe cache behavior, and show a usage event.

**TODO items:**

- [x] Add `./scripts/quickstart.sh`
- [x] Start the onebox Docker stack from the script
- [x] Register a demo `customer-support-agent`
- [x] Send a sample chat completion through the gateway
- [x] Show first request as cache miss and repeated request as cache hit
- [x] Print usage event summary after the request
- [x] Add quickstart output examples to README
- [x] Add troubleshooting section for ports, Docker, and missing API keys

---

### ✅ Public demo scenario

**What it means:** A small, realistic demo helps users understand why AgentVoir exists before they read the full architecture.

**TODO items:**

- [x] Create `examples/agents/customer-support-agent.yaml`
- [x] Include sample prompt, model route, budget, cache policy, and dependencies
- [x] Include a sample LLM request and response
- [x] Include a sample policy denial scenario
- [x] Include screenshot or terminal output in README
- [x] Add "demo walk-through" documentation under `docs/demo/`

---

### ✅ Contribution-ready issue backlog

**What it means:** New contributors should be able to find scoped tasks quickly, and AI-assisted coding workflows should have well-structured issues to consume.

**TODO items:**

- [x] Create labeled GitHub issues for gateway, registry-api, SDK, docs, security, and observability
- [x] Add `good-first-issue` and `help-wanted` labels to approachable tasks
- [x] Add issue templates for bug reports, feature requests, and AI coding tasks
- [x] Add pull request template with AI-assistance disclosure
- [x] Add `docs/AI_CONTRIBUTION_POLICY.md`
- [x] Add architecture decision records under `docs/adr/`
- [x] Add `CODEOWNERS` for key modules

Run `./scripts/bootstrap-github-labels.sh` and `./scripts/bootstrap-github-issues.sh` to apply labels and seed issues on GitHub.

---

### ✅ API documentation portal

**What it means:** Users should be able to explore AgentVoir APIs without reading source code.

**TODO items:**

- [x] Generate OpenAPI spec for registry API
- [x] Generate OpenAPI spec for gateway API extensions
- [x] Publish Swagger UI or Redoc locally in the Docker stack
- [x] Add API docs to GitHub Pages
- [x] Add examples for authentication, agent registration, gateway calls, usage queries, and policy simulation
- [x] Add SDK examples that map to each API section

Local Swagger UI: `docker compose -f deployments/docker/docker-compose.yml --profile docs up -d` → [http://localhost:8089](http://localhost:8089)

---

## Phase 1: Registry and exact cache

**Goal:** Give teams a working foundation — register agents, route LLM requests through a gateway, cache identical requests, track usage, and run everything locally with Docker.

---

### ✅ Agent registration API

**What it means:** A central catalog where every AI agent in your company is recorded — who owns it, what version it is, what environment it runs in, and its lifecycle stage (draft, staging, production, etc.). Think of it like an inventory system for software agents instead of leaving them scattered across repos and configs.

**TODO items:**

- [x] Define agent data model (ID, name, version, owner, environment, lifecycle, risk level)
- [x] Implement `POST /v1/agents` to register a new agent
- [x] Implement `GET /v1/agents` to list agents with optional filters
- [x] Implement `GET /v1/agents/{agentID}` to fetch a single agent
- [x] Implement `PATCH /v1/agents/{agentID}` to update agent metadata
- [x] Persist agents in PostgreSQL (not just in-memory)
- [x] Add HTTP handler tests for registration flows
- [x] Document agent fields in OpenAPI spec
- [x] Add pagination and sorting to list endpoint
- [x] Add lifecycle transition validation (e.g. draft → production requires review)

---

### ✅ Agent YAML manifest parser

**What it means:** Instead of calling APIs field-by-field, operators can describe an agent in a single YAML file (like a recipe card) and upload it. The system reads that file and registers the agent plus its related settings in one step.

**TODO items:**

- [x] Define manifest schema (`apiVersion`, `kind`, `metadata`, `spec`)
- [x] Parse YAML into structured Go types
- [x] Validate required fields (name, version, owner, lifecycle)
- [x] Implement `POST /v1/agents/register-from-manifest` endpoint
- [x] Map manifest fields to agent, budget, dependencies, and model route records
- [x] Add parser unit tests with example manifests
- [x] Return detailed validation errors (line numbers, field names)
- [x] Support manifest import from URL or Git repository
- [x] Add JSON Schema / CRD-compatible validation for manifests

---

### ✅ OpenAI-compatible gateway endpoint

**What it means:** Applications can talk to AgentVoir using the same client libraries and API shape they already use for OpenAI — no custom SDK required to get started. AgentVoir sits in the middle, adds governance, and forwards requests to the real model provider.

**TODO items:**

- [x] Implement `POST /v1/chat/completions` (non-streaming)
- [x] Implement `GET /v1/models`
- [x] Accept `Authorization: Bearer` API key authentication
- [x] Require `x-agent-id` header to identify which agent is calling
- [x] Support optional headers: `x-agent-version`, `x-tenant-id`, `x-user-id`, `x-trace-id`
- [x] Return OpenAI-shaped JSON responses and error format
- [x] Add mock provider for local testing without a real API key
- [x] Add OpenAI provider adapter for live model calls
- [x] Implement streaming (`stream: true`) end-to-end
- [x] Emit operational response headers (`x-cache-status`, `x-cost-usd`, etc.) on every response
- [x] Load agent config from registry API at request time (not headers only)

---

### ✅ Redis exact cache

**What it means:** If the exact same question is asked twice (same model, same messages, same settings), AgentVoir can return the previous answer instantly from memory instead of calling the AI model again — saving money and time. "Exact" means the request must match byte-for-byte after normalization.

**TODO items:**

- [x] Define cache key normalization (model + messages + relevant parameters)
- [x] Implement in-memory cache store for unit tests
- [x] Implement Redis cache store with TTL support
- [x] Integrate cache lookup before provider call in gateway handler
- [x] Write cache entries after successful provider responses
- [x] Record cache status (`hit`, `miss`, `bypass`) in usage events
- [x] Support configurable cache mode via environment (`exact_only`, `off`, etc.)
- [x] Load per-agent cache settings from registry (TTL, mode) instead of global config only
- [x] Add cache bypass rules for sensitive or non-deterministic requests
- [x] Expose cache hit-rate metrics

---

### ✅ PostgreSQL metadata schema

**What it means:** All agent registry information — agents, prompts, budgets, dependencies, and model routes — is stored in a reliable database (PostgreSQL) so it survives restarts and can be queried consistently by multiple services.

**TODO items:**

- [x] Create initial migration for core tables (`agents`, `prompts`, `dependencies`, `budgets`, `model_routes`)
- [x] Add constraints and indexes migration for query performance
- [x] Implement PostgreSQL store layer for each entity type
- [x] Wire registry API to use PostgreSQL when `POSTGRES_DSN` is set
- [x] Add database migration CLI (`make db-migrate`)
- [x] Include migrations in registry-api Docker image
- [x] Add down migrations tested in CI
- [x] Add seed data script for demo agents
- [x] Document schema ER diagram in docs

---

### ✅ Usage event ingestion

**What it means:** Every LLM request generates a receipt — how many tokens were used, how much it cost, how long it took, whether cache was used, and which agent made the call. These receipts are collected so finance and engineering teams can answer "who spent what?"

**TODO items:**

- [x] Define usage event schema (agent, model, tokens, cost, latency, cache status)
- [x] Build token-accounting HTTP service (`POST /v1/usage-events`, `GET /v1/usage-events`)
- [x] Implement in-memory store for local dev without ClickHouse
- [x] Implement ClickHouse store with auto-created table
- [x] Gateway emits usage events asynchronously after each request
- [x] Add ClickHouse DDL matching `db/clickhouse/001_usage_events.sql`
- [x] Docker-compose wiring for token-accounting + ClickHouse
- [x] Compute cost automatically from model pricing table (not caller-supplied)
- [x] Add daily/monthly rollup jobs for budget dashboards
- [x] Add retention policy and archival for old events

---

### ✅ Docker Compose environment

**What it means:** Anyone can start the full AgentVoir stack on their laptop with one command — no manual database installs, no port conflicts with existing Postgres, and no need to install Go or Node just to try the product.

**TODO items:**

- [x] Create developer Docker Compose stack (Postgres, Redis, ClickHouse, OPA, Prometheus, Grafana, OTel)
- [x] Create onebox stack (isolated, minimal host ports, self-contained)
- [x] Add Dockerfiles for gateway, registry-api, and token-accounting (developer stack)
- [x] Add unified onebox Dockerfile (`deployments/docker/Dockerfile`) — one GHCR package per release
- [x] Add healthchecks for infrastructure services
- [x] Add Makefile targets: `onebox-up`, `onebox-down`, `onebox-smoke`, `dev-up`, `dev-up-all`
- [x] Create `.env.onebox.example` for port and API key configuration
- [x] Write end-user install guide (`deployments/docker/INSTALL.md`)
- [x] Document onebox vs dev stack differences
- [x] Switch onebox to pre-built GHCR images (no local build for end users)
- [x] Add GitHub Actions workflow to build and push images on release (`.github/workflows/release-images.yml`)
- [x] Docker-only start path (`./scripts/onebox.sh` — no Make required)
- [ ] Publish first GitHub Release and make GHCR package public *(maintainer action — see [docs/RELEASE.md](RELEASE.md))*
- [x] Add docker-compose health wait script for smoother first-run UX

---

### ✅ Python and TypeScript SDK skeletons

**What it means:** Developer-friendly libraries so teams can register agents, call the gateway, and query usage from Python or JavaScript/TypeScript apps without writing raw HTTP code themselves.

**TODO items:**

- [x] Python: `AgentVoirClient` for registry API
- [x] Python: `GatewayClient` for chat completions and model listing
- [x] Python: typed request/response models (Pydantic)
- [x] Python: unit tests for client methods
- [x] TypeScript: registry client with typed models
- [x] TypeScript: gateway client for chat completions
- [x] TypeScript: unit tests
- [x] README and install instructions for both SDKs
- [x] Add usage/analytics client to both SDKs
- [x] Publish to PyPI and npm *(workflow + docs; requires maintainer secrets)*
- [x] Add retry, timeout, and error-handling best practices to docs
- [ ] Generate SDKs from OpenAPI spec (optional automation)

---

### ✅ Release security and software supply chain

**What it means:** Enterprises need to trust the artifacts they run. AgentVoir releases should include signed images, software bills of materials, vulnerability scans, and provenance so operators can verify what they deploy.

**TODO items:**

- [x] Generate SBOM for every Docker image
- [x] Sign container images with Sigstore/cosign
- [x] Publish provenance attestation for release builds
- [x] Add vulnerability scanning for images and dependencies
- [x] Add dependency review in CI
- [x] Add license scanning in CI
- [x] Pin GitHub Actions versions or use trusted reusable workflows
- [x] Add release checklist for maintainers
- [x] Document artifact verification steps for users
- [x] Add `SECURITY.md` release and disclosure expectations

---

## Phase 2: Enterprise controls

**Goal:** Make AgentVoir safe and manageable for real enterprise deployments — proper login, permissions, spending limits, audit trails, and operational visibility.

**Acronym quick reference** (used throughout this phase):

| Term | Plain English |
| ---- | ------------- |
| **API** | A way for software programs to talk to AgentVoir over the network. |
| **Gateway** | The front door that receives AI requests, applies rules, and forwards them to model providers. |
| **Registry** | The catalog service that stores agent definitions, budgets, policies, and dependencies. |
| **YAML manifest** | A human-readable config file that describes an agent in one place. |
| **OIDC** | OpenID Connect — log in with your company account (Okta, Google, Microsoft) instead of a shared password. |
| **JWT** | JSON Web Token — a signed digital pass that proves who you are for a short time. |
| **RBAC** | Role-Based Access Control — different people get different permissions (admin vs viewer). |
| **OPA** | Open Policy Agent — a rules engine that answers “is this request allowed?” |
| **MCP** | Model Context Protocol — a standard way for agents to call external tools and data sources. |
| **Redis** | A fast in-memory database used for caching and rate counting. |
| **PostgreSQL** | The main relational database where agent metadata is stored. |
| **ClickHouse** | A database optimized for usage and cost analytics over time. |
| **OTel** | OpenTelemetry — industry standard for tracing where time is spent inside a request. |
| **Prometheus / Grafana** | Prometheus collects metrics; Grafana draws charts and dashboards from them. |
| **SIEM** | Security Information and Event Management — central log systems like Splunk or Datadog. |
| **PII** | Personally Identifiable Information — names, emails, SSNs, etc. |
| **GHCR** | GitHub Container Registry — where pre-built AgentVoir Docker images are published. |
| **HITL** | Human-in-the-loop — a person must approve before a risky action proceeds. |
| **IdP** | Identity Provider — the system that issues login tokens (Okta, Azure AD, Keycloak). |

### GitHub showcase track (v2 — ready for release)

High-impact items for visitors evaluating AgentVoir on GitHub. Run **`make showcase`** after onebox is up.

**Governance (gateway + registry):**

- [x] Gateway OPA policy check before upstream calls (403 on deny) — *Run your governance rules before calling the AI provider; reject disallowed requests with “forbidden.”*
- [x] Gateway monthly budget enforcement (429 on exceed) — *Block requests when an agent has spent its monthly dollar allowance.*
- [x] Per-agent rate limits — Redis fixed-window, 429 + `Retry-After` — *Cap requests per minute per agent using Redis; tell callers how long to wait before retrying.*
- [x] Provider routing fallback — primary fails → backup (`x-routing-fallback`) — *Automatically switch to a backup AI provider when the primary one fails.*
- [x] Budget utilization API — `GET /v1/agents/{agentID}/budget/status` — *Let dashboards and scripts read how much budget an agent has left.*
- [x] Policy simulation API — `POST /v1/policies/simulate` — *Test “would this request be allowed?” without sending it to a real model.*
- [x] Persist agent policies, budgets, and model routes from YAML manifest — *Import rules, spending caps, and routing from the agent config file into the database.*

**Demos and docs:**

- [x] Demo scripts: `demo-policy-denial`, `demo-budget-block`, `demo-rate-limit`, `demo-fallback`, `demo-budget-status`, `demo-policy-simulate` — *Runnable scripts that show each governance feature working end-to-end.*
- [x] Example agents: `rate-limit-demo-agent`, `fallback-demo-agent` — *Sample agent configs tuned to trigger rate limits and provider fallback.*
- [x] Cache-friendly quickstart path (`cache-demo-agent`, miss → hit) — *A demo where the first request is slow (cache miss) and the repeat is instant (cache hit).*
- [x] Demo walkthrough — [docs/demo/README.md](demo/README.md) — *Step-by-step guide for evaluators walking through the demos.*

**Console and observability:**

- [x] Admin web console MVP (dashboard, agent list, agent detail) — *A browser UI to browse agents and see basic health without using raw APIs.*
- [x] Grafana overview dashboard panels (cache, policy, budget metrics) — *Charts showing cache hits, policy blocks, and budget usage over time.*

**Remaining showcase polish:**

- [ ] README screenshots / GIF of admin console — *Visual proof on the project homepage so visitors see the UI without running Docker.*
- [ ] Publish GitHub Release with showcase v2 features (GHCR tag + onebox bundle) — *Ship a versioned Docker image and one-command installer for the showcase features.*

---

### ⬜ OIDC authentication

**What it means:** Users log in with your company's existing identity system (Okta, Azure AD, Google Workspace, etc.) instead of shared static API keys. AgentVoir verifies "who is this person or service?" using industry-standard OpenID Connect (OIDC) tokens.

**TODO items:**

- [x] Add OIDC provider configuration (issuer URL, client ID, client secret) — *Tell AgentVoir how to connect to your company login system.*
- [x] Validate JWT access tokens on registry API requests — *Check that registry callers present a valid signed login token, not just a shared key.*
- [x] Validate JWT access tokens on gateway requests (or accept OIDC + API key hybrid) — *Same for AI traffic; allow both corporate login and legacy API keys during migration.*
- [x] Map OIDC claims (`sub`, `email`, `groups`) to AgentVoir user identity — *Turn token fields (user ID, email, team groups) into an AgentVoir user record.*
- [x] Add integration tests with a local OIDC mock (e.g. Dex) — *Automated tests that log in locally without needing a real Okta or Azure tenant.*
- [x] Document OIDC setup for common providers (Okta, Azure AD, Keycloak) — *Step-by-step guides for the identity systems enterprises already use.*
- [ ] Support machine-to-machine (client credentials) flow for automated agents — *Let unattended software agents authenticate without a human browser login.*
- [ ] Deprecate or gate static API keys behind admin-only bootstrap mode — *Reduce reliance on long-lived shared secrets once SSO is in place.*

---

### ⬜ RBAC and service accounts

**What it means:** Not everyone should be able to do everything. Role-Based Access Control (RBAC) defines who can register agents, change production settings, view costs, or administer policies. Service accounts give automated systems their own credentials with limited permissions.

**TODO items:**

- [ ] Define roles (e.g. `admin`, `agent-owner`, `viewer`, `auditor`) — *Name the job titles that map to different access levels.*
- [ ] Define permissions (register agent, promote lifecycle, view usage, edit budgets) — *List the specific actions each role may or may not perform.*
- [ ] Store roles and role bindings in PostgreSQL — *Save who has which role in the database so it survives restarts.*
- [ ] Enforce permissions on registry API endpoints — *Reject registry changes when the caller lacks the right role.*
- [ ] Create service account entity with scoped API tokens — *Give each automated system its own limited credential, separate from human logins.*
- [ ] Implement `POST /v1/service-accounts` and token rotation — *API to create service accounts and replace their tokens on a schedule.*
- [ ] Wire RBAC checks into gateway for sensitive operations — *Apply the same permission rules to high-risk gateway actions.*
- [ ] Add audit log entry on permission denied events — *Record when someone tried to do something they were not allowed to do.*
- [ ] Document default roles and recommended enterprise mappings — *Guidance for mapping your org chart to AgentVoir roles.*

---

### ⬜ Secret and credential governance

**What it means:** Agent capabilities often come from credentials. If an agent has a Slack token, GitHub token, cloud IAM role, SaaS OAuth grant, or banking API key, that credential becomes part of the agent's risk profile. AgentVoir should never store secrets directly, but it should track secret references, ownership, scope, rotation, and blast radius.

**TODO items:**

- [ ] Add `SecretRef` entity: provider, path, owner, environment, expiration, rotation interval, scopes — *Track *where* a password lives (e.g. Vault path) without storing the password itself.*
- [ ] Link secret refs to agents, tools, MCP servers, model providers, and external systems — *Show which agents depend on which credentials.*
- [ ] Add secret scope metadata: read-only, write, admin, payment, production, customer-data — *Describe how powerful each credential is if misused.*
- [ ] Add secret rotation status and expiry alerts — *Warn before passwords expire or miss their rotation schedule.*
- [ ] Add policy: production agents cannot depend on expired or unowned secrets — *Block production traffic if credentials are stale or have no accountable owner.*
- [ ] Add blast-radius query: "which agents depend on this secret?" — *Answer impact questions before revoking a shared API key.*
- [ ] Add emergency revoke workflow: revoke secret → quarantine affected agents → notify owners — *One-click response when a credential is compromised.*
- [ ] Add secret leakage detector in prompts, responses, traces, and logs — *Scan traffic for accidentally exposed API keys or passwords.*
- [ ] Add integration examples for Vault, AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, Doppler, and 1Password — *Show how to plug in common enterprise secret stores.*

---

### 🟡 Per-agent budgets

**What it means:** Set spending and usage caps per agent — for example, "this support bot can spend at most $1,000/month" or "no single request may exceed 12,000 input tokens." AgentVoir blocks or warns when limits are exceeded.

**TODO items:**

- [x] Define budget model (monthly USD, max tokens per request) — *Data shape for “how much money per month” and “how big one request can be.”*
- [x] Implement `GET/PUT /v1/agents/{agentID}/budget` registry endpoints — *API to read and update an agent’s spending limits.*
- [x] Persist budgets in PostgreSQL — *Store limits in the database so they are not lost on restart.*
- [x] Accept budget fields from agent YAML manifest — *Set budgets when importing an agent from a config file.*
- [x] Gateway loads budget for agent on each request — *Look up the current cap before every AI call.*
- [ ] Enforce max tokens per request before calling provider — *Reject oversized requests before they reach the billable model.*
- [x] Track cumulative spend per agent per month (from ClickHouse rollups) — *Roll up usage receipts into monthly totals per agent.*
- [x] Return `429` or structured error when monthly budget exceeded — *Stop traffic cleanly when the monthly cap is hit.*
- [x] Add budget utilization API (`GET /v1/agents/{agentID}/budget/status`) — *Expose “spent vs remaining” for dashboards and alerts.*
- [ ] Optional: soft limits (warn) vs hard limits (block) — *Let some agents warn at 80% but only block at 100%, if configured.*
- [ ] Notify owners when budget reaches 80% / 100% thresholds — *Email or Slack the agent owner before and when the cap is reached.*

---

### 🟡 Per-agent and per-tenant rate limits

**What it means:** Prevent any single agent or customer (tenant) from flooding the gateway with too many requests per minute. Protects shared infrastructure and prevents runaway automation loops from causing outages or surprise bills.

**TODO items:**

- [x] Add rate limit fields to budget/config model (requests per minute) — *Store “max requests per minute” alongside budget settings.*
- [x] Implement fixed-window limiter in gateway (Redis-backed) — *Count requests in one-minute buckets using Redis for speed.*
- [x] Apply limits per agent ID — *Each agent gets its own request cap.*
- [x] Apply limits per tenant ID (`x-tenant-id` header) — *Also cap traffic per customer or business unit sharing the gateway.*
- [x] Return `429 Too Many Requests` with `Retry-After` header — *Tell callers they are going too fast and when to try again.*
- [ ] Record rate-limit events in usage/analytics stream — *Log throttled requests for ops and billing dashboards.*
- [ ] Admin API to view current rate-limit state per agent — *Let operators see who is close to or over their limit right now.*
- [ ] Add tokens per minute limit field — *Throttle by AI token volume, not just request count.*
- [ ] Configurable burst allowance vs sustained rate — *Allow short spikes while still limiting average load over time.*
- [ ] Load test rate limiter under concurrent load — *Prove the limiter stays correct when many requests arrive at once.*

---

### ⬜ Audit logging

**What it means:** A tamper-evident record of who changed what and when — agent registrations, policy updates, budget changes, production promotions. Required for compliance, security reviews, and debugging "who broke production?"

**TODO items:**

- [ ] Define audit event schema (actor, action, resource, timestamp, before/after snapshot) — *Standard record: who did what, to which agent, when, and what changed.*
- [ ] Create `audit_events` table in PostgreSQL (append-only) — *Write-once history that cannot be silently edited.*
- [ ] Emit audit events from registry API on all mutating operations — *Log every create, update, and delete in the catalog.*
- [ ] Emit audit events from gateway on policy denials and budget blocks — *Log when traffic was stopped by rules or spending caps.*
- [ ] Implement `GET /v1/audit-events` with filters (agent, actor, date range) — *Searchable audit trail for security and compliance reviews.*
- [ ] Optional: ship audit logs to SIEM (Splunk, Datadog) via webhook — *Forward events to your company’s central security log system.*
- [ ] Retention policy configuration (e.g. 90 days hot, 7 years archive) — *Keep recent logs fast to query; archive older records for regulations.*
- [ ] Document audit log fields for compliance teams — *Explain each field so auditors know what evidence AgentVoir provides.*

---

### 🟡 Policy-as-code engine

**What it means:** A centralized policy layer decides whether an agent may call a model, use a tool, cache a response, access a dependency, or move to production. This makes AgentVoir a governance control plane instead of only a proxy.

**TODO items:**

- [ ] Define standard OPA input schema for gateway requests — *Agree on the facts (agent, model, user, cost estimate) that gateway rules can inspect.*
- [ ] Define standard OPA input schema for registry mutations — *Same for catalog changes like promoting an agent to production.*
- [x] Add default policy: deny semantic cache when PII is present — *Do not reuse cached answers when personal data might be involved.*
- [ ] Add default policy: deny production agents without owner/team — *Production agents must have a named owner before going live.*
- [x] Add default policy: deny unapproved model providers — *Only call AI vendors that your organization has approved.*
- [ ] Add default policy: deny high-risk agents without audit logging — *High-risk agents must write audit records or they cannot run.*
- [ ] Add default policy: deny tool access outside approved dependency list — *Agents may only call tools explicitly registered for them.*
- [x] Implement gateway policy check before provider call — *Evaluate rules on every AI request before money is spent.*
- [ ] Implement registry policy check before lifecycle promotion — *Evaluate rules before an agent moves to staging or production.*
- [ ] Add policy decision logs to audit events — *Record allow/deny decisions with the reason for later review.*
- [ ] Add policy test fixtures using `opa test` — *Automated tests that prove your Rego rules behave as expected.*
- [x] Add policy simulation endpoint: `POST /v1/policies/simulate` — *Dry-run rules against a sample request without calling a model.*
- [ ] Document "Writing AgentVoir policies" — *Author guide for security and platform teams writing governance rules.*

---

### 🟡 Provider routing and fallback

**What it means:** If the primary AI provider (e.g. OpenAI) is down, slow, or rejects a request, AgentVoir automatically tries a backup provider (e.g. Anthropic) according to rules defined for each agent — similar to how DNS failover works for websites.

**TODO items:**

- [x] Define model route schema (primary provider/model, fallback provider/model) — *Config format for “try OpenAI first, then Anthropic.”*
- [x] Implement registry API for model routes (`GET/PUT /v1/agents/{agentID}/model-route`) — *API to read and set routing per agent.*
- [x] Accept model routes from agent YAML manifest — *Declare routing in the agent config file.*
- [x] Gateway provider registry with OpenAI and mock adapters — *Built-in connectors for OpenAI and a fake provider for testing.*
- [x] Gateway loads model route from registry for each agent — *Pick the right providers on every request.*
- [x] Attempt primary provider first; on failure, try fallback — *Automatic failover when the main vendor errors or times out.*
- [x] Configurable routing policy (`primary_then_fallback`; `primary_only` supported) — *Choose failover vs single-provider-only behavior.*
- [ ] Add `round_robin` routing policy — *Spread load across multiple providers instead of always preferring one.*
- [ ] Add Anthropic, Azure OpenAI, and local model adapters — *Support more vendors and on-prem models.*
- [x] Record which provider was actually used in usage events and response headers — *Billing and debugging show the provider that answered.*
- [ ] Circuit breaker when provider error rate exceeds threshold — *Temporarily stop sending traffic to a failing vendor.*
- [ ] Admin UI or API to test routing without live traffic — *Safe way to verify routing config before production traffic hits it.*

---

### ⬜ Model/provider catalog and pricing drift monitor

**What it means:** Model prices, terms, context windows, regions, rate limits, and capabilities change frequently. Hardcoded pricing tables and static model assumptions will drift. AgentVoir should maintain a first-class catalog of provider/model capabilities and alert when price, capability, terms, or deprecation changes affect agents.

**TODO items:**

- [ ] Add `ModelCatalog` entity: provider, model, modality, context window, tool support, JSON mode, streaming, region support — *A reference list of what each AI model can do and where it runs.*
- [ ] Add pricing history table: input token price, output token price, cached-token price, image/audio/video price — *Track price changes over time so cost math stays accurate.*
- [ ] Add provider terms metadata: data-retention policy, training policy, region availability, enterprise plan requirement — *Record legal/ops constraints, not just technical specs.*
- [ ] Add model deprecation date and replacement model recommendation — *Know when a model will shut down and what to migrate to.*
- [ ] Add scheduled pricing update workflow with manual approval — *Review and approve vendor price changes before they affect budgets.*
- [ ] Add alert when provider price changes affect monthly budget projections — *Warn finance when a price hike breaks an agent’s monthly cap.*
- [ ] Add model capability diff: "new version changes context window/tool support" — *Highlight breaking or beneficial changes between model versions.*
- [ ] Add policy: block deprecated models for production agents after cutoff date — *Prevent production traffic to models past their end-of-life.*
- [ ] Add provider health dashboard: latency, error rate, rate-limit rate, cost trend, fallback usage — *One screen showing how each AI vendor is performing.*

---

### ⬜ Provider adapter conformance suite

**What it means:** Every model provider adapter should behave consistently so routing, caching, cost tracking, streaming, and fallback work the same way across OpenAI, Anthropic, Azure OpenAI, Gemini, Bedrock, and local models.

**TODO items:**

- [ ] Define provider adapter interface — *A common contract every AI vendor connector must implement.*
- [ ] Define normalized request and response structs — *Translate different vendor formats into one internal shape.*
- [ ] Add conformance tests for non-streaming chat — *Verify one-shot chat works the same on every provider.*
- [ ] Add conformance tests for streaming chat — *Verify live token streaming works the same on every provider.*
- [ ] Add conformance tests for tool calls — *Verify “call a function” behavior is consistent across vendors.*
- [ ] Add conformance tests for provider errors and timeouts — *Verify failures are handled uniformly for routing and retries.*
- [ ] Add conformance tests for token usage extraction — *Verify billing counts match what each vendor reports.*
- [ ] Add finish reason mapping across providers — *Map vendor-specific “why the model stopped” codes to one enum.*
- [ ] Add mock provider test harness — *Fake provider for running the full test suite without API keys.*
- [ ] Add adapter capability discovery (streaming, tools, JSON mode, embeddings) — *Advertise what each connector supports at runtime.*
- [ ] Add per-provider retry/backoff config — *Tune how aggressively to retry each vendor on transient errors.*
- [ ] Document how contributors can add a new provider — *Guide for adding the next AI vendor to AgentVoir.*

---

### 🟡 Dependency graph API

**What it means:** A map of what each agent depends on — other agents, tools (Zendesk, Salesforce), vector databases, MCP servers. Helps answer impact analysis: "If we change this tool, which agents break?"

**TODO items:**

- [x] Define dependency model (tools, vector stores, agents, APIs) — *Types of things an agent can depend on (CRM, database, another agent, etc.).*
- [x] Implement `GET/PUT /v1/agents/{agentID}/dependencies` registry endpoints — *API to list and update an agent’s dependency list.*
- [x] Persist dependencies in PostgreSQL — *Store dependency links in the database.*
- [x] Accept dependencies from agent YAML manifest — *Declare dependencies in the agent config file.*
- [ ] Implement graph query API (`GET /v1/dependency-graph?agent_id=...`) — *Fetch the full web of what an agent relies on and what relies on it.*
- [ ] Return upstream and downstream dependents (transitive closure) — *Include indirect dependencies, not just immediate neighbors.*
- [ ] Visualize graph in web console — *Draw the dependency map for non-technical stakeholders.*
- [ ] Detect circular agent dependencies and reject on registration — *Prevent agent A → B → A loops that can cause infinite calls.*
- [ ] Export graph as JSON/GraphML for external tools — *Let architecture or GRC tools import the map.*
- [ ] Link dependency changes to audit log — *Record who added or removed a dependency and when.*

---

### ⬜ Tool and MCP server registry

**What it means:** Agent risk depends heavily on the tools an agent can invoke. AgentVoir should track tools and MCP servers as governed dependencies with owners, permissions, risk levels, secrets, and audit requirements.

**TODO items:**

- [ ] Define tool registry model (`tool_id`, name, owner, protocol, risk level, allowed scopes) — *Catalog entry for each action an agent can perform (send email, query CRM, etc.).*
- [ ] Support tool protocols: HTTP, gRPC, MCP, and function-style tools — *Support common ways agents invoke external capabilities.*
- [ ] Define MCP server registry model — *Register whole MCP servers (bundles of tools) as governed assets.*
- [ ] Link tools and MCP servers to agent dependencies — *Show which agents are allowed to call which tools.*
- [ ] Enforce tool allowlist in gateway/policy layer — *Block tool calls that are not registered and approved.*
- [ ] Add tool-call audit events — *Log every tool invocation for security review.*
- [ ] Add tool-call traces using OpenTelemetry — *Show tool latency and errors inside distributed traces.*
- [ ] Add tool risk review workflow — *Require human approval before high-risk tools go to production.*
- [ ] Add "disable tool globally" kill switch — *Instantly stop one tool everywhere if it is compromised.*
- [ ] Add docs: "Registering MCP servers and tools" — *How-to for platform teams onboarding new tools.*
- [ ] Add example MCP server manifest — *Sample config teams can copy when registering an MCP server.*

---

### ⬜ Tool execution sandbox and permission broker

**What it means:** Listing tools is not enough. Tool calls are where agents can change real systems, leak data, spend money, or damage production. AgentVoir should mediate high-risk tool execution through a broker that enforces policy, validates arguments, limits runtime, controls egress, and records side effects.

**TODO items:**

- [ ] Define `ToolExecutionBroker` service boundary — *A middle layer that sits between agents and tools instead of direct calls.*
- [ ] Route registered tool calls through broker instead of direct agent-to-tool calls — *Every tool action passes through governance checks.*
- [ ] Add argument schema validation before execution — *Reject malformed or dangerous tool inputs before they run.*
- [ ] Add network egress allowlist per tool and per agent — *Limit which internet destinations a tool may contact.*
- [ ] Add filesystem sandbox policy for code/file tools — *Restrict which folders a code-running tool can read or write.*
- [ ] Add command allowlist/denylist for shell/code-execution tools — *Block dangerous shell commands like `rm -rf`.*
- [ ] Add per-tool timeout, memory, CPU, and output-size limits — *Prevent runaway tools from hanging or flooding output.*
- [ ] Add side-effect ledger: external system changed, before/after if available, rollback hint — *Record what changed in the real world when a tool ran.*
- [ ] Add dry-run mode for tools that support preview — *Show what a tool *would* do without actually doing it.*
- [ ] Add tool output classification: public/internal/confidential/PII/secret — *Label tool results so downstream policies can treat them safely.*
- [ ] Add global tool kill switch and per-agent tool quarantine — *Emergency stops for one tool or one agent’s tools.*
- [ ] Add sample sandbox adapter for MCP tools — *Reference implementation for sandboxing MCP-based tools.*

---

### ⬜ Agent discovery and shadow-agent scanner

**What it means:** Real organizations and personal environments will have agents that were never manually registered. AgentVoir should discover unknown agents, LLM callers, MCP servers, tool configs, and AI workflows from repos, Kubernetes workloads, gateway traffic, package manifests, local configs, and browser extensions.

**TODO items:**

- [ ] Add `AgentDiscoverySource` entity: GitHub repo, Kubernetes cluster, Docker container, browser extension, local file path, cloud logs, proxy logs, mobile device — *Places AgentVoir can look for agents nobody registered yet.*
- [ ] Build GitHub repository scanner for common frameworks: LangChain, LangGraph, CrewAI, AutoGen, LlamaIndex, Vercel AI SDK, OpenAI Agents SDK — *Find AI agent code in repos by detecting popular frameworks.*
- [ ] Detect MCP server configs and tool manifests from repos and local filesystem — *Spot tool definitions sitting in config files on disk or in Git.*
- [ ] Detect direct provider SDK usage: OpenAI, Anthropic, Gemini, Bedrock, OpenRouter, Azure OpenAI — *Find apps calling AI vendors without going through AgentVoir.*
- [ ] Detect missing `x-agent-id` gateway traffic and create "unregistered caller" candidates — *Flag anonymous traffic that should be tied to a registered agent.*
- [ ] Add Kubernetes scanner for deployments with LLM provider API keys or agent framework packages — *Find agent workloads running in your clusters.*
- [ ] Add Docker Compose scanner for local/personal mode — *Find agents in local developer stacks.*
- [ ] Add browser extension scanner for AI extensions in AgentVoir Personal — *Detect AI browser extensions on a user’s machine.*
- [ ] Add confidence score for discovered agent candidates — *Rank findings so reviewers tackle the most likely agents first.*
- [ ] Add UI workflow: review discovered agents → approve registration → assign owner/risk tier — *Human review queue before shadow agents become governed.*
- [ ] Add policy: block or warn on production LLM calls from unregistered agents — *Stop or alert on “invisible” agents in production.*
- [ ] Add scheduled discovery job and discovery history — *Re-scan periodically and keep a log of what was found when.*

---

### 🟡 OpenTelemetry traces and Prometheus metrics

**What it means:** Deep visibility into how AgentVoir behaves in production — how long each request takes, where time is spent (cache, policy check, provider call), and aggregate metrics for dashboards and alerts. OpenTelemetry (OTel) is the industry standard for distributed tracing; Prometheus collects numeric metrics.

**TODO items:**

- [x] Add OTel Collector and Prometheus to dev Docker Compose stack — *Ship tracing and metrics infrastructure with the local dev environment.*
- [x] Add placeholder Prometheus scrape config for gateway and registry-api — *Wire Prometheus to pull metrics from core services.*
- [x] Add placeholder Grafana dashboard JSON — *Starter dashboard file teams can import.*
- [ ] Instrument gateway with OTel traces (cache, provider, total latency spans) — *Show where time is spent inside each gateway request.*
- [ ] Instrument registry-api with OTel traces — *Trace registry operations for debugging slow catalog queries.*
- [ ] Expose `/metrics` endpoint on gateway (request count, latency histogram, cache hits) — *Publish numeric stats Prometheus can scrape from the gateway.*
- [ ] Expose `/metrics` endpoint on registry-api — *Same for the registry service.*
- [ ] Populate Grafana dashboards (latency, error rate, cache hit rate, cost) — *Fill in charts operators actually use day to day.*
- [ ] Propagate `trace_id` from gateway through to usage events — *Link billing receipts back to the trace for one-click debugging.*
- [ ] Document local observability setup for developers — *Guide for running traces and metrics on a laptop.*
- [ ] Add alerting rules (error rate > 5%, p99 latency > 10s) — *Automatic pages when the system misbehaves.*

---

### ⬜ Pre-flight token and cost estimation

**What it means:** AgentVoir should estimate token usage and maximum cost before calling a provider. This allows policy checks, budget enforcement, warnings, and routing decisions before money is spent.

**TODO items:**

- [ ] Add tokenizer abstraction per provider/model family — *Count tokens the way each AI vendor counts them, before the call.*
- [ ] Estimate input tokens before provider call — *Predict bill size from the messages about to be sent.*
- [ ] Estimate maximum possible output cost from `max_tokens` — *Worst-case cost if the model uses the full output allowance.*
- [ ] Compare estimated cost against per-request budget — *Block or warn before spending if the estimate exceeds limits.*
- [ ] Add `x-estimated-cost-usd` debug/response header when enabled — *Optional header showing predicted cost to developers.*
- [ ] Add dry-run endpoint: `POST /v1/chat/completions:estimate` — *“How much would this cost?” without calling the model.*
- [ ] Add policy rule: block request if estimated cost exceeds threshold — *Governance rule based on predicted spend, not just actual spend.*
- [ ] Add tests for tokenizer drift and unknown-model fallback behavior — *Stay safe when token counting or pricing data is incomplete.*
- [ ] Document model pricing table update workflow — *How finance/platform teams keep price tables current.*

---

### ⬜ Human-in-the-loop approval gates

**What it means:** Some actions should pause until an authorized human approves them — for example high-risk tool calls, production agent promotion, expensive requests, or sensitive data export.

**TODO items:**

- [ ] Define approval request model — *Data shape for “something needs a human yes/no.”*
- [ ] Add approval policy: require approval for high-risk tools — *Pause before dangerous tool calls until someone approves.*
- [ ] Add approval policy: require approval for production lifecycle promotion — *Require sign-off before an agent goes live.*
- [ ] Add approval policy: require approval when estimated cost exceeds threshold — *Require sign-off on unusually expensive requests.*
- [ ] Add approval policy: require approval for data export tools — *Require sign-off before data leaves the organization.*
- [ ] Implement `POST /v1/approvals` — *API to create an approval request.*
- [ ] Implement approve/reject endpoints — *API for managers to approve or deny pending actions.*
- [ ] Add approval audit log events — *Permanent record of who approved what and when.*
- [ ] Add Slack, email, or webhook notification integration — *Notify approvers where they already work.*
- [ ] Add web console approval queue — *Inbox UI for pending approvals.*
- [ ] Add timeout/expiry behavior — *Auto-deny or escalate if nobody responds in time.*

---

### ⬜ Prompt injection and tool-call security

**What it means:** Untrusted text can try to override system instructions, exfiltrate secrets, or trick an agent into unsafe tool calls. AgentVoir should make trusted/untrusted boundaries explicit and enforce tool-call safety policies.

**TODO items:**

- [ ] Mark input sources as trusted vs untrusted — *Treat user email differently from your own system instructions.*
- [ ] Add prompt-injection detector hook before tool execution — *Scan for tricks that try to override safety rules.*
- [ ] Add policy rule: untrusted content cannot authorize tool calls — *User-supplied text alone cannot trigger dangerous actions.*
- [ ] Add tool-call confirmation policy for high-risk tools — *Extra confirmation step before sensitive tools run.*
- [ ] Add allowlist/denylist for tool names and arguments — *Explicit lists of permitted tools and parameters.*
- [ ] Add argument schema validation before tool execution — *Reject tool calls with unexpected or oversized arguments.*
- [ ] Add secret redaction before model calls — *Strip API keys and passwords from text sent to the model.*
- [ ] Add response filtering for system prompt leakage — *Block models from revealing hidden instructions.*
- [ ] Add attack simulation test cases — *Automated tests mimicking known jailbreak and injection patterns.*
- [ ] Add docs: "Prompt injection threat model" — *Explain attack types and how AgentVoir mitigates them.*

---

### 🟡 Admin web console

**What it means:** A browser UI makes AgentVoir easier to demo and operate. Platform teams should be able to inspect agents, dependencies, cost, cache behavior, policies, eval results, and approvals without querying raw APIs.

**TODO items:**

- [x] Agent list and detail pages — *Browse all registered agents and open a detail view.*
- [ ] Agent registration form — *Create agents in the UI without calling the API manually.*
- [ ] Manifest upload and validation UI — *Upload a YAML file and see validation errors inline.*
- [ ] Dependency graph visualization — *Interactive map of what each agent depends on.*
- [x] Cost and token usage dashboard — *Charts of spend and token volume over time.*
- [x] Cache hit/miss dashboard — *See how often cached answers are reused.*
- [ ] Policy decision viewer — *Inspect why requests were allowed or denied.*
- [ ] Audit event explorer — *Search the audit trail from the browser.*
- [ ] Prompt version viewer and diff page — *Compare prompt versions side by side.*
- [ ] Eval results comparison page — *See quality test scores before and after a change.*
- [ ] Approval queue — *Approve or reject pending human-in-the-loop requests.*
- [ ] Provider health page — *Status of each AI vendor (latency, errors, fallback usage).*

---

### ⬜ Enhanced agent metadata (governed runtime asset)

**What it means:** Today the registry captures basics (owner, lifecycle, risk, policies, budget, cache, dependencies, model route). Strategy docs ([meta-data.md](meta-data.md), [future-of-agents.md](future-of-agents.md)) describe modeling each agent as a **governed runtime asset** — who owns it, what it can touch, what version runs, how to disable it. This section tracks metadata gaps vs. the current YAML manifest and PostgreSQL schema.

**Already captured (extend, don't duplicate):** `owner_team`, `cost_center`, `environment`, `framework`, `risk_level`, `lifecycle`, `data_classes`, `policies`, `budget`, `cache`, `dependencies`, `model routes`.

**TODO items — ownership and accountability:**

- [ ] Add `technical_owner`, `business_owner`, `security_reviewer`, `compliance_reviewer` fields — *Name the engineers, product owners, and reviewers accountable for each agent.*
- [ ] Add `oncall_contact`, `support_channel`, `escalation_policy` references — *Who to page when the agent breaks at 2 a.m.*
- [ ] Expose ownership block in manifest YAML and admin console — *Make ownership visible in config files and the UI.*
- [ ] Migration: `000005_agent_ownership.up.sql` — *Database change to store the new ownership fields.*

**TODO items — lifecycle and approval:**

- [ ] Add `approval_status`, `last_reviewed_at`, `next_review_due_at`, `retirement_date` — *Track review cadence and planned shutdown dates.*
- [ ] Add `change_ticket` and `release_notes` on agent version records — *Link each version to your change-management ticket and release notes.*
- [ ] API: lifecycle promotion requires approval metadata when target is `production` — *Cannot go live without documented approval.*
- [ ] Align with [agents-sunsets.md](agents-sunsets.md) degradation states (`cost_saving`, `degraded`, `read_only`, `suspended`) — *Standard states for winding down or limiting an agent safely.*

**TODO items — versioning (agent, prompt, policy, eval):**

- [ ] Add `AgentVersion` entity: `git_sha`, `container_image`, `prompt_version`, `policy_version`, `eval_suite_version` — *Pin exactly which code, prompt, and rules ran for each release.*
- [ ] Link prompt registry versions to agent versions — *Know which prompt text shipped with which agent version.*
- [ ] API: `GET /v1/agents/{id}/versions` with full version manifest — *List every deployed version and its full config snapshot.*

**TODO items — runtime and deployment:**

- [ ] Add `runtime` block: `hosting_platform`, `region`, `replicas`, `timeout_seconds`, `retry_policy`, `concurrency_limit` — *Where and how the agent runs (cloud, region, scale, timeouts).*
- [ ] Add `state_backend`, `queue_backend` references (secret refs only) — *Point to where state and job queues live without storing credentials.*
- [ ] Support runtime metadata in manifest import — *Load deployment facts from the YAML file.*

**TODO items — model policy (beyond model route):**

- [ ] Add `model_policy`: `allowed_models`, `forbidden_models`, `selection_strategy`, `max_context_tokens`, `max_output_tokens` — *Fine-grained rules on which models an agent may use and how large requests can be.*
- [ ] Add `requires_private_model`, `provider_region_constraint` — *Require on-prem or in-region models when data cannot leave a boundary.*
- [ ] Track non-LLM model roles per [non-llm-models.md](non-llm-models.md): embedding, reranker, classifier, speech, vision — *Register helper models (search embeddings, speech, vision), not just chat models.*
- [ ] Cost metric: `cost_per_successful_task` not tokens alone — *Measure cost per completed job, not just raw token count.*

**TODO items — tools, MCP, and side effects:**

- [ ] Extend dependency model with `side_effect_level`, `allowed_actions`, `forbidden_actions` — *Describe how dangerous each dependency is and what it may or may not do.*
- [ ] Add `requires_human_approval`, `approval_policy`, `max_call_count_per_run` per tool — *Per-tool safety: approvals required and call limits per request.*
- [ ] First-class `Tool` and `MCPServer` registry entities (see Tool/MCP registry section) — *Tools become full catalog objects, not loose strings.*
- [ ] Typed dependency graph nodes: model, prompt, policy, eval_suite, secret, cache, workflow — *Rich dependency map with labeled node types.*

**TODO items — data access and residency:**

- [ ] Add `data_access` block: typed sources, read/write scope, `retention_policy`, export allow/deny lists — *Document which databases and files the agent may read, write, or export.*
- [ ] Add `provider_risk` per [china-and-robots.md](china-and-robots.md): `provider_country`, `data_leaves_boundary`, `allowed_regions` — *Record cross-border data flow and residency constraints.*
- [ ] Query: agents touching a given data classification or external system — *“Show every agent that can access customer PII or Salesforce.”*

**TODO items — risk, HITL, and runtime controls:**

- [ ] Expand `risk` block: `impact_area`, capability flags (`can_spend_money`, `can_send_messages`, `can_modify_systems`) — *Explicit flags for the scariest things an agent might do.*
- [ ] Add `human_in_loop` block: triggers, approver roles, timeout, fallback on timeout — *When and how a human must approve before the agent continues.*
- [ ] Add `runtime_controls`: kill switch, `quarantine_mode`, `disabled_reason`, `disabled_by`, `disabled_at` — *Emergency stop metadata: who disabled the agent and why.*
- [ ] Gateway enforces quarantine modes (`read_only`, `no_external_tools`, `disabled`) — *Gateway respects “read only” or “no tools” without redeploying the agent.*

**TODO items — evals, observability, and quality:**

- [ ] Add `evals` block: `suite_id`, `last_score`, `required_score`, `promotion_gate_required`, `canary_percent` — *Quality test requirements before promoting or rolling out changes.*
- [ ] Add `observability` block: `otel_service_name`, SLO latency/success, `dashboard_url`, `alert_policy` — *Link each agent to its dashboards, SLOs (service level objectives), and alerts.*
- [ ] Add `quality_profile` per [agent-quality-review.md](agent-quality-review.md): multi-dimensional scores, trend, gate events — *Rolling quality grades from user feedback and automated review.*
- [ ] Wire negative user feedback → eval candidate pipeline — *Bad user ratings automatically become new regression tests.*

**TODO items — multilingual and voice (enterprise):**

- [ ] Add `language_profile` per [multilingual-agents.md](multilingual-agents.md): supported locales, quality-by-language, routing rules — *Which languages the agent supports and how well it performs in each.*
- [ ] Add `voice_pipeline` / `operational_agent_profile` per [voice-agents.md](voice-agents.md) for incident/responder subtypes — *Extra metadata for phone/voice and on-call incident agents.*
- [ ] Add `agent_subtype` field: `chat`, `workflow`, `copilot`, `voice_responder`, `digital_worker`, `embodied` — *Classify agents beyond generic “chat bot.”*

**TODO items — external systems registry:**

- [ ] New entity: `ExternalSystem` (API, DB, model provider, SaaS) with owner, auth type, secret ref, rate limits, SLA — *Catalog of every external system agents connect to.*
- [ ] Link agents to external systems with allowed/blocked lists — *Explicit allow/deny lists per agent for each external system.*
- [ ] Model performance rollups per [model-performance.md](model-performance.md) as dependency health nodes — *Health scores for models and dependencies on the graph.*

**TODO items — schema and docs:**

- [ ] Publish `docs/schemas/agent-metadata-v2.yaml` reference manifest — *Official example of the full metadata schema.*
- [ ] ADR: governed runtime asset metadata model — *Architecture decision record explaining the metadata design.*
- [ ] Admin console: tabbed agent detail (Ownership, Risk, Tools, Data, Evals, Controls) — *Organized UI tabs so operators find the right metadata quickly.*

---

## Phase 3: Semantic cache and evals

**Goal:** Smarter caching (similar questions get similar answers), systematic quality testing for agents, and safety hooks for sensitive data.

---

### ⬜ Cache correctness and safety framework

**What it means:** Caching must be safe before it is clever. AgentVoir should prove that cached responses cannot leak data across tenants, agents, prompts, policies, users, tools, or model versions.

**TODO items:**

- [ ] Define canonical cache key contract
- [ ] Include tenant, agent, model, prompt version, tools, response format, policy version, and context hash in cache key
- [ ] Add cache-key golden tests
- [ ] Add tenant-isolation tests
- [ ] Add policy-version invalidation tests
- [ ] Add prompt-version invalidation tests
- [ ] Add RAG-context invalidation tests
- [ ] Add cache poisoning tests
- [ ] Add cache replay tests
- [ ] Add request-level and agent-level `never_cache` policy
- [ ] Add cache explain endpoint: `GET /v1/cache/explain?trace_id=...`
- [ ] Document cache safety model and invalidation rules

---

### ⬜ RedisVL / Qdrant semantic cache

**What it means:** Unlike exact cache (identical request only), semantic cache recognizes when two questions mean the same thing even if worded differently — "What's the refund policy?" vs "How do I get my money back?" — and returns a cached answer. Uses vector embeddings stored in RedisVL or Qdrant.

**TODO items:**

- [ ] Choose vector store (RedisVL vs Qdrant) and add to Docker Compose
- [ ] Embed incoming request (model + messages) into vector on cache write
- [ ] Query vector store for similar prior requests above similarity threshold
- [ ] Enforce OPA policy: semantic cache only when agent allows it and no PII present
- [ ] Implement `semantic_safe` and `semantic_aggressive` cache modes in gateway
- [ ] Record `semantic-hit` vs `exact-hit` in usage events and response headers
- [ ] Add cache entry metadata (embedding model, similarity score)
- [ ] TTL and eviction policy for semantic cache entries
- [ ] Benchmark hit rate and latency vs exact cache
- [ ] Document when semantic cache is safe vs unsafe for enterprise data

---

### ⬜ Cache shadow mode

**What it means:** Test whether cached answers are still good without actually serving them to users. AgentVoir returns the live model answer but quietly compares it to what the cache would have returned — useful before turning cache on in production.

**TODO items:**

- [ ] Implement `shadow` cache mode in gateway config
- [ ] On cache hit in shadow mode: still call live provider, return live answer
- [ ] Compare cached vs live response (exact match, semantic similarity, token diff)
- [ ] Emit shadow comparison metrics (match rate, divergence score)
- [ ] Store shadow comparison samples for offline review
- [ ] Dashboard panel for shadow mode hit rate and divergence
- [ ] Document recommended shadow mode rollout playbook

---

### 🟡 Prompt registry

**What it means:** A version-controlled library of prompts used by agents — system prompts, templates, few-shot examples. Teams can track changes, roll back bad prompt updates, and tie prompt versions to agent versions for reproducibility.

**TODO items:**

- [x] Define prompt model (ID, agent, version, content, metadata)
- [x] Implement basic prompt CRUD registry API
- [x] Persist prompts in PostgreSQL
- [ ] Prompt versioning with immutable history (v1, v2, v3 — no overwrite)
- [ ] Link prompts to agent versions in manifest
- [ ] Gateway resolves prompt by agent + version at request time
- [ ] Prompt diff API (`GET /v1/prompts/{id}/diff?from=v1&to=v2`)
- [ ] Prompt approval workflow (draft → approved → production)
- [ ] Import/export prompts from Git repository
- [ ] Web console prompt editor with preview

---

### ⬜ Eval datasets and regression runner

**What it means:** Automated quality tests for agents — run a fixed set of example questions through an agent and check that answers still meet expectations after you change a prompt, model, or policy. Catches regressions before they reach users.

**TODO items:**

- [ ] Define eval dataset format (input, expected output or rubric, tags)
- [ ] Implement dataset storage (PostgreSQL + file import)
- [ ] Build evaluator service job: run agent against dataset via gateway
- [ ] Support eval metrics: exact match, LLM-as-judge, custom scorers
- [ ] Store eval run results with timestamps and config snapshot
- [ ] CLI: `agentvoir eval run --agent customer-support --dataset support-v1`
- [ ] Compare eval runs side-by-side (before/after prompt change)
- [ ] Fail CI pipeline if eval score drops below threshold
- [ ] Document eval dataset authoring guide

---

### ⬜ Data lineage, evidence, and provenance

**What it means:** Quality scores and evals say whether an agent performed well, but regulated and high-risk agents also need to prove where an answer came from. AgentVoir should capture the chain of influence: prompt version, model version, retrieved documents, tool responses, policy decisions, human approvals, and final output.

**TODO items:**

- [ ] Define `AgentOutputProvenance` schema
- [ ] Capture prompt version, model/provider version, policy version, eval suite version, and tool schema version per run
- [ ] Capture RAG document IDs, chunk IDs, embedding model, retrieval score, reranker score, and corpus version
- [ ] Capture external API response hashes instead of storing sensitive full payloads by default
- [ ] Add evidence bundle export: `GET /v1/runs/{runID}/evidence-bundle`
- [ ] Add output citation contract for agents that must provide source-backed answers
- [ ] Add policy: high-risk agents must attach evidence bundle before final response
- [ ] Add UI evidence timeline: request → retrieval → tool calls → policy decisions → output
- [ ] Add freshness metadata for knowledge sources and retrieved documents
- [ ] Add provenance redaction rules for PII, secrets, and confidential documents
- [ ] Add tamper-evident hash chain for critical run provenance records

---

### ⬜ Agent scorecards

**What it means:** A report card for each agent summarizing health — cost trend, error rate, cache hit rate, eval scores, policy violations, and budget utilization. Helps managers and owners see which agents are healthy and which need attention.

**TODO items:**

- [ ] Define scorecard schema (agent, period, KPIs, grade, recommendations)
- [ ] Aggregate KPIs from ClickHouse (cost, latency, error rate, cache hit rate)
- [ ] Pull latest eval scores from evaluator service
- [ ] Count policy denials from audit/usage logs
- [ ] Implement `GET /v1/agents/{agentID}/scorecard?period=30d`
- [ ] Render scorecard in web console
- [ ] Optional: email/Slack weekly scorecard digest to agent owners
- [ ] Benchmark and trend comparison vs previous period

---

### ⬜ Red-team and adversarial test harness

**What it means:** Prompt injection, jailbreaks, tool escalation, cache poisoning, and data exfiltration should be tested continuously, not only handled by runtime hooks. AgentVoir should provide repeatable security test packs and make them part of promotion gates.

**TODO items:**

- [ ] Define red-team scenario format: attack prompt, untrusted source, expected policy behavior, expected refusal/action
- [ ] Add built-in attack packs: prompt injection, secret exfiltration, tool escalation, cache poisoning, jailbreak, data export abuse
- [ ] Run red-team suites through gateway and policy engine
- [ ] Store red-team run results alongside eval runs
- [ ] Gate production promotion on passing required red-team pack
- [ ] Add regression cases automatically from real policy violations
- [ ] Dashboard: security pass rate, top failing attack categories
- [ ] CLI: `agentvoir redteam run --agent <id> --pack prompt-injection-basic`

---

### ⬜ PII / secret detection hooks

**What it means:** Automatically detect when requests or responses contain personally identifiable information (names, emails, SSNs) or secrets (API keys, passwords) — and block caching, redact content, or deny the request based on policy.

**TODO items:**

- [ ] Integrate PII detection library or service (regex + ML-based)
- [ ] Integrate secret detection (e.g. trufflehog patterns)
- [ ] Run detection on gateway request before cache lookup and provider call
- [ ] Run detection on provider response before cache write
- [ ] Set `contains_pii` / `contains_secret` flags for OPA policy input
- [ ] Wire OPA policies to deny or bypass cache when PII detected
- [ ] Optional: redact PII in audit logs while keeping structure
- [ ] Build `pii-redactor` service plugin for pluggable detectors
- [ ] Add false-positive tuning configuration per tenant
- [ ] Document compliance implications and data handling

---

### ⬜ Agent memory and knowledge-base governance

**What it means:** Agent memory can become sensitive, stale, or legally restricted. RAG corpora can contain outdated, private, or customer data. AgentVoir should treat memory stores and knowledge bases as governed assets with ownership, retention, freshness, deletion, portability, and access controls.

**TODO items:**

- [ ] Add `MemoryStore` entity: owner, backend, data classes, retention, deletion policy, embedding model
- [ ] Add `KnowledgeBase` entity: source systems, refresh schedule, corpus version, last indexed time, data classification
- [ ] Link agents to memory stores and knowledge bases in dependency graph
- [ ] Add policy: high-risk agents cannot use unreviewed memory stores
- [ ] Add memory deletion API for personal mode and privacy requests
- [ ] Add stale knowledge alert when corpus has not refreshed within SLA
- [ ] Add RAG ingestion audit: document added/removed, source, timestamp, actor
- [ ] Add memory export/import with redaction for backup and portability
- [ ] Add admin UI tab for Memory and Knowledge Sources

---

## Phase 4: Kubernetes-native control plane

**Goal:** Run AgentVoir the way large enterprises run production software — on Kubernetes, with declarative config, GitOps, and multi-region patterns.

---

### 🟡 Helm chart

**What it means:** A packaged, configurable installer for Kubernetes — one `helm install` deploys AgentVoir with sensible defaults, and operators tune settings via a values file instead of hand-editing dozens of YAML manifests.

**TODO items:**

- [x] Create initial Helm chart skeleton (`deployments/helm/agentvoir/`)
- [x] Add gateway and registry-api Deployment + Service templates
- [x] Add values.yaml with image tags, URLs, and resource limits placeholders
- [ ] Add token-accounting, worker, and evaluator deployments to chart
- [ ] Add subcharts or templates for Postgres, Redis, ClickHouse (or document external deps)
- [ ] Add Ingress templates with TLS support
- [ ] Add ConfigMaps and Secrets for environment configuration
- [ ] Add HorizontalPodAutoscaler templates for gateway
- [ ] Helm chart CI: lint, template render, kubeconform validation
- [ ] Publish chart to OCI registry or Helm repo
- [ ] Document production values examples (HA, external DB)

---

### 🟡 Kubernetes CRDs: Agent, Prompt, ModelRoute, AgentPolicy

**What it means:** Define agents and related config as native Kubernetes resources — `kubectl apply -f agent.yaml` registers an agent, and Kubernetes itself tracks desired state. Enables GitOps: config lives in Git, cluster reconciles automatically.

**TODO items:**

- [x] Draft initial Agent CRD schema (`infra/kubernetes/crds/agent.agentvoir.dev.yaml`)
- [ ] Finalize CRD schemas for Prompt, ModelRoute, and AgentPolicy
- [ ] Generate OpenAPI validation schema for each CRD
- [ ] Build Kubernetes controller/operator to reconcile CR → registry API
- [ ] Status subresource on CRDs (synced, error, last reconciled)
- [ ] Watch CR changes and update registry in real time
- [ ] Delete CR → retire agent in registry (soft delete)
- [ ] Document CRD field reference for platform teams
- [ ] Add example CR manifests under `examples/`

---

### ⬜ Admission controller

**What it means:** A gatekeeper that runs before AgentVoir resources are saved to Kubernetes — rejects invalid or unsafe configs (e.g. production agent without owner, PII agent with semantic cache enabled) before they reach the cluster.

**TODO items:**

- [ ] Build validating admission webhook service
- [ ] Register webhook in Kubernetes (`ValidatingWebhookConfiguration`)
- [ ] Validate Agent CR fields (required owner, valid lifecycle, budget limits)
- [ ] Validate AgentPolicy CR against OPA policy syntax
- [ ] Reject CRs that violate enterprise policy (e.g. no PII + semantic cache)
- [ ] Return clear rejection messages to `kubectl` users
- [ ] TLS cert management for webhook (cert-manager integration)
- [ ] Integration tests with envtest or kind cluster
- [ ] Document admission rules and bypass annotations for emergencies

---

### ⬜ GitOps examples

**What it means:** Show teams how to manage AgentVoir config the enterprise way — agent definitions stored in Git, automatically deployed to Kubernetes by tools like Argo CD or Flux when someone merges a pull request.

**TODO items:**

- [ ] Create example Git repo layout (agents/, prompts/, policies/ directories)
- [ ] Add Argo CD Application manifest pointing at example repo
- [ ] Add Flux Kustomization example
- [ ] Document PR-based workflow: propose agent change → review → merge → auto-deploy
- [ ] Add CI check: validate manifests and CRDs on every PR
- [ ] Example: promote agent from staging to production via Git branch merge
- [ ] Document rollback procedure (Git revert → auto-sync)
- [ ] Optional: integrate with GitHub Actions for eval-on-PR

---

### ⬜ Multi-region routing examples

**What it means:** Patterns for running AgentVoir across multiple geographic regions — route users to the nearest gateway, fail over if a region goes down, and keep usage analytics consistent across regions.

**TODO items:**

- [ ] Document multi-region reference architecture (active-active vs active-passive)
- [ ] Example: global load balancer → regional gateway deployments
- [ ] Example: regional gateways with shared registry (single Postgres) vs regional registry replicas
- [ ] ClickHouse replication or centralized analytics aggregation pattern
- [ ] Cross-region failover for provider routing
- [ ] Data residency considerations (EU agents stay in EU region)
- [ ] Example Terraform modules for two-region deployment
- [ ] Runbook: regional outage detection and traffic shift
- [ ] Load test cross-region latency and failover time

---

### ⬜ Backup, restore, and disaster recovery

**What it means:** AgentVoir becomes a control plane. Losing registry data, policies, audit history, usage records, or approval state could break operations and compliance. AgentVoir needs tested export/import, backup, restore, and disaster-recovery flows for both enterprise and personal deployments.

**TODO items:**

- [ ] Add full registry export: agents, prompts, policies, dependencies, budgets, model routes, tool registry, external systems
- [ ] Add selective export: one agent and its dependency bundle
- [ ] Add encrypted local backup for AgentVoir Personal
- [ ] Add Postgres backup/restore guide and scripts
- [ ] Add ClickHouse usage/audit backup and retention strategy
- [ ] Add restore smoke test in CI using sample backup
- [ ] Add disaster recovery runbook: registry down, gateway degraded, provider outage, database restore
- [ ] Add `agentvoir backup create` and `agentvoir backup restore` CLI commands
- [ ] Add backup integrity verification using hashes/signatures
- [ ] Add admin UI backup status and last successful restore-test timestamp

---

## Phase 5: Ecosystem and integrations

**Goal:** Make AgentVoir useful inside real enterprise AI stacks by integrating with agent frameworks, CI/CD systems, observability tools, and data platforms.

---

### ⬜ Framework integrations

**What it means:** Teams should be able to adopt AgentVoir without rewriting every agent. Framework adapters let LangChain, LangGraph, LlamaIndex, CrewAI, AutoGen, and custom agents send traffic through the gateway and register metadata.

**TODO items:**

- [ ] LangChain callback/tracing integration
- [ ] LangGraph metadata and checkpoint integration
- [ ] LlamaIndex callback integration
- [ ] CrewAI integration example
- [ ] AutoGen integration example
- [ ] OpenAI Agents SDK integration example if useful
- [ ] Framework compatibility matrix in docs
- [ ] Example apps for each supported framework

---

### ⬜ Agent contract and interoperability validation

**What it means:** AgentVoir will integrate with many frameworks and may allow agents to call other agents. Each agent should have a machine-readable contract: input shape, output shape, error shape, side effects, idempotency, timeout, and failure modes.

**TODO items:**

- [ ] Add `AgentContract` entity: input schema, output schema, error schema, side effects, idempotency, timeout
- [ ] Add contract validation API for agent outputs
- [ ] Add contract conformance tests for registered agents
- [ ] Add compatibility mapping for OpenAPI, JSON Schema, MCP tool schemas, App Intents, and Android App Functions
- [ ] Add policy: agents with downstream dependents cannot change contract without review
- [ ] Add semantic versioning rules for agent contracts
- [ ] Add contract diff UI and breaking-change warning

---

### ⬜ CI/CD integrations

**What it means:** Agent definitions, prompts, policies, and evals should fit into normal engineering workflows. Pull requests should validate changes before they reach production.

**TODO items:**

- [ ] GitHub Action to validate agent manifests
- [ ] GitHub Action to run AgentVoir evals on PR
- [ ] GitHub Action to publish prompt/agent config
- [ ] Pre-commit hook for manifest validation
- [ ] CI check for OPA policy tests
- [ ] CI check for prompt registry diffs
- [ ] Example PR workflow for agent promotion
- [ ] Example release workflow for production agent config

---

### ⬜ Data platform and notification integrations

**What it means:** Enterprises often centralize usage, cost, audit, and operational events in existing tools. AgentVoir should export data cleanly instead of becoming another silo.

**TODO items:**

- [ ] Snowflake usage export
- [ ] Datadog metrics/export
- [ ] Splunk audit log export
- [ ] S3/GCS/Azure Blob artifact export
- [ ] Slack notifications for budget thresholds
- [ ] Slack notifications for approval requests
- [ ] Webhook integration for eval failures and policy denials
- [ ] CSV/Parquet export for finance and governance teams

---

### ⬜ AgentVoir CLI

**What it means:** A first-class CLI makes AgentVoir scriptable, demoable, and easy to use from developer terminals and CI systems.

**TODO items:**

- [ ] `agentvoir login`
- [ ] `agentvoir agents list`
- [ ] `agentvoir agents apply -f agent.yaml`
- [ ] `agentvoir gateway test`
- [ ] `agentvoir eval run`
- [ ] `agentvoir cache inspect`
- [ ] `agentvoir policy test`
- [ ] `agentvoir usage summarize`
- [ ] `agentvoir issues scout` for local AI-suggested GitHub issues
- [ ] `agentvoir issues code` for local AI coder workflow

---

## Phase 6: AgentVoir Home (Personal Mode)

**Goal:** A lighter deployment for individuals and families — track every personal agent (coded, installed, or sourced from marketplaces like OpenClaw) with plain-English permissions, privacy controls, and cost visibility. Same registry/gateway concepts; smaller packaging (Docker Compose, SQLite option, local dashboard). See [agent-voir-home.md](agent-voir-home.md).

---

### ⬜ Personal deployment profile

**What it means:** One user (or family) runs AgentVoir locally without enterprise infra — no Kubernetes, SSO, or multi-tenant RBAC required.

**TODO items:**

- [ ] Define `deployment_mode`: `enterprise` | `personal` in config
- [ ] Personal onebox profile: SQLite or single-user Postgres, simplified OPA rules
- [ ] `docker-compose.personal.yml` (minimal services: registry, gateway, local UI)
- [ ] Document Personal vs Enterprise comparison table in INSTALL.md
- [ ] Default personal budget and privacy-safe policies

---

### ⬜ Agent source and marketplace metadata

**What it means:** Personal users install agents like browser extensions — from OpenClaw Marketplace, GitHub, npm, Docker, or friends. Provenance and update trust matter.

**TODO items:**

- [ ] Add `source` block to manifest: `origin_type`, `platform`, `publisher`, `source_url`, `installed_at`, `update_channel`, `auto_update_enabled`, `integrity_hash`
- [ ] OpenClaw import profile: `skills`, `channels`, `model_providers`, `voice_enabled`, `browser_control_enabled`
- [ ] Marketplace trust fields: `publisher_verified`, `user_rating`, `install_count`, `known_vulnerabilities`
- [ ] UI: "Where did this agent come from?" card on agent detail
- [ ] Warn on permission changes after agent update

---

### ⬜ Marketplace and third-party agent security scanner

**What it means:** Personal and enterprise users may import agents from OpenClaw-like marketplaces, GitHub, Docker images, npm packages, Python packages, or vendors. AgentVoir should inspect third-party agents before granting trust.

**TODO items:**

- [ ] Add marketplace import scanner for OpenClaw-like platforms, GitHub repos, Docker images, npm packages, and Python packages
- [ ] Add permission diff analyzer: new version requests email-send, file-write, browser-control, purchase, or home-device access
- [ ] Add suspicious endpoint detection: unknown domains, paste sites, disposable webhooks, raw IPs
- [ ] Add dependency vulnerability scan for imported agent packages
- [ ] Add license scan for imported agent code and assets
- [ ] Add SBOM ingestion for third-party agent packages
- [ ] Add publisher trust score and verification status
- [ ] Add user warning: "This agent can send emails and access browser forms"
- [ ] Add quarantine mode for newly imported agents until reviewed
- [ ] Add auto-update policy: disabled by default for high-risk agents

---

### ⬜ Personal permissions (plain English)

**What it means:** The most important home feature — what each agent can read, send, spend, or control on devices and services.

**TODO items:**

- [ ] Add `permissions` block: email, calendar, files, browser, money, home_devices (read/send/spend booleans + caps)
- [ ] Personal risk labels: Safe / Needs review / Sensitive / Dangerous / Disabled
- [ ] Gateway policy checks against personal permission manifest
- [ ] UI renders permissions as plain English (not raw YAML)
- [ ] `requires_confirmation_for` list for booking, purchases, outbound email

---

### ⬜ Privacy and personal budget metadata

**What it means:** Personal users care about data leaving the device, retention, and surprise token bills.

**TODO items:**

- [ ] Add `privacy` block: `data_leaves_device`, `external_model_provider`, `retention_days`, `can_use_data_for_training`
- [ ] Extend budget with personal alerts (`alert_at_usd` tiers), `cheaper_model_fallback`, `local_model_preferred`
- [ ] Dashboard: "This agent sends email content to OpenAI" disclosure
- [ ] Monthly spend by agent and by model provider

---

### ⬜ Home automation and physical safety

**What it means:** Agents connected to lights, locks, cameras, or appliances need stricter caps than email readers.

**TODO items:**

- [ ] Add `home_device_access` and `physical_safety` blocks (locks, cameras, garage, appliances)
- [ ] Policy: deny `can_unlock_doors`, `can_disable_alarm` by default
- [ ] Personal kill switch: **Pause all agents** (global + per-agent)
- [ ] See also embodied metadata in Phase 8 for robots

---

### ⬜ Personal dashboard (simple cards)

**What it means:** Home UI is not enterprise governance — simple cards: name, source, access, cost, risk, pause button.

**TODO items:**

- [ ] Personal web UI mode (or simplified `apps/web` view)
- [ ] Agent card: source, permissions summary, cost this month, last run, risk tier
- [ ] One-click pause / disable / require-approval mode
- [ ] Activity feed: what agents did today (lightweight audit)

---

### ⬜ Browser extension and desktop agent monitor

**What it means:** Before agents are deeply integrated into mobile operating systems, many personal agents will operate through browsers, browser extensions, local desktop apps, and web automation. AgentVoir should provide a companion that shows active agents, captures approvals, and monitors risky browser actions.

**TODO items:**

- [ ] Browser extension MVP for Chrome/Edge/Firefox
- [ ] Detect AI agents/extensions installed in browser where possible
- [ ] Add approval prompt for risky browser actions: form submit, purchase, file upload, password field interaction
- [ ] Add activity capture: website domain, action category, agent ID, approval status
- [ ] Add policy: block agents from entering passwords or payment details without explicit approval
- [ ] Add local desktop tray app for pause-all/approval inbox
- [ ] Pair browser extension with AgentVoir Personal server
- [ ] Show browser-agent activity in personal dashboard timeline

---

## Phase 7: AgentVoir Mobile

**Goal:** Mobile companion app — App Store–style agent inventory, permission manager, AI firewall, cost monitor, activity timeline, approval inbox, and kill switch for agents running on phones. See [mobile-version.md](mobile-version.md).

---

### ⬜ Mobile agent inventory

**What it means:** Track which agents are installed on which devices, from which marketplace, and whether the publisher is verified.

**TODO items:**

- [ ] Add `mobile_profile` top-level manifest section
- [ ] Fields: `installed_on[]`, `display_name`, `publisher`, `verified_publisher`, `status`, `version`
- [ ] Sync protocol: mobile app ↔ AgentVoir Home / Cloud / desktop registry
- [ ] API: register mobile-installed agent + device fingerprint

---

### ⬜ Mobile permissions and app integrations

**What it means:** Beyond OS app permissions — agent *action* permissions (contacts, calendar, SMS, purchases, cross-app actions).

**TODO items:**

- [ ] Add `mobile_permissions` block: contacts, calendar, email, photos, location, wallet, SMS, phone
- [ ] iOS App Intents allowlist: `ios_app_intents.allowed_intents[]`
- [ ] Android App Functions allowlist: `android_app_functions.allowed_functions[]`
- [ ] UI: permission diff when agent updates ("now requests email send")

---

### ⬜ Mobile activity timeline and approvals

**What it means:** Screen-time-style report of what agents did — apps used, data accessed, model called, cost, approval required.

**TODO items:**

- [ ] Extend usage events with `activity_event` shape: apps_used, data_accessed, user_approval_required
- [ ] Mobile API: `GET /v1/agents/{id}/activity?device_id=`
- [ ] Approval inbox: pending external actions (send email, book travel, purchase)
- [ ] Push notification on background agent action (configurable)

---

### ⬜ Mobile runtime controls

**What it means:** Emergency controls users expect on a phone — pause all agents, privacy mode, background limits.

**TODO items:**

- [ ] Add `mobile_runtime_controls`: `allow_background_execution`, `require_approval_for_external_actions`, `emergency_pause_enabled`, `current_mode`
- [ ] Add `background_behavior`: allowed windows, max runs/day, allowed/forbidden triggers, notify on background action
- [ ] Add `inference_mode`: on-device vs cloud, `data_leaves_device`, `private_cloud_supported`
- [ ] Global **Emergency privacy mode** toggles all agents to approval-required

---

### ⬜ AgentVoir Mobile app (MVP)

**What it means:** Native or cross-platform client paired with home server or cloud sync.

**TODO items:**

- [ ] Mobile app scaffold (React Native or Flutter — TBD)
- [ ] Screens: inventory, permissions, timeline, cost, approvals, kill switch
- [ ] Pairing flow with AgentVoir Home server (QR / local network)
- [ ] Offline read-only inventory cache
- [ ] Biometric lock for sensitive controls

---

## Phase 8: AI asset intelligence and extended types

**Goal:** Org-wide intelligence, quality loops, analytics, and asset types beyond chat agents — informed by [data-analytics.md](data-analytics.md), [agent-quality-review.md](agent-quality-review.md), [voice-agents.md](voice-agents.md), [agents-sunsets.md](agents-sunsets.md), and [china-and-robots.md](china-and-robots.md). Builds on Phase 2 metadata and Phase 6/7 profiles.

---

### ⬜ Managed AI asset types

**What it means:** Registry holds not only chat agents but workflows, copilots, MCP servers, tools, eval suites, and embodied agents as first-class assets.

**TODO items:**

- [ ] Add `asset_type` enum: agent, workflow, copilot, assistant, bot, mcp_server, tool, model, prompt, eval_suite, digital_worker, embodied_agent
- [ ] Shared asset base schema (owner, risk, lifecycle, kill switch, cost, quality score)
- [ ] Subtype-specific profile extensions (voice responder, robot, sales copilot)
- [ ] Dependency graph queries across asset types

---

### ⬜ AgentVoir Insights (org intelligence)

**What it means:** Conversation and usage analytics for teams — cost, training gaps, SME gaps, model load — without individual employee surveillance by default.

**TODO items:**

- [ ] Conversation event collector with topic tags, task_type, department (aggregated)
- [ ] Dashboards: department usage, training gap, model peak load, cost optimization, ROI
- [ ] Privacy defaults: team-level aggregation, configurable retention, opt-in raw capture
- [ ] Runaway-agent detection and deprecate/shutdown recommendations
- [ ] Demo agents from [data-analytics.md](data-analytics.md) (sales, marketing, compliance)

---

### ⬜ Quality feedback and reputation

**What it means:** Continuous quality scoring from user feedback, human review samples, and reviewer agents — gates for watch/quarantine/disable.

**TODO items:**

- [ ] Entities: `agent_feedback`, `agent_quality_scores`, `agent_review_jobs`, `quality_gate_events`
- [ ] Multi-dimensional scores: accuracy, grounding, safety, policy compliance, tool-use
- [ ] Quality gates: thresholds → mark_watch / quarantine / disable_external_tools
- [ ] Negative feedback → eval candidate → regression test (link Phase 3 evals)

---

### ⬜ Voice and operational agents

**What it means:** Incident responders and voice agents need escalation, runbook, war-room, and comms policies — not just chat completion metadata.

**TODO items:**

- [ ] `OperationalAgentProfile`: autonomy level, escalation_policy, runbook_access, war_room_behavior
- [ ] Entities: IncidentSession, VoiceCallTranscript, EscalationDecision, HumanHandoffRecord
- [ ] Metrics: MTTA, MTTR, escalation accuracy, unsafe action attempts
- [ ] Post-incident artifacts → eval and policy update pipeline

---

### ⬜ Consent, disclosure, and communication compliance

**What it means:** Voice, mobile, customer-support, healthcare, finance, and incident-response agents may communicate with real people. AgentVoir needs proof that AI identity was disclosed, recording/transcription consent was captured where required, and regulated communications followed approval policy.

**TODO items:**

- [ ] Add `ConsentRecord` entity: subject, channel, consent type, jurisdiction, timestamp, expiration, revocation status
- [ ] Add `AIDisclosurePolicy` entity: required wording, locale, channel, version
- [ ] Capture whether AI identity was disclosed at start of call/chat/email
- [ ] Capture recording/transcription consent for voice and meeting agents
- [ ] Add locale-specific disclosure templates for multilingual agents
- [ ] Add policy: voice agents cannot record/transcribe unless consent requirements are met
- [ ] Add policy: agents cannot contact external humans unless communication policy allows it
- [ ] Add customer-facing message approval workflow for regulated communications
- [ ] Add consent revocation handling: stop memory usage, delete eligible records, disable future outreach
- [ ] Add audit export for consent/disclosure history

---

### ⬜ Financial resilience and sunset

**What it means:** Graceful degradation when budgets or vendors fail; export packages for M&A or shutdown.

**TODO items:**

- [ ] `financial_resilience` and `budget_degradation_policy` metadata
- [ ] `continuity_plan`, `decommission_plan`, `liquidation_readiness` blocks
- [ ] Engine: auto-degrade agent at budget thresholds (cheaper model → read_only → suspended)
- [ ] Vendor payment failure impact report (which agents break)

---

### ⬜ Embodied and robot governance

**What it means:** Physical agents (warehouse, delivery, home robots) need movement permissions, safety zones, emergency stop, and firmware audit trails.

**TODO items:**

- [ ] Add `embodied_agent_profile` / `robot_governance` block
- [ ] Fields: robot_type, manufacturer, deployment_location, physical action permissions, safety zones, e-stop, firmware version
- [ ] Policy layer for physical actions (separate from LLM policy)
- [ ] Near-miss and human-override metrics

---

## How to use this roadmap

1. **Pick a phase** aligned with your deployment maturity (Phase 1 is the current baseline).
2. **Work top-to-bottom** within a phase — later items often depend on earlier ones.
3. **Turn large sections into small GitHub issues** with clear scope, acceptance criteria, and suggested files.
4. **Mark items in progress** in GitHub Issues or Projects; link PRs to specific TODO bullets.
5. **Use AgentVoir Scout** to suggest candidate issues, label them `ai-suggested`, and manually promote approved work with `ai-code`.
6. **Update this doc** when items are completed or scope changes.
7. **Read strategy docs** in the table at the top when scoping metadata or new product phases (Home, Mobile).

Recommended GitHub issue fields when converting a roadmap item:

- **Goal**
- **Scope**
- **Acceptance criteria**
- **Suggested files/modules**
- **Constraints**
- **Priority**
- **Labels**

For questions about what a technology does in AgentVoir, see [Tech Stack Usage](architecture/tech-stack-usage.md). For local setup, see [Docker Install Guide](../deployments/docker/INSTALL.md).