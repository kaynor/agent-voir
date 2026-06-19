# PostgreSQL metadata schema

AgentVoir stores control-plane metadata in PostgreSQL. Migrations live in `db/migrations/postgres/`.

## Entity relationship diagram

```mermaid
erDiagram
    agents ||--o{ agent_dependencies : has
    agents ||--o| budgets : has
    agents ||--o| model_routes : routes
    agents ||--o{ cache_entries : caches

    agents {
        uuid id PK
        text agent_id
        text version
        text environment
        text owner_team
        text lifecycle
        text cache_mode
        bigint cache_ttl_seconds
        boolean semantic_cache_allowed
    }

    agent_dependencies {
        uuid id PK
        text agent_id FK
        text dependency_type
        text dependency_name
    }

    prompts {
        uuid id PK
        text prompt_id
        text version
        text template
    }

    model_routes {
        uuid id PK
        text agent_id FK
        text primary_provider
        text primary_model
    }

    budgets {
        uuid id PK
        text agent_id FK
        numeric monthly_usd
    }

    cache_entries {
        text cache_key PK
        text agent_id FK
        text model
        timestamptz expires_at
    }
```

## Tables

| Table | Purpose |
|-------|---------|
| `agents` | Registered enterprise agents with lifecycle and cache settings |
| `agent_dependencies` | Tools, APIs, vector stores, MCP servers, and agent dependencies |
| `prompts` | Versioned prompt templates |
| `model_routes` | Primary/fallback model routing per agent version |
| `budgets` | Monthly spend and per-request token limits |
| `cache_entries` | Optional metadata for cache entries (exact cache uses Redis in Phase 1) |

Usage events are stored in ClickHouse (`usage_events` table), not PostgreSQL.

Apply migrations with `make db-migrate` or automatically on registry-api startup when `POSTGRES_DSN` is set.

Seed demo agents with `./scripts/seed-demo.sh` after the onebox stack is running.
