from __future__ import annotations

from typing import Any
from unittest.mock import Mock, call

import pytest

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql.collection_collections import CollectionCollections
from numerous._client.graphql.collection_create import CollectionCreate
from numerous._client.graphql.collection_documents import CollectionDocuments
from numerous._client.graphql.collection_files import CollectionFiles
from numerous._client.graphql.collection_tag_add import (
    CollectionTagAdd,
    CollectionTagAddCollectionTagAddCollection,
)
from numerous._client.graphql.collection_tag_delete import (
    CollectionTagDelete,
    CollectionTagDeleteCollectionTagDeleteCollection,
)
from numerous._client.graphql.collection_tags import (
    CollectionTags,
    CollectionTagsCollectionCollection,
)
from numerous._client.graphql.input_types import TagInput
from numerous._client.graphql_client import GraphQLClient
from numerous.collections import collection
from numerous.collections.exceptions import ParentCollectionNotFoundError


ORGANIZATION_ID = "test-organization-id"
COLLECTION_KEY = "test-collection-key"
COLLECTION_ID = "test-collection-id"
NESTED_COLLECTION_KEY = "test-nested-collection-key"
NESTED_COLLECTION_ID = "test-nested-collection-id"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}


def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "Collection", "key": key, "id": ref_id}}
    )


def _collection_create_collection_not_found(ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "CollectionNotFound", "id": ref_id}}
    )


def _collection_file_data(
    file_id: str,
    key: str,
    download_url: str | None = None,
    upload_url: str | None = None,
    tags: dict[str, str] | None = None,
) -> dict[str, Any]:
    return {
        "__typename": "CollectionFile",
        "id": file_id,
        "key": key,
        "downloadURL": download_url or "",
        "uploadURL": upload_url or "",
        "tags": [{"key": key, "value": value} for key, value in tags.items()]
        if tags
        else [],
    }


def _collection_collections(
    col_id: str,
    key: str,
    nested_references: list[tuple[str, str]],
    has_next_page: bool,  # noqa: FBT001
    end_cursor: str | None,
) -> CollectionCollections:
    return CollectionCollections.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "id": col_id,
                "key": key,
                "collections": {
                    "edges": [
                        {
                            "node": {
                                "__typename": "Collection",
                                "id": nested_col_id,
                                "key": nested_col_key,
                            }
                        }
                        for nested_col_id, nested_col_key in nested_references
                    ],
                    "pageInfo": {
                        "hasNextPage": has_next_page,
                        "endCursor": end_cursor,
                    },
                },
            }
        }
    )


def _collection_documents(
    col_id: str,
    col_key: str,
    doc_refs: list[tuple[str, str]],
    has_next_page: bool,  # noqa: FBT001
    end_cursor: str | None,
) -> CollectionDocuments:
    return CollectionDocuments.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "id": col_id,
                "key": col_key,
                "documents": {
                    "edges": [
                        {
                            "node": {
                                "__typename": "CollectionDocument",
                                "id": doc_id,
                                "key": doc_key,
                            }
                        }
                        for doc_id, doc_key in doc_refs
                    ],
                    "pageInfo": {
                        "hasNextPage": has_next_page,
                        "endCursor": end_cursor,
                    },
                },
            }
        }
    )


def _collection_files(
    col_id: str,
    col_key: str,
    file_nodes: list[dict[str, Any]],
    has_next_page: bool,  # noqa: FBT001
    end_cursor: str | None,
) -> CollectionFiles:
    return CollectionFiles.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "id": col_id,
                "key": col_key,
                "files": {
                    "edges": [{"node": file_node} for file_node in file_nodes],
                    "pageInfo": {
                        "hasNextPage": has_next_page,
                        "endCursor": end_cursor,
                    },
                },
            }
        }
    )


def _collection_tags_response(
    collection_id: str,
    collection_key: str,
    tags: list[dict[str, str]],
) -> CollectionTags:
    collection_response = CollectionTagsCollectionCollection.model_validate(
        {
            "__typename": "Collection",
            "id": collection_id,
            "key": collection_key,
            "tags": tags,
        }
    )
    return CollectionTags.model_validate({"collection": collection_response})


def _collection_tag_add_response(
    collection_id: str,
    collection_key: str,
) -> CollectionTagAdd:
    collection_response = CollectionTagAddCollectionTagAddCollection.model_validate(
        {
            "__typename": "Collection",
            "id": collection_id,
            "key": collection_key,
        }
    )
    return CollectionTagAdd.model_validate({"collectionTagAdd": collection_response})


def _collection_tag_delete_response(
    collection_id: str,
    collection_key: str,
) -> CollectionTagDelete:
    collection_response = (
        CollectionTagDeleteCollectionTagDeleteCollection.model_validate(
            {
                "__typename": "Collection",
                "id": collection_id,
                "key": collection_key,
            }
        )
    )
    return CollectionTagDelete.model_validate(
        {"collectionTagDelete": collection_response}
    )


@pytest.fixture
def gql() -> Mock:
    return Mock(GQLClient)


@pytest.fixture
def client(gql: Mock) -> GraphQLClient:
    return GraphQLClient(gql, ORGANIZATION_ID, "token")


def test_collection_returns_new_collection(gql: Mock, client: GraphQLClient) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )

    parent_key = None

    result = collection(COLLECTION_KEY, client)

    gql.collection_create.assert_called_once()
    gql.collection_create.assert_called_once_with(
        ORGANIZATION_ID, COLLECTION_KEY, parent_key, **HEADERS_WITH_AUTHORIZATION
    )
    assert result.key == COLLECTION_KEY
    assert result.id == COLLECTION_ID


def test_collection_returns_new_nested_collection(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        NESTED_COLLECTION_KEY,
        NESTED_COLLECTION_ID,
    )
    result = collection(COLLECTION_KEY, client)

    nested_result = result.collection(NESTED_COLLECTION_ID)

    assert nested_result is not None
    assert nested_result.key == NESTED_COLLECTION_KEY
    assert nested_result.id == NESTED_COLLECTION_ID


def test_nested_collection_not_found_raises_parent_not_found_error(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )

    result = collection(COLLECTION_KEY, client)
    gql.collection_create.return_value = _collection_create_collection_not_found(
        COLLECTION_ID
    )

    with pytest.raises(ParentCollectionNotFoundError) as exc_info:
        result.collection(NESTED_COLLECTION_KEY)

    assert exc_info.value.collection_id == COLLECTION_ID


def test_collection_collections_returns_expected_references(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    expected_references = [
        ("nested-id-1", "nested-key-1"),
        ("nested-id-2", "nested-key-2"),
        ("nested-id-3", "nested-key-3"),
    ]
    gql.collection_collections.return_value = _collection_collections(
        COLLECTION_ID,
        COLLECTION_KEY,
        expected_references,
        end_cursor="nested-id-3",
        has_next_page=False,
    )

    col = collection(COLLECTION_KEY, client)
    nested_collections = list(col.collections())

    actual_references = [
        (c.id, c.key) for c in sorted(nested_collections, key=lambda c: c.id)
    ]
    assert actual_references == expected_references
    gql.collection_collections.assert_called_once_with(
        COLLECTION_ID,
        after="",
        first=100,
        tag=None,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_collections_makes_expected_paging_calls(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_collections.side_effect = [
        _collection_collections(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"nested-id-{i}", f"nested-key-{i}") for i in range(1, 101)],
            end_cursor="nested-id-100",
            has_next_page=True,
        ),
        _collection_collections(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"nested-id-{i}", f"nested-key-{i}") for i in range(101, 201)],
            end_cursor="nested-id-200",
            has_next_page=True,
        ),
        _collection_collections(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"nested-id-{i}", f"nested-key-{i}") for i in range(201, 251)],
            end_cursor="nested-id-250",
            has_next_page=False,
        ),
    ]

    col = collection(COLLECTION_KEY, client)
    nested_collections = list(col.collections())

    assert len(nested_collections) == 250  # noqa: PLR2004
    gql.collection_collections.assert_has_calls(
        [
            call(
                COLLECTION_ID,
                after="",
                first=100,
                tag=None,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                after="nested-id-100",
                first=100,
                tag=None,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                after="nested-id-200",
                first=100,
                tag=None,
                **HEADERS_WITH_AUTHORIZATION,
            ),
        ]
    )


def test_collection_collections_passes_tag_filter_to_client(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    col = collection(COLLECTION_KEY, client)

    tag_key = "key"
    tag_value = "value"
    list(col.collections(tag_key=tag_key, tag_value=tag_value))

    gql.collection_collections.assert_called_once_with(
        COLLECTION_ID,
        after="",
        first=100,
        tag=TagInput(key=tag_key, value=tag_value),
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_documents_returns_expected_documents(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    expected_doc_refs = [
        ("doc_id_1", "doc_key_1"),
        ("doc_id_2", "doc_key_2"),
        ("doc_id_3", "doc_key_3"),
    ]
    gql.collection_documents.return_value = _collection_documents(
        COLLECTION_ID,
        COLLECTION_KEY,
        expected_doc_refs,
        has_next_page=False,
        end_cursor=None,
    )
    col = collection(COLLECTION_KEY, client)

    result = list(col.documents())
    actual_doc_refs = [(d.id, d.key) for d in sorted(result, key=lambda d: d.key)]

    assert expected_doc_refs == actual_doc_refs
    gql.collection_documents.assert_called_once_with(
        COLLECTION_ID,
        None,
        after="",
        first=100,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_documents_makes_expected_paging_calls(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_documents.side_effect = [
        _collection_documents(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"doc-id-{i}", f"doc-key-{i}") for i in range(1, 101)],
            end_cursor="doc-id-100",
            has_next_page=True,
        ),
        _collection_documents(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"doc-id-{i}", f"doc-key-{i}") for i in range(101, 201)],
            end_cursor="doc-id-200",
            has_next_page=True,
        ),
        _collection_documents(
            COLLECTION_ID,
            COLLECTION_KEY,
            [(f"doc-id-{i}", f"doc-key-{i}") for i in range(201, 251)],
            end_cursor="doc-id-250",
            has_next_page=False,
        ),
    ]

    col = collection(COLLECTION_KEY, client)
    nested_collections = list(col.documents())

    assert len(nested_collections) == 250  # noqa: PLR2004
    gql.collection_documents.assert_has_calls(
        [
            call(
                COLLECTION_ID,
                None,
                after="",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                None,
                after="doc-id-100",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                None,
                after="doc-id-200",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
        ]
    )


def test_collection_documents_passes_tag_filter_to_client(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    col = collection(COLLECTION_KEY, client)

    tag_key = "key"
    tag_value = "value"
    list(col.documents(tag_key=tag_key, tag_value=tag_value))

    gql.collection_documents.assert_called_once_with(
        COLLECTION_ID,
        TagInput(key=tag_key, value=tag_value),
        after="",
        first=100,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_files_returns_expected_files(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    expected_file_refs = [
        ("file_id_1", "file_key_1"),
        ("file_id_2", "file_key_2"),
        ("file_id_3", "file_key_3"),
    ]
    gql.collection_files.return_value = _collection_files(
        COLLECTION_ID,
        COLLECTION_KEY,
        [
            _collection_file_data(file_id, file_key)
            for file_id, file_key in expected_file_refs
        ],
        has_next_page=False,
        end_cursor=None,
    )
    col = collection(COLLECTION_KEY, client)

    result = list(col.files())
    actual_doc_refs = [(f.id, f.key) for f in sorted(result, key=lambda f: f.id)]

    assert expected_file_refs == actual_doc_refs
    gql.collection_files.assert_called_once_with(
        COLLECTION_ID,
        None,
        after="",
        first=100,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_files_makes_expected_paging_calls(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_files.side_effect = [
        _collection_files(
            COLLECTION_ID,
            COLLECTION_KEY,
            [
                _collection_file_data(f"file-id-{i}", "file-key-{i}")
                for i in range(1, 101)
            ],
            end_cursor="doc-id-100",
            has_next_page=True,
        ),
        _collection_files(
            COLLECTION_ID,
            COLLECTION_KEY,
            [
                _collection_file_data(f"file-id-{i}", "file-key-{i}")
                for i in range(101, 201)
            ],
            end_cursor="doc-id-200",
            has_next_page=True,
        ),
        _collection_files(
            COLLECTION_ID,
            COLLECTION_KEY,
            [
                _collection_file_data(f"file-id-{i}", "file-key-{i}")
                for i in range(201, 251)
            ],
            end_cursor="doc-id-250",
            has_next_page=False,
        ),
    ]

    col = collection(COLLECTION_KEY, client)
    nested_collections = list(col.files())

    assert len(nested_collections) == 250  # noqa: PLR2004
    gql.collection_files.assert_has_calls(
        [
            call(
                COLLECTION_ID,
                None,
                after="",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                None,
                after="doc-id-100",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
            call(
                COLLECTION_ID,
                None,
                after="doc-id-200",
                first=100,
                **HEADERS_WITH_AUTHORIZATION,
            ),
        ]
    )


def test_collection_files_passes_tag_filter_on_to_client(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    tag_key = "key"
    tag_value = "value"

    col = collection(COLLECTION_KEY, client)
    list(col.files(tag_key=tag_key, tag_value=tag_value))

    gql.collection_files.assert_called_once_with(
        COLLECTION_ID,
        TagInput(key=tag_key, value=tag_value),
        after="",
        first=100,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_tags_returns_expected_tags(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    expected_tags = [
        {"key": "tag-1", "value": "value-1"},
        {"key": "tag-2", "value": "value-2"},
    ]
    gql.collection_tags.return_value = _collection_tags_response(
        COLLECTION_ID, COLLECTION_KEY, expected_tags
    )

    col = collection(COLLECTION_KEY, client)
    tags = col.tags

    assert tags == {"tag-1": "value-1", "tag-2": "value-2"}
    gql.collection_tags.assert_called_once_with(
        COLLECTION_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_collection_tag_add_calls_client_with_expected_parameters(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_tag_add.return_value = _collection_tag_add_response(
        COLLECTION_ID, COLLECTION_KEY
    )

    col = collection(COLLECTION_KEY, client)
    col.tag("test-key", "test-value")

    gql.collection_tag_add.assert_called_once_with(
        COLLECTION_ID,
        TagInput(key="test-key", value="test-value"),
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_collection_tag_delete_calls_client_with_expected_parameters(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_tag_delete.return_value = _collection_tag_delete_response(
        COLLECTION_ID, COLLECTION_KEY
    )

    col = collection(COLLECTION_KEY, client)
    col.tag_delete("test-key")

    gql.collection_tag_delete.assert_called_once_with(
        COLLECTION_ID, "test-key", **HEADERS_WITH_AUTHORIZATION
    )
