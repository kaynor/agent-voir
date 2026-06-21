import Link from "next/link";
import { getPlatformOverview } from "../lib/api";

export default async function DashboardPage() {
  let overview;
  let error = "";
  try {
    overview = await getPlatformOverview();
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load dashboard";
  }

  return (
    <div className="shell">
      <section className="hero">
        <p className="eyebrow">Admin Console</p>
        <h1>Govern agents, cost, and cache in one place</h1>
        <p>
          Live view of registered agents, monthly spend, and cache efficiency from your local
          AgentVoir stack.
        </p>
      </section>

      {error ? (
        <div className="alert">
          <strong>Could not reach AgentVoir APIs.</strong> Start the stack with{" "}
          <code>./scripts/onebox.sh</code> and set <code>REGISTRY_API_URL</code> /{" "}
          <code>TOKEN_ACCOUNTING_URL</code> if needed. ({error})
        </div>
      ) : null}

      {overview ? (
        <>
          <section className="stats">
            <article>
              <h2>{overview.agentCount}</h2>
              <p>Registered agents</p>
            </article>
            <article>
              <h2>${overview.monthlyCostUSD.toFixed(4)}</h2>
              <p>Monthly LLM spend</p>
            </article>
            <article>
              <h2>{(overview.cacheHitRate * 100).toFixed(1)}%</h2>
              <p>Cache hit rate</p>
            </article>
            <article>
              <h2>{overview.monthlyEvents}</h2>
              <p>Usage events (30d)</p>
            </article>
          </section>

          <section className="panel">
            <div className="panel-header">
              <h2>Recent agents</h2>
              <Link href="/agents">View all</Link>
            </div>
            <table>
              <thead>
                <tr>
                  <th>Agent</th>
                  <th>Version</th>
                  <th>Environment</th>
                  <th>Lifecycle</th>
                  <th>Owner</th>
                </tr>
              </thead>
              <tbody>
                {overview.agents.slice(0, 8).map((agent) => (
                  <tr key={`${agent.agent_id}:${agent.version}:${agent.environment}`}>
                    <td>
                      <Link href={`/agents/${agent.agent_id}?version=${agent.version}&environment=${agent.environment}`}>
                        {agent.name}
                      </Link>
                    </td>
                    <td>{agent.version}</td>
                    <td>{agent.environment}</td>
                    <td>
                      <span className={`badge badge-${agent.lifecycle}`}>{agent.lifecycle}</span>
                    </td>
                    <td>{agent.owner_team}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </>
      ) : null}
    </div>
  );
}
