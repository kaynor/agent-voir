"""Tests for git identity switching during issue processing."""

from __future__ import annotations

from agentvoir_local_bot.config import Settings
from agentvoir_local_bot.git_ops import GitOps


def test_temporary_bot_identity_restores_previous_name(tmp_path, monkeypatch):
    """Bot identity is applied for the block and reverted afterward."""
    repo = tmp_path / "repo"
    repo.mkdir()
    monkeypatch.chdir(repo)

    for key in ("GIT_USER_NAME", "GIT_USER_EMAIL"):
        monkeypatch.delenv(key, raising=False)

    git = GitOps(
        Settings(
            _env_file=None,
            github_token="x",
            github_repo="acme/repo",
            cursor_api_key="y",
            repo_path=repo,
            git_user_name="kailash-coder-bot-01",
        )
    )
    git.run(["git", "init"])
    git.run(["git", "config", "--local", "user.name", "Kailash Aynor"])
    git.run(["git", "config", "--local", "user.email", "kailash@example.com"])

    with git.temporary_bot_identity():
        during = git._get_local_config("user.name")
        assert during == "kailash-coder-bot-01"

    after = git._get_local_config("user.name")
    assert after == "Kailash Aynor"
