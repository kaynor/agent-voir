# OIDC setup for AgentVoir

AgentVoir supports **OpenID Connect (OIDC)** JWT validation on the registry API and gateway, with optional **hybrid mode** that also accepts bootstrap API keys (`REGISTRY_API_KEY`, `GATEWAY_API_KEY`).

## Environment variables

| Variable | Service | Description |
| -------- | ------- | ----------- |
| `OIDC_ISSUER_URL` | registry, gateway | OIDC issuer URL (enables JWT validation) |
| `OIDC_CLIENT_ID` | registry, gateway | Expected JWT audience when `OIDC_AUDIENCE` is unset |
| `OIDC_AUDIENCE` | registry, gateway | Optional explicit `aud` claim |
| `OIDC_GROUPS_CLAIM` | registry, gateway | JWT claim for groups (default: `groups`) |
| `OIDC_TENANT_CLAIM` | registry, gateway | Optional claim mapped to `x-tenant-id` |
| `REGISTRY_API_KEY` | registry | Bootstrap static key (hybrid with JWT) |
| `GATEWAY_API_KEY` | gateway | Bootstrap static key (hybrid with JWT) |

When `OIDC_ISSUER_URL` and static keys are **unset**, behavior matches Phase 1:

- Registry API: open (no auth)
- Gateway: requires `GATEWAY_API_KEY` Bearer token (default in onebox)

When `OIDC_ISSUER_URL` is set, valid JWT access tokens are accepted. Identity claims are mapped to:

- `sub` → authenticated user (`x-user-id` on gateway requests)
- `email`, `groups` → available in request context for future RBAC
- optional tenant claim → `x-tenant-id`

## Local Dex (onebox)

Start onebox with the OIDC overlay:

```bash
cp deployments/docker/.env.onebox.example deployments/docker/.env.onebox
docker compose --env-file deployments/docker/.env.onebox \
  -f deployments/docker/docker-compose.onebox.yml \
  -f deployments/docker/docker-compose.onebox.oidc.yml up -d
```

Dex demo user (password grant):

- Email: `admin@agentvoir.local`
- Password: `password`
- Client ID: `agentvoir`
- Client secret: `agentvoir-dev-secret`

Run the demo:

```bash
chmod +x scripts/demo-oidc.sh
./scripts/demo-oidc.sh
```

## Production providers

Point `OIDC_ISSUER_URL` at your IdP discovery URL:

| Provider | Issuer URL pattern |
| -------- | ------------------ |
| Okta | `https://<tenant>.okta.com/oauth2/default` |
| Azure AD | `https://login.microsoftonline.com/<tenant-id>/v2.0` |
| Keycloak | `https://<host>/realms/<realm>` |
| Google | `https://accounts.google.com` |

Set `OIDC_CLIENT_ID` to your OAuth client ID. Use `OIDC_AUDIENCE` when the access token `aud` differs from the client ID (common with custom API scopes).

## Machine-to-machine

Use your IdP's **client credentials** grant to obtain access tokens for automated agents. Dex supports this locally:

```bash
curl -sS -X POST http://localhost:5556/dex/token \
  -d grant_type=client_credentials \
  -d client_id=agentvoir \
  -d client_secret=agentvoir-dev-secret \
  -d scope=openid
```

## Next steps (RBAC)

OIDC establishes **who** is calling. Phase 2 RBAC will map `groups` claims to roles and enforce permissions on registry mutations and sensitive gateway operations.

See [development-roadmap.md](../development-roadmap.md) for RBAC and audit logging tasks.
