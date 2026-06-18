"""Main worker loop: poll GitHub, run agent, test, and open pull requests."""

from __future__ import annotations

import logging
import time

from agentvoir_local_bot.agent_runner import AgentRunner
from agentvoir_local_bot.config import Settings
from agentvoir_local_bot.git_ops import GitError, GitOps
from agentvoir_local_bot.github_client import GitHubClient, GitHubIssue
from agentvoir_local_bot.test_runner import TestRunner

logger = logging.getLogger(__name__)


class IssueWorker:
    """Orchestrates the end-to-end ai-code issue pipeline."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._github = GitHubClient(settings)
        self._git = GitOps(settings)
        self._agent = AgentRunner(settings)
        self._tests = TestRunner(settings)

    def close(self) -> None:
        """Release GitHub client resources."""
        self._github.close()

    def __enter__(self) -> IssueWorker:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()

    def poll_once(self) -> bool:
        """Process at most one issue. Returns True if an issue was selected."""
        issues = self._github.list_candidate_issues()
        if not issues:
            logger.debug("No candidate issues found")
            return False

        issue = issues[0]
        logger.info("Selected issue #%s: %s", issue.number, issue.title)
        try:
            self._process_issue(issue)
        except Exception as err:
            logger.exception("Failed processing issue #%s", issue.number)
            self._github.mark_failed(issue.number, str(err))
            self._reset_to_base()
        return True

    def run_forever(self) -> None:
        """Poll continuously, sleeping between idle cycles."""
        logger.info(
            "Worker %s polling %s every %ss for label %s",
            self._settings.worker_id,
            self._settings.github_repo,
            self._settings.poll_interval_seconds,
            self._settings.label_trigger,
        )
        while True:
            worked = self.poll_once()
            if not worked:
                time.sleep(self._settings.poll_interval_seconds)

    def _process_issue(self, issue: GitHubIssue) -> None:
        """Execute the full claim → branch → agent → test → PR pipeline for one issue."""
        existing_pr = self._github.find_open_pr_for_branch(
            self._git.branch_name_for_issue(issue.number, issue.title)
        )
        if existing_pr:
            logger.info("Open PR already exists for issue #%s, skipping", issue.number)
            return

        self._github.claim_issue(issue.number)
        refreshed = self._github.get_issue(issue.number)
        if self._settings.label_claimed not in refreshed.labels:
            raise RuntimeError("Failed to claim issue — label not present after claim")

        branch_name = self._git.branch_name_for_issue(issue.number, issue.title)
        logger.info("Creating branch %s", branch_name)
        self._git.create_branch(branch_name)

        self._agent.run_for_issue(issue)

        test_result = self._tests.run()
        if not test_result.passed:
            # Keep only the tail of output to avoid huge GitHub comments.
            output = (test_result.stdout + test_result.stderr).strip()
            raise RuntimeError(f"Tests failed (exit {test_result.returncode}):\n{output[-4000:]}")

        commit_message = f"feat: resolve #{issue.number} — {issue.title}"
        self._git.commit_all(commit_message)

        logger.info("Pushing branch %s", branch_name)
        self._git.push_branch(branch_name)

        pr_body = self._build_pr_body(issue)
        pr = self._github.create_pull_request(
            title=f"#{issue.number}: {issue.title}",
            body=pr_body,
            head=branch_name,
            base=self._settings.base_branch,
        )
        pr_url = pr["html_url"]
        logger.info("Created pull request %s", pr_url)
        self._github.mark_done(issue.number, pr_url)
        self._reset_to_base()

    def _build_pr_body(self, issue: GitHubIssue) -> str:
        """Format the pull request description with issue context and close keyword."""
        return "\n".join(
            [
                f"Automated PR for #{issue.number}.",
                "",
                f"Closes #{issue.number}",
                "",
                "## Issue",
                issue.title,
                "",
                f"Original issue: {issue.html_url}",
                "",
                f"_Opened by AgentVoir local bot (`{self._settings.worker_id}`)._",
            ]
        )

    def _reset_to_base(self) -> None:
        """Return the local checkout to the base branch after success or failure."""
        try:
            self._git.ensure_clean_base()
        except GitError as err:
            logger.warning("Could not reset to base branch: %s", err)
