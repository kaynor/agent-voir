# ADR 0002: Onebox distribution via pre-built GHCR images

## Status

Accepted

## Context

End users should run AgentVoir with Docker only — no Make, no local Go toolchain, and no fragile image builds on first install.

## Decision

- Publish a single unified **`ghcr.io/<owner>/agent-voir:<tag>`** image (gateway + registry-api + token-accounting) to **GHCR** on each GitHub Release
- `docker-compose.onebox.yml` pulls the pre-built image by default
- Contributors use `docker-compose.onebox.build.yml` or `make onebox-up-build` to build `deployments/docker/Dockerfile` locally
- Install docs live in `deployments/docker/INSTALL.md`; demo in `scripts/quickstart.sh`

## Consequences

- Faster, more reliable first-run experience (one `docker pull` for AgentVoir app code)
- Maintainers must publish releases and keep the GHCR package public
- Source-only zip installs still work but require pull access to GHCR
