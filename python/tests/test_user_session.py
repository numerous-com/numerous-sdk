import base64
import json
from unittest import mock

import pytest

from numerous.user_session import Session


class MockUser:
    def __init__(self, user_info: dict[str, str]):
        self.id = user_info["id"]
        self.name = user_info["name"]

class MockCookieGetter:
    def __init__(self, cookies):
        self._cookies = cookies

    def cookies(self):
        return self._cookies

def test_session_initializes_with_no_user():
    cg = MockCookieGetter({})
    session = Session(cg)
    assert session._user is None

def test_user_property_raises_value_error_when_no_cookie():
    cg = MockCookieGetter({})
    session = Session(cg)
    with pytest.raises(ValueError, match="Invalid user info in cookie or cookie is missing"):
        session.user

def test_user_property_returns_user_when_valid_cookie():
    user_info = {"id": 1, "name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info).encode("utf-8")).decode("utf-8")
    cg = MockCookieGetter({"numerous_user_info": encoded_info})

    with mock.patch("numerous.user_session.User", MockUser):
        session = Session(cg)
        assert isinstance(session.user, MockUser)
        assert session.user.id == 1
        assert session.user.name == "Test User"


def test_user_info_returns_decoded_info_for_valid_cookie():
    user_info = {"id": 1, "name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info).encode("utf-8")).decode("utf-8")
    cg = MockCookieGetter({"numerous_user_info": encoded_info})
    session = Session(cg)
    assert session._user_info() == user_info
