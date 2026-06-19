ALTER TABLE agents
  DROP COLUMN IF EXISTS semantic_cache_allowed,
  DROP COLUMN IF EXISTS cache_ttl_seconds,
  DROP COLUMN IF EXISTS cache_mode;
