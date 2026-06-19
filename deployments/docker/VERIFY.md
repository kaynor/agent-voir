# Verifying AgentVoir container images

AgentVoir release images are published to GitHub Container Registry (GHCR) and signed with [Sigstore cosign](https://docs.sigstore.dev/).

## Prerequisites

- [cosign](https://docs.sigstore.dev/cosign/installation/) v2+
- Docker or another OCI client

## Verify image signature

Replace `<owner>`, `<tag>`, and `<service>` (`gateway`, `registry-api`, or `token-accounting`):

```bash
export IMAGE="ghcr.io/<owner>/agent-voir/gateway:<tag>"
cosign verify "$IMAGE" \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  --certificate-identity-regexp='https://github.com/<owner>/agent-voir/.github/workflows/release-images.yml@refs/tags/.*'
```

A successful verification prints the signed payload digest.

## Inspect SBOM

Download the SBOM artifact from the GitHub Actions run for the release, or generate locally:

```bash
syft "$IMAGE" -o spdx-json > agentvoir-gateway.sbom.json
```

## Scan for vulnerabilities

```bash
trivy image --severity HIGH,CRITICAL "$IMAGE"
```

Review findings before deploying to production. Some base-image CVEs may require rebuilding on a newer distro tag.

## Runtime smoke test

After pulling release images:

```bash
export AGENTVOIR_IMAGE_TAG=<tag>
./scripts/onebox.sh up
./scripts/wait-for-onebox.sh
./scripts/onebox-smoke.sh
```

See [INSTALL.md](./INSTALL.md) for full setup instructions.
