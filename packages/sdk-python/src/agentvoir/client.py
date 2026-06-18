from __future__ import annotations

from typing import Any

import httpx

from .errors import AgentVoirError
from .types import (
    Agent,
    CreateDependencyRequest,
    HealthResponse,
    JsonObject,
    Prompt,
    RegisterAgentRequest,
    RegisterPromptRequest,
    UpsertBudgetRequest,
    UpsertModelRouteRequest,
)

SDK_VERSION = "0.1.0"


class AgentVoirClient:
    """Client for the AgentVoir registry API."""

    def __init__(self, base_url: str, api_key: str | None = None, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout

    def _headers(self) -> dict[str, str]:
        headers = {"user-agent": f"agentvoir-python-sdk/{SDK_VERSION}"}
        if self.api_key:
            headers["authorization"] = f"Bearer {self.api_key}"
        return headers

    def _request(
        self,
        method: str,
        path: str,
        *,
        params: dict[str, Any] | None = None,
        json: dict[str, Any] | None = None,
    ) -> Any:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.request(
                method,
                f"{self.base_url}{path}",
                headers=self._headers(),
                params=params,
                json=json,
            )
        if response.status_code >= 400:
            message = response.text.strip() or f"request failed with status {response.status_code}"
            raise AgentVoirError(message, response.status_code)
        if response.status_code == 204 or not response.content:
            return None
        return response.json()

    def health(self) -> HealthResponse:
        return HealthResponse.model_validate(self._request("GET", "/healthz"))

    def list_agents(self) -> list[Agent]:
        return [Agent.model_validate(item) for item in self._request("GET", "/v1/agents")]

    def get_agent(self, agent_id: str, *, version: str, environment: str = "dev") -> Agent:
        data = self._request(
            "GET",
            f"/v1/agents/{agent_id}",
            params={"version": version, "environment": environment},
        )
        return Agent.model_validate(data)

    def register_agent(self, request: RegisterAgentRequest) -> Agent:
        data = self._request("POST", "/v1/agents", json=request.model_dump())
        return Agent.model_validate(data)

    def update_agent(
        self,
        agent_id: str,
        *,
        version: str,
        environment: str = "dev",
        **fields: Any,
    ) -> Agent:
        data = self._request(
            "PUT",
            f"/v1/agents/{agent_id}",
            params={"version": version, "environment": environment},
            json=fields,
        )
        return Agent.model_validate(data)

    def delete_agent(self, agent_id: str, *, version: str, environment: str = "dev") -> None:
        self._request(
            "DELETE",
            f"/v1/agents/{agent_id}",
            params={"version": version, "environment": environment},
        )

    def list_prompts(self) -> list[Prompt]:
        return [Prompt.model_validate(item) for item in self._request("GET", "/v1/prompts")]

    def register_prompt(self, request: RegisterPromptRequest) -> Prompt:
        data = self._request("POST", "/v1/prompts", json=request.model_dump())
        return Prompt.model_validate(data)

    def list_dependencies(self, agent_id: str, *, version: str) -> list[JsonObject]:
        return self._request(
            "GET",
            f"/v1/agents/{agent_id}/dependencies",
            params={"version": version},
        )

    def create_dependency(
        self,
        agent_id: str,
        *,
        version: str,
        request: CreateDependencyRequest,
    ) -> JsonObject:
        return self._request(
            "POST",
            f"/v1/agents/{agent_id}/dependencies",
            params={"version": version},
            json=request.model_dump(),
        )

    def get_dependency_graph(self, agent_id: str, *, version: str) -> JsonObject:
        return self._request(
            "GET",
            f"/v1/agents/{agent_id}/dependency-graph",
            params={"version": version},
        )

    def get_budget(self, agent_id: str, *, version: str) -> JsonObject:
        return self._request(
            "GET",
            f"/v1/agents/{agent_id}/budget",
            params={"version": version},
        )

    def upsert_budget(
        self,
        agent_id: str,
        *,
        version: str,
        request: UpsertBudgetRequest,
    ) -> JsonObject:
        return self._request(
            "PUT",
            f"/v1/agents/{agent_id}/budget",
            params={"version": version},
            json=request.model_dump(),
        )

    def get_model_route(self, agent_id: str, *, version: str) -> JsonObject:
        return self._request(
            "GET",
            f"/v1/agents/{agent_id}/model-route",
            params={"version": version},
        )

    def upsert_model_route(
        self,
        agent_id: str,
        *,
        version: str,
        request: UpsertModelRouteRequest,
    ) -> JsonObject:
        return self._request(
            "PUT",
            f"/v1/agents/{agent_id}/model-route",
            params={"version": version},
            json=request.model_dump(),
        )

    def parse_manifest(self, manifest_yaml: str) -> JsonObject:
        return self._request_raw("POST", "/v1/agents/parse-manifest", content=manifest_yaml)

    def register_from_manifest(self, manifest_yaml: str) -> JsonObject:
        return self._request_raw("POST", "/v1/agents/from-manifest", content=manifest_yaml)

    def _request_raw(self, method: str, path: str, *, content: str) -> Any:
        headers = self._headers()
        headers["content-type"] = "application/yaml"
        with httpx.Client(timeout=self.timeout) as client:
            response = client.request(
                method,
                f"{self.base_url}{path}",
                headers=headers,
                content=content.encode(),
            )
        if response.status_code >= 400:
            message = response.text.strip() or f"request failed with status {response.status_code}"
            raise AgentVoirError(message, response.status_code)
        return response.json()
