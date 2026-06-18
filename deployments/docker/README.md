# AgentVoir Docker deployments

**End users:** see **[INSTALL.md](INSTALL.md)** — Docker only, pre-built images, no Make required.

## Onebox (end users / try-outs)

**Goal:** minimum friction — pull pre-built images, run `docker compose up`, no local compile.

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
./scripts/onebox.sh
./scripts/onebox-smoke.sh
```

Uses `deployments/docker/docker-compose.onebox.yml` with project name `agentvoir-onebox`.

App images (default):

- `ghcr.io/kaynor/agent-voir/gateway`
- `ghcr.io/kaynor/agent-voir/registry-api`
- `ghcr.io/kaynor/agent-voir/token-accounting`

Published on each [GitHub Release](https://github.com/kaynor/agent-voir/releases) via `.github/workflows/release-images.yml`.

| What | Onebox | Developer stack (`make dev-up-all`) |
| ---- | ------ | ----------------------------------- |
| Local image build | No — pulls from GHCR | Yes — `--build` |
| Requires Make | No | Optional |
| Postgres on host `:5432` | No — internal only | Yes |
| Redis on host `:6379` | No — internal only | Yes |
| ClickHouse on host `:8123` | No — internal only | Yes |
| Grafana / Prometheus / OTel | Not included | Included |
| App ports `:8080-8082` | Yes (configurable) | Yes |
| Volume / network isolation | `agentvoir-onebox` project | `docker` default project |

Configure image tag, host ports, and API key in `deployments/docker/.env.onebox` (copy from `.env.onebox.example`).

### Contributors: build locally

If GHCR images are unavailable or you are changing Go source:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml \
  -f deployments/docker/docker-compose.onebox.build.yml up -d --build
```

Or: `make onebox-up-build`

## Developer stack

Infrastructure only:

```bash
make dev-up
```

Full stack with apps (infra ports exposed for hybrid local development):

```bash
make dev-up-all
```

## Why not a single container?

A true all-in-one image (one container, many processes) is possible but harder to operate: mixed logs, coupled restarts, and heavier rebuilds. The onebox Compose file gives a similar **black-box experience** with standard Docker commands while keeping each service independently healthy inside Docker.

For production, use separate services (Helm/Kubernetes) so gateway, registry, and analytics can scale independently.
