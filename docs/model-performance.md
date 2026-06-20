are we capturing performance or response times from a model that agent depends on?


Not explicitly enough.

Based on what we defined so far, AgentVoir is capturing **token usage, cost, dependencies, evals, observability, and quality feedback**, but we should add a more explicit section for:

> **Model dependency performance metadata**

This means: for every model/provider an agent depends on, AgentVoir should track latency, reliability, cost, errors, fallback behavior, and quality impact.

## Add this metadata to each agent

```yaml
model_performance:
  enabled: true

  default_model: "gpt-4.1"
  provider: "openai"

  latency:
    p50_ms: 850
    p95_ms: 3200
    p99_ms: 6200
    avg_ms_24h: 1100
    avg_ms_7d: 1250

  reliability:
    success_rate_24h: 0.992
    error_rate_24h: 0.008
    timeout_rate_24h: 0.003
    rate_limit_rate_24h: 0.002

  throughput:
    requests_24h: 18420
    tokens_in_24h: 9200000
    tokens_out_24h: 2100000

  cost:
    cost_24h_usd: 184.52
    cost_7d_usd: 1220.19
    avg_cost_per_run_usd: 0.031

  fallback:
    fallback_enabled: true
    fallback_model: "gpt-4.1-mini"
    fallback_invocations_24h: 91
    fallback_reason_counts:
      timeout: 42
      rate_limit: 31
      provider_error: 18
```

## Capture performance per model dependency, not only per agent

An agent may depend on multiple models:

```yaml
model_dependencies:
  - model_id: "openai:gpt-4.1"
    role: "primary_reasoning_model"
    p95_latency_ms: 3200
    success_rate_24h: 0.992
    cost_24h_usd: 184.52

  - model_id: "openai:gpt-4.1-mini"
    role: "fallback_model"
    p95_latency_ms: 1200
    success_rate_24h: 0.997
    cost_24h_usd: 21.44

  - model_id: "qwen:qwen3-embedding-0.6b"
    role: "embedding_model"
    p95_latency_ms: 220
    success_rate_24h: 0.999
    cost_24h_usd: 3.10
```

This is important because a single agent may use:

```text
reasoning model
summarization model
embedding model
reranker model
classification model
fallback model
judge/evaluator model
```

Each of those dependencies should have its own performance profile.

## Metrics AgentVoir should capture

At minimum:

| Metric                   | Why it matters                                    |
| ------------------------ | ------------------------------------------------- |
| `time_to_first_token_ms` | User-perceived responsiveness for streaming       |
| `total_latency_ms`       | End-to-end model response time                    |
| `queue_time_ms`          | Provider/platform congestion                      |
| `prompt_tokens`          | Input cost and context usage                      |
| `completion_tokens`      | Output cost                                       |
| `total_tokens`           | Total consumption                                 |
| `cost_usd`               | Spend tracking                                    |
| `success/failure`        | Reliability                                       |
| `error_type`             | Timeout, rate limit, provider error, policy block |
| `retry_count`            | Detect unstable providers                         |
| `fallback_used`          | See when primary model failed                     |
| `cache_hit`              | See whether AgentVoir avoided model call          |
| `model_region`           | Data residency/performance debugging              |
| `provider_request_id`    | Debugging with provider support                   |

## Add model-level SLOs

Each agent should define expected performance.

```yaml
model_slo:
  p95_latency_ms: 5000
  max_timeout_rate: 0.01
  min_success_rate: 0.99
  max_cost_per_run_usd: 0.10
  max_time_to_first_token_ms: 1500
```

Then AgentVoir can trigger governance actions:

```yaml
model_performance_gates:
  - condition: "p95_latency_ms > 5000 for 15 minutes"
    action: "route_to_fallback_model"

  - condition: "success_rate_5m < 0.97"
    action: "disable_primary_model_temporarily"

  - condition: "cost_per_run_usd > 0.25"
    action: "require_cheaper_model_for_low_risk_tasks"

  - condition: "timeout_rate_5m > 0.05"
    action: "open_incident"
```

## Store model performance as dependency health

In the dependency graph, model nodes should have health status:

```yaml
dependency_health:
  dependency_id: "model:openai:gpt-4.1"
  dependency_type: "model"
  status: "degraded"
  affected_agents:
    - "agent:compliance-position-limit-checker"
    - "agent:fundamental-analysis-agent"
  p95_latency_ms: 7900
  success_rate_5m: 0.941
  last_checked_at: "2026-06-19T09:15:00Z"
```

This lets AgentVoir answer:

```text
Which agents are slow because GPT-4.1 latency increased?
Which agents failed because Anthropic rate-limited us?
Which agents should be moved to fallback models?
Which model dependency is driving cost today?
```

## Suggested database tables

I would add:

```text
model_dependency
model_call_trace
model_performance_rollup
model_slo
model_routing_event
model_provider_health
```

Core trace table:

```sql
CREATE TABLE model_call_trace (
    id UUID PRIMARY KEY,
    agent_id TEXT NOT NULL,
    agent_version TEXT,
    run_id TEXT NOT NULL,
    trace_id TEXT,
    model_provider TEXT NOT NULL,
    model_name TEXT NOT NULL,
    model_role TEXT,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    cost_usd NUMERIC,
    time_to_first_token_ms INTEGER,
    total_latency_ms INTEGER,
    status TEXT,
    error_type TEXT,
    retry_count INTEGER DEFAULT 0,
    fallback_used BOOLEAN DEFAULT false,
    cache_hit BOOLEAN DEFAULT false,
    provider_request_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
```

Rollup table:

```sql
CREATE TABLE model_performance_rollup (
    id UUID PRIMARY KEY,
    agent_id TEXT,
    model_provider TEXT NOT NULL,
    model_name TEXT NOT NULL,
    window_start TIMESTAMP NOT NULL,
    window_end TIMESTAMP NOT NULL,
    request_count INTEGER,
    success_rate NUMERIC,
    error_rate NUMERIC,
    timeout_rate NUMERIC,
    rate_limit_rate NUMERIC,
    p50_latency_ms INTEGER,
    p95_latency_ms INTEGER,
    p99_latency_ms INTEGER,
    total_tokens INTEGER,
    total_cost_usd NUMERIC,
    fallback_count INTEGER,
    cache_hit_rate NUMERIC
);
```

## My recommendation

Yes, AgentVoir should absolutely capture this.

I would make **model performance** a first-class part of the dependency graph:

```text
Agent
  depends on Model
  depends on Tool
  depends on Prompt
  depends on Policy
  depends on DataSource
```

And every model dependency should expose:

```text
latency
reliability
cost
token usage
error rate
fallback usage
cache hit rate
quality score impact
```

This is important because sometimes an agent is “bad” not because the agent logic is bad, but because:

```text
the model is slow
the provider is rate-limiting
the fallback model is weaker
the context window is too small
the expensive model is overused
the cheap model is causing quality drops
```

So the answer is: **we discussed observability generally, but we should explicitly add model dependency performance tracking as a core AgentVoir feature.**
