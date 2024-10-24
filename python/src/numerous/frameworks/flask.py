"""Module for integrating Numerous with Flask."""

from typing import Dict

from flask import request

from numerous import user_session


class FlaskCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        return {str(key): str(val) for key, val in request.cookies.items()}


session = user_session.Session(cg=FlaskCookieGetter())
