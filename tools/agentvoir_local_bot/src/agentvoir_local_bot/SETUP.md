**Use a dedicated virtual environment.** 
The bot has its own dependencies (`cursor-sdk`, `pydantic-settings`, etc.) separate from the rest of the monorepo (`packages/sdk-python`). A venv keeps versions isolated and matches what `.gitignore` already expects (`.venv/` is ignored).

## Setup (Windows / PowerShell)

From the repo root:

```powershell
cd d:\Projects\agentvoir\tools\agentvoir_local_bot

# 1. Create and activate a venv (Python 3.11+)
python -m venv .venv
.\.venv\Scripts\Activate.ps1
You should see `(.venv)` in your prompt after activation.

# 2. Upgrade pip and install the bot + dev tools
python -m pip install --upgrade pip
pip install -e ".[dev]"

# 3. Configure secrets (needed to run the bot, not for unit tests)
Copy-Item .env.example .env
# Edit .env — set at minimum:
#   GITHUB_TOKEN=...
#   GITHUB_REPO=kaynor/agent-voir
#   CURSOR_API_KEY=...
```

## Run tests

With the venv still active:

```powershell
cd d:\Projects\agentvoir\tools\agentvoir_local_bot

pytest
```

Or explicitly:

```powershell
python -m pytest -q
```

Optional lint:

```powershell
python -m ruff check .
```

The unit tests mock GitHub HTTP calls and build `Settings` in code, so **you don't need a real `.env` just to run `pytest`**.

## Run the bot (after tests pass)

```powershell
# Still in tools/agentvoir_local_bot with .venv active
agentvoir-local-bot --once    # process one issue and exit
agentvoir-local-bot           # poll continuously
```

## Day-to-day

Each new terminal session:

```powershell
cd d:\Projects\agentvoir\tools\agentvoir_local_bot
.\.venv\Scripts\Activate.ps1
```

To leave the venv:

```powershell
deactivate
```

## Quick checklist

| Step | Command |
|------|---------|
| Create venv | `python -m venv .venv` |
| Activate | `.\.venv\Scripts\Activate.ps1` |
| Install | `pip install -e ".[dev]"` |
| Test | `pytest` |
| Run bot | `agentvoir-local-bot --once` (needs `.env` + GitHub/Cursor keys) |

**Note:** Default `TEST_COMMAND` is `make test`, which runs the **whole monorepo** test suite when the bot processes an issue. For bot development only, `pytest` inside `tools/agentvoir_local_bot` is enough. If you want the bot to only run its own tests after coding, set `TEST_COMMAND=pytest` in `.env`.