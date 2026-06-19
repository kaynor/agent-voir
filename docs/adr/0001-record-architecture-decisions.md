# ADR 0001: Record architecture decisions

## Status

Accepted

## Context

AgentVoir is an evolving enterprise control plane. We need a lightweight way to document significant technical decisions without bloating the README.

## Decision

We use **Architecture Decision Records (ADRs)** in `docs/adr/`:

- One markdown file per decision
- Numbered sequentially (`0001`, `0002`, …)
- Short sections: Status, Context, Decision, Consequences

## Consequences

- Contributors can understand *why* choices were made
- ADRs are immutable once accepted; supersede with a new ADR rather than editing history
- Roadmap items may reference ADRs when implementation depends on a settled design

## Template

```markdown
# ADR NNNN: Title

## Status
Proposed | Accepted | Deprecated | Superseded by ADR-XXXX

## Context
What problem are we solving?

## Decision
What did we choose?

## Consequences
What becomes easier or harder?
```
