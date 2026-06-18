import assert from "node:assert/strict";
import test from "node:test";

import { AgentVoirClient } from "../dist/client.js";
import { AgentVoirError } from "../dist/errors.js";
import { GatewayClient, UsageClient } from "../dist/gateway.js";

test("health returns registry status", async () => {
  const client = new AgentVoirClient({
    baseUrl: "http://localhost:8081",
    fetchImpl: async (input) => {
      assert.equal(String(input), "http://localhost:8081/healthz");
      return new Response(
        JSON.stringify({
          service: "agentvoir-registry-api",
          status: "ok",
          time_utc: "2026-06-18T20:00:00Z",
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      );
    },
  });

  const health = await client.health();
  assert.equal(health.status, "ok");
});

test("listAgents parses agent records", async () => {
  const client = new AgentVoirClient({
    baseUrl: "http://localhost:8081",
    fetchImpl: async () =>
      new Response(
        JSON.stringify([
          {
            id: "1",
            agent_id: "support-agent",
            name: "Support Agent",
            version: "0.1.0",
            owner_team: "support",
            environment: "staging",
            risk_level: "low",
            lifecycle: "draft",
            data_classes: [],
            created_at: "2026-06-18T20:00:00Z",
            updated_at: "2026-06-18T20:00:00Z",
          },
        ]),
        { status: 200, headers: { "content-type": "application/json" } },
      ),
  });

  const agents = await client.listAgents();
  assert.equal(agents[0]?.agent_id, "support-agent");
});

test("gateway chatCompletions sends agent headers", async () => {
  const gateway = new GatewayClient({
    baseUrl: "http://localhost:8080",
    apiKey: "agentvoir-local-dev-key",
    agentId: "customer-support-agent",
    tenantId: "acme",
    fetchImpl: async (_input, init) => {
      const headers = new Headers(init?.headers);
      assert.equal(headers.get("x-agent-id"), "customer-support-agent");
      return new Response(
        JSON.stringify({
          choices: [{ message: { role: "assistant", content: "ok" } }],
        }),
        { status: 200, headers: { "content-type": "application/json" } },
      );
    },
  });

  const response = await gateway.chatCompletions({
    model: "gpt-4.1-mini",
    messages: [{ role: "user", content: "Hello" }],
  });
  assert.equal(response.choices?.[0]?.message?.content, "ok");
});

test("usage client ingests events", async () => {
  const usage = new UsageClient({
    baseUrl: "http://localhost:8082",
    fetchImpl: async (input, init) => {
      assert.match(String(input), /\/v1\/usage-events$/);
      assert.equal(init?.method, "POST");
      return new Response(JSON.stringify({ agent_id: "support-agent" }), {
        status: 201,
        headers: { "content-type": "application/json" },
      });
    },
  });

  const event = await usage.ingestEvent({
    agent_id: "support-agent",
    model: "gpt-4.1-mini",
  });
  assert.equal(event.agent_id, "support-agent");
});

test("AgentVoirError exposes status code", () => {
  const error = new AgentVoirError("bad request", 400);
  assert.equal(error.statusCode, 400);
});
