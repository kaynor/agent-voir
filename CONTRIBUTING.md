# Contributing to AgentVoir

Thank you for your interest in contributing to AgentVoir.

## Development principles

- Security and correctness before cleverness.
- Enterprise defaults should be safe by default.
- Cache behavior must be explicit and auditable.
- All model/tool access should be attributable to an agent identity.
- Prefer open standards: OpenTelemetry, OpenAPI, OIDC, Rego, Kubernetes APIs.

## Local setup

```bash
cp .env.example .env
make dev-up
make run-api
make run-gateway
```

## Pull requests

Before opening a PR, run:

```bash
make fmt
make lint
make test
```

## Areas for contribution

- Gateway provider adapters
- Redis exact cache implementation
- Agent registry schema and APIs
- OpenTelemetry instrumentation
- OPA policies
- SDKs
- Admin console
- Documentation
