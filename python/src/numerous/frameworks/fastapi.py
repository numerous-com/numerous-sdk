"""Module for integrating Numerous with FastAPI."""

from fastapi import Request

from numerous import user_session
from numerous.local import is_local_mode, local_user


class FastAPICookieGetter:
    def __init__(self, request: Request) -> None:
        self.request = request

    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Update the cookies on the fastapi server
            user_session.set_user_info_cookie(self.request.cookies, local_user)

        return {str(key): str(val) for key, val in self.request.cookies.items()}


def get_session(request: Request) -> user_session.Session:
    """
    Get the session for the current user.

    Returns:
        Session: The session for the current user.

    """
    return user_session.Session(cg=FastAPICookieGetter(request))
