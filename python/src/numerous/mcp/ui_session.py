"""In-memory store linking live browser UI sessions to backend MCP tool calls."""

from __future__ import annotations

from threading import RLock
from typing import Any


class UISessionStore:
    """
    Thread-safe in-memory store keyed by verified Numerous UI session IDs.

    Apps can publish live UI state (selected filters, active scenario, etc.)
    under ``session.ui_session_id`` so MCP tools receiving the same verified
    ``caller.ui_session_id`` can read or update the exact live tab state.
    """

    def __init__(self) -> None:
        self._lock = RLock()
        self._states: dict[str, dict[str, Any]] = {}

    def set(self, session_id: str, state: dict[str, Any]) -> None:
        """Replace the stored state for a UI session ID."""
        with self._lock:
            self._states[session_id] = dict(state)

    def update(self, session_id: str, **values: Any) -> dict[str, Any]:  # noqa: ANN401
        """Merge values into the stored state for a UI session ID."""
        with self._lock:
            current = self._states.setdefault(session_id, {})
            current.update(values)
            return dict(current)

    def get(self, session_id: str | None) -> dict[str, Any] | None:
        """Return a shallow copy of the stored state for a UI session ID, if any."""
        if not session_id:
            return None
        with self._lock:
            state = self._states.get(session_id)
            return dict(state) if state is not None else None

    def delete(self, session_id: str) -> None:
        """Remove stored state for a UI session ID."""
        with self._lock:
            self._states.pop(session_id, None)

    def clear(self) -> None:
        """Remove all stored UI session states."""
        with self._lock:
            self._states.clear()


ui_sessions = UISessionStore()
