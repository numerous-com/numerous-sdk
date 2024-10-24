"""Module for integrating Numerous with FastAPI."""

from typing import Dict

from fastapi import Request

from numerous import user_session


class FastapiCookieGetter:

    def __init__(self, request: Request) -> None:
        self.request = request

    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        return {str(key): str(val) for key, val in self.request.cookies.items()}


def get_session(request: Request) -> user_session.Session:
    return user_session.Session(cg=FastapiCookieGetter(request))
