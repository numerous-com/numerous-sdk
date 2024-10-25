from unittest.mock import Mock

import pytest

from numerous._client._graphql_client import GraphQLClient
from numerous.collection import NumerousCollection
from numerous.generated.graphql.client import Client as GQLClient
from numerous.user import User


ORGANIZATION_ID = "test_org"


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


@pytest.fixture
def mock_gql_client() -> GQLClient:
    return Mock(GQLClient)


mock_key = "mock_key"
mock_id = "mock_id"


@pytest.fixture
def mock_graphql_client(mock_gql_client: GQLClient) -> GraphQLClient:
    mock_graphql_client = GraphQLClient(mock_gql_client)

    # Create a mock CollectionReference
    mock_collection_ref = Mock()
    mock_collection_ref.id = mock_id
    mock_collection_ref.key = mock_key

    # Set up the mock to return the mock CollectionReference
    mock_graphql_client._create_collection_ref = Mock(return_value=mock_collection_ref)  # type: ignore[method-assign] # noqa: SLF001

    return mock_graphql_client


def test_user_collection_property_returns_numerous_collection(
    mock_graphql_client: GraphQLClient,
) -> None:
    user = User(id=mock_id, name="John Doe", _client=mock_graphql_client)
    assert isinstance(user.collection, NumerousCollection)


def test_user_collection_property_uses_user_id(
    mock_graphql_client: GraphQLClient,
) -> None:
    user = User(id=mock_id, name="John Doe", _client=mock_graphql_client)
    assert user.collection is not None
    assert user.collection.key == mock_key


def test_from_user_info_returns_user_with_correct_attributes(
    mock_graphql_client: GraphQLClient,
) -> None:
    user_info = {"user_id": mock_id, "name": "Jane Smith"}
    user = User.from_user_info(user_info, _client=mock_graphql_client)

    assert user.id == mock_id
    assert user.name == "Jane Smith"


def test_from_user_info_returns_user_instance(
    mock_graphql_client: GraphQLClient,
) -> None:
    user_info = {"user_id": mock_id, "name": "Alice Johnson"}
    user = User.from_user_info(user_info, _client=mock_graphql_client)
    assert isinstance(user, User)
