"""Module for integrating Numerous with Dash."""

from numerous import user_session
from numerous.frameworks.flask import FlaskCookieGetter


class DashCookieGetter(FlaskCookieGetter):
    pass


session = user_session.Session(cg=DashCookieGetter())
