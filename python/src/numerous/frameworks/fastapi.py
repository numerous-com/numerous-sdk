"""Module for integrating Numerous with FastAPI."""

from typing import Dict

from fastapi import Request

from numerous import user_session
from numerous.utils import is_local_mode


class FastapiCookieGetter:
    def __init__(self, request: Request) -> None:
        self.request = request

    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Update the cookies on the fastapi server
            self.request.cookies.update(user_session.get_encoded_user_info())

        return {str(key): str(val) for key, val in self.request.cookies.items()}


def get_session(request: Request) -> user_session.Session:
    return user_session.Session(cg=FastapiCookieGetter(request))
