"""Module for integrating Numerous with Marimo."""

from typing import Any

from numerous import user_session
from numerous.experimental import marimo
from numerous.local import is_local_mode, local_user


class MarimoCookieGetter:
    def cookies(self) -> dict[str, Any]:
        """Get the cookies associated with the current request."""
        cookies = marimo.cookies()
        if is_local_mode():
            user_session.set_user_info_cookie(cookies, local_user)
        return cookies


def get_session() -> user_session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return user_session.Session(cg=MarimoCookieGetter())
