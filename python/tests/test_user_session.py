import base64
import json
from unittest import mock

import pytest

from numerous._client._graphql_client import GraphQLClient
from numerous.generated.graphql.client import Client as GQLClient
from numerous.user import User
from numerous.user_session import Session


class CookieGetterStub:
    def __init__(self, cookies: dict[str, str]) -> None:
        self._cookies = cookies

    def cookies(self) -> dict[str, str]:
        """Get the cookies associated with the current request."""
        return self._cookies


ORGANIZATION_ID = "test_org"


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


@pytest.fixture
def mock_gql_client() -> GQLClient:
    return mock.Mock(GQLClient)


@pytest.fixture
def mock_graphql_client(mock_gql_client: GQLClient) -> GraphQLClient:
    return GraphQLClient(mock_gql_client)


def test_user_property_raises_value_error_when_no_cookie(
    mock_graphql_client: GraphQLClient,
) -> None:
    cg = CookieGetterStub({})
    session = Session(cg, _client=mock_graphql_client)
    with pytest.raises(
        ValueError, match="Invalid user info in cookie or cookie is missing"
    ):
        # ruff: noqa: B018
        session.user


def test_user_property_returns_user_when_valid_cookie(
    mock_graphql_client: GraphQLClient,
) -> None:
    user_info = {"user_id": "1", "name": "Test User"}
    encoded_info = base64.b64encode(json.dumps(user_info).encode()).decode()
    cg = CookieGetterStub({"numerous_user_info": encoded_info})

    session = Session(cg, _client=mock_graphql_client)

    assert isinstance(session.user, User)

    assert session.user is not None

    assert session.user.id == "1"

    assert session.user.name == "Test User"
