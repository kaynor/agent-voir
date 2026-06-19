import Link from "next/link";
import { listAgents, type AgentListResult } from "../../lib/api";

export default async function AgentsPage() {
  let agents: AgentListResult = { items: [], total: 0, limit: 0, offset: 0 };
  let error = "";
  try {
    agents = await listAgents();
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load agents";
  }

  return (
    <>
      <section className="hero compact">
        <p className="eyebrow">Registry</p>
        <h1>Agents</h1>
        <p>{agents.total} agents registered across environments.</p>
      </section>

      {error ? <div className="alert">{error}</div> : null}

      <section className="panel">
        <table>
          <thead>
            <tr>
              <th>Agent ID</th>
              <th>Name</th>
              <th>Version</th>
              <th>Environment</th>
              <th>Lifecycle</th>
              <th>Risk</th>
              <th>Cache</th>
            </tr>
          </thead>
          <tbody>
            {agents.items.map((agent) => (
              <tr key={`${agent.agent_id}:${agent.version}:${agent.environment}`}>
                <td>
                  <Link href={`/agents/${agent.agent_id}?version=${agent.version}&environment=${agent.environment}`}>
                    {agent.agent_id}
                  </Link>
                </td>
                <td>{agent.name}</td>
                <td>{agent.version}</td>
                <td>{agent.environment}</td>
                <td>
                  <span className={`badge badge-${agent.lifecycle}`}>{agent.lifecycle}</span>
                </td>
                <td>{agent.risk_level}</td>
                <td>{agent.cache_mode ?? "exact_only"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </>
  );
}
