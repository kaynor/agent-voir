# AgentVoir onebox bundle

Minimal files to run AgentVoir with Docker **without cloning the monorepo**.

Published on each GitHub Release as:

- `agentvoir-onebox-<tag>.zip` — compose, OPA policies, run scripts
- `run-agentvoir.sh` — one-command installer

## End user: one command

Use the **exact release tag** in the URL (from [GitHub Releases](https://github.com/kaynor/agent-voir/releases)):

```bash
curl -fsSL https://github.com/kaynor/agent-voir/releases/download/v0.2.6/run-agentvoir.sh | bash
```

The downloaded script is pinned to that release — no `latest` redirect.

## End user: download zip

1. Open [GitHub Releases](https://github.com/kaynor/agent-voir/releases)
2. Download `agentvoir-onebox-vX.Y.Z.zip`
3. Unzip and run:

```bash
unzip agentvoir-onebox-v0.2.4.zip -d agentvoir-onebox
cd agentvoir-onebox
chmod +x onebox.sh onebox-smoke.sh
./onebox.sh
./onebox-smoke.sh
```

## What's in the bundle

| File | Purpose |
|------|---------|
| `docker-compose.yml` | Starts Postgres, Redis, ClickHouse, OPA + AgentVoir image |
| `policies/opa/` | OPA Rego policies (mounted into OPA container) |
| `.version` | Release tag this bundle matches |
| `onebox.sh` | Pull + start |
| `onebox-smoke.sh` | Health checks |

The **AgentVoir app** itself comes from GHCR (`ghcr.io/kaynor/agent-voir:<tag>`). This bundle is the "wheels" around it.

## Maintainers: pack locally

```bash
./scripts/pack-onebox-bundle.sh v0.2.4 ghcr.io/kaynor/agent-voir
ls dist/
```
