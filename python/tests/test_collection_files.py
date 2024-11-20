from __future__ import annotations

from typing import TYPE_CHECKING, Any, Generator
from unittest.mock import MagicMock, Mock, patch

import pytest

from numerous import collection
from numerous._client._graphql_client import COLLECTED_OBJECTS_NUMBER, GraphQLClient
from numerous.collection.file_reference import FileReference
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_create import CollectionCreate
from numerous.generated.graphql.collection_file import CollectionFile
from numerous.generated.graphql.collection_file_create import CollectionFileCreate
from numerous.generated.graphql.collection_file_delete import CollectionFileDelete
from numerous.generated.graphql.collection_file_tag_add import CollectionFileTagAdd
from numerous.generated.graphql.collection_file_tag_delete import (
    CollectionFileTagDelete,
)
from numerous.generated.graphql.collection_files import CollectionFiles
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import dict_to_base64


if TYPE_CHECKING:
    from pathlib import Path


ORGANIZATION_ID = "test-org-id"
COLLECTION_KEY = "test-collection-key"
NESTED_COLLECTION_ID = "nested_test_collection"
COLLECTION_REFERENCE_KEY = "test_key"
COLLECTION_REFERENCE_ID = "test_id"
NESTED_COLLECTION_REFERENCE_KEY = "nested_test_key"
NESTED_COLLECTION_REFERENCE_ID = "nested_test_id"
COLLECTION_DOCUMENT_KEY = "test_document"
COLLECTION_FILE_KEY = "test-file.txt"
DOCUMENT_DATA = {"test": "test"}
BASE64_DOCUMENT_DATA = dict_to_base64(DOCUMENT_DATA)
TEST_FILE_ID = "ce5aba38-842d-4ee0-877b-4af9d426c848"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}
_REQUEST_TIMEOUT_SECONDS = 1.5


TEST_DOWNLOAD_URL = "http://127.0.0.1:8082/download/collection_files/" + TEST_FILE_ID
TEST_UPLOAD_URL = "http://127.0.0.1:8082/upload/collection_files/" + TEST_FILE_ID


TEST_FILE_TEXT_CONTENT = "File content 1;2;3;4;\n1;2;3;4"
TEST_FILE_BYTES_CONTENT = TEST_FILE_TEXT_CONTENT.encode()


def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "Collection", "key": key, "id": ref_id}}
    )


def _collection_file_tag_delete_found(file_id: str) -> CollectionFileTagDelete:
    return CollectionFileTagDelete.model_validate(
        {
            "collectionFileTagDelete": _collection_file_data(
                file_id,
                "t22",
                "http://127.0.0.1:8082/download/collection_files/0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                "http://127.0.0.1:8082/upload/collection_files/0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
            )
        }
    )


def _collection_file_tag_add_found(file_id: str) -> CollectionFileTagAdd:
    return CollectionFileTagAdd.model_validate(
        {
            "collectionFileTagAdd": _collection_file_data(
                file_id,
                "t22",
                TEST_DOWNLOAD_URL,
                TEST_UPLOAD_URL,
                tags={"key": "test"},
            )
        }
    )


def _collection_file_delete_found(file_id: str) -> CollectionFileDelete:
    return CollectionFileDelete.model_validate(
        {
            "collectionFileDelete": _collection_file_data(
                file_id,
                "t21",
                TEST_DOWNLOAD_URL,
                TEST_UPLOAD_URL,
            )
        }
    )


def _collection_files_reference() -> CollectionFiles:
    return CollectionFiles.model_validate(
        {
            "collection": {
                "__typename": "Collection",
                "id": "0d2f82fa-1546-49a4-a034-3392eefc3e4e",
                "key": "t1",
                "files": {
                    "edges": [
                        {
                            "node": _collection_file_data(
                                "0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                                "t22",
                                TEST_DOWNLOAD_URL,
                                TEST_UPLOAD_URL,
                            )
                        },
                        {
                            "node": _collection_file_data(
                                "14ea9afd-41ba-42eb-8a55-314d161e32c6",
                                "t21",
                                "http://127.0.0.1:8082/download/collection_files/14ea9afd-41ba-42eb-8a55-314d161e32c6",
                                "http://127.0.0.1:8082/upload/collection_files/14ea9afd-41ba-42eb-8a55-314d161e32c6",
                            ),
                        },
                    ],
                    "pageInfo": {
                        "hasNextPage": "false",
                        "endCursor": "14ea9afd-41ba-42eb-8a55-314d161e32c6",
                    },
                },
            }
        }
    )


def _collection_file_reference(
    key: str, tags: dict[str, str] | None = None
) -> CollectionFile:
    return CollectionFile.model_validate(
        {
            "collectionFile": _collection_file_data(
                TEST_FILE_ID, key, TEST_DOWNLOAD_URL, TEST_UPLOAD_URL, tags
            )
        }
    )


def _collection_file_reference_not_found() -> CollectionFile:
    return CollectionFile.model_validate(
        {"collectionFile": {"__typename": "CollectionFileNotFound", "id": TEST_FILE_ID}}
    )


def _collection_file_create_reference(key: str) -> CollectionFileCreate:
    return CollectionFileCreate.model_validate(
        {
            "collectionFileCreate": _collection_file_data(
                TEST_FILE_ID, key, TEST_DOWNLOAD_URL, TEST_UPLOAD_URL
            )
        }
    )


def _collection_file_reference_no_urls(key: str) -> CollectionFileCreate:
    return CollectionFileCreate.model_validate(
        {"collectionFileCreate": _collection_file_data(TEST_FILE_ID, key)}
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


@pytest.fixture
def mock_get() -> Generator[MagicMock, None, None]:
    with patch("requests.get") as m:
        yield m


@pytest.fixture
def mock_put() -> Generator[MagicMock, None, None]:
    with patch("requests.put") as m:
        yield m


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


@pytest.fixture
def base_path(tmp_path: Path) -> Path:
    return tmp_path


def test_exists_is_true_when_file_exists_and_has_download_url() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_reference_no_urls(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)

    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert file.exists is True


def test_file_returns_file_exists_after_load(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.content = TEST_FILE_BYTES_CONTENT
    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)

    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert file.exists is True


def test_read_file_returns_expected_text(
    mock_get: MagicMock,
) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.text = TEST_FILE_TEXT_CONTENT

    col = collection(COLLECTION_KEY, client)

    file = col.file(COLLECTION_FILE_KEY)
    text = file.read_text()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert text == TEST_FILE_TEXT_CONTENT


def test_read_bytes_returns_expected_bytes(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.content = TEST_FILE_BYTES_CONTENT

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    bytes_data = file.read_bytes()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == TEST_FILE_BYTES_CONTENT


def test_open_read_returns_expected_file_content(
    mock_get: MagicMock,
) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.content = TEST_FILE_BYTES_CONTENT

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    with file.open() as fd:
        bytes_data = fd.read()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == TEST_FILE_BYTES_CONTENT


def test_save_with_bytes_makes_put_request(mock_put: MagicMock) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    file.save(TEST_FILE_BYTES_CONTENT)

    mock_put.assert_called_once_with(
        TEST_UPLOAD_URL,
        timeout=_REQUEST_TIMEOUT_SECONDS,
        headers={
            "Content-Type": "application/octet-stream",
            "Content-Length": str(len(TEST_FILE_BYTES_CONTENT)),
        },
        data=TEST_FILE_BYTES_CONTENT,
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert isinstance(file, FileReference)


def test_save_makes_expected_put_request(mock_put: MagicMock) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    file.save(TEST_FILE_TEXT_CONTENT)

    mock_put.assert_called_once_with(
        TEST_UPLOAD_URL,
        timeout=_REQUEST_TIMEOUT_SECONDS,
        headers={
            "Content-Type": "text/plain",
            "Content-Length": str(len(TEST_FILE_TEXT_CONTENT)),
        },
        data=TEST_FILE_BYTES_CONTENT,
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_save_file_makes_expected_put_request(mock_put: MagicMock) -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    col.save_file(COLLECTION_FILE_KEY, TEST_FILE_TEXT_CONTENT)

    mock_put.assert_called_once_with(
        TEST_UPLOAD_URL,
        timeout=_REQUEST_TIMEOUT_SECONDS,
        headers={
            "Content-Type": "text/plain",
            "Content-Length": str(len(TEST_FILE_TEXT_CONTENT)),
        },
        data=TEST_FILE_BYTES_CONTENT,
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_delete_calls_expected_mutation() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file_delete.return_value = _collection_file_delete_found(
        TEST_FILE_ID
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    file.delete()

    gql.collection_file_delete.assert_called_once_with(
        TEST_FILE_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_collection_files_makes_expected_query_and_returns_expected_file_count() -> (
    None
):
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_files.return_value = _collection_files_reference()

    col = collection(COLLECTION_KEY, client)
    result = list(col.files())

    expected_number_of_files = 2
    assert len(result) == expected_number_of_files
    gql.collection_files.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        None,
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_tag_add_makes_expected_mutation() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file_tag_add.return_value = _collection_file_tag_add_found(
        TEST_FILE_ID
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    file.tag("key", "test")

    gql.collection_file_tag_add.assert_called_once_with(
        TEST_FILE_ID, TagInput(key="key", value="test"), **HEADERS_WITH_AUTHORIZATION
    )


def test_tag_delete_makes_expected_mutation() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    gql.collection_file_tag_delete.return_value = _collection_file_tag_delete_found(
        TEST_FILE_ID
    )
    tag_key = "key"

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    file.tag_delete(tag_key)

    gql.collection_file_tag_delete.assert_called_once_with(
        TEST_FILE_ID, tag_key, **HEADERS_WITH_AUTHORIZATION
    )


def test_collection_files_passes_tag_filter_on_to_client() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_files.return_value = _collection_files_reference()
    tag_key = "key"
    tag_value = "value"

    col = collection(COLLECTION_KEY, client)
    list(col.files(tag_key=tag_key, tag_value=tag_value))

    gql.collection_files.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        TagInput(key=tag_key, value=tag_value),
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_tags_property_queries_and_returns_expected_tags() -> None:
    gql = Mock(GQLClient)
    client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file_create.return_value = _collection_file_create_reference(
        COLLECTION_FILE_KEY
    )
    expected_tags = {"tag_1_key": "tag_1_value", "tag_2_key": "tag_2_value"}
    gql.collection_file.return_value = _collection_file_reference(
        COLLECTION_FILE_KEY, tags=expected_tags
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(COLLECTION_FILE_KEY)
    tags = file.tags

    assert tags == expected_tags
