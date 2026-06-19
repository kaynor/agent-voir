# AgentVoir API documentation

Interactive and reference documentation for AgentVoir HTTP APIs.

## OpenAPI specifications

| API | Spec file |
| --- | --------- |
| Registry (agents, prompts, budgets, manifests) | [openapi.yaml](openapi.yaml) |
| Gateway (OpenAI-compatible + AgentVoir headers) | [gateway-openapi.yaml](gateway-openapi.yaml) |
| Token accounting (usage events) | [token-accounting-openapi.yaml](token-accounting-openapi.yaml) |

## Local Swagger UI (developer stack)

Start the developer Docker stack with the `docs` profile:

```bash
docker compose -f deployments/docker/docker-compose.yml --profile docs up -d
```

Open **http://localhost:8089** — Swagger UI with all three specs.

## Examples

See [examples.md](examples.md) for curl and SDK snippets covering:

- Authentication
- Agent registration and manifest import
- Gateway chat completions and cache headers
- Usage event queries
- OPA policy simulation

## GitHub Pages

API specs and Redoc HTML are published on push to `main` via `.github/workflows/docs-pages.yml`.

After the workflow runs, browse: `https://kaynor.github.io/agent-voir/api/` (enable Pages under repo Settings → Pages if needed).

## SDK mapping

| API | Python | TypeScript |
| --- | ------ | ---------- |
| Registry | `AgentVoirClient` | `AgentVoirClient` |
| Gateway | `GatewayClient` | `GatewayClient` |
| Usage | HTTP / future client | HTTP / future client |

See [packages/sdk-python/README.md](../../packages/sdk-python/README.md) and [packages/sdk-typescript/README.md](../../packages/sdk-typescript/README.md).
