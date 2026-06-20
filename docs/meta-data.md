Based on above .... and understanding all features that can be coded ... what additional metadata about agents or other dependent systems we need to capture ... that we dont already have. 


Yes. Based on the commercial feature landscape, AgentVoir should capture more than “agent name + token usage + dependencies.” The metadata should let you answer questions like:

> Who owns this agent? What can it access? What tools can it call? What model does it use? What policies apply? What happens if it fails? What changed between versions? Can we safely disable it?

Assuming AgentVoir already has **agent ID, name, description/traits, token usage, cost, dependencies, and basic registry fields**, I would add the following metadata categories.

---

# 1. Ownership and accountability metadata

This is critical for enterprise adoption.

| Metadata              | Why it matters                                 |
| --------------------- | ---------------------------------------------- |
| `owner_team`          | Which business/engineering team owns the agent |
| `technical_owner`     | Person responsible for code/runtime            |
| `business_owner`      | Person responsible for business behavior       |
| `security_reviewer`   | Who approved data/tool access                  |
| `compliance_reviewer` | Who approved regulated usage                   |
| `oncall_contact`      | Who gets paged when agent misbehaves           |
| `support_channel`     | Slack/Teams/email for issues                   |
| `escalation_policy`   | What to do during production incidents         |

Example:

```yaml
ownership:
  owner_team: "Trading Technology"
  technical_owner: "agent-platform-team"
  business_owner: "portfolio-analytics"
  security_reviewer: "appsec-ai-review"
  compliance_reviewer: "compliance-ai-governance"
  support_channel: "#agentvoir-prod-alerts"
  oncall_contact: "agent-platform-oncall"
```

Without this, the registry becomes a static catalog. With this, it becomes a governance system.

---

# 2. Agent lifecycle metadata

You need to know whether an agent is experimental, approved, deprecated, or blocked.

| Metadata             | Example values                                              |
| -------------------- | ----------------------------------------------------------- |
| `lifecycle_stage`    | `draft`, `dev`, `staging`, `prod`, `deprecated`, `disabled` |
| `approval_status`    | `pending`, `approved`, `rejected`, `expired`                |
| `created_at`         | Timestamp                                                   |
| `last_reviewed_at`   | Timestamp                                                   |
| `next_review_due_at` | Timestamp                                                   |
| `retirement_date`    | Optional                                                    |
| `change_ticket`      | Jira/GitHub/ServiceNow reference                            |
| `release_notes`      | What changed in this version                                |

Example:

```yaml
lifecycle:
  stage: "prod"
  approval_status: "approved"
  created_at: "2026-06-18T10:00:00Z"
  last_reviewed_at: "2026-06-18T12:00:00Z"
  next_review_due_at: "2026-09-18T12:00:00Z"
  change_ticket: "AGENT-482"
```

This becomes important when auditors ask: “Who approved this agent and when?”

---

# 3. Version metadata

Commercial tools are strong at prompt/model/eval versioning. AgentVoir should track versions for the whole agent, not just code.

| Metadata               | Why                             |
| ---------------------- | ------------------------------- |
| `agent_version`        | Current released version        |
| `code_version`         | Git SHA or container image tag  |
| `prompt_version`       | Prompt template version         |
| `tool_schema_version`  | Tool contract version           |
| `policy_version`       | OPA/guardrail policy version    |
| `model_config_version` | Model routing/settings version  |
| `eval_suite_version`   | Eval pack used before promotion |

Example:

```yaml
versioning:
  agent_version: "1.4.2"
  git_sha: "8af32c9"
  container_image: "ghcr.io/kaynor/agent-voir/compliance-agent:1.4.2"
  prompt_version: "compliance-risk-prompt:v7"
  policy_version: "opa/compliance-agent:v3"
  eval_suite_version: "compliance-agent-evals:v5"
```

This lets you answer: “Which exact version caused the bad output?”

---

# 4. Runtime and deployment metadata

AgentVoir should know where and how the agent runs.

| Metadata            | Example                                                                         |
| ------------------- | ------------------------------------------------------------------------------- |
| `runtime_type`      | `langgraph`, `crewai`, `autogen`, `vercel-ai-sdk`, `custom-python`, `mcp-agent` |
| `runtime_language`  | `python`, `typescript`, `java`, `go`                                            |
| `deployment_env`    | `dev`, `staging`, `prod`                                                        |
| `hosting_platform`  | `kubernetes`, `ecs`, `lambda`, `vercel`, `databricks`, `local`                  |
| `region`            | `us-east-1`, `us-west-2`, etc.                                                  |
| `replicas`          | Number of instances                                                             |
| `timeout_seconds`   | Runtime timeout                                                                 |
| `retry_policy`      | Retry count/backoff                                                             |
| `concurrency_limit` | Max parallel runs                                                               |
| `state_backend`     | Redis/Postgres/S3/etc.                                                          |
| `queue_backend`     | Kafka/SQS/RabbitMQ/etc.                                                         |

Example:

```yaml
runtime:
  framework: "langgraph"
  language: "python"
  hosting_platform: "kubernetes"
  environment: "prod"
  region: "us-west-2"
  timeout_seconds: 120
  max_concurrency: 25
  state_backend: "postgres://agent-state"
  queue_backend: "kafka://agent-events"
```

This is useful for debugging and operations.

---

# 5. Model usage metadata

Do not just store “uses GPT-4.” Store the full allowed model policy.

| Metadata                     | Why                                                                 |
| ---------------------------- | ------------------------------------------------------------------- |
| `allowed_models`             | Which models the agent may use                                      |
| `default_model`              | Default inference model                                             |
| `fallback_models`            | Backup models                                                       |
| `forbidden_models`           | Models not allowed                                                  |
| `model_selection_strategy`   | `fixed`, `cost_optimized`, `latency_optimized`, `quality_optimized` |
| `max_context_tokens`         | Safety/cost control                                                 |
| `max_output_tokens`          | Prevent runaway usage                                               |
| `temperature`                | Reproducibility                                                     |
| `requires_private_model`     | Whether data cannot leave private infra                             |
| `provider_region_constraint` | Data residency                                                      |

Example:

```yaml
model_policy:
  default_model: "gpt-4.1"
  allowed_models:
    - "gpt-4.1"
    - "claude-3.5-sonnet"
    - "gemini-2.5-pro"
  fallback_models:
    - "gpt-4.1-mini"
  selection_strategy: "quality_then_cost"
  max_context_tokens: 128000
  max_output_tokens: 4000
  temperature: 0.2
  private_model_required: false
```

This makes AgentVoir closer to LiteLLM, Portkey, Helicone, and Vercel AI Gateway.

---

# 6. Tool and MCP metadata

This is one of the most important areas. Agents become risky because they can call tools.

For every tool/MCP server/API dependency, capture:

| Metadata                     | Why                                                                                        |
| ---------------------------- | ------------------------------------------------------------------------------------------ |
| `tool_name`                  | Human-readable name                                                                        |
| `tool_type`                  | `read_only`, `write`, `payment`, `email`, `trading`, `database`, `filesystem`, `code_exec` |
| `mcp_server`                 | MCP source, if applicable                                                                  |
| `allowed_actions`            | Specific functions permitted                                                               |
| `forbidden_actions`          | Explicitly blocked functions                                                               |
| `requires_human_approval`    | HITL gate                                                                                  |
| `approval_policy`            | Who must approve                                                                           |
| `max_call_count_per_run`     | Prevent loops                                                                              |
| `timeout_seconds`            | Prevent hanging                                                                            |
| `rate_limit`                 | Prevent abuse                                                                              |
| `data_return_classification` | Public/internal/confidential/PII                                                           |
| `side_effect_level`          | None/read/write/destructive/financial                                                      |

Example:

```yaml
tools:
  - name: "github"
    type: "source_control"
    mcp_server: "github-mcp-prod"
    allowed_actions:
      - "read_issue"
      - "create_issue"
      - "comment_on_pr"
    forbidden_actions:
      - "merge_pr"
      - "delete_branch"
    requires_human_approval: true
    approval_policy: "engineering-owner-approval"
    side_effect_level: "write"

  - name: "edgar"
    type: "financial_data"
    allowed_actions:
      - "search_filings"
      - "read_10k"
      - "read_10q"
    requires_human_approval: false
    side_effect_level: "read_only"
```

This is where AgentVoir can become very valuable.

---

# 7. Data access metadata

AgentVoir should know what data each agent can touch.

| Metadata                        | Why                                                                           |
| ------------------------------- | ----------------------------------------------------------------------------- |
| `data_sources`                  | APIs, databases, files, SaaS tools                                            |
| `data_classification`           | `public`, `internal`, `confidential`, `restricted`, `PII`, `PHI`, `financial` |
| `read_access`                   | What it can read                                                              |
| `write_access`                  | What it can modify                                                            |
| `data_residency`                | Region/legal constraint                                                       |
| `retention_policy`              | How long logs/prompts/responses are stored                                    |
| `redaction_required`            | Whether PII redaction must run                                                |
| `allowed_export_destinations`   | Where outputs can be sent                                                     |
| `forbidden_export_destinations` | Where outputs cannot be sent                                                  |

Example:

```yaml
data_access:
  classification:
    - "confidential"
    - "financial"
  sources:
    - name: "positions-db"
      type: "postgres"
      access: "read_only"
      tables:
        - "portfolio_positions"
        - "futures_exposure"
    - name: "sec-edgar"
      type: "external_api"
      access: "read_only"
  pii_redaction_required: true
  retention_days: 30
  allowed_export_destinations:
    - "internal-dashboard"
    - "approved-email-domain"
```

This lets you answer: “Which agents can access confidential portfolio data?”

---

# 8. Policy and guardrail metadata

Since AgentVoir is planning OPA, policies should be first-class metadata.

| Metadata                     | Example                           |
| ---------------------------- | --------------------------------- |
| `policy_bundle`              | OPA policy package                |
| `policy_version`             | Version of policy rules           |
| `pre_prompt_filters`         | Filters before LLM call           |
| `post_response_filters`      | Filters after LLM output          |
| `tool_call_policies`         | Rules before tool execution       |
| `data_loss_policies`         | DLP checks                        |
| `jailbreak_detection`        | Enabled/disabled                  |
| `prompt_injection_detection` | Enabled/disabled                  |
| `blocked_topics`             | Things the agent must not answer  |
| `allowed_output_formats`     | JSON, markdown, PDF, ticket, etc. |

Example:

```yaml
governance:
  opa_policy_bundle: "policies/trading-agent.rego"
  policy_version: "v12"
  prompt_injection_detection: true
  pii_redaction: true
  jailbreak_detection: true
  blocked_actions:
    - "place_trade"
    - "send_external_email_without_approval"
  human_approval_required_for:
    - "write_database"
    - "send_email"
    - "create_github_issue"
```

This is the core of runtime governance.

---

# 9. Risk metadata

Each agent should have an explicit risk profile.

| Metadata                    | Example values                                                         |
| --------------------------- | ---------------------------------------------------------------------- |
| `risk_tier`                 | `low`, `medium`, `high`, `critical`                                    |
| `impact_area`               | `customer_support`, `trading`, `legal`, `finance`, `engineering`, `hr` |
| `can_take_external_action`  | Boolean                                                                |
| `can_modify_systems`        | Boolean                                                                |
| `can_access_sensitive_data` | Boolean                                                                |
| `can_spend_money`           | Boolean                                                                |
| `can_send_messages`         | Boolean                                                                |
| `requires_periodic_review`  | Boolean                                                                |
| `review_frequency_days`     | 30/60/90/180                                                           |

Example:

```yaml
risk:
  tier: "high"
  impact_area:
    - "financial_compliance"
    - "portfolio_risk"
  can_access_sensitive_data: true
  can_modify_systems: false
  can_send_external_messages: false
  can_spend_money: false
  review_frequency_days: 90
```

This is very useful for dashboards.

---

# 10. Budget and quota metadata

You already have token/cost tracking, but you should add forward-looking controls.

| Metadata              | Why                               |
| --------------------- | --------------------------------- |
| `daily_token_budget`  | Prevent runaway cost              |
| `monthly_cost_budget` | Team-level cost control           |
| `per_run_token_limit` | Stop infinite loops               |
| `per_user_budget`     | User-level fairness               |
| `per_tenant_budget`   | SaaS/multi-tenant support         |
| `alert_thresholds`    | Notify at 50%, 80%, 100%          |
| `hard_stop_enabled`   | Kill calls after exceeding budget |
| `budget_owner`        | Person/team that owns spend       |

Example:

```yaml
budget:
  monthly_usd_limit: 5000
  daily_token_limit: 2000000
  per_run_token_limit: 50000
  alert_thresholds:
    - 0.5
    - 0.8
    - 1.0
  hard_stop_enabled: true
```

This moves AgentVoir from passive tracking to active cost governance.

---

# 11. Evaluation metadata

Commercial tools like LangSmith, Braintrust, Arize, Galileo, and Datadog are strong here. AgentVoir should at least track eval status.

| Metadata                  | Why                                               |
| ------------------------- | ------------------------------------------------- |
| `eval_suite_id`           | Which tests apply                                 |
| `last_eval_run_id`        | Latest eval execution                             |
| `last_eval_score`         | Overall score                                     |
| `eval_threshold`          | Minimum passing score                             |
| `golden_dataset`          | Regression dataset                                |
| `failure_categories`      | Hallucination, wrong tool, policy violation, etc. |
| `promotion_gate_required` | Whether evals block prod deployment               |
| `shadow_mode_enabled`     | Run new agent version silently                    |
| `canary_percent`          | Percentage of traffic on new version              |

Example:

```yaml
evals:
  suite_id: "financial-compliance-agent-evals"
  golden_dataset: "s3://agentvoir-evals/compliance/v5.jsonl"
  last_score: 0.91
  required_score: 0.88
  promotion_gate_required: true
  shadow_mode_enabled: true
  canary_percent: 5
```

This is important if AgentVoir wants to support safe deployment.

---

# 12. Observability metadata

The agent should advertise how it is traced, logged, and monitored.

| Metadata            | Why                          |
| ------------------- | ---------------------------- |
| `trace_enabled`     | Whether traces are collected |
| `otel_service_name` | OpenTelemetry service name   |
| `log_level`         | Debug/info/warn/error        |
| `metrics_enabled`   | Whether metrics are emitted  |
| `slo_latency_ms`    | Latency objective            |
| `slo_success_rate`  | Reliability objective        |
| `dashboard_url`     | Link to Datadog/Grafana/etc. |
| `alert_policy`      | Alerting configuration       |
| `sample_rate`       | Trace sampling rate          |

Example:

```yaml
observability:
  tracing_enabled: true
  otel_service_name: "agentvoir.compliance-agent"
  metrics_enabled: true
  dashboard_url: "https://grafana.example.com/d/agent-compliance"
  slo:
    p95_latency_ms: 8000
    success_rate: 0.98
  alert_policy: "prod-agent-critical-alerts"
```

This makes AgentVoir integrate cleanly with Datadog, Grafana, OpenTelemetry, Arize, etc.

---

# 13. Cache metadata

Since AgentVoir has an LLM cache/gateway concept, cache behavior should be explicit.

| Metadata                  | Why                                          |
| ------------------------- | -------------------------------------------- |
| `cache_enabled`           | Whether responses can be cached              |
| `cache_type`              | Exact, semantic, embedding-based             |
| `cache_ttl_seconds`       | Expiration                                   |
| `cache_scope`             | Global/team/user/session                     |
| `cache_key_strategy`      | Prompt hash, normalized prompt, semantic key |
| `sensitive_cache_allowed` | Whether sensitive data can be cached         |
| `cache_bypass_conditions` | When not to cache                            |
| `cache_hit_policy`        | Return exact cached output vs revalidate     |

Example:

```yaml
cache:
  enabled: true
  type: "semantic"
  ttl_seconds: 86400
  scope: "team"
  cache_sensitive_data: false
  bypass_conditions:
    - "contains_pii"
    - "high_risk_tool_call"
    - "real_time_market_data"
```

Important: for trading/market/compliance agents, caching must be careful because stale data can be dangerous.

---

# 14. Human-in-the-loop metadata

For enterprise agents, HITL should be a first-class concept.

| Metadata              | Example                                                |
| --------------------- | ------------------------------------------------------ |
| `approval_required`   | Boolean                                                |
| `approval_triggers`   | Tool call, risk score, cost threshold, external action |
| `approver_role`       | Manager, compliance, engineer, business owner          |
| `approval_timeout`    | How long to wait                                       |
| `fallback_on_timeout` | Reject, continue, escalate                             |
| `approval_audit_log`  | Where approval records are stored                      |

Example:

```yaml
human_in_loop:
  enabled: true
  triggers:
    - action: "send_external_email"
      approver_role: "business_owner"
    - action: "write_database"
      approver_role: "technical_owner"
    - risk_score_above: 80
      approver_role: "compliance_reviewer"
  timeout_minutes: 60
  on_timeout: "reject"
```

This is especially useful for agents that touch GitHub, emails, tickets, trades, payments, or production systems.

---

# 15. Dependency graph metadata

This is one of AgentVoir’s biggest potential differentiators.

Each dependency should not just be a string. It should be a typed node.

Dependency types:

```yaml
dependency_types:
  - agent
  - model
  - prompt
  - tool
  - mcp_server
  - database
  - api
  - file_store
  - queue
  - vector_store
  - policy
  - eval_suite
  - secret
  - cache
  - workflow
  - dashboard
```

Example:

```yaml
dependencies:
  - id: "model:gpt-4.1"
    type: "model"
    criticality: "high"
    fallback: "model:gpt-4.1-mini"

  - id: "tool:sec-edgar"
    type: "external_api"
    criticality: "medium"
    fallback: null

  - id: "policy:trading-risk-policy-v3"
    type: "policy"
    criticality: "critical"

  - id: "vectorstore:compliance-knowledge-base"
    type: "vector_store"
    criticality: "high"
```

This enables blast-radius queries:

```text
Show all production agents that depend on Claude.
Show all high-risk agents that can call email tools.
Show all agents affected if the SEC EDGAR API is down.
Show all agents using stale prompt versions.
Show all agents without evals in the last 30 days.
```

That would be extremely useful.

---

# 16. External system metadata

AgentVoir should also register dependent systems, not only agents.

For every external API/tool/database/model provider, capture:

| Metadata              | Why                                             |
| --------------------- | ----------------------------------------------- |
| `system_id`           | Unique ID                                       |
| `system_type`         | API, DB, model provider, SaaS, queue, vector DB |
| `owner`               | Owning team                                     |
| `environment`         | Dev/staging/prod                                |
| `auth_type`           | API key, OAuth, IAM, mTLS                       |
| `secret_reference`    | Pointer to secret, not the secret itself        |
| `rate_limits`         | Provider limits                                 |
| `cost_model`          | Cost per token/request/GB/etc.                  |
| `sla`                 | Availability expectations                       |
| `data_classification` | What data passes through it                     |
| `allowed_agents`      | Which agents can use it                         |
| `blocked_agents`      | Which agents cannot use it                      |
| `deprecation_status`  | Whether it is being retired                     |

Example:

```yaml
external_system:
  id: "api:sec-edgar"
  type: "external_api"
  owner_team: "data-platform"
  auth_type: "api_key"
  secret_reference: "vault://agentvoir/sec-edgar-api-key"
  rate_limit:
    requests_per_minute: 100
  data_classification:
    - "public"
  allowed_agents:
    - "agent:fundamental-analysis"
    - "agent:compliance-review"
```

Important: store a **secret reference**, never the actual secret.

---

# 17. Incident and kill-switch metadata

AgentVoir should be able to disable agents, tools, models, or dependencies.

| Metadata              | Why                                         |
| --------------------- | ------------------------------------------- |
| `kill_switch_enabled` | Can this agent be disabled quickly?         |
| `disabled_reason`     | Why it was disabled                         |
| `disabled_by`         | Who disabled it                             |
| `disabled_at`         | Timestamp                                   |
| `quarantine_mode`     | Read-only, no-tools, no-model, full-disable |
| `last_incident_id`    | Link to incident                            |
| `incident_count_30d`  | Risk signal                                 |

Example:

```yaml
runtime_controls:
  kill_switch_enabled: true
  current_mode: "active"
  allowed_modes:
    - "active"
    - "read_only"
    - "no_external_tools"
    - "quarantined"
    - "disabled"
```

This is a must-have for enterprise confidence.

---

# 18. Recommended MVP metadata model

For the first serious version of AgentVoir, I would not try to build everything. I would define these core entities:

```text
Agent
AgentVersion
Tool
Model
Prompt
Policy
DataSource
EvalSuite
Deployment
ExternalSystem
RuntimeTraceSummary
ApprovalRecord
Incident
```

And this minimal but strong metadata per agent:

```yaml
agent:
  id: "agent:compliance-position-limit-checker"
  name: "Position Limit Compliance Agent"
  description: "Checks futures positions against exchange and regulatory limits."
  owner_team: "Compliance Technology"
  lifecycle_stage: "prod"
  risk_tier: "high"

versioning:
  agent_version: "1.0.0"
  git_sha: "abc123"
  prompt_version: "position-limit:v4"
  policy_version: "opa-position-limit:v2"

runtime:
  framework: "langgraph"
  environment: "prod"
  hosting_platform: "kubernetes"

model_policy:
  default_model: "gpt-4.1"
  allowed_models:
    - "gpt-4.1"
    - "gpt-4.1-mini"
  max_context_tokens: 128000
  monthly_usd_limit: 3000

data_access:
  classification:
    - "confidential"
    - "financial"
  sources:
    - "positions-db"
    - "cme-position-limits-api"
    - "ice-position-limits-api"

tools:
  - name: "read_positions"
    side_effect_level: "read_only"
  - name: "create_warning_report"
    side_effect_level: "write"
    requires_human_approval: false
  - name: "email_compliance_team"
    side_effect_level: "external_message"
    requires_human_approval: true

governance:
  pii_redaction: true
  prompt_injection_detection: true
  opa_policy_bundle: "policies/position-limit-agent.rego"

evals:
  suite_id: "position-limit-evals"
  last_score: 0.94
  required_score: 0.90
  promotion_gate_required: true

observability:
  tracing_enabled: true
  otel_service_name: "agentvoir.position-limit-agent"
  slo:
    p95_latency_ms: 10000
    success_rate: 0.98

runtime_controls:
  kill_switch_enabled: true
  current_mode: "active"
```

---

# My strongest recommendation

The biggest metadata gap to fix is this:

> AgentVoir should model every agent as a **governed runtime asset**, not just a registered software component.

That means each agent should have metadata for:

1. **Who owns it**
2. **What version is running**
3. **What model it can use**
4. **What tools it can call**
5. **What data it can touch**
6. **What policies guard it**
7. **What evals it passed**
8. **What it costs**
9. **What it depends on**
10. **How to disable it**

If you build those metadata fields cleanly, AgentVoir becomes much more than a registry. It becomes an **open-source enterprise agent control plane**.
