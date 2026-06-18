from __future__ import annotations

from typing import Any

from pydantic import BaseModel, Field


class HealthResponse(BaseModel):
    service: str
    status: str
    time_utc: str


class RegisterAgentRequest(BaseModel):
    agent_id: str
    name: str
    version: str
    owner_team: str
    cost_center: str = ""
    environment: str = "dev"
    framework: str = ""
    risk_level: str = "low"
    lifecycle: str = "draft"
    data_classes: list[str] = Field(default_factory=list)


class Agent(RegisterAgentRequest):
    id: str
    created_at: str
    updated_at: str


class RegisterPromptRequest(BaseModel):
    prompt_id: str
    name: str
    version: str
    owner_team: str
    template: str
    risk_level: str = "low"
    approved_models: list[str] = Field(default_factory=list)


class Prompt(RegisterPromptRequest):
    id: str
    created_at: str
    updated_at: str


class CreateDependencyRequest(BaseModel):
    dependency_type: str
    dependency_name: str
    dependency_version: str = ""
    required: bool = True


class UpsertBudgetRequest(BaseModel):
    monthly_usd: float = 0
    max_prompt_tokens_per_request: int = 0
    max_completion_tokens_per_request: int = 0


class UpsertModelRouteRequest(BaseModel):
    primary_provider: str
    primary_model: str
    fallback_provider: str = ""
    fallback_model: str = ""
    routing_policy: str = "primary_then_fallback"


class ChatMessage(BaseModel):
    role: str
    content: str


class ChatCompletionRequest(BaseModel):
    model: str
    messages: list[ChatMessage]
    temperature: float | None = None
    stream: bool = False


class UsageEventRequest(BaseModel):
    agent_id: str
    trace_id: str = ""
    tenant_id: str = "default"
    agent_version: str = "0.1.0"
    user_id: str = ""
    provider: str = ""
    model: str = ""
    cache_status: str = ""
    prompt_tokens: int = 0
    completion_tokens: int = 0
    cached_tokens: int = 0
    cost_usd: float = 0
    latency_ms: int = 0
    status_code: int = 200
    error_code: str = ""


JsonObject = dict[str, Any]
