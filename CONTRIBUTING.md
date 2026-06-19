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

See [docs/AI_CONTRIBUTION_POLICY.md](docs/AI_CONTRIBUTION_POLICY.md) for AI-assisted contribution rules.

### GitHub labels and backlog issues

Maintainers can seed labels and starter issues (requires [GitHub CLI](https://cli.github.com/)):

```bash
./scripts/bootstrap-github-labels.sh
./scripts/bootstrap-github-issues.sh
```

Issue templates: bug report, feature request, and **AI coding task** under `.github/ISSUE_TEMPLATE/`.

## Areas for contribution

- Gateway provider adapters
- Redis exact cache implementation
- Agent registry schema and APIs
- OpenTelemetry instrumentation
- OPA policies
- SDKs
- Admin console
- Documentation
