import pytest

from numerous._client.graphql_client import GraphQLClient
from numerous.collections._get_client import get_client


@pytest.fixture(autouse=True)
def _clear_client() -> None:
    import numerous.collections._get_client

    numerous.collections._get_client._client = None  # noqa: SLF001


def test_given_graphql_environment_variables_returns_graphql_client(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", "organization-id")

    client = get_client()

    assert isinstance(client, GraphQLClient)


def test_given_graphql_environment_variables_without_url_returns_graphql_client(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("NUMEROUS_API_URL", raising=False)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", "organization-id")

    client = get_client()

    assert isinstance(client, GraphQLClient)
