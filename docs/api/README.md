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

Open **http://localhost:8089/** — Swagger UI with all three specs (not `/healthz`).

### Swagger “Try it out” with different API URLs

Swagger UI runs on **port 8089**. Each AgentVoir API runs on its **own port** (8080 / 8081 / 8082). Two things make Try it out work:

1. **Pick the right spec** — use the dropdown at the top (Registry API / Gateway API / Token accounting).
2. **Pick the right server** — under each operation, open **Servers** and choose the URL that matches your stack, e.g. `http://localhost:8081` for Registry.

Each OpenAPI file defines multiple `servers` entries (localhost, 127.0.0.1, alternate ports). Edit `docs/api/*.yaml` to add staging/production URLs.

**Gateway auth:** click **Authorize**, enter `agentvoir-onebox-key` (or your `GATEWAY_API_KEY`), and add header `x-agent-id` on chat requests.

**CORS / “URL scheme must be http or https” error**

This usually means Swagger could not resolve a valid server URL. Fix:

1. Recreate the docs container after spec changes:
   ```bash
   docker compose -f deployments/docker/docker-compose.yml --profile docs up -d --force-recreate api-docs
   ```
2. Hard-refresh the browser (`Ctrl+Shift+R`).
3. Pick a spec from the **top dropdown** (Registry / Gateway / Token accounting).
4. Under **Servers**, select a full URL such as `http://localhost:8081` — not empty, not `/`.
5. Ensure the target API is running (`./scripts/onebox-smoke.sh`).

Specs use OpenAPI **3.0.3** for Swagger UI compatibility. Each spec lists `http://localhost:…` server entries.

**CORS:** Browsers block cross-origin calls from `:8089` to `:8080` unless APIs allow it. Set on app containers:

```bash
CORS_ALLOWED_ORIGINS=http://localhost:8089,http://127.0.0.1:8089
```

This is already in Docker Compose for onebox/dev. **Rebuild app images** (`make onebox-up-build` or a new GHCR release) after pulling CORS support.

Restart after spec changes:

```bash
docker compose -f deployments/docker/docker-compose.yml --profile docs up -d --force-recreate api-docs
```

Ensure onebox/dev app stack is running on 8080–8082 when using Try it out.

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
