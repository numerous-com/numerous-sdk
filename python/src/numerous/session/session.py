"""Module for managing user sessions and cookie-based authentication."""

from __future__ import annotations

import base64
import json
from typing import TYPE_CHECKING, Any, Protocol


if TYPE_CHECKING:
    from numerous._client.graphql_client import GraphQLClient

from .user import User


class CookieGetter(Protocol):
    def cookies(self) -> dict[str, Any]:
        """Get the cookies associated with the current session."""


def encode_user_info(user_id: str, name: str) -> str:
    user_info_json = json.dumps({"user_id": user_id, "user_full_name": name})
    return base64.b64encode(user_info_json.encode()).decode()


def set_user_info_cookie(cookies: dict[str, str], user: User) -> None:
    cookies["numerous_user_info"] = encode_user_info(user.id, user.name)


class Session:
    """A session with Numerous."""

    def __init__(self, cg: CookieGetter, _client: GraphQLClient | None = None) -> None:
        self._cg = cg
        self._user: User | None = None
        self._client = _client

    def _user_info(self) -> dict[str, str]:
        cookies = self._cg.cookies()
        user_info_b64 = cookies.get("numerous_user_info")
        if not user_info_b64:
            msg = "Invalid user info in cookie or cookie is missing"
            raise ValueError(msg)
        try:
            user_info_json = base64.b64decode(user_info_b64).decode("utf-8")
            return {
                str(key): str(val) for key, val in json.loads(user_info_json).items()
            }
        except ValueError as err:
            msg = "Invalid user info in cookie or cookie is missing"
            raise ValueError(msg) from err

    @property
    def user(self) -> User:
        """The user associated with the current session."""
        if self._user is None:
            user_info = self._user_info()
            self._user = User.from_user_info(user_info, self._client)
        return self._user

    @property
    def cookies(self) -> dict[str, str]:
        """The cookies associated with the current session."""
        return self._cg.cookies()
