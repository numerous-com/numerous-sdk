"""Module for integrating Numerous with FastAPI."""

from fastapi import Request

from numerous._utils.local import is_local_mode, local_user
from numerous.session import Session, session


class FastAPICookieGetter:
    def __init__(self, request: Request) -> None:
        self.request = request

    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        if is_local_mode():
            # Update the cookies on the fastapi server
            session.set_user_info_cookie(self.request.cookies, local_user)

        return {str(key): str(val) for key, val in self.request.cookies.items()}


def get_session(request: Request) -> Session:
    """
    Get the session for the current user.

    Returns:
        The session for the current user.

    """
    return Session(cg=FastAPICookieGetter(request))
