import pytest

from numerous._client.get_client import get_client
from numerous._client.graphql_client import GraphQLClient


def test_open_client_with_graphql_environment_returns_graphql_client(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", "organization-id")

    client = get_client()

    assert isinstance(client, GraphQLClient)
