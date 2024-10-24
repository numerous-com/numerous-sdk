"""Module for integrating Numerous with FastAPI."""

from typing import Any, Dict

from fastapi import Request

from numerous import user_session


class FastapiCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        cookies = {str(key): str(val) for key, val in Request.cookies.items()}
        
        return cookies

session = user_session.Session(cg=FastapiCookieGetter())
