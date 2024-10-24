"""Module for integrating Numerous with Panel."""

from typing import Dict

import panel as pn

from numerous import user_session


class PanelCookieGetter:
    @staticmethod
    def cookies() -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        if pn.state.curdoc and pn.state.curdoc.session_context:
            return {key: str(val) for key, val in pn.state.cookies.items()}
        return {}


def get_session() -> user_session.Session:
    return user_session.Session(cg=PanelCookieGetter())
