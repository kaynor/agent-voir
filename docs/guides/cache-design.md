# Cache Design

AgentVoir supports exact cache first and semantic cache later.

## Exact cache key inputs

- tenant ID
- agent ID and version
- provider and model
- normalized messages hash
- system prompt hash
- tool schema hash
- tool choice
- temperature/top_p/max_tokens
- response format
- RAG context hash
- authorization context hash when answer visibility depends on permissions
- policy version
- prompt template version

## Unsafe cache scenarios

Do not cache by default when:

- request contains PII or secrets
- request asks for latest/current facts
- user permissions affect the answer
- tool outputs are missing from the key
- temperature is non-deterministic
- response is legal, medical, financial, HR, or compliance-sensitive
