# Release checklist

Use this checklist when publishing AgentVoir container images and SDK packages.

## Pre-release

- [ ] All CI checks pass on `main`
- [ ] `docs/development-roadmap.md` reflects shipped scope
- [ ] OpenAPI specs updated for any API changes
- [ ] Migration scripts tested (`scripts/test-migrations.sh`)
- [ ] `./scripts/onebox-smoke.sh` and `./scripts/quickstart.sh` pass locally

## GitHub Release

1. Create a [GitHub Release](https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository) with a semver tag (for example `v0.1.0`).
2. The `Release container images` workflow builds and pushes a single image:
   - `ghcr.io/<owner>/agent-voir:<tag>`
4. The same workflow prepends a **Docker one-liner** to the GitHub Release page and uploads run assets:
   - `agentvoir-onebox-<tag>.zip` — compose, OPA policies, run scripts (~10 KB)
   - `run-agentvoir.sh` — one-command installer (`curl -fsSL .../run-agentvoir.sh | bash`)
5. Make the GHCR package **public** under **Package settings → Change visibility** so end users can pull without auth.
4. Attach release notes summarizing user-visible changes.

## Supply chain artifacts

Each release image workflow run produces:

- **SBOM** (Syft) uploaded as a workflow artifact
- **Vulnerability scan** (Trivy) — review HIGH/CRITICAL findings before promoting
- **Cosign signature** on the pushed image tag
- **SLSA provenance** attestation from Docker Buildx

See [deployments/docker/VERIFY.md](../deployments/docker/VERIFY.md) for verification commands.

## SDK publishing (optional)

Python and TypeScript SDKs are published manually or via `.github/workflows/publish-sdks.yml` on release:

```bash
cd packages/sdk-python && python -m build && twine upload dist/*
cd packages/sdk-typescript && npm publish --access public
```

Requires `PYPI_API_TOKEN` and `NPM_TOKEN` repository secrets.

## Post-release

- [ ] Verify images with cosign (see VERIFY.md)
- [ ] Run onebox pull + smoke test against the release tag
- [ ] Update README badges / install docs if the default tag changed
