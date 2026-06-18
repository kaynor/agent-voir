"""Git subprocess helpers for branch management and publishing changes."""

from __future__ import annotations

import re
import subprocess
from dataclasses import dataclass
from pathlib import Path

from agentvoir_local_bot.config import Settings


class GitError(RuntimeError):
    """Raised when a git command exits with a non-zero status."""


@dataclass(frozen=True)
class GitResult:
    """Captured output from a git subprocess invocation."""

    returncode: int
    stdout: str
    stderr: str


def _slugify(text: str, max_length: int = 40) -> str:
    """Convert issue titles into URL-safe branch name segments."""
    slug = re.sub(r"[^a-z0-9]+", "-", text.lower()).strip("-")
    return slug[:max_length].rstrip("-") or "task"


class GitOps:
    """Run git commands against the configured local repository checkout."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._cwd = settings.repo_path

    def branch_name_for_issue(self, issue_number: int, title: str) -> str:
        """Derive a deterministic branch name from issue metadata."""
        slug = _slugify(title)
        return f"{self._settings.branch_prefix}-{issue_number}-{slug}"

    def run(self, args: list[str], *, check: bool = True) -> GitResult:
        """Execute a git command in the repo checkout and optionally raise on failure."""
        completed = subprocess.run(
            args,
            cwd=self._cwd,
            capture_output=True,
            text=True,
            check=False,
        )
        result = GitResult(
            returncode=completed.returncode,
            stdout=completed.stdout,
            stderr=completed.stderr,
        )
        if check and result.returncode != 0:
            cmd = " ".join(args)
            raise GitError(
                f"git command failed ({cmd}): {result.stderr.strip() or result.stdout.strip()}"
            )
        return result

    def ensure_clean_base(self) -> None:
        """Fetch remote state and fast-forward the configured base branch."""
        self.run(["git", "fetch", self._settings.git_push_remote])
        self.run(["git", "checkout", self._settings.base_branch])
        self.run(["git", "pull", "--ff-only", self._settings.git_push_remote, self._settings.base_branch])

    def create_branch(self, branch_name: str) -> None:
        """Reset to base, then create or reset the working branch from it."""
        self.ensure_clean_base()
        # -B creates the branch if missing or resets it to the current HEAD (base).
        self.run(["git", "checkout", "-B", branch_name])

    def has_changes(self) -> bool:
        """Return True when the working tree has staged or unstaged changes."""
        result = self.run(["git", "status", "--porcelain"], check=False)
        return bool(result.stdout.strip())

    def commit_all(self, message: str) -> None:
        """Stage all changes and commit when the working tree is dirty."""
        if not self.has_changes():
            return
        self.run(["git", "add", "-A"])
        self.run(["git", "commit", "-m", message])

    def push_branch(self, branch_name: str) -> None:
        """Push a branch to the configured remote and set upstream tracking."""
        self.run(
            [
                "git",
                "push",
                "-u",
                self._settings.git_push_remote,
                branch_name,
            ]
        )

    def current_branch(self) -> str:
        """Return the short name of the checked-out branch."""
        result = self.run(["git", "rev-parse", "--abbrev-ref", "HEAD"])
        return result.stdout.strip()

    def repo_root(self) -> Path:
        """Return the absolute path to the git repository root."""
        result = self.run(["git", "rev-parse", "--show-toplevel"])
        return Path(result.stdout.strip())
