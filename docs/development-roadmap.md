# AgentVoir Development Roadmap (Detailed)

This document expands the high-level roadmap in [README.md](../README.md) into concrete work items. Each section includes:

- **What it means** — a plain-language explanation for readers who are not deeply familiar with the underlying technology.
- **TODO items** — specific, actionable tasks for implementers.

Status legend:

| Symbol | Meaning |
| ------ | ------- |
| ✅ | Done (Phase 1 baseline) |
| 🟡 | Partially done — foundation exists, more work needed |
| ⬜ | Not started |

For infrastructure component context (ClickHouse, OPA, Prometheus, Grafana), see [Tech Stack Usage](architecture/tech-stack-usage.md).

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
- [ ] Add pagination and sorting to list endpoint
- [ ] Add lifecycle transition validation (e.g. draft → production requires review)

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
- [ ] Return detailed validation errors (line numbers, field names)
- [ ] Support manifest import from URL or Git repository
- [ ] Add JSON Schema / CRD-compatible validation for manifests

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
- [ ] Implement streaming (`stream: true`) end-to-end
- [ ] Emit operational response headers (`x-cache-status`, `x-cost-usd`, etc.) on every response
- [ ] Load agent config from registry API at request time (not headers only)

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
- [ ] Load per-agent cache settings from registry (TTL, mode) instead of global config only
- [ ] Add cache bypass rules for sensitive or non-deterministic requests
- [ ] Expose cache hit-rate metrics

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
- [ ] Add down migrations tested in CI
- [ ] Add seed data script for demo agents
- [ ] Document schema ER diagram in docs

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
- [ ] Compute cost automatically from model pricing table (not caller-supplied)
- [ ] Add daily/monthly rollup jobs for budget dashboards
- [ ] Add retention policy and archival for old events

---

### ✅ Docker Compose environment

**What it means:** Anyone can start the full AgentVoir stack on their laptop with one command — no manual database installs, no port conflicts with existing Postgres, and no need to install Go or Node just to try the product.

**TODO items:**

- [x] Create developer Docker Compose stack (Postgres, Redis, ClickHouse, OPA, Prometheus, Grafana, OTel)
- [x] Create onebox stack (isolated, minimal host ports, self-contained)
- [x] Add Dockerfiles for gateway, registry-api, and token-accounting
- [x] Add healthchecks for infrastructure services
- [x] Add Makefile targets: `onebox-up`, `onebox-down`, `onebox-smoke`, `dev-up`, `dev-up-all`
- [x] Create `.env.onebox.example` for port and API key configuration
- [x] Write end-user install guide (`deployments/docker/INSTALL.md`)
- [x] Document onebox vs dev stack differences
- [ ] Publish pre-built images to a container registry (skip local `--build`)
- [ ] Add GitHub Actions workflow to build and push images on release
- [ ] Add docker-compose health wait script for smoother first-run UX

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
- [ ] Add usage/analytics client to both SDKs
- [ ] Publish to PyPI and npm
- [ ] Add retry, timeout, and error-handling best practices to docs
- [ ] Generate SDKs from OpenAPI spec (optional automation)

---

## Phase 2: Enterprise controls

**Goal:** Make AgentVoir safe and manageable for real enterprise deployments — proper login, permissions, spending limits, audit trails, and operational visibility.

---

### ⬜ OIDC authentication

**What it means:** Users log in with your company's existing identity system (Okta, Azure AD, Google Workspace, etc.) instead of shared static API keys. AgentVoir verifies "who is this person or service?" using industry-standard OpenID Connect (OIDC) tokens.

**TODO items:**

- [ ] Add OIDC provider configuration (issuer URL, client ID, client secret)
- [ ] Validate JWT access tokens on registry API requests
- [ ] Validate JWT access tokens on gateway requests (or accept OIDC + API key hybrid)
- [ ] Map OIDC claims (`sub`, `email`, `groups`) to AgentVoir user identity
- [ ] Support machine-to-machine (client credentials) flow for automated agents
- [ ] Document OIDC setup for common providers (Okta, Azure AD, Keycloak)
- [ ] Add integration tests with a local OIDC mock (e.g. Dex)
- [ ] Deprecate or gate static API keys behind admin-only bootstrap mode

---

### ⬜ RBAC and service accounts

**What it means:** Not everyone should be able to do everything. Role-Based Access Control (RBAC) defines who can register agents, change production settings, view costs, or administer policies. Service accounts give automated systems their own credentials with limited permissions.

**TODO items:**

- [ ] Define roles (e.g. `admin`, `agent-owner`, `viewer`, `auditor`)
- [ ] Define permissions (register agent, promote lifecycle, view usage, edit budgets)
- [ ] Store roles and role bindings in PostgreSQL
- [ ] Enforce permissions on registry API endpoints
- [ ] Create service account entity with scoped API tokens
- [ ] Implement `POST /v1/service-accounts` and token rotation
- [ ] Wire RBAC checks into gateway for sensitive operations
- [ ] Add audit log entry on permission denied events
- [ ] Document default roles and recommended enterprise mappings

---

### 🟡 Per-agent budgets

**What it means:** Set spending and usage caps per agent — for example, "this support bot can spend at most $1,000/month" or "no single request may exceed 12,000 input tokens." AgentVoir blocks or warns when limits are exceeded.

**TODO items:**

- [x] Define budget model (monthly USD, max tokens per request)
- [x] Implement `GET/PUT /v1/agents/{agentID}/budget` registry endpoints
- [x] Persist budgets in PostgreSQL
- [x] Accept budget fields from agent YAML manifest
- [ ] Gateway loads budget for agent on each request
- [ ] Enforce max tokens per request before calling provider
- [ ] Track cumulative spend per agent per month (from ClickHouse rollups)
- [ ] Return `429` or structured error when monthly budget exceeded
- [ ] Add budget utilization API (`GET /v1/agents/{agentID}/budget/status`)
- [ ] Optional: soft limits (warn) vs hard limits (block)
- [ ] Notify owners when budget reaches 80% / 100% thresholds

---

### ⬜ Per-agent and per-tenant rate limits

**What it means:** Prevent any single agent or customer (tenant) from flooding the gateway with too many requests per minute. Protects shared infrastructure and prevents runaway automation loops from causing outages or surprise bills.

**TODO items:**

- [ ] Add rate limit fields to budget/config model (requests per minute, tokens per minute)
- [ ] Implement token-bucket or sliding-window limiter in gateway (Redis-backed)
- [ ] Apply limits per agent ID
- [ ] Apply limits per tenant ID (`x-tenant-id` header)
- [ ] Return `429 Too Many Requests` with `Retry-After` header
- [ ] Record rate-limit events in usage/analytics stream
- [ ] Admin API to view current rate-limit state per agent
- [ ] Configurable burst allowance vs sustained rate
- [ ] Load test rate limiter under concurrent load

---

### ⬜ Audit logging

**What it means:** A tamper-evident record of who changed what and when — agent registrations, policy updates, budget changes, production promotions. Required for compliance, security reviews, and debugging "who broke production?"

**TODO items:**

- [ ] Define audit event schema (actor, action, resource, timestamp, before/after snapshot)
- [ ] Create `audit_events` table in PostgreSQL (append-only)
- [ ] Emit audit events from registry API on all mutating operations
- [ ] Emit audit events from gateway on policy denials and budget blocks
- [ ] Implement `GET /v1/audit-events` with filters (agent, actor, date range)
- [ ] Optional: ship audit logs to SIEM (Splunk, Datadog) via webhook
- [ ] Retention policy configuration (e.g. 90 days hot, 7 years archive)
- [ ] Document audit log fields for compliance teams

---

### 🟡 Provider routing and fallback

**What it means:** If the primary AI provider (e.g. OpenAI) is down, slow, or rejects a request, AgentVoir automatically tries a backup provider (e.g. Anthropic) according to rules defined for each agent — similar to how DNS failover works for websites.

**TODO items:**

- [x] Define model route schema (primary provider/model, fallback provider/model)
- [x] Implement registry API for model routes (`GET/PUT /v1/agents/{agentID}/model-route`)
- [x] Accept model routes from agent YAML manifest
- [x] Gateway provider registry with OpenAI and mock adapters
- [ ] Gateway loads model route from registry for each agent
- [ ] Attempt primary provider first; on failure, try fallback
- [ ] Configurable routing policy (`primary_only`, `primary_then_fallback`, `round_robin`)
- [ ] Add Anthropic, Azure OpenAI, and local model adapters
- [ ] Record which provider was actually used in usage events and response headers
- [ ] Circuit breaker when provider error rate exceeds threshold
- [ ] Admin UI or API to test routing without live traffic

---

### 🟡 Dependency graph API

**What it means:** A map of what each agent depends on — other agents, tools (Zendesk, Salesforce), vector databases, MCP servers. Helps answer impact analysis: "If we change this tool, which agents break?"

**TODO items:**

- [x] Define dependency model (tools, vector stores, agents, APIs)
- [x] Implement `GET/PUT /v1/agents/{agentID}/dependencies` registry endpoints
- [x] Persist dependencies in PostgreSQL
- [x] Accept dependencies from agent YAML manifest
- [ ] Implement graph query API (`GET /v1/dependency-graph?agent_id=...`)
- [ ] Return upstream and downstream dependents (transitive closure)
- [ ] Visualize graph in web console
- [ ] Detect circular agent dependencies and reject on registration
- [ ] Export graph as JSON/GraphML for external tools
- [ ] Link dependency changes to audit log

---

### 🟡 OpenTelemetry traces and Prometheus metrics

**What it means:** Deep visibility into how AgentVoir behaves in production — how long each request takes, where time is spent (cache, policy check, provider call), and aggregate metrics for dashboards and alerts. OpenTelemetry (OTel) is the industry standard for distributed tracing; Prometheus collects numeric metrics.

**TODO items:**

- [x] Add OTel Collector and Prometheus to dev Docker Compose stack
- [x] Add placeholder Prometheus scrape config for gateway and registry-api
- [x] Add placeholder Grafana dashboard JSON
- [ ] Instrument gateway with OTel traces (cache, provider, total latency spans)
- [ ] Instrument registry-api with OTel traces
- [ ] Expose `/metrics` endpoint on gateway (request count, latency histogram, cache hits)
- [ ] Expose `/metrics` endpoint on registry-api
- [ ] Populate Grafana dashboards (latency, error rate, cache hit rate, cost)
- [ ] Propagate `trace_id` from gateway through to usage events
- [ ] Document local observability setup for developers
- [ ] Add alerting rules (error rate > 5%, p99 latency > 10s)

---

## Phase 3: Semantic cache and evals

**Goal:** Smarter caching (similar questions get similar answers), systematic quality testing for agents, and safety hooks for sensitive data.

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

## How to use this roadmap

1. **Pick a phase** aligned with your deployment maturity (Phase 1 is the current baseline).
2. **Work top-to-bottom** within a phase — later items often depend on earlier ones.
3. **Mark items in progress** in GitHub Issues or Projects; link PRs to specific TODO bullets.
4. **Update this doc** when items are completed or scope changes.

For questions about what a technology does in AgentVoir, see [Tech Stack Usage](architecture/tech-stack-usage.md). For local setup, see [Docker Install Guide](../deployments/docker/INSTALL.md).
