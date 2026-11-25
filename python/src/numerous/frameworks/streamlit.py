"""Module for integrating Numerous with Streamlit."""

from __future__ import annotations

from typing import TYPE_CHECKING, Any

import streamlit as st

from numerous._utils.local import is_local_mode, local_user
from numerous.agent import agent_context_script
from numerous.session import Session, session


if TYPE_CHECKING:
    from collections.abc import Mapping


class StreamlitCookieGetter:
    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        cookies = {key: str(val) for key, val in st.context.cookies.items()}
        if is_local_mode():
            # Update the cookies on the streamlit server
            session.set_user_info_cookie(cookies, local_user)
        return cookies

    def headers(self) -> dict[str, str]:
        """Get the HTTP headers associated with the current request."""
        raw_headers = getattr(st.context, "headers", None)
        if not raw_headers:
            return {}
        return {str(key): str(val) for key, val in raw_headers.items()}


def get_session() -> Session:
    """
    Get the session for the current user.

    Returns:
        The session for the current user.

    """
    getter = StreamlitCookieGetter()
    return Session(cg=getter, hg=getter)


def publish_agent_context(context: Mapping[str, Any]) -> None:
    """Publish structured app context to the Numerous Agent UI bridge."""
    script = agent_context_script(context)
    st.html(script)
