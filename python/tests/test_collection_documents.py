from unittest.mock import Mock

import pytest

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql.collection_create import CollectionCreate
from numerous._client.graphql.collection_document import CollectionDocument
from numerous._client.graphql.collection_document_in_collection import (
    CollectionDocumentInCollection,
)
from numerous._client.graphql.collection_document_set import CollectionDocumentSet
from numerous._client.graphql.collection_document_tag_add import (
    CollectionDocumentTagAdd,
)
from numerous._client.graphql.collection_document_tag_delete import (
    CollectionDocumentTagDelete,
)
from numerous._client.graphql.input_types import TagInput
from numerous._utils.jsonbase64 import dict_to_base64
from numerous.collections import collection
from numerous.collections._client import Client
from numerous.collections.document_reference import DocumentDoesNotExistError


ORGANIZATION_ID = "test-organization-id"
COLLECTION_KEY = "test-collection-name"
COLLECTION_KEY = "test-collection-reference-key"
COLLECTION_ID = "test-collection-reference-id"
DOCUMENT_ID = "test-document-id"
DOCUMENT_KEY = "test-document-key"
DOCUMENT_DATA = {"data": "test data"}
DOCUMENT_DATA_ENCODED = dict_to_base64(DOCUMENT_DATA)
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}


def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "Collection", "key": key, "id": ref_id}}
    )


def _collection_document_set_reference(key: str) -> CollectionDocumentSet:
    return CollectionDocumentSet.model_validate(
        {
            "collectionDocumentSet": {
                "__typename": "CollectionDocument",
                "id": DOCUMENT_ID,
                "key": key,
                "data": DOCUMENT_DATA_ENCODED,
                "tags": [],
            }
        }
    )


def _collection_document_tag_delete_found(_id: str) -> CollectionDocumentTagDelete:
    return CollectionDocumentTagDelete.model_validate(
        {
            "collectionDocumentTagDelete": {
                "__typename": "CollectionDocument",
                "id": _id,
                "key": "t21",
                "data": DOCUMENT_DATA_ENCODED,
                "tags": [],
            }
        }
    )


def _collection_document_tag_add_found(_id: str) -> CollectionDocumentTagAdd:
    return CollectionDocumentTagAdd.model_validate(
        {
            "collectionDocumentTagAdd": {
                "__typename": "CollectionDocument",
                "id": _id,
                "key": "t21",
                "data": DOCUMENT_DATA_ENCODED,
                "tags": [{"key": "key", "value": "test"}],
            }
        }
    )


def _collection_document_in_collection(
    doc_id: str, key: str
) -> CollectionDocumentInCollection:
    return CollectionDocumentInCollection.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "document": {
                    "__typename": "CollectionDocument",
                    "id": doc_id,
                    "key": key,
                },
            }
        }
    )


def _collection_document_in_collection_missing() -> CollectionDocumentInCollection:
    return CollectionDocumentInCollection.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "document": None,
            }
        }
    )


def _collection_document(
    doc_id: str,
    key: str,
    b64data: str,
    tags: list[dict[str, str]],
) -> CollectionDocument:
    return CollectionDocument.model_validate(
        {
            "collectionDocument": {
                "__typename": "CollectionDocument",
                "id": doc_id,
                "key": key,
                "data": b64data,
                "tags": tags,
            }
        }
    )


@pytest.fixture
def gql() -> Mock:
    return Mock(GQLClient)


@pytest.fixture
def client(gql: Mock) -> Client:
    from numerous._client.graphql_client import GraphQLClient

    return GraphQLClient(gql, ORGANIZATION_ID, "token")


def test_document_calls_client(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    col = collection(COLLECTION_KEY, client)

    col.document(DOCUMENT_KEY)

    gql.collection_document_in_collection.assert_called_once_with(
        COLLECTION_ID,
        DOCUMENT_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_document_returns_document_reference(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    col = collection(COLLECTION_KEY, client)

    doc = col.document(DOCUMENT_KEY)

    assert doc.id == DOCUMENT_ID
    assert doc.key == DOCUMENT_KEY
    assert doc.collection_id == COLLECTION_ID
    assert doc.collection_key == COLLECTION_KEY


def test_set_calls_client(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    gql.collection_document_set.return_value = _collection_document_set_reference(
        DOCUMENT_KEY
    )
    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    document.set(DOCUMENT_DATA)

    gql.collection_document_set.assert_called_once_with(
        COLLECTION_ID,
        DOCUMENT_KEY,
        DOCUMENT_DATA_ENCODED,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_get_returns_none_for_nonexisting_document(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection_missing()
    )
    gql.collection_document.return_value = _collection_document(
        DOCUMENT_ID, DOCUMENT_KEY, DOCUMENT_DATA_ENCODED, []
    )
    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)

    result = doc.get()

    assert result is None
    gql.collection_document.assert_not_called()


def test_get_makes_expected_calls(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    gql.collection_document.return_value = _collection_document(
        DOCUMENT_ID, DOCUMENT_KEY, DOCUMENT_DATA_ENCODED, []
    )
    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)

    doc.get()

    gql.collection_document.assert_called_once_with(
        DOCUMENT_ID,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_get_returns_expected_data(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    gql.collection_document.return_value = _collection_document(
        DOCUMENT_ID, DOCUMENT_KEY, DOCUMENT_DATA_ENCODED, []
    )

    data = document.get()

    assert data == DOCUMENT_DATA


def test_delete_calls_client(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    document.delete()

    gql.collection_document_delete.assert_called_once_with(
        DOCUMENT_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_delete_for_nonexisting_raises_document_does_not_exist(
    gql: Mock, client: Client
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection_missing()
    )
    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    with pytest.raises(DocumentDoesNotExistError):
        document.delete()

    gql.collection_document_delete.assert_not_called()


def test_tag_add_calls_client(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)
    gql.collection_document_tag_add.return_value = _collection_document_tag_add_found(
        DOCUMENT_ID
    )

    doc.tag("key", "test")

    gql.collection_document_tag_add.assert_called_once_with(
        DOCUMENT_ID, TagInput(key="key", value="test"), **HEADERS_WITH_AUTHORIZATION
    )


def test_tag_add_for_nonexisting_raises_document_does_not_exist(
    gql: Mock, client: Client
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection_missing()
    )
    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)
    gql.collection_document_tag_add.return_value = _collection_document_tag_add_found(
        DOCUMENT_ID
    )

    with pytest.raises(DocumentDoesNotExistError):
        doc.tag("key", "test")

    gql.collection_document_tag_add.assert_not_called()


def test_tag_delete_calls_client(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    gql.collection_document_tag_delete.return_value = (
        _collection_document_tag_delete_found(DOCUMENT_ID)
    )

    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    document.tag_delete("key")

    gql.collection_document_tag_delete.assert_called_once_with(
        DOCUMENT_ID, "key", **HEADERS_WITH_AUTHORIZATION
    )


def test_tag_delete_for_nonexisting_raises_document_does_not_exist(
    gql: Mock, client: Client
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection_missing()
    )

    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)

    with pytest.raises(DocumentDoesNotExistError):
        doc.tag_delete("key")

    gql.collection_document_tag_delete.assert_not_called()


def test_tags_returns_expected_tags(gql: Mock, client: Client) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection(DOCUMENT_ID, DOCUMENT_KEY)
    )
    gql.collection_document.return_value = _collection_document(
        DOCUMENT_ID,
        DOCUMENT_KEY,
        DOCUMENT_DATA_ENCODED,
        [{"key": "tag-1", "value": "value-1"}, {"key": "tag-2", "value": "value-2"}],
    )

    col = collection(COLLECTION_KEY, client)
    document = col.document(DOCUMENT_KEY)

    tags = document.tags

    assert tags == {"tag-1": "value-1", "tag-2": "value-2"}
    gql.collection_document.assert_called_once_with(
        DOCUMENT_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_tags_for_nonexisting_document_returns_empty_dict(
    gql: Mock, client: Client
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_document_in_collection.return_value = (
        _collection_document_in_collection_missing()
    )

    col = collection(COLLECTION_KEY, client)
    doc = col.document(DOCUMENT_KEY)

    assert doc.tags == {}
    gql.collection_document.assert_not_called()
