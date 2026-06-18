import { AgentVoirError } from "./errors.js";
import type { ChatCompletionRequest, JsonObject, UsageEventRequest } from "./types.js";

const SDK_VERSION = "0.1.0";

export interface GatewayClientOptions {
  baseUrl: string;
  apiKey: string;
  agentId: string;
  agentVersion?: string;
  tenantId?: string;
  userId?: string;
  fetchImpl?: typeof fetch;
}

export class GatewayClient {
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly agentId: string;
  private readonly agentVersion: string;
  private readonly tenantId: string;
  private readonly userId?: string;
  private readonly fetchImpl: typeof fetch;

  constructor(options: GatewayClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.apiKey = options.apiKey;
    this.agentId = options.agentId;
    this.agentVersion = options.agentVersion ?? "0.1.0";
    this.tenantId = options.tenantId ?? "default";
    this.userId = options.userId;
    this.fetchImpl = options.fetchImpl ?? fetch;
  }

  async chatCompletions(body: ChatCompletionRequest): Promise<JsonObject> {
    const response = await this.fetchImpl(`${this.baseUrl}/v1/chat/completions`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      throw new AgentVoirError(await response.text(), response.status);
    }
    return (await response.json()) as JsonObject;
  }

  async listModels(): Promise<JsonObject> {
    const response = await this.fetchImpl(`${this.baseUrl}/v1/models`, {
      headers: this.headers(false),
    });
    if (!response.ok) {
      throw new AgentVoirError(await response.text(), response.status);
    }
    return (await response.json()) as JsonObject;
  }

  private headers(includeContentType = true): Record<string, string> {
    const headers: Record<string, string> = {
      authorization: `Bearer ${this.apiKey}`,
      "x-agent-id": this.agentId,
      "x-agent-version": this.agentVersion,
      "x-tenant-id": this.tenantId,
      "user-agent": `agentvoir-typescript-sdk/${SDK_VERSION}`,
    };
    if (includeContentType) {
      headers["content-type"] = "application/json";
    }
    if (this.userId) {
      headers["x-user-id"] = this.userId;
    }
    return headers;
  }
}

export interface UsageClientOptions {
  baseUrl: string;
  fetchImpl?: typeof fetch;
}

export class UsageClient {
  private readonly baseUrl: string;
  private readonly fetchImpl: typeof fetch;

  constructor(options: UsageClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.fetchImpl = options.fetchImpl ?? fetch;
  }

  async ingestEvent(body: UsageEventRequest): Promise<JsonObject> {
    const response = await this.fetchImpl(`${this.baseUrl}/v1/usage-events`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!response.ok) {
      throw new AgentVoirError(await response.text(), response.status);
    }
    return (await response.json()) as JsonObject;
  }

  async listEvents(options: {
    agentId?: string;
    tenantId?: string;
    limit?: number;
  } = {}): Promise<JsonObject[]> {
    const url = new URL(`${this.baseUrl}/v1/usage-events`);
    if (options.agentId) url.searchParams.set("agent_id", options.agentId);
    if (options.tenantId) url.searchParams.set("tenant_id", options.tenantId);
    url.searchParams.set("limit", String(options.limit ?? 100));

    const response = await this.fetchImpl(url);
    if (!response.ok) {
      throw new AgentVoirError(await response.text(), response.status);
    }
    return (await response.json()) as JsonObject[];
  }
}
