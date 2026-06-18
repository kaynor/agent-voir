import json

import httpx
import pytest

from agentvoir import (
    AgentVoirClient,
    AgentVoirError,
    ChatCompletionRequest,
    ChatMessage,
    RegisterAgentRequest,
    UsageEventRequest,
)


def test_health() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.path == "/healthz"
        return httpx.Response(
            200,
            json={
                "service": "agentvoir-registry-api",
                "status": "ok",
                "time_utc": "2026-06-18T20:00:00Z",
            },
        )

    client = AgentVoirClient("http://localhost:8081")
    original_request = client._request

    def mock_request(method: str, path: str, **kwargs):  # type: ignore[no-untyped-def]
        with httpx.Client(
            transport=httpx.MockTransport(handler),
            base_url=client.base_url,
        ) as http:
            response = http.request(method, path, **kwargs)
            if response.status_code >= 400:
                raise AgentVoirError(response.text, response.status_code)
            return response.json()

    client._request = mock_request  # type: ignore[method-assign]
    health = client.health()
    assert health.status == "ok"
    client._request = original_request  # type: ignore[method-assign]


def test_list_agents_transport() -> None:
    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(
            200,
            json=[
                {
                    "id": "1",
                    "agent_id": "support-agent",
                    "name": "Support Agent",
                    "version": "0.1.0",
                    "owner_team": "support",
                    "environment": "staging",
                    "risk_level": "low",
                    "lifecycle": "draft",
                    "data_classes": [],
                    "created_at": "2026-06-18T20:00:00Z",
                    "updated_at": "2026-06-18T20:00:00Z",
                }
            ],
        )

    with httpx.Client(transport=httpx.MockTransport(handler), base_url="http://localhost:8081") as http:
        response = http.get("/v1/agents")
        agents = response.json()

    assert agents[0]["agent_id"] == "support-agent"


def test_register_agent_request_model() -> None:
    request = RegisterAgentRequest(
        agent_id="support-agent",
        name="Support Agent",
        version="0.1.0",
        owner_team="support",
        environment="staging",
    )
    assert request.environment == "staging"


def test_gateway_chat_request_model() -> None:
    request = ChatCompletionRequest(
        model="gpt-4.1-mini",
        messages=[ChatMessage(role="user", content="Hello")],
    )
    assert request.messages[0].content == "Hello"


def test_usage_event_request_defaults() -> None:
    event = UsageEventRequest(agent_id="support-agent", model="gpt-4.1-mini")
    assert event.tenant_id == "default"
    assert event.status_code == 200


def test_agentvoir_error_status_code() -> None:
    error = AgentVoirError("bad request", 400)
    assert error.status_code == 400


def test_register_agent_roundtrip_json() -> None:
    payload = RegisterAgentRequest(
        agent_id="support-agent",
        name="Support Agent",
        version="0.1.0",
        owner_team="support",
    ).model_dump()
    assert json.loads(json.dumps(payload))["agent_id"] == "support-agent"
