export interface HealthResponse {
  service: string;
  status: string;
  time_utc: string;
}

export interface RegisterAgentRequest {
  agent_id: string;
  name: string;
  version: string;
  owner_team: string;
  cost_center?: string;
  environment?: string;
  framework?: string;
  risk_level?: string;
  lifecycle?: string;
  data_classes?: string[];
}

export interface Agent extends RegisterAgentRequest {
  id: string;
  created_at: string;
  updated_at: string;
}

export interface RegisterPromptRequest {
  prompt_id: string;
  name: string;
  version: string;
  owner_team: string;
  template: string;
  risk_level?: string;
  approved_models?: string[];
}

export interface Prompt extends RegisterPromptRequest {
  id: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDependencyRequest {
  dependency_type: string;
  dependency_name: string;
  dependency_version?: string;
  required?: boolean;
}

export interface UpsertBudgetRequest {
  monthly_usd?: number;
  max_prompt_tokens_per_request?: number;
  max_completion_tokens_per_request?: number;
}

export interface UpsertModelRouteRequest {
  primary_provider: string;
  primary_model: string;
  fallback_provider?: string;
  fallback_model?: string;
  routing_policy?: string;
}

export interface ChatMessage {
  role: string;
  content: string;
}

export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  stream?: boolean;
}

export interface UsageEventRequest {
  agent_id: string;
  trace_id?: string;
  tenant_id?: string;
  agent_version?: string;
  user_id?: string;
  provider?: string;
  model?: string;
  cache_status?: string;
  prompt_tokens?: number;
  completion_tokens?: number;
  cached_tokens?: number;
  cost_usd?: number;
  latency_ms?: number;
  status_code?: number;
  error_code?: string;
}

export type JsonObject = Record<string, unknown>;

export interface AgentQuery extends Record<string, string | undefined> {
  version: string;
  environment?: string;
}
