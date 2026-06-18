import { AgentVoirClient } from "@agentvoir/sdk";

const client = new AgentVoirClient({ baseUrl: "http://localhost:8081", apiKey: "dev" });
console.log(await client.health());
console.log(await client.listAgents());
