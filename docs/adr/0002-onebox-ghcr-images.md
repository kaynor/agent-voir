# ADR 0002: Onebox distribution via pre-built GHCR images

## Status

Accepted

## Context

End users should run AgentVoir with Docker only — no Make, no local Go toolchain, and no fragile image builds on first install.

## Decision

- Publish `gateway`, `registry-api`, and `token-accounting` images to **GHCR** on each GitHub Release
- `docker-compose.onebox.yml` pulls pre-built images by default
- Contributors use `docker-compose.onebox.build.yml` or `make onebox-up-build` for local builds
- Install docs live in `deployments/docker/INSTALL.md`; demo in `scripts/quickstart.sh`

## Consequences

- Faster, more reliable first-run experience
- Maintainers must publish releases and keep GHCR packages public
- Source-only zip installs still work but require pull access to GHCR
