export interface AgentVoirClientOptions {
  baseUrl: string;
  apiKey?: string;
}

export class AgentVoirClient {
  private readonly baseUrl: string;
  private readonly apiKey?: string;

  constructor(options: AgentVoirClientOptions) {
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.apiKey = options.apiKey;
  }

  async health(): Promise<unknown> {
    const response = await fetch(`${this.baseUrl}/healthz`, {
      headers: this.headers()
    });
    if (!response.ok) {
      throw new Error(`AgentVoir health check failed: ${response.status}`);
    }
    return response.json();
  }

  async listAgents(): Promise<unknown[]> {
    const response = await fetch(`${this.baseUrl}/v1/agents`, {
      headers: this.headers()
    });
    if (!response.ok) {
      throw new Error(`AgentVoir listAgents failed: ${response.status}`);
    }
    return response.json() as Promise<unknown[]>;
  }

  private headers(): Record<string, string> {
    const headers: Record<string, string> = {
      "user-agent": "agentvoir-typescript-sdk/0.1.0"
    };
    if (this.apiKey) {
      headers.authorization = `Bearer ${this.apiKey}`;
    }
    return headers;
  }
}
