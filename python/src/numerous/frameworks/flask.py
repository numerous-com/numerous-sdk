"""Module for integrating Numerous with Flask."""

from flask import request

from numerous import user_session
from numerous.local import is_local_mode, local_user


class FlaskCookieGetter:
    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        cookies = {key: str(val) for key, val in request.cookies.items()}
        if is_local_mode():
            # Update the cookies on the flask server
            user_session.set_user_info_cookie(cookies, local_user)
        return cookies


def get_session() -> user_session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return user_session.Session(cg=FlaskCookieGetter())
