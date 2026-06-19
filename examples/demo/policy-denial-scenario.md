# Policy denial demo scenario

This demo shows AgentVoir denying a gateway request when OPA policy rules are not satisfied.

## Scenario

A **draft** agent sends a request with `x-agent-environment: production`. OPA denies the call before any model provider is contacted.

## Automated demo

```bash
./scripts/onebox.sh
./scripts/seed-demo.sh          # register customer-support-agent if needed
./scripts/demo-policy-denial.sh # expect HTTP 403
```

## Manual curl

```bash
curl -sD - -o /tmp/out.json \
  -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-onebox-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -H "x-agent-version: 0.1.0" \
  -H "x-agent-environment: production" \
  --data-binary @examples/demo/sample-chat-request.json
```

## OPA policy

Rego rules live in `policies/opa/agentvoir.rego`. Draft agents are allowed in `dev` and `staging` environments; production environment requires production lifecycle.

## Direct OPA query

```bash
curl -s http://localhost:8181/v1/data/agentvoir/authz/allow \
  -H "Content-Type: application/json" \
  -d @examples/demo/policy-denial-input.json
```

## Related examples

- Agent manifest: `examples/agents/customer-support-agent.yaml`
- Agent policy manifest: `examples/policies/no-semantic-cache-for-pii.yaml`
