from pathlib import Path
from unittest.mock import MagicMock, Mock, patch

import pytest

from numerous import collection
from numerous._client._graphql_client import COLLECTED_OBJECTS_NUMBER, GraphQLClient
from numerous.collection.numerous_file import NumerousFile
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_create import CollectionCreate
from numerous.generated.graphql.collection_file import CollectionFile
from numerous.generated.graphql.collection_file_delete import CollectionFileDelete
from numerous.generated.graphql.collection_file_tag_add import CollectionFileTagAdd
from numerous.generated.graphql.collection_file_tag_delete import (
    CollectionFileTagDelete,
)
from numerous.generated.graphql.collection_files import CollectionFiles
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
COLLECTION_FILE_KEY = "test-file.txt"
DOCUMENT_DATA = {"test": "test"}
BASE64_DOCUMENT_DATA = dict_to_base64(DOCUMENT_DATA)
FILE_ID = "ce5aba38-842d-4ee0-877b-4af9d426c848"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}
_REQUEST_TIMEOUT_SECONDS_ = 1.5


_TEST_DOWNLOAD_URL_ = "http://127.0.0.1:8082/download/collection_files/" + FILE_ID
_TEST_UPLOAD_URL_ = "http://127.0.0.1:8082/upload/collection_files/" + FILE_ID


_TEST_FILE_CONTENT_TEXT_ = "File content 1;2;3;4;\n1;2;3;4"
_TEST_FILE_CONTENT_TEXT_BYTE_ = _TEST_FILE_CONTENT_TEXT_.encode()


def _collection_create_collection_reference(key: str, ref_id: str) -> CollectionCreate:
    return CollectionCreate.model_validate(
        {"collectionCreate": {"typename__": "Collection", "key": key, "id": ref_id}}
    )


def _collection_file_tag_delete_found(_id: str) -> CollectionFileTagDelete:
    return CollectionFileTagDelete.model_validate(
        {
            "collectionFileTagDelete": {
                "__typename": "CollectionFile",
                "id": "0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                "key": "t22",
                "downloadURL": "http://127.0.0.1:8082/download/collection_files/0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                "uploadURL": "http://127.0.0.1:8082/upload/collection_files/0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                "tags": [],
            }
        }
    )


def _collection_file_tag_add_found(_id: str) -> CollectionFileTagAdd:
    return CollectionFileTagAdd.model_validate(
        {
            "collectionFileTagAdd": {
                "__typename": "CollectionFile",
                "id": "0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                "key": "t22",
                "downloadURL": _TEST_DOWNLOAD_URL_,
                "uploadURL": _TEST_UPLOAD_URL_,
                "tags": [{"key": "key", "value": "test"}],
            }
        }
    )


def _collection_file_delete_found(_id: str) -> CollectionFileDelete:
    return CollectionFileDelete.model_validate(
        {
            "collectionFileDelete": {
                "__typename": "CollectionFile",
                "id": _id,
                "key": "t21",
                "downloadURL": _TEST_DOWNLOAD_URL_,
                "uploadURL": _TEST_UPLOAD_URL_,
                "tags": [],
            }
        }
    )


def _collection_files_reference() -> CollectionFiles:
    return CollectionFiles.model_validate(
        {
            "collectionCreate": {
                "__typename": "Collection",
                "id": "0d2f82fa-1546-49a4-a034-3392eefc3e4e",
                "key": "t1",
                "files": {
                    "edges": [
                        {
                            "node": {
                                "__typename": "CollectionFile",
                                "id": "0ac6436b-f044-4616-97c6-2bb5a8dbf7a1",
                                "key": "t22",
                                "downloadURL": _TEST_DOWNLOAD_URL_,
                                "uploadURL": _TEST_UPLOAD_URL_,
                                "tags": [],
                            }
                        },
                        {
                            "node": {
                                "__typename": "CollectionFile",
                                "id": "14ea9afd-41ba-42eb-8a55-314d161e32c6",
                                "key": "t21",
                                "downloadURL": "http://127.0.0.1:8082/download/collection_files/14ea9afd-41ba-42eb-8a55-314d161e32c6",
                                "uploadURL": "http://127.0.0.1:8082/upload/collection_files/14ea9afd-41ba-42eb-8a55-314d161e32c6",
                                "tags": [],
                            }
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


def _collection_file_reference(key: str) -> CollectionFile:
    return CollectionFile.model_validate(
        {
            "collectionFileCreate": {
                "__typename": "CollectionFile",
                "id": FILE_ID,
                "key": key,
                "downloadURL": _TEST_DOWNLOAD_URL_,
                "uploadURL": _TEST_UPLOAD_URL_,
                "tags": [],
            }
        }
    )


def _collection_file_reference_no_urls(key: str) -> CollectionFile:
    return CollectionFile.model_validate(
        {
            "collectionFileCreate": {
                "__typename": "CollectionFile",
                "id": FILE_ID,
                "key": key,
                "downloadURL": "",
                "uploadURL": "",
                "tags": [],
            }
        }
    )


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


@pytest.fixture
def base_path(tmp_path: Path) -> Path:
    return tmp_path


@patch("requests.get")
def test_collection_file_new_file_returns_exists_false(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = ""
    mock_get.return_value = mock_response

    gql.collection_file.return_value = _collection_file_reference_no_urls(
        COLLECTION_FILE_KEY
    )

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)

    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )
    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is False


@patch("requests.get")
def test_collection_file_returns_file_exists_after_load(
    mock_get: MagicMock
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)

    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.get")
def test_collection_file_returns_file_text_content_after_load(
    mock_get: MagicMock,
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_response.text = _TEST_FILE_CONTENT_TEXT_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)
    text = fileref.read_text()

    mock_get.assert_called_once_with(
        _TEST_DOWNLOAD_URL_, timeout=_REQUEST_TIMEOUT_SECONDS_
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert "".join(text) == _TEST_FILE_CONTENT_TEXT_
    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.get")
def test_collection_file_returns_file_byte_content_after_load(
    mock_get: MagicMock,
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)
    bytes_data = fileref.read_bytes()

    mock_get.assert_called_once_with(
        _TEST_DOWNLOAD_URL_, timeout=_REQUEST_TIMEOUT_SECONDS_
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == _TEST_FILE_CONTENT_TEXT_BYTE_
    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.get")
def test_collection_file_returns_file_can_be_opened_after_load(
    mock_get: MagicMock
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)

    with fileref.open() as file:
        bytes_data = file.read()

    mock_get.assert_called_once_with(
        _TEST_DOWNLOAD_URL_, timeout=_REQUEST_TIMEOUT_SECONDS_
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert bytes_data == _TEST_FILE_CONTENT_TEXT_BYTE_
    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.put")
@patch("requests.get")
def test_collection_bytefile_can_be_uploaded_on_save(
    mock_get: MagicMock, mock_put: MagicMock
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_put.return_value = mock_response

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)
    fileref.save(_TEST_FILE_CONTENT_TEXT_BYTE_)

    mock_put.assert_called_once_with(
        _TEST_UPLOAD_URL_,
        files={"file": _TEST_FILE_CONTENT_TEXT_BYTE_},
        timeout=_REQUEST_TIMEOUT_SECONDS_,
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.put")
@patch("requests.get")
def test_collection_textfile_can_be_uploaded_on_save(
    mock_get: MagicMock, mock_put: MagicMock
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_put.return_value = mock_response

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)
    fileref.save(_TEST_FILE_CONTENT_TEXT_)

    mock_put.assert_called_once_with(
        _TEST_UPLOAD_URL_,
        files={"file": _TEST_FILE_CONTENT_TEXT_BYTE_},
        timeout=_REQUEST_TIMEOUT_SECONDS_,
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.put")
@patch("requests.get")
def test_collection_file_can_be_save_from_collection(
    mock_get: MagicMock, mock_put: MagicMock
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_put.return_value = mock_response

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)
    test_collection.save_file(COLLECTION_FILE_KEY, _TEST_FILE_CONTENT_TEXT_)

    mock_put.assert_called_once_with(
        _TEST_UPLOAD_URL_,
        files={"file": _TEST_FILE_CONTENT_TEXT_BYTE_},
        timeout=_REQUEST_TIMEOUT_SECONDS_,
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )


@patch("requests.put")
@patch("requests.get")
def test_collection_file_can_be_uploaded_on_save_open(
    mock_get: MagicMock, mock_put: MagicMock, base_path: Path
) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_put.return_value = mock_response

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)
    fileref = test_collection.file(COLLECTION_FILE_KEY)
    file_name = "file_name"
    _create_test_file_system_file(base_path, file_name, _TEST_FILE_CONTENT_TEXT_BYTE_)

    with Path.open(base_path / f"{file_name}") as f:
        fileref.save_file(f)

    mock_put.assert_called_once_with(
        _TEST_UPLOAD_URL_,
        files={"file": _TEST_FILE_CONTENT_TEXT_BYTE_},
        timeout=_REQUEST_TIMEOUT_SECONDS_,
    )
    gql.collection_file.assert_called_once_with(
        COLLECTION_REFERENCE_ID,
        COLLECTION_FILE_KEY,
        **HEADERS_WITH_AUTHORIZATION,
    )

    assert isinstance(fileref, NumerousFile)
    assert fileref.exists is True


@patch("requests.get")
def test_collection_file_delete_marks_file_exists_false(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    test_collection = collection(COLLECTION_NAME, _client)
    fileref = test_collection.file(COLLECTION_FILE_KEY)
    assert fileref.exists is True
    gql.collection_file_delete.return_value = _collection_file_delete_found(FILE_ID)

    fileref.delete()

    gql.collection_file_delete.assert_called_once_with(
        FILE_ID, **HEADERS_WITH_AUTHORIZATION
    )
    assert fileref.exists is False


def _create_test_file_system_file(
    directory_path: Path, file_name: str, data: bytes
) -> None:
    directory_path.mkdir(exist_ok=True, parents=True)
    path = directory_path / f"{file_name}"
    path.write_bytes(data)


@patch("requests.get")
def test_collection_files_return_more_than_one(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_files.return_value = _collection_files_reference()
    test_collection = collection(COLLECTION_NAME, _client)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    result = []
    expected_number_of_filess = 2
    for file in test_collection.files():
        assert file.exists
        result.append(file)

    assert len(result) == expected_number_of_filess
    gql.collection_files.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        None,
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )


@patch("requests.get")
def test_collection_document_tag_add(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_
    mock_get.return_value = mock_response

    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)

    gql.collection_file_tag_add.return_value = _collection_file_tag_add_found(FILE_ID)
    assert fileref.exists

    fileref.tag("key", "test")

    gql.collection_file_tag_add.assert_called_once_with(
        FILE_ID, TagInput(key="key", value="test"), **HEADERS_WITH_AUTHORIZATION
    )
    assert fileref.tags == {"key": "test"}


@patch("requests.get")
def test_collection_document_tag_delete(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )

    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_
    mock_get.return_value = mock_response

    gql.collection_file.return_value = _collection_file_reference(COLLECTION_FILE_KEY)

    test_collection = collection(COLLECTION_NAME, _client)

    fileref = test_collection.file(COLLECTION_FILE_KEY)

    gql.collection_file_tag_add.return_value = _collection_file_tag_add_found(FILE_ID)
    gql.collection_file_tag_delete.return_value = _collection_file_tag_delete_found(
        FILE_ID
    )
    assert fileref.exists
    fileref.tag("key", "test")
    assert fileref.tags == {"key": "test"}

    fileref.tag_delete("key")

    assert fileref.tags == {}
    gql.collection_file_tag_delete.assert_called_once_with(
        FILE_ID, "key", **HEADERS_WITH_AUTHORIZATION
    )


@patch("requests.get")
def test_collection_files_query_tag_specific_file(mock_get: MagicMock) -> None:
    gql = Mock(GQLClient)
    _client = GraphQLClient(gql)
    gql.collection_create.return_value = _collection_create_collection_reference(
        COLLECTION_REFERENCE_KEY, COLLECTION_REFERENCE_ID
    )
    gql.collection_files.return_value = _collection_files_reference()
    test_collection = collection(COLLECTION_NAME, _client)
    mock_response = MagicMock()
    mock_response.status_code = 200
    mock_response.content = _TEST_FILE_CONTENT_TEXT_BYTE_
    mock_get.return_value = mock_response

    tag_key = "key"
    tag_value = "value"
    for document in test_collection.files(tag_key=tag_key, tag_value=tag_value):
        assert document.exists

    gql.collection_files.assert_called_once_with(
        ORGANIZATION_ID,
        COLLECTION_REFERENCE_KEY,
        TagInput(key=tag_key, value=tag_value),
        after="",
        first=COLLECTED_OBJECTS_NUMBER,
        **HEADERS_WITH_AUTHORIZATION,
    )
