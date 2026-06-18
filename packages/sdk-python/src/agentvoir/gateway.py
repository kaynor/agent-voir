from __future__ import annotations

from typing import Any

import httpx

from .errors import AgentVoirError
from .types import ChatCompletionRequest, JsonObject, UsageEventRequest

SDK_VERSION = "0.1.0"


class GatewayClient:
    """OpenAI-compatible client for the AgentVoir LLM gateway."""

    def __init__(
        self,
        base_url: str,
        api_key: str,
        *,
        agent_id: str,
        agent_version: str = "0.1.0",
        tenant_id: str = "default",
        user_id: str = "",
        timeout: float = 60.0,
    ) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.agent_id = agent_id
        self.agent_version = agent_version
        self.tenant_id = tenant_id
        self.user_id = user_id
        self.timeout = timeout

    def _headers(self) -> dict[str, str]:
        headers = {
            "authorization": f"Bearer {self.api_key}",
            "content-type": "application/json",
            "x-agent-id": self.agent_id,
            "x-agent-version": self.agent_version,
            "x-tenant-id": self.tenant_id,
            "user-agent": f"agentvoir-python-sdk/{SDK_VERSION}",
        }
        if self.user_id:
            headers["x-user-id"] = self.user_id
        return headers

    def chat_completions(self, request: ChatCompletionRequest) -> JsonObject:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.post(
                f"{self.base_url}/v1/chat/completions",
                headers=self._headers(),
                json=request.model_dump(exclude_none=True),
            )
        if response.status_code >= 400:
            raise AgentVoirError(response.text.strip(), response.status_code)
        return response.json()

    def list_models(self) -> JsonObject:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.get(
                f"{self.base_url}/v1/models",
                headers=self._headers(),
            )
        if response.status_code >= 400:
            raise AgentVoirError(response.text.strip(), response.status_code)
        return response.json()


class UsageClient:
    """Client for the AgentVoir token-accounting / usage ingestion API."""

    def __init__(self, base_url: str, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def ingest_event(self, request: UsageEventRequest) -> JsonObject:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.post(
                f"{self.base_url}/v1/usage-events",
                headers={"content-type": "application/json"},
                json=request.model_dump(exclude_none=True),
            )
        if response.status_code >= 400:
            raise AgentVoirError(response.text.strip(), response.status_code)
        return response.json()

    def list_events(
        self,
        *,
        agent_id: str | None = None,
        tenant_id: str | None = None,
        limit: int = 100,
    ) -> list[JsonObject]:
        params: dict[str, Any] = {"limit": limit}
        if agent_id:
            params["agent_id"] = agent_id
        if tenant_id:
            params["tenant_id"] = tenant_id
        with httpx.Client(timeout=self.timeout) as client:
            response = client.get(
                f"{self.base_url}/v1/usage-events",
                params=params,
            )
        if response.status_code >= 400:
            raise AgentVoirError(response.text.strip(), response.status_code)
        return response.json()
