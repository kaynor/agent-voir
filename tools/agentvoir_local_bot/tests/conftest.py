"""Shared pytest fixtures for agentvoir_local_bot tests."""

from __future__ import annotations

import pytest

from agentvoir_local_bot.config import Settings

_ENV_KEYS = (
    "GITHUB_TOKEN",
    "GITHUB_REPO",
    "GITHUB_API_URL",
    "LABEL_TRIGGER",
    "LABEL_CLAIMED",
    "LABEL_DONE",
    "LABEL_FAILED",
    "CURSOR_API_KEY",
    "REPO_PATH",
)


@pytest.fixture
def settings(tmp_path, monkeypatch):
    """Settings isolated from the developer's .env and process environment."""
    for key in _ENV_KEYS:
        monkeypatch.delenv(key, raising=False)

    return Settings(
        _env_file=None,
        github_token="test-token",
        github_repo="acme/repo",
        cursor_api_key="cursor-key",
        repo_path=tmp_path,
    )
