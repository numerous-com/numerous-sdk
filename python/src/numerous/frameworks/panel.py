"""Module for integrating Numerous with Panel."""

import panel as pn

from numerous._utils.local import is_local_mode, local_user
from numerous.session import session


class PanelCookieGetter:
    @staticmethod
    def cookies() -> dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Add the user info to the cookies on panel server
            session.set_user_info_cookie(pn.state.cookies, local_user)

        if pn.state.curdoc and pn.state.curdoc.session_context:
            return {key: str(val) for key, val in pn.state.cookies.items()}
        return {}


def get_session() -> session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return session.Session(cg=PanelCookieGetter())
