import Link from "next/link";
import { getAgent, getBudget, getUsageSummary, listDependencies } from "../../../lib/api";

type PageProps = {
  params: { agentId: string };
  searchParams: { version?: string; environment?: string };
};

export default async function AgentDetailPage({ params, searchParams }: PageProps) {
  const version = searchParams.version ?? "0.1.0";
  const environment = searchParams.environment ?? "staging";
  let agent;
  let budget = null;
  let dependencies: Awaited<ReturnType<typeof listDependencies>> = [];
  let usage = { cost_usd: 0, cache_hit_rate: 0, event_count: 0, period: "monthly", prompt_tokens: 0, completion_tokens: 0 };
  let error = "";

  try {
    [agent, budget, dependencies, usage] = await Promise.all([
      getAgent(params.agentId, version, environment),
      getBudget(params.agentId, version),
      listDependencies(params.agentId, version),
      getUsageSummary(params.agentId),
    ]);
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load agent";
  }

  if (error || !agent) {
    return (
      <div className="alert">
        {error || "Agent not found"}{" "}
        <Link href="/agents">Back to agents</Link>
      </div>
    );
  }

  return (
    <>
      <section className="hero compact">
        <p className="eyebrow">Agent detail</p>
        <h1>{agent.name}</h1>
        <p>
          {agent.agent_id} · v{agent.version} · {agent.environment}
        </p>
      </section>

      <section className="stats">
        <article>
          <h2>{agent.lifecycle}</h2>
          <p>Lifecycle</p>
        </article>
        <article>
          <h2>${usage.cost_usd.toFixed(4)}</h2>
          <p>Monthly spend</p>
        </article>
        <article>
          <h2>{(usage.cache_hit_rate * 100).toFixed(1)}%</h2>
          <p>Cache hit rate</p>
        </article>
        <article>
          <h2>{usage.event_count}</h2>
          <p>Requests (30d)</p>
        </article>
      </section>

      <section className="grid two-col">
        <article className="panel">
          <h2>Metadata</h2>
          <dl className="detail-list">
            <dt>Owner team</dt>
            <dd>{agent.owner_team}</dd>
            <dt>Risk level</dt>
            <dd>{agent.risk_level}</dd>
            <dt>Cache mode</dt>
            <dd>{agent.cache_mode ?? "exact_only"}</dd>
            <dt>Cache TTL</dt>
            <dd>{agent.cache_ttl_seconds ?? 86400}s</dd>
            <dt>Data classes</dt>
            <dd>{agent.data_classes?.join(", ") || "none"}</dd>
          </dl>
        </article>

        <article className="panel">
          <h2>Budget & policy</h2>
          <dl className="detail-list">
            <dt>Monthly budget</dt>
            <dd>{budget?.monthly_usd != null ? `$${budget.monthly_usd}` : "not set"}</dd>
            <dt>Max prompt tokens</dt>
            <dd>{budget?.max_prompt_tokens_per_request ?? "not set"}</dd>
            <dt>Allowed providers</dt>
            <dd>{agent.policies?.allowed_providers?.join(", ") || "not set"}</dd>
            <dt>PII allowed</dt>
            <dd>{agent.policies?.pii_allowed ? "yes" : "no"}</dd>
            <dt>Audit log required</dt>
            <dd>{agent.policies?.require_audit_log ? "yes" : "no"}</dd>
          </dl>
        </article>
      </section>

      <section className="panel">
        <h2>Dependencies</h2>
        {dependencies.length === 0 ? (
          <p className="muted">No dependencies registered.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Type</th>
                <th>Name</th>
              </tr>
            </thead>
            <tbody>
              {dependencies.map((dep) => (
                <tr key={dep.id}>
                  <td>{dep.dependency_type}</td>
                  <td>{dep.dependency_name}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <p>
        <Link href="/agents">← Back to agents</Link>
      </p>
    </>
  );
}
