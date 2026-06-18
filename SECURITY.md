# Security Policy

AgentVoir is intended for enterprise AI governance and may process sensitive metadata, prompts, completions, and usage records.

## Reporting vulnerabilities

Please do not create public GitHub issues for security vulnerabilities.

Until a dedicated security contact is configured, use a private maintainer channel for disclosure.

## Security expectations

- Never log raw secrets.
- Avoid caching sensitive requests by default.
- Keep tenant data isolated.
- Encrypt sensitive cache entries and logs where supported.
- Treat prompts, completions, tool outputs, and RAG context as potentially sensitive.
- Require policy checks before model and tool access.
