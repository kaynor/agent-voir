# AgentVoir Local Bot

Local worker that polls GitHub for issues labeled `ai-code`, runs a Cursor coding agent on your machine, runs tests, and opens a pull request for human review.

## Flow

```text
GitHub Issue (label: ai-code)
        ↓
Worker polls GitHub API
        ↓
Worker claims issue (label: ai-code-claimed)
        ↓
Creates local branch
        ↓
Runs Cursor agent locally (cursor-sdk)
        ↓
Runs tests (make test by default)
        ↓
Commits, pushes branch
        ↓
Creates PR via GitHub API
        ↓
You review and merge
```

## Prerequisites

- Python 3.11+
- Git with push access to the target repo
- [Cursor API key](https://cursor.com/docs) (`CURSOR_API_KEY`)
- GitHub personal access token with `repo` scope (`GITHUB_TOKEN`)

## Setup

```bash
cd tools/agentvoir_local_bot
cp .env.example .env
# Edit .env with your tokens and repo

pip install -e ".[dev]"
```

## Run

Poll continuously (default):

```bash
agentvoir-local-bot
```

Process one issue and exit:

```bash
agentvoir-local-bot --once
```

Or without installing the console script:

```bash
python -m agentvoir_local_bot
```

## GitHub labels

| Label | Purpose |
| ----- | ------- |
| `ai-code` | Trigger — issue is ready for the bot |
| `ai-code-claimed` | Bot is working on it |
| `ai-code-done` | PR opened successfully |
| `ai-code-failed` | Bot could not complete the issue |

Label names are configurable via `.env`.

## Configuration

See [.env.example](.env.example). Key settings:

| Variable | Description |
| -------- | ----------- |
| `GITHUB_TOKEN` | GitHub API token |
| `GITHUB_REPO` | `owner/repo` to watch |
| `CURSOR_API_KEY` | Cursor SDK API key |
| `REPO_PATH` | Local git checkout (defaults to repo root) |
| `TEST_COMMAND` | Shell command to validate changes (default: `make test`) |
| `POLL_INTERVAL_SECONDS` | Poll interval when idle (default: 60) |

## Issue workflow

1. Create a GitHub issue describing the task.
2. Add the `ai-code` label.
3. Start the worker locally.
4. The worker claims the issue, implements it with the Cursor agent, runs tests, and opens a PR.
5. Review and merge the PR.

Failed runs are labeled `ai-code-failed` with an error comment on the issue.

## Development

```bash
cd tools/agentvoir_local_bot
pip install -e ".[dev]"
pytest
ruff check .
```
