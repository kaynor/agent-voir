"""GitHub REST API client for issue polling, claiming, and pull request creation."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any
from urllib.parse import quote

import httpx

from agentvoir_local_bot.config import Settings


@dataclass(frozen=True)
class GitHubIssue:
    """Minimal issue snapshot used by the worker pipeline."""

    number: int
    title: str
    body: str | None
    html_url: str
    labels: tuple[str, ...]


class GitHubClient:
    """Thin wrapper around GitHub's REST API for the ai-code issue workflow."""

    def __init__(self, settings: Settings, client: httpx.Client | None = None) -> None:
        self._settings = settings
        # Track ownership so we only close clients we created (useful in tests).
        self._owns_client = client is None
        self._client = client or httpx.Client(
            base_url=settings.github_api_url.rstrip("/"),
            headers={
                "Authorization": f"Bearer {settings.github_token}",
                "Accept": "application/vnd.github+json",
                "X-GitHub-Api-Version": "2022-11-28",
                "User-Agent": "agentvoir-local-bot/0.1.0",
            },
            timeout=60.0,
        )

    def close(self) -> None:
        """Release the underlying HTTP client when owned by this instance."""
        if self._owns_client:
            self._client.close()

    def __enter__(self) -> GitHubClient:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()

    def _repo_path(self, suffix: str = "") -> str:
        """Build a repo-scoped API path, e.g. `/repos/owner/repo/issues`."""
        owner = self._settings.github_owner
        repo = self._settings.github_repo_name
        return f"/repos/{owner}/{repo}{suffix}"

    def list_candidate_issues(self) -> list[GitHubIssue]:
        """Return open trigger-labeled issues that are not claimed, done, or failed.

        Oldest issues are returned first. Pull requests are excluded because GitHub
        exposes them through the issues endpoint as well.
        """
        params = {
            "labels": self._settings.label_trigger,
            "state": "open",
            "per_page": 30,
            "sort": "created",
            "direction": "asc",
        }
        response = self._client.get(self._repo_path("/issues"), params=params)
        response.raise_for_status()
        skip_labels = {
            self._settings.label_claimed,
            self._settings.label_done,
            self._settings.label_failed,
        }
        issues: list[GitHubIssue] = []
        for item in response.json():
            if "pull_request" in item:
                continue
            labels = tuple(label["name"] for label in item.get("labels", []))
            if skip_labels.intersection(labels):
                continue
            issues.append(
                GitHubIssue(
                    number=item["number"],
                    title=item["title"],
                    body=item.get("body"),
                    html_url=item["html_url"],
                    labels=labels,
                )
            )
        return issues

    def get_issue(self, issue_number: int) -> GitHubIssue:
        """Fetch a single issue by number (used to verify claim succeeded)."""
        response = self._client.get(self._repo_path(f"/issues/{issue_number}"))
        response.raise_for_status()
        item = response.json()
        labels = tuple(label["name"] for label in item.get("labels", []))
        return GitHubIssue(
            number=item["number"],
            title=item["title"],
            body=item.get("body"),
            html_url=item["html_url"],
            labels=labels,
        )

    def claim_issue(self, issue_number: int) -> None:
        """Mark an issue in-progress: swap trigger label for claimed."""
        self._replace_labels(
            issue_number,
            remove=[self._settings.label_trigger],
            add=[self._settings.label_claimed],
        )
        self.add_comment(
            issue_number,
            (
                f"🤖 **AgentVoir local bot** (`{self._settings.worker_id}`) claimed this issue.\n\n"
                "Working on a branch locally; a pull request will follow when tests pass."
            ),
        )

    def mark_done(self, issue_number: int, pr_url: str) -> None:
        """Record successful completion and link the opened pull request."""
        self._replace_labels(
            issue_number,
            remove=[self._settings.label_claimed, self._settings.label_trigger],
            add=[self._settings.label_done],
        )
        self.add_comment(
            issue_number,
            f"✅ **AgentVoir local bot** opened a pull request: {pr_url}",
        )

    def mark_failed(self, issue_number: int, reason: str) -> None:
        """Record a failed run so the issue is not picked up again automatically."""
        self._replace_labels(
            issue_number,
            remove=[self._settings.label_claimed, self._settings.label_trigger],
            add=[self._settings.label_failed],
        )
        self.add_comment(
            issue_number,
            f"❌ **AgentVoir local bot** could not complete this issue.\n\n```\n{reason}\n```",
        )

    def _replace_labels(self, issue_number: int, *, remove: list[str], add: list[str]) -> None:
        """Remove workflow labels and apply the next-step label."""
        for label in remove:
            self.remove_label(issue_number, label)
        if add:
            self.add_labels(issue_number, add)

    def remove_label(self, issue_number: int, label: str) -> None:
        """Remove a single label from an issue (no-op if GitHub returns 404)."""
        encoded = quote(label, safe="")
        response = self._client.delete(
            self._repo_path(f"/issues/{issue_number}/labels/{encoded}"),
        )
        if response.status_code == 404:
            return
        response.raise_for_status()

    def add_labels(self, issue_number: int, labels: list[str]) -> None:
        """Append labels to an issue without removing existing ones."""
        response = self._client.post(
            self._repo_path(f"/issues/{issue_number}/labels"),
            json={"labels": labels},
        )
        response.raise_for_status()

    def add_comment(self, issue_number: int, body: str) -> None:
        """Post a comment on an issue."""
        response = self._client.post(
            self._repo_path(f"/issues/{issue_number}/comments"),
            json={"body": body},
        )
        response.raise_for_status()

    def create_pull_request(
        self,
        *,
        title: str,
        body: str,
        head: str,
        base: str,
    ) -> dict[str, Any]:
        """Open a pull request from a pushed branch. Returns the GitHub API response."""
        response = self._client.post(
            self._repo_path("/pulls"),
            json={"title": title, "body": body, "head": head, "base": base},
        )
        response.raise_for_status()
        return response.json()

    def find_open_pr_for_branch(self, branch_name: str) -> dict[str, Any] | None:
        """Return an open PR for `owner:branch_name`, or None if none exists."""
        response = self._client.get(
            self._repo_path("/pulls"),
            params={"state": "open", "head": f"{self._settings.github_owner}:{branch_name}"},
        )
        response.raise_for_status()
        pulls = response.json()
        return pulls[0] if pulls else None
