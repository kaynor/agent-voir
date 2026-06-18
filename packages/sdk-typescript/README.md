# AgentVoir TypeScript SDK

TypeScript client libraries for the AgentVoir registry API, LLM gateway, and usage ingestion service.

## Install

```bash
cd packages/sdk-typescript
npm install
npm run build
```

## Registry API

```typescript
import { AgentVoirClient } from "@agentvoir/sdk";

const client = new AgentVoirClient({ baseUrl: "http://localhost:8081" });

console.log(await client.health());
console.log(await client.listAgents());

const agent = await client.registerAgent({
  agent_id: "customer-support-agent",
  name: "Customer Support Agent",
  version: "0.1.0",
  owner_team: "support-platform",
  environment: "staging",
});
console.log(agent.agent_id);
```

## LLM gateway

```typescript
import { GatewayClient } from "@agentvoir/sdk";

const gateway = new GatewayClient({
  baseUrl: "http://localhost:8080",
  apiKey: "agentvoir-local-dev-key",
  agentId: "customer-support-agent",
  tenantId: "acme",
});

const response = await gateway.chatCompletions({
  model: "gpt-4.1-mini",
  messages: [{ role: "user", content: "Summarize this ticket." }],
});
console.log(response.choices?.[0]?.message?.content);
```

## Usage ingestion

```typescript
import { UsageClient } from "@agentvoir/sdk";

const usage = new UsageClient({ baseUrl: "http://localhost:8082" });
await usage.ingestEvent({
  agent_id: "customer-support-agent",
  provider: "openai",
  model: "gpt-4.1-mini",
  cache_status: "miss",
  prompt_tokens: 120,
  completion_tokens: 45,
});
console.log(await usage.listEvents({ agentId: "customer-support-agent" }));
```

## Test

```bash
npm run build
npm test
```
