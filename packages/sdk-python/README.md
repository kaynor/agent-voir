# AgentVoir Python SDK

Python client libraries for the AgentVoir registry API, LLM gateway, and usage ingestion service.

## Install

```bash
cd packages/sdk-python
pip install -e ".[dev]"
```

## Registry API

```python
from agentvoir import AgentVoirClient, RegisterAgentRequest

client = AgentVoirClient(base_url="http://localhost:8081")

print(client.health())
print(client.list_agents())

agent = client.register_agent(
    RegisterAgentRequest(
        agent_id="customer-support-agent",
        name="Customer Support Agent",
        version="0.1.0",
        owner_team="support-platform",
        environment="staging",
    )
)
print(agent.agent_id)
```

## LLM gateway

```python
from agentvoir import ChatCompletionRequest, ChatMessage, GatewayClient

gateway = GatewayClient(
    base_url="http://localhost:8080",
    api_key="agentvoir-local-dev-key",
    agent_id="customer-support-agent",
    tenant_id="acme",
)

response = gateway.chat_completions(
    ChatCompletionRequest(
        model="gpt-4.1-mini",
        messages=[ChatMessage(role="user", content="Summarize this ticket.")],
    )
)
print(response["choices"][0]["message"]["content"])
```

## Usage ingestion

```python
from agentvoir import UsageClient, UsageEventRequest

usage = UsageClient(base_url="http://localhost:8082")
usage.ingest_event(
    UsageEventRequest(
        agent_id="customer-support-agent",
        provider="openai",
        model="gpt-4.1-mini",
        cache_status="miss",
        prompt_tokens=120,
        completion_tokens=45,
    )
)
print(usage.list_events(agent_id="customer-support-agent"))
```

## Test

```bash
pytest
```
