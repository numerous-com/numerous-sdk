import base64
import json
from unittest import mock

import pytest

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql_client import GraphQLClient
from numerous.session import Session


class CookieGetterStub:
    def __init__(self, cookies: dict[str, str]) -> None:
        self._cookies = cookies

    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        return self._cookies


ORGANIZATION_ID = "test_org"


@pytest.fixture
def gql() -> GQLClient:
    return mock.Mock(GQLClient)


@pytest.fixture
def client(gql: GQLClient) -> GraphQLClient:
    return GraphQLClient(gql, ORGANIZATION_ID, "token")


def test_user_property_raises_value_error_when_no_cookie(client: GraphQLClient) -> None:
    cg = CookieGetterStub({})

    session = Session(cg, _client=client)

    with pytest.raises(
        ValueError, match="Invalid user info in cookie or cookie is missing"
    ):
        # ruff: noqa: B018
        session.user


def test_user_property_returns_user_when_valid_cookie(
    client: GraphQLClient,
) -> None:
    user_info = {
        "user_id": "1",
        "user_full_name": "Test User",
        "user_email": "test@example.com",
    }
    encoded_info = base64.b64encode(json.dumps(user_info).encode()).decode()
    cg = CookieGetterStub({"numerous_user_info": encoded_info})

    session = Session(cg, _client=client)

    assert session.user is not None
    assert session.user.id == "1"
    assert session.user.name == "Test User"
    assert session.user.email == "test@example.com"


def test_user_property_returns_user_when_valid_cookie_and_no_email(
    client: GraphQLClient,
) -> None:
    user_info = {"user_id": "1", "user_full_name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info).encode()).decode()
    cg = CookieGetterStub({"numerous_user_info": encoded_info})

    session = Session(cg, _client=client)

    assert session.user is not None
    assert session.user.id == "1"
    assert session.user.name == "Test User"
    assert session.user.email is None
