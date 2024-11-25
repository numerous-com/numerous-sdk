"""Module for integrating Numerous with Marimo."""

from typing import Any

from numerous._utils.local import is_local_mode, local_user
from numerous.experimental import marimo
from numerous.session import Session, session


class MarimoCookieGetter:
    def cookies(self) -> dict[str, Any]:
        """Get the cookies associated with the current request."""
        cookies = marimo.cookies()
        if is_local_mode():
            session.set_user_info_cookie(cookies, local_user)
        return cookies


def get_session() -> Session:
    """
    Get the session for the current user.

    Returns:
        The session for the current user.

    """
    return Session(cg=MarimoCookieGetter())
