# AgentVoir Docker deployments

**End users:** see **[INSTALL.md](INSTALL.md)** for step-by-step onebox setup on a fresh machine.

## Onebox (end users / try-outs)

**Goal:** minimum overhead, no conflicts with existing Postgres/Redis containers.

```bash
make onebox-up
make onebox-smoke
```

Uses `deployments/docker/docker-compose.onebox.yml` with project name `agentvoir-onebox`.

| What | Onebox | Developer stack (`make dev-up-all`) |
| ---- | ------ | ----------------------------------- |
| Postgres on host `:5432` | No — internal only | Yes |
| Redis on host `:6379` | No — internal only | Yes |
| ClickHouse on host `:8123` | No — internal only | Yes |
| Grafana / Prometheus / OTel | Not included | Included |
| App ports `:8080-8082` | Yes (configurable) | Yes |
| Volume / network isolation | `agentvoir-onebox` project | `docker` default project |

Configure host ports and API key in `deployments/docker/.env.onebox` (copy from `.env.onebox.example`).

```bash
# If 8080 is already in use on your machine:
AGENTVOIR_GATEWAY_PORT=18080
AGENTVOIR_REGISTRY_PORT=18081
AGENTVOIR_USAGE_PORT=18082
```

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

A true all-in-one image (one container, many processes) is possible but harder to operate: mixed logs, coupled restarts, and heavier rebuilds. The onebox Compose file gives a similar **black-box experience** with one command while keeping each service independently healthy inside Docker.

For production, use separate services (Helm/Kubernetes) so gateway, registry, and analytics can scale independently.
