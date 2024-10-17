from unittest.mock import Mock, call

import pytest

from numerous import collection
from numerous._client._graphql_client import COLLECTED_OBJECTS_NUMBER, GraphQLClient
from numerous.collection.exceptions import ParentCollectionNotFoundError
from numerous.collection.numerous_document import NumerousDocument
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_collections import CollectionCollections
from numerous.generated.graphql.collection_create import CollectionCreate
from numerous.generated.graphql.collection_document import CollectionDocument
from numerous.generated.graphql.collection_document_delete import (
    CollectionDocumentDelete,
)
from numerous.generated.graphql.collection_document_set import CollectionDocumentSet
from numerous.generated.graphql.collection_document_tag_add import (
    CollectionDocumentTagAdd,
)
from numerous.generated.graphql.collection_document_tag_delete import (
    CollectionDocumentTagDelete,
)
from numerous.generated.graphql.collection_documents import CollectionDocuments
from numerous.generated.graphql.input_types import TagInput
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


def _collection_document_set_reference(key: str) -> CollectionDocumentSet:
    return CollectionDocumentSet.model_validate(
        {
            "collectionDocumentSet": {
                "__typename": "CollectionDocument",
                "id": DOCUMENT_ID,
                "key": key,
                "data": BASE64_DOCUMENT_DATA,
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
                "data": BASE64_DOCUMENT_DATA,
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
                "data": BASE64_DOCUMENT_DATA,
                "tags": [{"key": "key", "value": "test"}],
            }
        }
    )


def _collection_document_delete_found(_id: str) -> CollectionDocumentDelete:
    return CollectionDocumentDelete.model_validate(
        {
            "collectionDocumentDelete": {
                "__typename": "CollectionDocument",
                "id": _id,
                "key": "t21",
                "data": BASE64_DOCUMENT_DATA,
                "tags": [],
            }
        }
    )


def _collection_collections(_id: str) -> CollectionCollections:
    return CollectionCollections.model_validate(
        {
            "collectionCreate": {
                "__typename": "Collection",
                "id": "1a9299d1-5c81-44bb-b94f-ba40afc05f3a",
                "key": "root_collection",
                "collections": {
                    "edges": [
                        {
                            "node": {
                                "__typename": "Collection",
                                "id": "496da1f7-5378-4962-8373-5c30663848cf",
                                "key": "collection0",
                            }
                        },
                        {
                            "node": {
                                "__typename": "Collection",
                                "id": "6ae8ee18-8ebb-4206-aba1-8d2b44c22682",
                                "key": "collection1",
                            }
                        },
                        {
                            "node": {
                                "__typename": "Collection",
                                "id": "deb5ee57-e4ba-470c-a913-a6a619e9661d",
                                "key": "collection2",
                            }
                        },
                    ],
                    "pageInfo": {
                        "hasNextPage": "false",
                        "endCursor": "deb5ee57-e4ba-470c-a913-a6a619e9661d",
                    },
                },
            }
        }
    )


def _collection_documents_reference(key: str) -> CollectionDocuments:
    return CollectionDocuments.model_validate(
        {
            "collectionCreate": {
                "__typename": "Collection",
                "id": "0d2f82fa-1546-49a4-a034-3392eefc3e4e",
                "key": "t1",
                "documents": {
                    "edges": [
                        {
                            "node": {
                                "__typename": "CollectionDocument",
                                "id": "10634601-67b5-4015-840c-155d9faf9591",
                                "key": key,
                                "data": "ewogICJoZWxsbyI6ICJ3b3JsZCIKfQ==",
                                "tags": [{"key": "key", "value": "test"}],
                            }
                        },
                        {
                            "node": {
                                "__typename": "CollectionDocument",
                                "id": "915b75c5-9e95-4fa7-aaa2-2214c8d251ce",
                                "key": key + "1",
                                "data": "ewogICJoZWxsbyI6ICJ3b3JsZCIKfQ==",
                                "tags": [],
                            }
                        },
                    ],
                    "pageInfo": {
                        "hasNextPage": "false",
                        "endCursor": "915b75c5-9e95-4fa7-aaa2-2214c8d251ce",
                    },
                },
            }
        }
    )


def _collection_document_reference(key: str) -> CollectionDocument:
    return CollectionDocument.model_validate(
        {
            "collectionCreate": {
                "__typename": "Collection",
                "document": {
                    "__typename": "CollectionDocument",
                    "id": DOCUMENT_ID,
                    "key": key,
                    "data": BASE64_DOCUMENT_DATA,
                    "tags": [],
                },
            }
        }
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


def test_collection_document_returns_new_document() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    test_collection = collection(COLLECTION_NAME, _client)

    document = test_collection.document(COLLECTION_DOCUMENT_KEY)

    gql.collection_document.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        COLLECTION_DOCUMENT_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert isinstance(document, NumerousDocument)
    assert document.exists is False


def test_collection_document_returns_existing_document() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document.return_value = _collection_document_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)

    document = test_collection.document(COLLECTION_DOCUMENT_KEY)

    gql.collection_document.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        COLLECTION_DOCUMENT_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert isinstance(document, NumerousDocument)
    assert document.exists


def test_collection_document_set_data_uploads_document() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document_set.return_value = _collection_document_set_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    document = test_collection.document(COLLECTION_DOCUMENT_KEY)
    assert isinstance(document, NumerousDocument)
    assert document.exists is False

    document.set({"test": "test"})

    gql.collection_document_set.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_DOCUMENT_KEY,
        BASE64_DOCUMENT_DATA,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert document.exists


def test_collection_document_get_returns_dict() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document.return_value = _collection_document_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    document = test_collection.document(COLLECTION_DOCUMENT_KEY)

    data = document.get()

    assert isinstance(document, NumerousDocument)
    gql.collection_document.assert_has_calls(
        [
            call(
                ORGANIZATION_ID,
                COLLECTION_REFERENCE_KEY,
                COLLECTION_DOCUMENT_KEY,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                ORGANIZATION_ID,
                COLLECTION_REFERENCE_KEY,
                COLLECTION_DOCUMENT_KEY,
                **HEADERS_WITH_AUTHORIZATION,
            ),
        ]
    )
    assert document.exists
    assert data == DOCUMENT_DATA


def test_collection_document_delete_marks_document_exists_false() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document.return_value = _collection_document_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    document = test_collection.document(COLLECTION_DOCUMENT_KEY)
    assert document.document_id is not None
    gql.collection_document_delete.return_value = _collection_document_delete_found(
        document.document_id
    )
    assert document.exists

    document.delete()

    gql.collection_document_delete.assert_called_once_with(
        DOCUMENT_ID, **HEADERS_WITH_AUTHORIZATION
    )
    assert document.exists is False


def test_collection_document_tag_add() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document.return_value = _collection_document_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    document = test_collection.document(COLLECTION_DOCUMENT_KEY)
    assert document.document_id is not None
    gql.collection_document_tag_add.return_value = _collection_document_tag_add_found(
        document.document_id
    )
    assert document.exists

    document.tag("key", "test")

    gql.collection_document_tag_add.assert_called_once_with(
        DOCUMENT_ID, TagInput(key="key", value="test"), **HEADERS_WITH_AUTHORIZATION
    )
    assert document.tags == {"key": "test"}


def test_collection_document_tag_delete() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_document.return_value = _collection_document_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    document = test_collection.document(COLLECTION_DOCUMENT_KEY)
    assert document.document_id is not None
    gql.collection_document_tag_add.return_value = _collection_document_tag_add_found(
        document.document_id
    )
    gql.collection_document_tag_delete.return_value = (
        _collection_document_tag_delete_found(document.document_id)
    )
    assert document.exists
    document.tag("key", "test")
    assert document.tags == {"key": "test"}

    document.tag_delete("key")

    assert document.tags == {}
    gql.collection_document_tag_delete.assert_called_once_with(
        DOCUMENT_ID, "key", **HEADERS_WITH_AUTHORIZATION
    )


def test_collection_documents_return_more_than_one() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_documents.return_value = _collection_documents_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)

    result = []
    expected_number_of_documents = 2
    for document in test_collection.documents():
        assert document.exists
        result.append(document)

    assert len(result) == expected_number_of_documents
    gql.collection_documents.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        None,
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_documents_query_tag_specific_document() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_documents.return_value = _collection_documents_reference(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)

    tag_key = "key"
    tag_value = "value"
    for document in test_collection.documents(tag_key=tag_key, tag_value=tag_value):
        assert document.exists

    gql.collection_documents.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        TagInput(key=tag_key, value=tag_value),
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_collections_return_more_than_one() -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_collections.return_value = _collection_collections(
        COLLECTION_DOCUMENT_KEY
    )
    test_collection = collection(COLLECTION_NAME, _client)
    result = []
    expected_number_of_collections = 3
    for collection_element in test_collection.collections():
        assert collection_element.key
        result.append(collection_element)

    assert len(result) == expected_number_of_collections
    gql.collection_collections.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )
