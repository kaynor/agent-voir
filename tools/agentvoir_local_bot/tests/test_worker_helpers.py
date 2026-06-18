"""Unit tests for branch naming, slugify, and issue prompt helpers."""

from __future__ import annotations

from agentvoir_local_bot.agent_runner import build_issue_prompt
from agentvoir_local_bot.github_client import GitHubIssue
from agentvoir_local_bot.git_ops import GitOps, _slugify


def test_slugify():
    """Slugify lowercases text, replaces non-alphanumerics, and falls back to 'task'."""
    assert _slugify("Add User Auth!!!") == "add-user-auth"
    assert _slugify("---") == "task"


def test_branch_name_for_issue():
    """Branch names combine prefix, issue number, and a slugified title."""
    from pathlib import Path

    from agentvoir_local_bot.config import Settings

    settings = Settings(
        _env_file=None,
        github_token="x",
        github_repo="acme/repo",
        cursor_api_key="y",
        repo_path=Path("."),
    )
    git = GitOps(settings)
    name = git.branch_name_for_issue(42, "Add retry logic")
    assert name == "ai/issue-42-add-retry-logic"


def test_build_issue_prompt_includes_title_and_body():
    """Prompts include issue metadata and instruct the agent not to commit."""
    issue = GitHubIssue(
        number=7,
        title="Fix cache TTL",
        body="The TTL should respect agent config.",
        html_url="https://github.com/acme/repo/issues/7",
        labels=("ai-code",),
    )
    prompt = build_issue_prompt(issue)
    assert "#7" in prompt
    assert "Fix cache TTL" in prompt
    assert "The TTL should respect agent config." in prompt
    assert "Do not commit" in prompt
