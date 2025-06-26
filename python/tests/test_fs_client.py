from __future__ import annotations

import json
from pathlib import Path
from typing import Any

import pytest

from numerous._client.fs_client import FileSystemClient
from numerous._utils.jsonbase64 import dict_to_base64
from numerous.collections._client import (
    CollectionDocumentIdentifier,
    CollectionIdentifier,
    Tag,
)


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


def test_collection_tag_add_adds_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)

    client.collection_tag_add(
        TEST_COLLECTION_ID, Tag(key="added-tag-key", value="added-tag-value")
    )

    metadata_path = base_path / TEST_COLLECTION_ID / ".collection.meta.json"
    assert metadata_path.exists() is True
    metadata_content = json.loads(metadata_path.read_text())
    assert metadata_content["tags"] == [
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_collection_tag_add_adds_tag_to_existing_metadata(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)
    metadata_path = base_path / TEST_COLLECTION_ID / ".collection.meta.json"
    existing_metadata = {
        "collection_id": TEST_COLLECTION_ID,
        "collection_key": _TEST_COLLECTION_KEY,
        "tags": [{"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"}],
    }
    metadata_path.write_text(json.dumps(existing_metadata))

    client.collection_tag_add(
        TEST_COLLECTION_ID, Tag(key="added-tag-key", value="added-tag-value")
    )

    metadata_content = json.loads(metadata_path.read_text())
    assert metadata_content["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_collection_tag_delete_deletes_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)
    metadata_path = base_path / TEST_COLLECTION_ID / ".collection.meta.json"
    existing_metadata = {
        "collection_id": TEST_COLLECTION_ID,
        "collection_key": _TEST_COLLECTION_KEY,
        "tags": [
            {"key": "tag-key", "value": "tag-value"},
            {"key": "tag-to-be-deleted-key", "value": "tag-to-be-deleted-value"},
        ],
    }
    metadata_path.write_text(json.dumps(existing_metadata))

    client.collection_tag_delete(TEST_COLLECTION_ID, "tag-to-be-deleted-key")

    metadata_content = json.loads(metadata_path.read_text())
    assert metadata_content["tags"] == [
        {"key": "tag-key", "value": "tag-value"},
    ]


def test_collection_tag_delete_for_nonexisting_metadata_succeeds(
    client: FileSystemClient,
) -> None:
    client.collection_tag_delete(TEST_COLLECTION_ID, "nonexisting-tag-key")


def test_collection_tags_returns_expected_tags(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)
    metadata_path = base_path / TEST_COLLECTION_ID / ".collection.meta.json"
    existing_metadata = {
        "collection_id": TEST_COLLECTION_ID,
        "collection_key": _TEST_COLLECTION_KEY,
        "tags": [
            {"key": "tag-1", "value": "value-1"},
            {"key": "tag-2", "value": "value-2"},
        ],
    }
    metadata_path.write_text(json.dumps(existing_metadata))

    tags = client.collection_tags(TEST_COLLECTION_ID)

    assert tags == [
        Tag(key="tag-1", value="value-1"),
        Tag(key="tag-2", value="value-2"),
    ]


def test_collection_tags_returns_empty_list_for_nonexisting_metadata(
    client: FileSystemClient,
) -> None:
    tags = client.collection_tags(TEST_COLLECTION_ID)
    assert tags == []


def test_document_reference_returns_expected_existing_document_reference(
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

    doc = client.document_reference(TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc == CollectionDocumentIdentifier(
        id=str(Path(_TEST_COLLECTION_KEY) / _TEST_DOCUMENT_KEY),
        key=_TEST_DOCUMENT_KEY,
    )


def test_document_reference_returns_expected_nested_document_reference(
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

    doc = client.document_reference(_TEST_NESTED_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc == CollectionDocumentIdentifier(
        id=str(Path(_TEST_NESTED_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
        key=_TEST_DOCUMENT_KEY,
    )


def test_document_reference_returns_none_for_nonexisting_document(
    client: FileSystemClient, base_path: Path
) -> None:
    (base_path / _TEST_COLLECTION_KEY).mkdir()

    doc = client.document_reference(TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY)

    assert doc is None


def test_set_document_creates_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    (base_path / _TEST_COLLECTION_KEY).mkdir()

    encoded_data = dict_to_base64(data)
    client.document_set(TEST_COLLECTION_ID, _TEST_DOCUMENT_KEY, encoded_data)

    stored_doc_path = (
        base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    )
    assert stored_doc_path.exists() is True
    assert stored_doc_path.read_text() == json.dumps({"data": data, "tags": []})


def test_document_get_returns_none_for_nonexisting_file(
    client: FileSystemClient,
) -> None:
    assert client.document_get("nonexisting-doc-id") is None


def test_document_get_returns_expected_data_for_existing_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=[]
    )
    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)

    actual = client.document_get(doc_id)

    assert actual == dict_to_base64(data)


def test_document_delete_removes_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = {"field1": 123, "field2": "text"}
    _create_test_document(
        base_path / _TEST_COLLECTION_KEY, _TEST_DOCUMENT_KEY, data=data, tags=[]
    )

    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.document_delete(doc_id)

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    assert doc_path.exists() is False


def test_document_delete_for_nonexisting_succeeds(
    client: FileSystemClient,
) -> None:
    doc_id = str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY)
    client.document_delete(doc_id)


def test_document_tag_add_adds_expected_tag(
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
    client.document_tag_add(doc_id, Tag(key="added-tag-key", value="added-tag-value"))

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_document_tag_delete_deletes_expected_tag(
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
    client.document_tag_delete(doc_id, "tag-to-be-deleted-key")

    doc_path = base_path / _TEST_COLLECTION_KEY / f"{_TEST_DOCUMENT_KEY}.doc.json"
    assert json.loads(doc_path.read_text())["tags"] == [
        {"key": "tag-key", "value": "tag-value"},
    ]


def test_collection_documents_returns_all_documents(
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

    result, has_next_page, end_cursor = client.collection_documents(
        _TEST_COLLECTION_KEY, "", None
    )

    expected = [
        CollectionDocumentIdentifier(
            id=str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
            key=_TEST_DOCUMENT_KEY,
        ),
        CollectionDocumentIdentifier(
            id=str(Path(TEST_COLLECTION_ID) / _TEST_ANOTHER_DOCUMENT_KEY),
            key=_TEST_ANOTHER_DOCUMENT_KEY,
        ),
    ]
    assert sorted(result, key=lambda d: d.key, reverse=True) == expected
    assert has_next_page is False
    assert end_cursor == ""


def test_collection_documents_returns_documents_with_tag(
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

    result, has_next_page, end_cursor = client.collection_documents(
        _TEST_COLLECTION_KEY,
        "",
        Tag(key="tag-key", value="tag-value"),
    )

    assert result == [
        CollectionDocumentIdentifier(
            id=str(Path(TEST_COLLECTION_ID) / _TEST_DOCUMENT_KEY),
            key=_TEST_DOCUMENT_KEY,
        ),
    ]
    assert has_next_page is False
    assert end_cursor == ""


def test_collection_collections_returns_collections_with_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir()
    (base_path / _TEST_NESTED_COLLECTION_ID).mkdir()
    (base_path / _TEST_ANOTHER_NESTED_COLLECTION_ID).mkdir()

    client.collection_tag_add(
        _TEST_NESTED_COLLECTION_ID, Tag(key="environment", value="production")
    )

    collections, has_next_page, end_cursor = client.collection_collections(
        _TEST_COLLECTION_KEY, "", Tag(key="environment", value="production")
    )

    assert has_next_page is False
    assert end_cursor == ""
    assert len(collections) == 1
    assert collections[0].id == _TEST_NESTED_COLLECTION_ID
    assert collections[0].key == _TEST_NESTED_COLLECTION_KEY


def test_collection_collections_returns_expected_collections(
    base_path: Path, client: FileSystemClient
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir()
    (base_path / _TEST_NESTED_COLLECTION_ID).mkdir()
    (base_path / _TEST_ANOTHER_NESTED_COLLECTION_ID).mkdir()

    collections, has_next_page, end_cursor = client.collection_collections(
        _TEST_COLLECTION_KEY, "", None
    )
    expected_number_of_files = 2

    assert has_next_page is False
    assert end_cursor == ""
    assert (
        CollectionIdentifier(
            id=_TEST_NESTED_COLLECTION_ID, key=_TEST_NESTED_COLLECTION_KEY
        )
        in collections
    )
    assert (
        CollectionIdentifier(
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
    _create_test_file(base_path, data=data)

    file = client.file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert file is not None
    assert file.id == TEST_FILE_ID
    assert file.key == TEST_FILE_KEY


def test_get_collection_file_returns_expected_nonexisting_file_reference(
    client: FileSystemClient, base_path: Path
) -> None:
    (base_path / TEST_COLLECTION_ID).mkdir(parents=True)

    file = client.file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert file is not None
    assert file.key == TEST_FILE_KEY


def test_collection_files_returns_all_files(
    base_path: Path, client: FileSystemClient
) -> None:
    test_files = {
        TEST_FILE_ID + "_1": (TEST_FILE_KEY + "_1"),
        TEST_FILE_ID + "_2": (TEST_FILE_KEY + "_2"),
    }
    for file_id, file_key in test_files.items():
        _create_test_file(
            base_path, TEST_COLLECTION_ID, file_key, file_id, "file content"
        )

    result, has_next_page, end_cursor = client.collection_files(
        TEST_COLLECTION_ID, "", None
    )

    assert result is not None
    result_files = {file.id: file.key for file in result if file}
    assert result_files == test_files
    assert has_next_page is False
    assert end_cursor == ""


def test_file_delete_removes_expected_file(
    client: FileSystemClient, base_path: Path
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    _create_test_file(base_path, data=data)
    data_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.data"
    meta_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.meta.json"

    client.file_delete(TEST_FILE_ID)

    assert meta_path.exists() is False
    assert data_path.exists() is False


def test_file_tag_add_adds_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    tags = [{"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"}]
    _create_test_file(base_path, data=data, tags=tags)

    client.file_tag_add(TEST_FILE_ID, Tag(key="added-tag-key", value="added-tag-value"))

    meta_path = base_path / _TEST_COLLECTION_KEY / f"{TEST_FILE_KEY}.file.meta.json"
    assert json.loads(meta_path.read_text())["tags"] == [
        {"key": "pre-existing-tag-key", "value": "pre-existing-tag-value"},
        {"key": "added-tag-key", "value": "added-tag-value"},
    ]


def test_file_delete_tag_deletes_expected_tag(
    base_path: Path, client: FileSystemClient
) -> None:
    data = "File content 1;2;3;4;\n1;2;3;4"
    tags = [
        {"key": "tag-key", "value": "tag-value"},
        {"key": "tag-to-be-deleted-key", "value": "tag-to-be-deleted-value"},
    ]
    _create_test_file(base_path, data=data, tags=tags)

    client.file_delete_tag(TEST_FILE_ID, "tag-to-be-deleted-key")

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
    f = client.file_reference(TEST_COLLECTION_ID, TEST_FILE_KEY)

    assert f is not None
    assert client.file_exists(f.id) is False


def test_file_tags_returns_expected_tags(
    base_path: Path, client: FileSystemClient
) -> None:
    _create_test_file(
        base_path,
        tags=[
            {"key": "tag-1", "value": "value-1"},
            {"key": "tag-2", "value": "value-2"},
        ],
    )

    tags = client.file_tags(TEST_FILE_ID)

    assert tags == {"tag-1": "value-1", "tag-2": "value-2"}


def test_file_tags_returns_non_for_nonexisting_file(
    client: FileSystemClient,
) -> None:
    assert client.file_tags(TEST_FILE_ID) is None


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
