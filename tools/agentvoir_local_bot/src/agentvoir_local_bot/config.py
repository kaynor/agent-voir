"""Environment-backed configuration for the local bot worker."""

from __future__ import annotations

from pathlib import Path

from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


def _default_repo_path() -> Path:
    """Resolve the monorepo root from this package location (tools/agentvoir_local_bot)."""
    return Path(__file__).resolve().parents[4]


class Settings(BaseSettings):
    """Worker settings loaded from environment variables and optional `.env` file."""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
        populate_by_name=True,
    )

    github_token: str = Field(validation_alias="GITHUB_TOKEN")
    github_repo: str = Field(validation_alias="GITHUB_REPO")
    github_api_url: str = Field(default="https://api.github.com", validation_alias="GITHUB_API_URL")

    label_trigger: str = Field(default="ai-code", validation_alias="LABEL_TRIGGER")
    label_claimed: str = Field(default="ai-code-claimed", validation_alias="LABEL_CLAIMED")
    label_done: str = Field(default="ai-code-done", validation_alias="LABEL_DONE")
    label_failed: str = Field(default="ai-code-failed", validation_alias="LABEL_FAILED")

    poll_interval_seconds: int = Field(default=60, ge=5, validation_alias="POLL_INTERVAL_SECONDS")
    worker_id: str = Field(default="local-bot-1", validation_alias="WORKER_ID")

    repo_path: Path = Field(default_factory=_default_repo_path, validation_alias="REPO_PATH")
    base_branch: str = Field(default="main", validation_alias="BASE_BRANCH")
    branch_prefix: str = Field(default="ai/issue", validation_alias="BRANCH_PREFIX")

    cursor_api_key: str = Field(validation_alias="CURSOR_API_KEY")
    agent_model: str = Field(default="composer-2.5", validation_alias="AGENT_MODEL")

    test_command: str = Field(default="make test", validation_alias="TEST_COMMAND")
    git_push_remote: str = Field(default="origin", validation_alias="GIT_PUSH_REMOTE")
    git_user_name: str | None = Field(default=None, validation_alias="GIT_USER_NAME")
    git_user_email: str | None = Field(default=None, validation_alias="GIT_USER_EMAIL")

    @field_validator("repo_path", mode="before")
    @classmethod
    def _coerce_repo_path(cls, value: str | Path | None) -> Path:
        """Fall back to the inferred repo root when REPO_PATH is unset or empty."""
        if value in (None, ""):
            return _default_repo_path()
        return Path(value).resolve()

    @property
    def github_owner(self) -> str:
        """Owner segment of `GITHUB_REPO` (e.g. `acme` from `acme/repo`)."""
        owner, _ = self.github_repo.split("/", maxsplit=1)
        return owner

    @property
    def github_repo_name(self) -> str:
        """Repository name segment of `GITHUB_REPO` (e.g. `repo` from `acme/repo`)."""
        _, name = self.github_repo.split("/", maxsplit=1)
        return name
