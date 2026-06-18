"""CLI entry point for the AgentVoir local bot worker."""

from __future__ import annotations

import argparse
import logging
import sys

from agentvoir_local_bot.config import Settings
from agentvoir_local_bot.worker import IssueWorker


def _configure_logging(verbose: bool) -> None:
    """Set root log level and a simple timestamped format."""
    level = logging.DEBUG if verbose else logging.INFO
    logging.basicConfig(
        level=level,
        format="%(asctime)s %(levelname)s %(name)s: %(message)s",
    )


def main(argv: list[str] | None = None) -> int:
    """Parse CLI args, load settings, and start the worker loop."""
    parser = argparse.ArgumentParser(
        description="Poll GitHub for ai-code issues and run a local coding agent.",
    )
    parser.add_argument(
        "--once",
        action="store_true",
        help="Process at most one issue and exit.",
    )
    parser.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="Enable debug logging.",
    )
    args = parser.parse_args(argv)
    _configure_logging(args.verbose)

    try:
        settings = Settings()  # type: ignore[call-arg]
    except Exception as err:
        logging.error("Invalid configuration: %s", err)
        return 1

    with IssueWorker(settings) as worker:
        if args.once:
            worker.poll_once()
        else:
            worker.run_forever()
    return 0


if __name__ == "__main__":
    sys.exit(main())
