# Security Policy

AgentVoir is intended for enterprise AI governance and may process sensitive metadata, prompts, completions, and usage records.

## Reporting vulnerabilities

Please do not create public GitHub issues for security vulnerabilities.

Until a dedicated security contact is configured, use a private maintainer channel for disclosure. Include:

- Affected component and version (image tag or git SHA)
- Reproduction steps or proof-of-concept
- Impact assessment (confidentiality, integrity, availability)

We aim to acknowledge reports within **5 business days** and provide a remediation timeline for confirmed issues.

## Supported releases

Security fixes are applied to the **latest semver release** and `main`. Older release tags are not supported unless explicitly noted in release notes.

Verify images before deployment — see [deployments/docker/VERIFY.md](deployments/docker/VERIFY.md).

## Release expectations

Maintainers follow [docs/RELEASE.md](docs/RELEASE.md):

- SBOM generation for each service image
- Cosign signatures and SLSA provenance attestations
- Trivy vulnerability scans (HIGH/CRITICAL reviewed before release)
- Dependency and license review in CI

## Security expectations

- Never log raw secrets or API keys.
- Avoid caching sensitive requests by default; use `X-Cache-Bypass: true` for sensitive calls.
- Keep tenant data isolated via `x-tenant-id`.
- Encrypt sensitive cache entries and logs where supported.
- Treat prompts, completions, tool outputs, and RAG context as potentially sensitive.
- Require policy checks before model and tool access (Phase 2+).

## Usage data retention

Configure usage event retention with `USAGE_RETENTION_DAYS` on the token-accounting service (default: 365). See `services/token-accounting/README.md`.
