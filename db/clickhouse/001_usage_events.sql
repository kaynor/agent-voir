CREATE TABLE IF NOT EXISTS usage_events
(
  event_time DateTime64(3),
  trace_id String,
  tenant_id String,
  agent_id String,
  agent_version String,
  user_id String,
  provider String,
  model String,
  cache_status LowCardinality(String),
  prompt_tokens UInt64,
  completion_tokens UInt64,
  cached_tokens UInt64,
  cost_usd Float64,
  latency_ms UInt64,
  status_code UInt16,
  error_code String
)
ENGINE = MergeTree
PARTITION BY toDate(event_time)
ORDER BY (tenant_id, agent_id, event_time);
