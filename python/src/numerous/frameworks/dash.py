"""Module for integrating Numerous with Dash."""

from numerous import user_session
from numerous.frameworks.flask import FlaskCookieGetter


class DashCookieGetter(FlaskCookieGetter):
    pass


session = user_session.Session(cg=DashCookieGetter())


def get_session() -> user_session.Session:
    return user_session.Session(cg=DashCookieGetter())
