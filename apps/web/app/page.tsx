export default function HomePage() {
  return (
    <main className="shell">
      <section className="hero">
        <p className="eyebrow">AgentVoir</p>
        <h1>Enterprise AI agent registry and LLM gateway</h1>
        <p>
          Register agents, govern model usage, enforce policies, track token cost,
          map dependencies, and cache repeated LLM requests.
        </p>
      </section>

      <section className="grid">
        <article>
          <h2>Agent Registry</h2>
          <p>Identity, ownership, lifecycle, capabilities, and dependencies.</p>
        </article>
        <article>
          <h2>LLM Gateway</h2>
          <p>OpenAI-compatible proxy with exact and semantic cache support.</p>
        </article>
        <article>
          <h2>Governance</h2>
          <p>Budgets, policy-as-code, audit logs, and enterprise controls.</p>
        </article>
      </section>
    </main>
  );
}
