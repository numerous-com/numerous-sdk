"""Module for integrating Numerous with Dash."""

from numerous.frameworks.flask import FlaskCookieGetter
from numerous.session import Session


class DashCookieGetter(FlaskCookieGetter):
    pass


def get_session() -> Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return Session(cg=DashCookieGetter())
