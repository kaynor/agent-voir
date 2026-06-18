from __future__ import annotations

import httpx


class AgentVoirClient:
    def __init__(self, base_url: str, api_key: str | None = None, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout

    def _headers(self) -> dict[str, str]:
        headers = {"user-agent": "agentvoir-python-sdk/0.1.0"}
        if self.api_key:
            headers["authorization"] = f"Bearer {self.api_key}"
        return headers

    def health(self) -> dict:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.get(f"{self.base_url}/healthz", headers=self._headers())
            response.raise_for_status()
            return response.json()

    def list_agents(self) -> list[dict]:
        with httpx.Client(timeout=self.timeout) as client:
            response = client.get(f"{self.base_url}/v1/agents", headers=self._headers())
            response.raise_for_status()
            return response.json()
