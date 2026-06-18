"""Cursor SDK integration for running a local coding agent against an issue."""

from __future__ import annotations

import logging
import sys
from dataclasses import dataclass

from cursor_sdk import Agent, AgentOptions, Client, CursorAgentError, LocalAgentOptions

from agentvoir_local_bot.config import Settings
from agentvoir_local_bot.cursor_bridge import apply_windows_bridge_patch
from agentvoir_local_bot.github_client import GitHubIssue

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class AgentRunResult:
    """Outcome metadata from a completed Cursor agent run."""

    status: str
    run_id: str | None
    summary: str | None


def build_issue_prompt(issue: GitHubIssue) -> str:
    """Compose the agent prompt from issue title and body.

    The worker commits separately after tests pass, so the prompt tells the agent
    to leave changes uncommitted in the working tree.
    """
    body = (issue.body or "").strip()
    parts = [
        "You are an autonomous coding agent working in this repository.",
        f"Implement the following GitHub issue (#{issue.number}):",
        "",
        f"Title: {issue.title}",
    ]
    if body:
        parts.extend(["", "Description:", body])
    parts.extend(
        [
            "",
            "Requirements:",
            "- Make focused, minimal changes that solve the issue.",
            "- Follow existing project conventions and style.",
            "- Do not commit; leave changes in the working tree.",
            f"- Reference issue #{issue.number} in any commit message you would use.",
        ]
    )
    return "\n".join(parts)


class AgentRunner:
    """Launch a local Cursor agent scoped to the repository checkout."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings

    def run_for_issue(self, issue: GitHubIssue) -> AgentRunResult:
        """Run the coding agent for a single issue and stream assistant output.

        Raises RuntimeError when the agent fails to start or completes with error
        status. Startup failures (auth, config) are distinguished from mid-run
        failures via CursorAgentError handling.
        """
        prompt = build_issue_prompt(issue)
        logger.info("Starting local Cursor agent for issue #%s", issue.number)

        apply_windows_bridge_patch()

        local = LocalAgentOptions(cwd=str(self._settings.repo_path))
        options = AgentOptions(
            api_key=self._settings.cursor_api_key,
            model=self._settings.agent_model,
            local=local,
        )
        repo_path = str(self._settings.repo_path)

        try:
            with Client.launch_bridge(workspace=repo_path, local=local) as client:
                with Agent.create(options=options, client=client) as agent:
                    run = agent.send(prompt)
                    for message in run.messages():
                        if message.type == "assistant":
                            for block in message.message.content:
                                if block.type == "text":
                                    sys.stdout.write(block.text)
                                    sys.stdout.flush()
                    result = run.wait()
        except CursorAgentError as err:
            raise RuntimeError(
                f"Agent failed to start: {err.message} (retryable={err.is_retryable})"
            ) from err
        except OSError as err:
            raise RuntimeError(
                "Cursor local bridge failed on this platform. "
                "On Windows, ensure cursor-sdk is up to date; "
                "or run the worker from WSL/Linux. "
                f"Details: {err}"
            ) from err

        if result.status == "error":
            raise RuntimeError(f"Agent run failed (run_id={result.id})")

        summary = getattr(result, "result", None)
        logger.info("Agent finished with status=%s run_id=%s", result.status, result.id)
        return AgentRunResult(status=result.status, run_id=result.id, summary=summary)
