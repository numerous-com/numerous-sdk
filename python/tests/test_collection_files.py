from __future__ import annotations

from typing import TYPE_CHECKING, Any, Generator
from unittest.mock import MagicMock, Mock, patch

import pytest

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql.collection_create import CollectionCreate
from numerous._client.graphql.collection_file import CollectionFile
from numerous._client.graphql.collection_file_create import CollectionFileCreate
from numerous._client.graphql.collection_file_delete import CollectionFileDelete
from numerous._client.graphql.collection_file_tag_add import CollectionFileTagAdd
from numerous._client.graphql.collection_file_tag_delete import (
    CollectionFileTagDelete,
)
from numerous._client.graphql.input_types import TagInput
from numerous._client.graphql_client import GraphQLClient
from numerous._utils.jsonbase64 import dict_to_base64
from numerous.collections import collection
from numerous.collections.file_reference import FileReference


if TYPE_CHECKING:
    from pathlib import Path


ORGANIZATION_ID = "test-org-id"
COLLECTION_KEY = "test-collection-key"
NESTED_COLLECTION_ID = "nested_test_collection"
COLLECTION_KEY = "test_key"
COLLECTION_ID = "test_id"
NESTED_COLLECTION_REFERENCE_KEY = "nested_test_key"
NESTED_COLLECTION_REFERENCE_ID = "nested_test_id"
COLLECTION_DOCUMENT_KEY = "test_document"
FILE_KEY = "test-file.txt"
DOCUMENT_DATA = {"test": "test"}
BASE64_DOCUMENT_DATA = dict_to_base64(DOCUMENT_DATA)
FILE_ID = "ce5aba38-842d-4ee0-877b-4af9d426c848"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}
_REQUEST_TIMEOUT_SECONDS = 1.5


TEST_DOWNLOAD_URL = "http://127.0.0.1:8082/download/collection_files/" + FILE_ID
TEST_UPLOAD_URL = "http://127.0.0.1:8082/upload/collection_files/" + FILE_ID


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


def _collection_file(
    file_id: str,
    key: str,
    download_url: str | None = TEST_DOWNLOAD_URL,
    upload_url: str | None = TEST_UPLOAD_URL,
    tags: dict[str, str] | None = None,
) -> CollectionFile:
    return CollectionFile.model_validate(
        {
            "collectionFile": _collection_file_data(
                file_id, key, download_url, upload_url, tags
            )
        }
    )


def _collection_file_create(
    file_id: str,
    key: str,
    download_url: str | None = TEST_DOWNLOAD_URL,
    upload_url: str | None = TEST_UPLOAD_URL,
    tags: dict[str, str] | None = None,
) -> CollectionFileCreate:
    return CollectionFileCreate.model_validate(
        {
            "collectionFileCreate": _collection_file_data(
                file_id, key, download_url, upload_url, tags
            )
        }
    )


def _collection_file_reference_no_urls(key: str) -> CollectionFileCreate:
    return CollectionFileCreate.model_validate(
        {"collectionFileCreate": _collection_file_data(FILE_ID, key)}
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


@pytest.fixture
def base_path(tmp_path: Path) -> Path:
    return tmp_path


@pytest.fixture
def gql() -> Mock:
    return Mock(GQLClient)


@pytest.fixture
def client(gql: Mock) -> GraphQLClient:
    return GraphQLClient(gql, "test-organization-id", "token")


def test_exists_is_false_when_download_url_is_undefined(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_reference_no_urls(
        FILE_KEY
    )
    gql.collection_file.return_value = _collection_file(
        FILE_ID, FILE_KEY, download_url=None
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)

    assert file.exists is False


def test_exists_is_false_when_download_url_is_defined(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_reference_no_urls(
        FILE_KEY
    )
    gql.collection_file.return_value = _collection_file(
        FILE_ID, FILE_KEY, download_url=TEST_DOWNLOAD_URL
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)

    assert file.exists is True


def test_read_text_returns_expected_text(
    mock_get: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.text = TEST_FILE_TEXT_CONTENT

    col = collection(COLLECTION_KEY, client)

    file = col.file(FILE_KEY)
    text = file.read_text()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert text == TEST_FILE_TEXT_CONTENT


def test_read_bytes_returns_expected_bytes(
    mock_get: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.content = TEST_FILE_BYTES_CONTENT

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    bytes_data = file.read_bytes()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == TEST_FILE_BYTES_CONTENT


def test_open_read_returns_expected_file_content(
    mock_get: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_get.return_value.status_code = 200
    mock_get.return_value.content = TEST_FILE_BYTES_CONTENT

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    with file.open() as fd:
        bytes_data = fd.read()

    mock_get.assert_called_once_with(
        TEST_DOWNLOAD_URL, timeout=_REQUEST_TIMEOUT_SECONDS
    )
    gql.collection_file_create.assert_called_once_with(
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == TEST_FILE_BYTES_CONTENT


def test_save_with_bytes_makes_expected_put_request(
    mock_put: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
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
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert isinstance(file, FileReference)


def test_save_makes_expected_put_request(
    mock_put: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
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
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_file_save_makes_expected_put_request(
    mock_put: MagicMock, gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file.return_value = _collection_file(FILE_ID, FILE_KEY)
    mock_put.return_value.status_code = 200

    col = collection(COLLECTION_KEY, client)
    col.save_file(FILE_KEY, TEST_FILE_TEXT_CONTENT)

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
        COLLECTION_ID,
        FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


def test_delete_calls_expected_mutation(gql: Mock, client: GraphQLClient) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file_delete.return_value = _collection_file_delete_found(FILE_ID)

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    file.delete()

    gql.collection_file_delete.assert_called_once_with(
        FILE_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_tag_add_makes_expected_mutation(gql: Mock, client: GraphQLClient) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file_tag_add.return_value = _collection_file_tag_add_found(FILE_ID)

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    file.tag("key", "test")

    gql.collection_file_tag_add.assert_called_once_with(
        FILE_ID, TagInput(key="key", value="test"), **HEADERS_WITH_AUTHORIZATION
    )


def test_tag_delete_makes_expected_mutation(gql: Mock, client: GraphQLClient) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    gql.collection_file_tag_delete.return_value = _collection_file_tag_delete_found(
        FILE_ID
    )
    tag_key = "key"

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    file.tag_delete(tag_key)

    gql.collection_file_tag_delete.assert_called_once_with(
        FILE_ID, tag_key, **HEADERS_WITH_AUTHORIZATION
    )


def test_tags_property_queries_and_returns_expected_tags(
    gql: Mock, client: GraphQLClient
) -> None:
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_KEY, COLLECTION_ID
    )
    gql.collection_file_create.return_value = _collection_file_create(FILE_ID, FILE_KEY)
    expected_tags = {"tag_1_key": "tag_1_value", "tag_2_key": "tag_2_value"}
    gql.collection_file.return_value = _collection_file(
        FILE_ID, FILE_KEY, tags=expected_tags
    )

    col = collection(COLLECTION_KEY, client)
    file = col.file(FILE_KEY)
    tags = file.tags

    assert tags == expected_tags
