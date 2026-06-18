# AgentVoir onebox — Docker install guide

This guide is for **end users** who want to run AgentVoir locally with Docker. You do not need Go, Node.js, or a local database — only Docker.

## What you get

**Onebox** is a self-contained AgentVoir stack: one command starts everything you need to try the product.

| Exposed on your machine | URL (defaults) |
| ----------------------- | -------------- |
| Gateway (OpenAI-compatible API) | http://localhost:8080 |
| Registry API | http://localhost:8081 |
| Token accounting (usage events) | http://localhost:8082 |

Postgres, Redis, ClickHouse, and OPA run **inside Docker only**. They do not bind to `:5432`, `:6379`, or `:8123` on your host, so onebox will not conflict with databases you already run.

`docker ps` will show **7 containers** (all named `agentvoir-onebox-*`). That is expected — you interact with AgentVoir through the three URLs above, not by managing each container yourself.

---

## Prerequisites

Install these on your machine before starting:

1. **Docker Engine** with **Docker Compose v2** (`docker compose`, not legacy `docker-compose`)
   - [Docker Desktop](https://docs.docker.com/get-docker/) (macOS, Windows, Linux)
   - Or Docker Engine + Compose plugin on Linux
2. **Git** — to clone the repository
3. **Make** (optional) — shortcuts like `make onebox-up`; plain `docker compose` works without it
4. **curl** (optional) — for smoke tests below

Verify Docker is running:

```bash
docker info
docker compose version
```

---

## Install and start

### 1. Clone the repository

```bash
git clone https://github.com/your-org/agentvoir.git
cd agentvoir
```

Replace the clone URL with your fork or internal mirror if applicable.

### 2. (Optional) Configure ports and API key

On first start, a config file is created automatically from the example. To customize **before** starting:

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
```

Edit `deployments/docker/.env.onebox`:

```bash
# Change these if 8080/8081/8082 are already in use
AGENTVOIR_GATEWAY_PORT=8080
AGENTVOIR_REGISTRY_PORT=8081
AGENTVOIR_USAGE_PORT=8082

# API key for gateway clients (default is fine for local try-outs)
GATEWAY_API_KEY=agentvoir-onebox-key

# Optional: real OpenAI key if you want live model responses
# (mock provider works without this)
OPENAI_API_KEY=
```

### 3. Start onebox

**With Make (recommended):**

```bash
make onebox-up
```

**Without Make:**

```bash
cp -n deployments/docker/.env.onebox.example deployments/docker/.env.onebox
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml up -d --build
```

**Or use the helper script:**

```bash
./scripts/onebox.sh
```

The first run builds app images and may take several minutes. Later starts are much faster.

---

## Verify it works

Wait until services are healthy (usually 30–60 seconds on first boot), then run:

```bash
make onebox-smoke
```

Or check manually:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8081/healthz
curl http://localhost:8082/healthz
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer agentvoir-onebox-key"
```

All should succeed without connection errors.

---

## Use AgentVoir

### OpenAI-compatible clients

Point any OpenAI SDK or tool at the local gateway:

```bash
export OPENAI_BASE_URL="http://localhost:8080/v1"
export OPENAI_API_KEY="agentvoir-onebox-key"
```

If you changed ports or the API key in `.env.onebox`, use those values instead.

### Example chat completion

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer agentvoir-onebox-key" \
  -H "Content-Type: application/json" \
  -H "x-agent-id: customer-support-agent" \
  -d '{
    "model": "gpt-4.1-mini",
    "messages": [{"role": "user", "content": "Hello from AgentVoir onebox"}]
  }'
```

### Python SDK

```bash
pip install -e packages/sdk-python
python -c "
from agentvoir import GatewayClient
g = GatewayClient(base_url='http://localhost:8080', api_key='agentvoir-onebox-key')
print(g.list_models())
"
```

See [packages/sdk-python/README.md](../../packages/sdk-python/README.md) for full SDK docs.

---

## Day-to-day commands

| Action | Command |
| ------ | ------- |
| Start | `make onebox-up` |
| Stop (keep data) | `make onebox-down` |
| Follow logs | `make onebox-logs` |
| Stop and wipe all onebox data | `make onebox-reset` |
| Health checks | `make onebox-smoke` |

Without Make, replace `make onebox-*` with:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml <up|down|logs|...>
```

Add `-v` to `down` to remove volumes (same as `onebox-reset`).

---

## Troubleshooting

### Docker daemon not running

```text
Cannot connect to the Docker daemon
```

Start Docker Desktop, or on Linux:

```bash
sudo systemctl start docker
```

### Port already in use

```text
Bind for 0.0.0.0:8080 failed: port is already allocated
```

Edit `deployments/docker/.env.onebox` and pick free ports, for example:

```bash
AGENTVOIR_GATEWAY_PORT=18080
AGENTVOIR_REGISTRY_PORT=18081
AGENTVOIR_USAGE_PORT=18082
```

Then run `make onebox-up` again and use the new URLs.

### Services not ready yet

After `make onebox-up`, infra containers need time to pass health checks before apps start. If smoke tests fail, wait a minute and retry, or check logs:

```bash
make onebox-logs
```

### Build fails pulling Go modules (DNS / network)

Inside Docker builds, module downloads can fail on restricted networks. Ensure the host can reach the internet, or configure Docker DNS (see project maintainer docs). Retry:

```bash
make onebox-down
make onebox-up
```

### Wrong stack running

If you previously ran the **developer** stack (`make dev-up-all`), stop it before starting onebox:

```bash
make dev-down
make onebox-up
```

Onebox containers are named `agentvoir-onebox-*`. Developer stack containers are typically named `docker-*`.

### Start fresh

To reset all onebox data (Postgres, Redis, ClickHouse volumes):

```bash
make onebox-reset
make onebox-up
```

---

## Onebox vs developer Docker stack

| | Onebox (this guide) | Developer stack |
| -- | ------------------- | ----------------- |
| Command | `make onebox-up` | `make dev-up-all` |
| Best for | Trying AgentVoir, demos, SDK tests | Building AgentVoir locally |
| Postgres/Redis on host ports | No | Yes |
| Grafana / Prometheus | No | Yes |
| Requires Go / Node on host | No | Yes (for local app dev) |

For contributors hacking on Go services or the web console, see the main [README.md](../../README.md) developer setup section.

---

## Uninstall

Stop the stack and remove volumes:

```bash
make onebox-reset
```

Remove cloned source and built images if you no longer need them:

```bash
cd ..
rm -rf agentvoir
docker image prune -a   # optional: reclaim disk from built images
```
