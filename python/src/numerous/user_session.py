"""Module for managing user sessions and cookie-based authentication."""

import base64
import json
from typing import Any, Dict, Optional, Protocol

from numerous._client._graphql_client import GraphQLClient
from numerous.user import User


class CookieGetter(Protocol):
    def cookies(self) -> Dict[str, Any]:
        """Get the cookies associated with the current session."""
        ...


def encode_user_info(user_id: str, name: str) -> str:
    user_info_json = json.dumps({"user_id": user_id, "name": name})
    return base64.b64encode(user_info_json.encode("utf-8")).decode("utf-8")


def get_encoded_user_info(user: User) -> Dict[str, str]:
    return {"numerous_user_info": encode_user_info(user.id, user.name)}


class Session:
    def __init__(
        self, cg: CookieGetter, client: Optional[GraphQLClient] = None
    ) -> None:
        self._cg = cg
        self._user: Optional[User] = None
        self._client = client

    def _user_info(self) -> Dict[str, str]:
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
    def user(self) -> Optional[User]:
        """Get the User instance associated with the current session."""
        if self._user is None:
            user_info = self._user_info()
            self._user = User.from_user_info(user_info, self._client)
        return self._user

    @property
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current session."""
        return self._cg.cookies()
