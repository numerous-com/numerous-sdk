"""Module for integrating Numerous with Flask."""

from typing import Dict

from flask import request

from numerous import user_session
from numerous.utils import is_local_mode


class FlaskCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        local_cookies = {}
        if is_local_mode():
            # Update the cookies on the flask server
            local_cookies = user_session.get_encoded_user_info()
        return {
            **local_cookies,
            **{str(key): str(val) for key, val in request.cookies.items()},
        }


session = user_session.Session(cg=FlaskCookieGetter())
