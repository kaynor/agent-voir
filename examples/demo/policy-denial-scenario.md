# Policy denial demo scenario

This example shows how AgentVoir **would** deny a request when OPA policy integration is fully wired (Phase 2). Today the OPA container runs with example Rego policies; the gateway does not yet call OPA on every request.

## Scenario

An agent in **draft** lifecycle tries to call a provider that is not approved for production traffic.

## Setup

Agent manifest: `examples/agents/customer-support-agent.yaml` (lifecycle: `draft`)

OPA policy: `policies/opa/agentvoir.rego` — production agents must use allowed providers.

## Simulated policy input

```json
{
  "agent": {
    "lifecycle": "draft",
    "policies": {
      "allowedProviders": ["openai", "anthropic"],
      "piiAllowed": true
    },
    "cache": { "mode": "exact_only", "semanticCacheAllowed": false }
  },
  "request": {
    "provider": "openai",
    "contains_pii": false,
    "contains_secret": false
  },
  "environment": "production"
}
```

## Expected OPA result (today)

Query `allow` against the policy server:

```bash
curl -s http://localhost:8181/v1/data/agentvoir/authz/allow \
  -H "Content-Type: application/json" \
  -d @examples/demo/policy-denial-input.json
```

For a **draft** agent in **production** environment, `allow` is **false** unless lifecycle is `staging` and environment matches.

## Related examples

- Agent policy manifest: `examples/policies/no-semantic-cache-for-pii.yaml`
- Prompt: `examples/prompts/support-ticket-summary.yaml`

When gateway ↔ OPA integration lands (Phase 2), denied requests will return HTTP 403 with a structured error before the upstream model is called.
