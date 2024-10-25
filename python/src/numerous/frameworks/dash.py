"""Module for integrating Numerous with Dash."""

from numerous import user_session
from numerous.frameworks.flask import FlaskCookieGetter


class DashCookieGetter(FlaskCookieGetter):
    pass


def get_session() -> user_session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return user_session.Session(cg=DashCookieGetter())
