"""Module for integrating Numerous with FastAPI."""

from typing import Dict

from fastapi import Request

from numerous import user_session


class FastapiCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        return {str(key): str(val) for key, val in Request.cookies.items()}


session = user_session.Session(cg=FastapiCookieGetter())
