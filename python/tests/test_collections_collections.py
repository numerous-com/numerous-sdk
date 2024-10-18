from unittest.mock import Mock

import pytest

from numerous import collection
from numerous.client._graphql_client import GraphQLClient
from numerous.collection.exceptions import ParentCollectionNotFoundError
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_create import CollectionCreate
from numerous.jsonbase64 import dict_to_base64


ORGANIZATION_ID = "test_org"
COLLECTION_NAME = "test_collection"
NESTED_COLLECTION_ID = "nested_test_collection"
COLLECTION_REFERENCE_KEY = "test_key"
COLLECTION_REFERENCE_ID = "test_id"
NESTED_COLLECTION_REFERENCE_KEY = "nested_test_key"
NESTED_COLLECTION_REFERENCE_ID = "nested_test_id"
COLLECTION_DOCUMENT_KEY = "test_document"
DOCUMENT_DATA = {"test": "test"}
BASE64_DOCUMENT_DATA = dict_to_base64(DOCUMENT_DATA)
DOCUMENT_ID = "915b75c5-9e95-4fa7-aaa2-2214c8d251ce"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}


def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "Collection", "key": key, "id": ref_id}}
    )


def _collection_create_collection_not_found(ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "CollectionNotFound", "id": ref_id}}
    )


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


def test_collection_returns_new_collection() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    parent_key = None

    result = collection(COLLECTION_NAME, _client)

    gql.collection_create.assert_called_once()
    gql.collection_create.assert_called_once_with(
        ORGANIZATION_ID, COLLECTION_NAME, parent_key, **HEADERS_WITH_AUTHORIZATION
    )
    assert result.key == COLLECTION_REFERENCE_KEY
    assert result.id == COLLECTION_REFERENCE_ID


def test_collection_returns_new_nested_collection() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        NESTED_COLLECTION_REFERENCE_KEY,
        NESTED_COLLECTION_REFERENCE_ID,
    )
    result = collection(COLLECTION_NAME, _client)

    nested_result = result.collection(NESTED_COLLECTION_ID)

    assert nested_result is not None
    assert nested_result.key == NESTED_COLLECTION_REFERENCE_KEY
    assert nested_result.id == NESTED_COLLECTION_REFERENCE_ID


def test_nested_collection_not_found_raises_parent_not_found_error() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    result = collection(COLLECTION_NAME, _client)
    gql.collection_create.return_value = _collection_create_collection_not_found(
        COLLECTION_REFERENCE_ID
    )

    with pytest.raises(ParentCollectionNotFoundError) as exc_info:
        result.collection(NESTED_COLLECTION_REFERENCE_KEY)

    assert exc_info.value.collection_id == COLLECTION_REFERENCE_ID
