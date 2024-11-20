from __future__ import annotations

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
TEST_COLLECTION_ID = _TEST_COLLECTION_KEY

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

TEST_FILE_KEY = "file_key"
TEST_FILE_ID = "file_id"


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
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=tags
    )

    doc = client.get_collection_document(TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

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
    _create_test_document(
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

    doc = client.get_collection_document(TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc is None


def test_set_document_creates_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    (base_path / _TEST_COLLECTION_KEY).mkdir()

    encoded_data = dict_to_base64(data)
    doc = client.set_collection_document(
        TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY, encoded_data
    )

    assert doc is not None
    assert doc.data == dict_to_base64(data)
    stored_doc_path = (
        base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    )
    assert stored_doc_path.exists() is True
    assert stored_doc_path.read_text() == json.dumps({"data": data, "tags": []})


def test_delete_collection_document_removes_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=[]
    )

    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    doc = client.delete_collection_document(doc_id)

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
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
    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    doc = client.delete_collection_document(doc_id)

    assert doc is None


def test_add_collection_document_tag_adds_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=data,
        tags=[{"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"}],
    )

    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.add_collection_document_tag(
        doc_id, TagInput(key="added-tag-key", value="added-tag-value")
    )

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_delete_collection_document_tag_deletes_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=data,
        tags=[
            {"key": "tag-key", "value": "tag-value"},
            {"key": "tag-to-be-deleted-key", "value": "tag-to-be-deleted-value"},
        ],
    )

    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.delete_collection_document_tag(doc_id, "tag-to-be-deleted-key")

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "tag-key", "value": "tag-value"},
    ]


def test_get_collection_documents_returns_all_documents(
    base_path: Path, client: FileSystemClient
) -> None:
    test_data = {"name": "test document"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=test_data, tags=[]
    )
    test_another_data = {"name": "another test document"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_ANOTHER_DOCUMENT_KEY,
        data=test_another_data,
        tags=[],
    )
    expected_number_of_files = 2

    result, has_next_page, end_cursor = client.get_collection_documents(
        _TEST_COLLECTION_KEY, "", None
    )

    assert (
        CollectionDocumentReference(
            id=str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
            key=_TEST_DOCUMENT_KEY,
            data=dict_to_base64(test_data),
            tags=[],
        )
        in result
    )
    assert (
        CollectionDocumentReference(
            id=str(Path(TEST_COLLECTION_ID) / _TEST_ANOTHER_DOCUMENT_KEY),
            key=_TEST_ANOTHER_DOCUMENT_KEY,
            data=dict_to_base64(test_another_data),
            tags=[],
        )
        in result
    )
    assert len(result) == expected_number_of_files
    assert has_next_page is False
    assert end_cursor == ""


def test_get_collection_documents_returns_documents_with_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    test_tagged_data = {"name": "test document"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY,
        _TEST_DOCUMENT_KEY,
        data=test_tagged_data,
        tags=[{"key": "tag-key", "value": "tag-value"}],
    )
    test_untagged_data = {"name": "another test document"}
    _create_test_document(
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
            id=str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
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
    (base_path / TEST_COLLECTION_ID).mkdir()
    (base_path / _TEST_NESTED_COLLECTION_ID).mkdir()
    (base_path / _TEST_ANOTHER_NESTED_COLLECTION_ID).mkdir()

    collections, has_next_page, end_cursor = client.get_collection_collections(
        _TEST_COLLECTION_KEY, ""
    )
    expected_number_of_files = 2

    assert has_next_page is False
    assert end_cursor == ""
    assert (
        CollectionReference(
            id=_TEST_NESTED_COLLECTION_ID, key=_TEST_NESTED_COLLECTION_KEY
        )
        in collections
    )
    assert (
        CollectionReference(
            id=_TEST_ANOTHER_NESTED_COLLECTION_ID,
            key=_TEST_ANOTHER_NESTED_COLLECTION_KEY,
        )
        in collections
    )
    assert len(collections) == expected_number_of_files


def test_get_collection_file_returns_expected_existing_file_reference(
    client: FileSystemClient, base_path: Path
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    tags = [
        {"key": "tag-1-key", "value": "tag-1-value"},
        {"key": "tag-2-key", "value": "tag-2-value"},
    ]
    _create_test_file(base_path, data=data, tags=tags)

    file = client.create_collection_file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert file is not None
    assert file.file_id == TEST_FILE_ID
    assert file.key == TEST_FILE_KEY
    assert file.exists is True
    assert file.tags == {"tag-1-key": "tag-1-value", "tag-2-key": "tag-2-value"}


def test_get_collection_file_returns_expected_nonexisting_file_reference(
    client: FileSystemClient, base_path: Path
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)

    file = client.create_collection_file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert file is not None
    assert file.key == TEST_FILE_KEY
    assert file.exists is False
    assert file.tags == {}


def test_get_collection_files_returns_all_files(
    base_path: Path, client: FileSystemClient
) -> None:
    test_files = {
        TEST_FILE_ID + "_1": (TEST_FILE_KEY + "_1", "File content 1;2;3;4;\n1;2;3;4"),
        TEST_FILE_ID + "_2": (TEST_FILE_KEY + "_2", "File content 4;5;6;7;\n4;5;6;7"),
    }
    for file_id, (file_key, data) in test_files.items():
        _create_test_file(base_path, TEST_COLLECTION_ID, file_key, file_id, data)

    result, has_next_page, end_cursor = client.get_collection_files(
        TEST_COLLECTION_ID, "", None
    )

    assert result is not None
    result_files = {
        file.file_id: (file.key, file.read_text()) for file in result if file
    }
    assert result_files == test_files
    assert has_next_page is False
    assert end_cursor == ""


def test_delete_collection_file_removes_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    _create_test_file(base_path, data=data)
    data_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.data"
    meta_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.meta.json"

    client.delete_collection_file(TEST_FILE_ID)

    assert meta_path.exists() is False
    assert data_path.exists() is False


def test_add_collection_file_tag_adds_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    tags = [{"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"}]

    _create_test_file(base_path, data=data, tags=tags)

    meta_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.meta.json"

    client.add_collection_file_tag(
        TEST_FILE_ID, TagInput(key="added-tag-key", value="added-tag-value")
    )

    assert json.loads(meta_path.read_text())["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_delete_collection_file_tag_deletes_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    tags = [
        {"key": "tag-key", "value": "tag-value"},
        {"key": "tag-to-be-deleted-key", "value": "tag-to-be-deleted-value"},
    ]
    _create_test_file(base_path, data=data, tags=tags)

    client.delete_collection_file_tag(TEST_FILE_ID, "tag-to-be-deleted-key")

    meta_path = base_path / TEST_COLLECTION_ID / f"{TEST_FILE_KEY}.file.meta.json"
    assert json.loads(meta_path.read_text())["tags"] == [
        {"key": "tag-key", "value": "tag-value"},
    ]


def test_file_exists_returns_true_for_existing_file(
    base_path: Path, client: FileSystemClient
) -> None:
    _create_test_file(base_path, file_id=TEST_FILE_ID, data="some data")

    assert client.file_exists(TEST_FILE_ID) is True


def test_file_exists_returns_false_for_nonexisting_file(
    client: FileSystemClient,
) -> None:
    assert client.file_exists(TEST_FILE_ID) is False


def test_file_exists_returns_false_for_nonexisting_referenced_file(
    base_path: Path,
    client: FileSystemClient,
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)
    f = client.create_collection_file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert f is not None
    assert client.file_exists(f.file_id) is False


def test_collection_file_tags_returns_expected_tags(
    base_path: Path, client: FileSystemClient
) -> None:
    _create_test_file(
        base_path,
        tags=[
            {"key": "tag-1", "value": "value-1"},
            {"key": "tag-2", "value": "value-2"},
        ],
    )

    tags = client.collection_file_tags(TEST_FILE_ID)

    assert tags == {"tag-1": "value-1", "tag-2": "value-2"}


def test_collection_file_tags_returns_non_for_nonexisting_file(
    client: FileSystemClient,
) -> None:
    assert client.collection_file_tags(TEST_FILE_ID) is None


def _create_test_document(
    collection_path: Path,
    document_key: str,
    data: dict[str, Any],
    tags: list[dict[str, str]],
) -> None:
    collection_path.mkdir(exist_ok=True, parents=True)
    stored_doc_data = json.dumps({"data": data, "tags": tags})
    doc_path = collection_path / f"{document_key}.doc.json"
    doc_path.write_text(stored_doc_data)


def _create_test_file(  # noqa: PLR0913
    base_path: Path,
    collection_id: str = TEST_COLLECTION_ID,
    file_key: str = TEST_FILE_KEY,
    file_id: str = TEST_FILE_ID,
    data: str | None = None,
    tags: list[dict[str, str]] | None = None,
) -> None:
    index_path = base_path / FileSystemClient.FILE_INDEX_DIR
    index_path.mkdir(parents=True, exist_ok=True)
    index_entry_path = index_path / file_id
    index_entry_path.write_text(
        json.dumps({"file_key": file_key, "collection_id": collection_id})
    )
    collection_path = base_path / collection_id
    collection_path.mkdir(parents=True, exist_ok=True)
    meta_path = collection_path / f"{file_key}.file.meta.json"
    meta_path.write_text(
        json.dumps({"file_id": file_id, "file_key": file_key, "tags": tags or []})
    )
    data_path = collection_path / f"{file_key}.file.data"
    if data:
        data_path.write_text(data)
