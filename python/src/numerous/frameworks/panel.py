"""Module for integrating Numerous with Panel."""

import panel as pn

from numerous import user_session
from numerous.local import is_local_mode, local_user


class PanelCookieGetter:
    @staticmethod
    def cookies() -> dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Add the user info to the cookies on panel server
            user_session.set_user_info_cookie(pn.state.cookies, local_user)

        if pn.state.curdoc and pn.state.curdoc.session_context:
            return {key: str(val) for key, val in pn.state.cookies.items()}
        return {}


def get_session() -> user_session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return user_session.Session(cg=PanelCookieGetter())
