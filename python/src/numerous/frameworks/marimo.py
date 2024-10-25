"""Module for integrating Numerous with Marimo."""

from typing import Any, Dict

from numerous import user_session
from numerous.experimental import marimo
from numerous.utils import is_local_mode


class MarimoCookieGetter:
    def cookies(self) -> Dict[str, Any]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Update the cookies on the marimo server
            marimo.cookies().update(user_session.get_encoded_user_info())
        return {key: str(val) for key, val in marimo.cookies().items()}


session = user_session.Session(cg=MarimoCookieGetter())
