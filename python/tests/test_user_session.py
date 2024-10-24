import base64
import json
from unittest import mock

import pytest

from numerous.user_session import Session


class MockUser:
    def __init__(self, user_info: dict[str, str]) -> None:
        self.id = user_info["user_id"]
        self.name = user_info["name"]

    def from_user_info(self: dict[str, str]) -> "MockUser":
        """Create a MockUser instance from a user info dictionary."""
        return MockUser(self)

class MockCookieGetter:
    def __init__(self, cookies: dict[str, str]) -> None:
        self._cookies = cookies

    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        return self._cookies


def test_user_property_raises_value_error_when_no_cookie() -> None:
    cg = MockCookieGetter({})
    session = Session(cg)
    with pytest.raises(ValueError, \
                       match="Invalid user info in cookie or cookie is missing"):
        user = session.user
    assert user is None

def test_user_property_returns_user_when_valid_cookie() -> None:
    user_info = {"user_id": "1", "name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info)\
                                    .encode("utf-8")).decode("utf-8")
    cg = MockCookieGetter({"numerous_user_info": encoded_info})

    with mock.patch("numerous.user_session.User", MockUser):
        session = Session(cg)
        assert isinstance(session.user, MockUser)
        if session.user is None:
            msg = "User is None"
            raise ValueError(msg)
        assert session.user.id == "1"
        assert session.user.name == "Test User"


def test_user_info_returns_decoded_info_for_valid_cookie() -> None:
    user_info = {"user_id": "1", "name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info)\
                                    .encode("utf-8")).decode("utf-8")
    cg = MockCookieGetter({"numerous_user_info": encoded_info})
    session = Session(cg)
    if session.user is None:
        msg = "User is None"
        raise ValueError(msg)
    assert session.user.id == "1"
    assert session.user.name == "Test User"
