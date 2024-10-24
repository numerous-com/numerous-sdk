"""Module for integrating Numerous with Marimo."""

from typing import Any, Dict

from numerous import user_session
from numerous.experimental import marimo


class MarimoCookieGetter:
    def cookies(self) -> Dict[str, Any]:
        """Get the cookies associated with the current request."""
        cookies = {key: str(val) for key, val in marimo.cookies().items()}
        return cookies


session = user_session.Session(cg=MarimoCookieGetter())
