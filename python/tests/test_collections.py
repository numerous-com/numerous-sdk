from unittest.mock import Mock

import pytest
from numerous import collection
from numerous._client import Client
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_create import CollectionCreate
from numerous.generated.graphql.fragments import CollectionReference


ORGANIZATION_ID = "test_org"
COLLECTION_NAME = "test_collection"
NESTED_COLLECTION_ID = "nested_test_collection"
COLLECTION_REFERENCE = CollectionReference(key="test_key", id="test_id")
NESTED_COLLECTION_REFERENCE = CollectionReference(key="nested_test_key",
                                                 id="nested_test_id")

def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate({"collectionCreate":
                                             {"typename__": "Collection",
                                              "key": key,
                                              "id": ref_id}})

def _collection_create_collection_not_found(ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate({"collectionCreate":
                                             {"typename__":
                                              "CollectionNotFound",
                                              "id": ref_id}})


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch:pytest.MonkeyPatch)->None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


def test_collection_returns_new_collection()->None:
    gql = Mock(GQLClient)
    _client = Client(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE.key,COLLECTION_REFERENCE.id)
    organization_id = ""
    parent_key = None
    kwargs={"headers": {"Authorization": "Bearer token"}}

    result = collection(COLLECTION_NAME, _client)


    gql.collection_create.assert_called_once()
    gql.collection_create.assert_called_once_with(organization_id,COLLECTION_NAME,parent_key,
                                                  kwargs=kwargs)
    assert result.key == COLLECTION_REFERENCE.key
    assert result.id == COLLECTION_REFERENCE.id


def test_collection_returns_new_nested_collection()->None:
    gql = Mock(GQLClient)
    _client = Client(gql)
    gql.collection_create.return_value =  _collection_create_collection_reference(
        NESTED_COLLECTION_REFERENCE.key,NESTED_COLLECTION_REFERENCE.id)
    result = collection(COLLECTION_NAME, _client)

    nested_result = result.collection(NESTED_COLLECTION_ID)

    assert nested_result is not None
    assert nested_result.key == NESTED_COLLECTION_REFERENCE.key
    assert nested_result.id == NESTED_COLLECTION_REFERENCE.id




def test_nested_collection_not_found_returns_none()->None:
    gql = Mock(GQLClient)
    _client = Client(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        NESTED_COLLECTION_REFERENCE.key,NESTED_COLLECTION_REFERENCE.id)

    result = collection(COLLECTION_NAME, _client)
    gql.collection_create.return_value = _collection_create_collection_not_found(
        NESTED_COLLECTION_ID)

    nested_result = result.collection(NESTED_COLLECTION_ID)

    assert nested_result is None
