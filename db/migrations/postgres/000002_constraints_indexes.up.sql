ALTER TABLE model_routes
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE UNIQUE INDEX IF NOT EXISTS budgets_agent_version_uidx
  ON budgets (agent_id, agent_version);

CREATE UNIQUE INDEX IF NOT EXISTS model_routes_agent_version_uidx
  ON model_routes (agent_id, agent_version);

CREATE INDEX IF NOT EXISTS agent_dependencies_agent_idx
  ON agent_dependencies (agent_id, agent_version);

CREATE INDEX IF NOT EXISTS agents_agent_id_idx
  ON agents (agent_id);

CREATE INDEX IF NOT EXISTS prompts_prompt_id_idx
  ON prompts (prompt_id);

CREATE INDEX IF NOT EXISTS cache_entries_agent_idx
  ON cache_entries (agent_id, tenant_id);
