# AgentVoir onebox — Docker install guide

This guide is for **end users** who want to run AgentVoir locally. You need **Docker only** — no Go, Node.js, Make, or local builds.

Pre-built AgentVoir images are published to [GitHub Container Registry (GHCR)](https://github.com/kaynor/agent-voir/pkgs/container/agent-voir) when maintainers create a GitHub Release. Docker pulls one unified image; nothing is compiled on your machine.

## What you get

**Onebox** is a self-contained AgentVoir stack: a few Docker commands start everything you need to try the product.

| Exposed on your machine | URL (defaults) |
| ----------------------- | -------------- |
| Gateway (OpenAI-compatible API) | http://localhost:8080 |
| Registry API | http://localhost:8081 |
| Token accounting (usage events) | http://localhost:8082 |

Postgres, Redis, ClickHouse, and OPA run **inside Docker only**. They do not bind to `:5432`, `:6379`, or `:8123` on your host, so onebox will not conflict with databases you already run.

`docker ps` will show **5 containers** (all named `agentvoir-onebox-*`): Postgres, Redis, ClickHouse, OPA, and the unified AgentVoir app image. That is expected — you interact with AgentVoir through the three URLs above.

---

## Prerequisites

1. **Docker Engine** with **Docker Compose v2** (`docker compose`, not legacy `docker-compose`)
   - [Docker Desktop](https://docs.docker.com/get-docker/) (macOS, Windows, Linux)
   - Or Docker Engine + Compose plugin on Linux
2. **curl** (optional) — for smoke tests below

Verify Docker is running:

```bash
docker info
docker compose version
```

---

## Install and start

### Option A — GitHub Release zip (no Git required)

1. Open [GitHub Releases](https://github.com/kaynor/agent-voir/releases) and download the **Source code (zip)** for the version you want (e.g. `v1.0.0`).

2. Unzip and enter the folder:

```bash
unzip agent-voir-1.0.0.zip
cd agent-voir-1.0.0
```

3. Create config (optional — defaults work for a first try):

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
```

4. Pull images and start:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml pull

docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml up -d
```

**Or use the helper script** (same steps, no typing compose flags):

```bash
chmod +x scripts/onebox.sh
./scripts/onebox.sh
```

### Option B — Git clone

```bash
git clone https://github.com/kaynor/agent-voir.git
cd agent-voir
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
./scripts/onebox.sh
```

### Pin a specific release

Edit `deployments/docker/.env.onebox` before starting:

```bash
AGENTVOIR_IMAGE=ghcr.io/kaynor/agent-voir
AGENTVOIR_VERSION=v1.0.0
```

Use the same `docker compose pull` and `up -d` commands above.

---

## Configure (optional)

Edit `deployments/docker/.env.onebox`:

```bash
# Image version (match a GitHub Release tag, or use latest)
AGENTVOIR_IMAGE=ghcr.io/kaynor/agent-voir
AGENTVOIR_VERSION=latest

# Change these if 8080/8081/8082 are already in use
AGENTVOIR_GATEWAY_PORT=8080
AGENTVOIR_REGISTRY_PORT=8081
AGENTVOIR_USAGE_PORT=8082

# Gateway auth key for OpenAI-compatible clients
GATEWAY_API_KEY=agentvoir-onebox-key

# Optional: real OpenAI key for live model responses
# (mock provider works without this)
OPENAI_API_KEY=
```

After changing ports or version, run `docker compose ... pull` again if you changed `AGENTVOIR_VERSION`.

---

## Verify it works

Wait 30–60 seconds for services to become healthy, then:

```bash
After `up -d`, wait for services to become healthy:

```bash
./scripts/wait-for-onebox.sh
./scripts/onebox-smoke.sh
```

Optionally seed demo agents:

```bash
./scripts/seed-demo.sh
```

See [VERIFY.md](./VERIFY.md) for verifying signed release images.
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

```bash
export OPENAI_BASE_URL="http://localhost:8080/v1"
export OPENAI_API_KEY="agentvoir-onebox-key"
```

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

---

## Day-to-day commands (Docker only)

Set these once per shell session for shorter commands:

```bash
export ONEBOX="docker compose --env-file deployments/docker/.env.onebox -f deployments/docker/docker-compose.onebox.yml"
```

| Action | Command |
| ------ | ------- |
| Start | `$ONEBOX pull && $ONEBOX up -d` |
| Stop (keep data) | `$ONEBOX down` |
| Follow logs | `$ONEBOX logs -f` |
| Stop and wipe data | `$ONEBOX down -v` |
| Health checks | `./scripts/onebox-smoke.sh` |

---

## Troubleshooting

### Image pull fails / not found

Pre-built images are published when a maintainer creates a **GitHub Release**. If pull fails:

1. Confirm a release exists at [GitHub Releases](https://github.com/kaynor/agent-voir/releases).
2. Set `AGENTVOIR_VERSION` in `.env.onebox` to that release tag (e.g. `v1.0.0`).
3. Ensure GHCR packages are **public** (maintainer setting under GitHub → Packages).

**Contributors** building from source without published images:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml \
  -f deployments/docker/docker-compose.onebox.build.yml up -d --build
```

### Docker daemon not running

```text
Cannot connect to the Docker daemon
```

Start Docker Desktop, or on Linux: `sudo systemctl start docker`

### Port already in use

Edit `deployments/docker/.env.onebox`:

```bash
AGENTVOIR_GATEWAY_PORT=18080
AGENTVOIR_REGISTRY_PORT=18081
AGENTVOIR_USAGE_PORT=18082
```

Then `$ONEBOX up -d` again.

### Services not ready yet

Wait a minute and retry `./scripts/onebox-smoke.sh`, or inspect logs:

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml logs -f
```

### Wrong stack running

Stop the developer stack if you ran it earlier:

```bash
docker compose -f deployments/docker/docker-compose.yml --profile apps down
```

Onebox containers are named `agentvoir-onebox-*`.

---

## Onebox vs developer stack

| | Onebox (this guide) | Developer stack |
| -- | ------------------- | ----------------- |
| Start | `docker compose pull && up -d` | `make dev-up-all` (builds locally) |
| Best for | Trying AgentVoir, demos | Hacking on Go/Node source |
| Requires Make | No | Optional |
| Compiles on your machine | No | Yes |
| Grafana / Prometheus | No | Yes |

---

## Uninstall

```bash
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml down -v
```

Optional — remove downloaded images:

```bash
docker image rm ghcr.io/kaynor/agent-voir:latest
```

---

## For maintainers: publish images

Creating a GitHub Release (e.g. tag `v1.0.0`) triggers [.github/workflows/release-images.yml](../../.github/workflows/release-images.yml), which builds and pushes:

- `ghcr.io/kaynor/agent-voir:<tag>`

After the first publish, set the package to **Public** under GitHub → Packages so anonymous `docker pull` works for end users.

Manual publish (any tag):

```bash
# GitHub → Actions → Release container images → Run workflow
```
