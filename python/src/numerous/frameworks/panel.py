"""Module for integrating Numerous with Panel."""

from typing import Any, Dict

import panel as pn

from numerous import user_session


class PanelCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        cookies = {key: str(val) for key, val in pn.request.cookies.items()}
        
        return cookies

session = user_session.Session(cg=PanelCookieGetter())
