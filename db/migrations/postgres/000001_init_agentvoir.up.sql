CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS agents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  owner_team TEXT NOT NULL,
  cost_center TEXT,
  environment TEXT NOT NULL DEFAULT 'dev',
  framework TEXT,
  risk_level TEXT NOT NULL DEFAULT 'low',
  lifecycle TEXT NOT NULL DEFAULT 'draft',
  data_classes TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(agent_id, version, environment)
);

CREATE TABLE IF NOT EXISTS agent_dependencies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id TEXT NOT NULL,
  agent_version TEXT NOT NULL,
  dependency_type TEXT NOT NULL,
  dependency_name TEXT NOT NULL,
  dependency_version TEXT,
  required BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS prompts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  prompt_id TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT NOT NULL,
  owner_team TEXT NOT NULL,
  template TEXT NOT NULL,
  risk_level TEXT NOT NULL DEFAULT 'low',
  approved_models TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(prompt_id, version)
);

CREATE TABLE IF NOT EXISTS model_routes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id TEXT NOT NULL,
  agent_version TEXT NOT NULL,
  primary_provider TEXT NOT NULL,
  primary_model TEXT NOT NULL,
  fallback_provider TEXT,
  fallback_model TEXT,
  routing_policy TEXT NOT NULL DEFAULT 'primary_then_fallback',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS budgets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id TEXT NOT NULL,
  agent_version TEXT NOT NULL,
  monthly_usd NUMERIC(12, 4),
  max_prompt_tokens_per_request BIGINT,
  max_completion_tokens_per_request BIGINT,
  requests_per_minute BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cache_entries (
  cache_key TEXT PRIMARY KEY,
  tenant_id TEXT NOT NULL,
  agent_id TEXT NOT NULL,
  provider TEXT NOT NULL,
  model TEXT NOT NULL,
  prompt_hash TEXT NOT NULL,
  context_hash TEXT,
  response_hash TEXT,
  expires_at TIMESTAMPTZ,
  hit_count BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
