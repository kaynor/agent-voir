const DEFAULT_REGISTRY_URL = "http://localhost:8081";
const DEFAULT_USAGE_URL = "http://localhost:8082";

export function registryUrl(): string {
  return process.env.REGISTRY_API_URL ?? DEFAULT_REGISTRY_URL;
}

export function usageUrl(): string {
  return process.env.TOKEN_ACCOUNTING_URL ?? DEFAULT_USAGE_URL;
}

export type Agent = {
  id: string;
  agent_id: string;
  name: string;
  version: string;
  owner_team: string;
  environment: string;
  lifecycle: string;
  risk_level: string;
  cache_mode?: string;
  cache_ttl_seconds?: number;
  semantic_cache_allowed?: boolean;
  policies?: {
    allowed_providers?: string[];
    pii_allowed?: boolean;
    require_audit_log?: boolean;
  };
  data_classes?: string[];
  created_at: string;
  updated_at: string;
};

export type AgentListResult = {
  items: Agent[];
  total: number;
  limit: number;
  offset: number;
};

export type Budget = {
  monthly_usd?: number;
  max_prompt_tokens_per_request?: number;
  max_completion_tokens_per_request?: number;
};

export type Dependency = {
  id: string;
  dependency_type: string;
  dependency_name: string;
};

export type UsageSummary = {
  period: string;
  event_count: number;
  prompt_tokens: number;
  completion_tokens: number;
  cost_usd: number;
  cache_hit_rate: number;
};

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(url, { next: { revalidate: 5 } });
  if (!response.ok) {
    throw new Error(`Request failed (${response.status}) for ${url}`);
  }
  return response.json() as Promise<T>;
}

export async function listAgents(limit = 100): Promise<AgentListResult> {
  return fetchJSON<AgentListResult>(`${registryUrl()}/v1/agents?limit=${limit}`);
}

export async function getAgent(agentId: string, version: string, environment = "dev"): Promise<Agent> {
  const params = new URLSearchParams({ version, environment });
  return fetchJSON<Agent>(`${registryUrl()}/v1/agents/${encodeURIComponent(agentId)}?${params}`);
}

export async function getBudget(agentId: string, version: string): Promise<Budget | null> {
  try {
    return await fetchJSON<Budget>(
      `${registryUrl()}/v1/agents/${encodeURIComponent(agentId)}/budget?version=${encodeURIComponent(version)}`,
    );
  } catch {
    return null;
  }
}

export async function listDependencies(agentId: string, version: string): Promise<Dependency[]> {
  try {
    return await fetchJSON<Dependency[]>(
      `${registryUrl()}/v1/agents/${encodeURIComponent(agentId)}/dependencies?version=${encodeURIComponent(version)}`,
    );
  } catch {
    return [];
  }
}

export async function getUsageSummary(agentId?: string): Promise<UsageSummary> {
  const params = new URLSearchParams({ period: "monthly" });
  if (agentId) {
    params.set("agent_id", agentId);
  }
  return fetchJSON<UsageSummary>(`${usageUrl()}/v1/usage-events/summary?${params}`);
}

export async function getPlatformOverview() {
  const [agents, usage] = await Promise.all([listAgents(), getUsageSummary()]);
  return {
    agentCount: agents.total,
    monthlyCostUSD: usage.cost_usd,
    monthlyEvents: usage.event_count,
    cacheHitRate: usage.cache_hit_rate,
    agents: agents.items,
  };
}
