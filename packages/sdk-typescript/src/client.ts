import { AgentVoirError } from "./errors.js";
import type {
  Agent,
  AgentQuery,
  ChatCompletionRequest,
  CreateDependencyRequest,
  HealthResponse,
  JsonObject,
  Prompt,
  RegisterAgentRequest,
  RegisterPromptRequest,
  UpsertBudgetRequest,
  UpsertModelRouteRequest,
  UsageEventRequest,
} from "./types.js";

const SDK_VERSION = "0.1.0";

export interface AgentVoirClientOptions {
  baseUrl: string;
  apiKey?: string;
  fetchImpl?: typeof fetch;
}

export class AgentVoirClient {
  private readonly baseUrl: string;
  private readonly apiKey?: string;
  private readonly fetchImpl: typeof fetch;

  constructor(options: AgentVoirClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.apiKey = options.apiKey;
    this.fetchImpl = options.fetchImpl ?? fetch;
  }

  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>("GET", "/healthz");
  }

  async listAgents(): Promise<Agent[]> {
    return this.request<Agent[]>("GET", "/v1/agents");
  }

  async getAgent(agentId: string, query: AgentQuery): Promise<Agent> {
    return this.request<Agent>("GET", `/v1/agents/${encodeURIComponent(agentId)}`, {
      query,
    });
  }

  async registerAgent(body: RegisterAgentRequest): Promise<Agent> {
    return this.request<Agent>("POST", "/v1/agents", { body });
  }

  async updateAgent(
    agentId: string,
    query: AgentQuery,
    body: Partial<RegisterAgentRequest>,
  ): Promise<Agent> {
    return this.request<Agent>("PUT", `/v1/agents/${encodeURIComponent(agentId)}`, {
      query,
      body,
    });
  }

  async deleteAgent(agentId: string, query: AgentQuery): Promise<void> {
    await this.request<void>("DELETE", `/v1/agents/${encodeURIComponent(agentId)}`, { query });
  }

  async listPrompts(): Promise<Prompt[]> {
    return this.request<Prompt[]>("GET", "/v1/prompts");
  }

  async registerPrompt(body: RegisterPromptRequest): Promise<Prompt> {
    return this.request<Prompt>("POST", "/v1/prompts", { body });
  }

  async listDependencies(agentId: string, version: string): Promise<JsonObject[]> {
    return this.request<JsonObject[]>(
      "GET",
      `/v1/agents/${encodeURIComponent(agentId)}/dependencies`,
      { query: { version } },
    );
  }

  async createDependency(
    agentId: string,
    version: string,
    body: CreateDependencyRequest,
  ): Promise<JsonObject> {
    return this.request<JsonObject>(
      "POST",
      `/v1/agents/${encodeURIComponent(agentId)}/dependencies`,
      { query: { version }, body },
    );
  }

  async getDependencyGraph(agentId: string, version: string): Promise<JsonObject> {
    return this.request<JsonObject>(
      "GET",
      `/v1/agents/${encodeURIComponent(agentId)}/dependency-graph`,
      { query: { version } },
    );
  }

  async getBudget(agentId: string, version: string): Promise<JsonObject> {
    return this.request<JsonObject>("GET", `/v1/agents/${encodeURIComponent(agentId)}/budget`, {
      query: { version },
    });
  }

  async upsertBudget(
    agentId: string,
    version: string,
    body: UpsertBudgetRequest,
  ): Promise<JsonObject> {
    return this.request<JsonObject>("PUT", `/v1/agents/${encodeURIComponent(agentId)}/budget`, {
      query: { version },
      body,
    });
  }

  async getModelRoute(agentId: string, version: string): Promise<JsonObject> {
    return this.request<JsonObject>(
      "GET",
      `/v1/agents/${encodeURIComponent(agentId)}/model-route`,
      { query: { version } },
    );
  }

  async upsertModelRoute(
    agentId: string,
    version: string,
    body: UpsertModelRouteRequest,
  ): Promise<JsonObject> {
    return this.request<JsonObject>(
      "PUT",
      `/v1/agents/${encodeURIComponent(agentId)}/model-route`,
      { query: { version }, body },
    );
  }

  async parseManifest(manifestYaml: string): Promise<JsonObject> {
    return this.requestRaw("POST", "/v1/agents/parse-manifest", manifestYaml);
  }

  async registerFromManifest(manifestYaml: string): Promise<JsonObject> {
    return this.requestRaw("POST", "/v1/agents/from-manifest", manifestYaml);
  }

  private headers(contentType = "application/json"): Record<string, string> {
    const headers: Record<string, string> = {
      "user-agent": `agentvoir-typescript-sdk/${SDK_VERSION}`,
    };
    if (this.apiKey) {
      headers.authorization = `Bearer ${this.apiKey}`;
    }
    if (contentType) {
      headers["content-type"] = contentType;
    }
    return headers;
  }

  private buildUrl(path: string, query?: Record<string, string | undefined>): string {
    const url = new URL(`${this.baseUrl}${path}`);
    if (query) {
      for (const [key, value] of Object.entries(query)) {
        if (value !== undefined) {
          url.searchParams.set(key, value);
        }
      }
    }
    return url.toString();
  }

  private async request<T>(
    method: string,
    path: string,
    options: {
      query?: Record<string, string | undefined>;
      body?: unknown;
    } = {},
  ): Promise<T> {
    const response = await this.fetchImpl(this.buildUrl(path, options.query), {
      method,
      headers: this.headers(),
      body: options.body ? JSON.stringify(options.body) : undefined,
    });
    return this.parseResponse<T>(response);
  }

  private async requestRaw(method: string, path: string, manifestYaml: string): Promise<JsonObject> {
    const response = await this.fetchImpl(this.buildUrl(path), {
      method,
      headers: this.headers("application/yaml"),
      body: manifestYaml,
    });
    return this.parseResponse<JsonObject>(response);
  }

  private async parseResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const message = (await response.text()).trim() || `request failed with status ${response.status}`;
      throw new AgentVoirError(message, response.status);
    }
    if (response.status === 204) {
      return undefined as T;
    }
    return (await response.json()) as T;
  }
}
