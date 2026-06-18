"""Test command execution after the coding agent finishes."""

from __future__ import annotations

import logging
import subprocess
from dataclasses import dataclass

from agentvoir_local_bot.config import Settings

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class TestRunResult:
    """Captured output from the configured test command."""

    returncode: int
    stdout: str
    stderr: str

    @property
    def passed(self) -> bool:
        """True when the test command exited successfully."""
        return self.returncode == 0


class TestRunner:
    """Run the repository test suite (or any shell command) before opening a PR."""

    def __init__(self, settings: Settings) -> None:
        self._settings = settings

    def run(self) -> TestRunResult:
        """Execute TEST_COMMAND in the repo checkout and capture output."""
        command = self._settings.test_command
        logger.info("Running tests: %s", command)
        completed = subprocess.run(
            command,
            cwd=self._settings.repo_path,
            shell=True,
            capture_output=True,
            text=True,
            check=False,
        )
        if completed.stdout:
            logger.debug("test stdout:\n%s", completed.stdout)
        if completed.stderr:
            logger.debug("test stderr:\n%s", completed.stderr)
        return TestRunResult(
            returncode=completed.returncode,
            stdout=completed.stdout,
            stderr=completed.stderr,
        )
