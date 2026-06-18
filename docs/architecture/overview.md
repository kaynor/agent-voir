# AgentVoir Architecture Overview

AgentVoir is organized around three planes:

1. **Control plane**: registry API, agent metadata, prompt metadata, policies, lifecycle, budgets.
2. **Data plane**: low-latency LLM gateway/proxy, cache, routing, provider adapters.
3. **Observability plane**: traces, usage events, cost, cache analytics, eval results, dashboards.

```text
Agents / Apps
   |
   v
AgentVoir Gateway
   |-- AuthN/AuthZ
   |-- Policy checks
   |-- Cache lookup
   |-- Provider routing
   |-- Usage/event emission
   v
Model providers / local models

Registry API <-> PostgreSQL
Gateway      <-> Redis / Semantic cache
Analytics    <-> ClickHouse
Policy       <-> OPA
Telemetry    <-> OpenTelemetry / Prometheus / Grafana
```

See [Tech Stack Usage](tech-stack-usage.md) for details on ClickHouse, OPA, Prometheus, and Grafana.
