"""Windows-compatible workarounds for cursor-sdk local bridge startup."""

from __future__ import annotations

import sys
import time
from collections.abc import Mapping
from typing import Any

_PATCHED = False


def apply_windows_bridge_patch() -> None:
    """Patch cursor-sdk bridge discovery to avoid select() on pipes (Windows).

    cursor-sdk 0.1.x uses ``selectors.DefaultSelector`` while waiting for the
    bridge subprocess to print its discovery line. On Windows, ``select`` only
    works on sockets, not pipe handles, which raises ``OSError: [WinError 10038]``.
    """
    global _PATCHED
    if _PATCHED or sys.platform != "win32":
        return

    import subprocess

    from cursor_sdk._bridge import parse_discovery_line
    from cursor_sdk.errors import CursorSDKError

    def _read_discovery_readline(
        process: subprocess.Popen[str], timeout: float
    ) -> Mapping[str, Any]:
        if process.stderr is None:
            raise CursorSDKError("Bridge process stderr is unavailable")

        deadline = time.monotonic() + timeout
        stderr_lines: list[str] = []
        while time.monotonic() < deadline:
            line = process.stderr.readline()
            if line:
                stderr_lines.append(line)
                discovery = parse_discovery_line(line)
                if discovery is not None:
                    return discovery

            exit_code = process.poll()
            if exit_code is not None:
                raise CursorSDKError(
                    f"Bridge exited before discovery with status {exit_code}: "
                    + "".join(stderr_lines)
                )

            if not line:
                time.sleep(0.05)

        raise CursorSDKError("Timed out waiting for bridge discovery")

    import cursor_sdk._bridge as bridge_module

    bridge_module._read_discovery = _read_discovery_readline
    _PATCHED = True
