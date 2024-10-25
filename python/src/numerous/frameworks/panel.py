"""Module for integrating Numerous with Panel."""

from typing import Dict

import panel as pn

from numerous import user_session
from numerous.utils import is_local_mode


class PanelCookieGetter:
    @staticmethod
    def cookies() -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Add the user info to the cookies on panel server
            pn.state.cookies.update(user_session.get_encoded_user_info())

        if pn.state.curdoc and pn.state.curdoc.session_context:
            return {key: str(val) for key, val in pn.state.cookies.items()}
        return {}


def get_session() -> user_session.Session:
    return user_session.Session(cg=PanelCookieGetter())
