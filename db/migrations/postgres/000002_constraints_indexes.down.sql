DROP INDEX IF EXISTS cache_entries_agent_idx;
DROP INDEX IF EXISTS prompts_prompt_id_idx;
DROP INDEX IF EXISTS agents_agent_id_idx;
DROP INDEX IF EXISTS agent_dependencies_agent_idx;
DROP INDEX IF EXISTS model_routes_agent_version_uidx;
DROP INDEX IF EXISTS budgets_agent_version_uidx;

ALTER TABLE model_routes DROP COLUMN IF EXISTS updated_at;
