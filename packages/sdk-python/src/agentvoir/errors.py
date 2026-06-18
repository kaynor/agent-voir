from __future__ import annotations


class AgentVoirError(Exception):
    """Raised when the AgentVoir API returns an error response."""

    def __init__(self, message: str, status_code: int | None = None) -> None:
        super().__init__(message)
        self.status_code = status_code
