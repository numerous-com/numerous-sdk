import json
from pathlib import Path
from typing import Any

import pytest

from numerous._client._fs_client import FileSystemClient
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionDocumentReferenceTags,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import dict_to_base64


_TEST_COLLECTION_KEY = "collection_key"
_TEST_COLLECTION_ID = _TEST_COLLECTION_KEY

_TEST_NESTED_COLLECTION_KEY = "nested_collection_key"
_TEST_NESTED_COLLECTION_ID = str(
    Path(_TEST_COLLECTION_KEY) / _TEST_NESTED_COLLECTION_KEY
)

_TEST_ANOTHER_NESTED_COLLECTION_KEY = "another_nested_collection_key"
_TEST_ANOTHER_NESTED_COLLECTION_ID = str(
    Path(_TEST_COLLECTION_KEY) / _TEST_ANOTHER_NESTED_COLLECTION_KEY
)

_TEST_ANOTHER_COLLECTION_KEY = "another_collection_key"
_TEST_ANOTHER_COLLECTION_ID = _TEST_ANOTHER_COLLECTION_KEY

_TEST_DOCUMENT_KEY = "document_key"
_TEST_ANOTHER_DOCUMENT_KEY = "another_document_key"


@pytest.fixture
def base_path(tmp_path: Path) -> Path:
    return tmp_path


@pytest.fixture
def client(base_path: Path) -> FileSystemClient:
    return FileSystemClient(base_path)


def test_get_document_returns_expected_existing_document_reference(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    tags = [
        {"key": "tag-1-key", "value": "tag-1-value"},
        {"key": "tag-2-key", "value": "tag-2-value"},
    ]
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=tags
    )

    doc = client.get_collection_document(_TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc == CollectionDocumentReference(
        id=str(Path(_TEST_COLLECTION_KEY) / _TEST_DOCUMENT_KEY),
        key=_TEST_DOCUMENT_KEY,
        data=dict_to_base64(data),
        tags=[
            CollectionDocumentReferenceTags(key="tag-1-key", value="tag-1-value"),
            CollectionDocumentReferenceTags(key="tag-2-key", value="tag-2-value"),
        ],
    )


def test_get_document_returns_expected_nested_existing_document_reference(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    tags = [
        {"key": "tag-1-key", "value": "tag-1-value"},
        {"key": "tag-2-key", "value": "tag-2-value"},
    ]
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY / _TEST_NESTED_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=data,
        tags=tags,
    )

    doc = client.get_collection_document(_TEST_NESTED_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc == CollectionDocumentReference(
        id=str(Path(_TEST_NESTED_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
        key=_TEST_DOCUMENT_KEY,
        data=dict_to_base64(data),
        tags=[
            CollectionDocumentReferenceTags(key="tag-1-key", value="tag-1-value"),
            CollectionDocumentReferenceTags(key="tag-2-key", value="tag-2-value"),
        ],
    )


def test_get_document_returns_expected_none_for_nonexisting_document(
    client: FileSystemClient, base_path: Path
) -> None:
    (base_path / _TEST_COLLECTION_KEY).mkdir()

    doc = client.get_collection_document(_TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc is None


def test_set_document_creates_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    (base_path / _TEST_COLLECTION_KEY).mkdir()

    encoded_data = dict_to_base64(data)
    doc = client.set_collection_document(
        _TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY, encoded_data
    )

    assert doc is not None
    assert doc.data == dict_to_base64(data)
    stored_doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.json"
    assert stored_doc_path.exists() is True
    assert stored_doc_path.read_text() == json.dumps({"data": data, "tags": []})


def test_delete_collection_document_removes_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=[]
    )

    doc_id = str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    doc = client.delete_collection_document(doc_id)

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.json"
    assert doc_path.exists() is False
    assert doc == CollectionDocumentReference(
        id=doc_id,
        data=dict_to_base64(data),
        key=_TEST_DOCUMENT_KEY,
        tags=[],
    )


def test_delete_collection_document_for_nonexisting_returns_none(
    client: FileSystemClient,
) -> None:
    doc_id = str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    doc = client.delete_collection_document(doc_id)

    assert doc is None


def test_add_collection_document_tag_adds_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=data,
        tags=[{"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"}],
    )

    doc_id = str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.add_collection_document_tag(
        doc_id, TagInput(key="added-tag-key", value="added-tag-value")
    )

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_delete_collection_document_tag_deletes_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=data,
        tags=[
            {"key": "tag-key", "value": "tag-value"},
            {"key": "tag-to-be-deleted-key", "value": "tag-to-be-deleted-value"},
        ],
    )

    doc_id = str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.delete_collection_document_tag(doc_id, "tag-to-be-deleted-key")

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "tag-key", "value": "tag-value"},
    ]


def test_get_collection_documents_returns_all_documents(
    base_path: Path, client: FileSystemClient
) -> None:
    test_data = {"name": "test document"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=test_data, tags=[]
    )
    test_another_data = {"name": "another test document"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_ANOTHER_DOCUMENT_KEY,
        data=test_another_data,
        tags=[],
    )

    result, has_next_page, end_cursor = client.get_collection_documents(
        _TEST_COLLECTION_KEY, "", None
    )

    assert result == [
        CollectionDocumentReference(
            id=str(Path(_TEST_COLLECTION_ID) / _TEST_ANOTHER_DOCUMENT_KEY),
            key=_TEST_ANOTHER_DOCUMENT_KEY,
            data=dict_to_base64(test_another_data),
            tags=[],
        ),
        CollectionDocumentReference(
            id=str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
            key=_TEST_DOCUMENT_KEY,
            data=dict_to_base64(test_data),
            tags=[],
        ),
    ]
    assert has_next_page is False
    assert end_cursor == ""


def test_get_collection_documents_returns_documents_with_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    test_tagged_data = {"name": "test document"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=test_tagged_data,
        tags=[{"key": "tag-key", "value": "tag-value"}],
    )
    test_untagged_data = {"name": "another test document"}
    _create_test_file_system_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_ANOTHER_DOCUMENT_KEY,
        data=test_untagged_data,
        tags=[],
    )

    result, has_next_page, end_cursor = client.get_collection_documents(
        _TEST_COLLECTION_KEY,
        "",
        TagInput(key="tag-key", value="tag-value"),
    )

    assert result == [
        CollectionDocumentReference(
            id=str(Path(_TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
            key=_TEST_DOCUMENT_KEY,
            data=dict_to_base64(test_tagged_data),
            tags=[CollectionDocumentReferenceTags(key="tag-key", value="tag-value")],
        ),
    ]
    assert has_next_page is False
    assert end_cursor == ""


def test_get_collection_collections_returns_expected_collections(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / _TEST_COLLECTION_ID).mkdir()
    (base_path / _TEST_NESTED_COLLECTION_ID).mkdir()
    (base_path / _TEST_ANOTHER_NESTED_COLLECTION_ID).mkdir()

    collections, has_next_page, end_cursor = client.get_collection_collections(
        _TEST_COLLECTION_KEY, ""
    )

    assert has_next_page is False
    assert end_cursor == ""
    assert [
        CollectionReference(
            id=_TEST_ANOTHER_NESTED_COLLECTION_ID,
            key=_TEST_ANOTHER_NESTED_COLLECTION_KEY,
        ),
        CollectionReference(
            id=_TEST_NESTED_COLLECTION_ID, key=_TEST_NESTED_COLLECTION_KEY
        ),
    ] == collections


def _create_test_file_system_document(
    collection_path: Path,
    document_key: str,
    data: dict[str, Any],
    tags: list[dict[str, str]],
) -> None:
    collection_path.mkdir(exist_ok=True, parents=True)
    stored_doc_data = json.dumps({"data": data, "tags": tags})
    doc_path = collection_path / f"{document_key}.json"
    doc_path.write_text(stored_doc_data)
