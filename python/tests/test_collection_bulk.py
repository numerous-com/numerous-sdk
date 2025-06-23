from __future__ import annotations

import json
import logging  # For capturing log messages
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest  # Changed from 'from pytest import LogCaptureFixture'
from pathspec import pattern as pathspec_pattern  # For isinstance check

# Removed: from pytest import LogCaptureFixture - use pytest.LogCaptureFixture
# Functions to test from the bulk operations module
from numerous.collections import (
    CollectionReference,
    DocumentReference,
    FileReference,
)
from numerous.collections.bulk.helpers import (
    _ensure_local_directory,
    _get_document_key_from_filename,
    _is_document_file,
    _is_valid_collection_key,
    _save_document_as_json,
    _warn_invalid_key,
)
from numerous.collections.bulk.ignore import (
    _load_ignore_patterns,
    _should_ignore_path,
)
from numerous.collections.bulk.main import (
    bulk_download,
    bulk_upload,
)


"""Tests for the collection bulk operations feature."""

# Test constants
EXPECTED_PATTERNS_COUNT = 3
EXPECTED_FILE_CALL_COUNT = 2


# --- Tests for _ensure_local_directory ---
def test_ensure_local_directory_creates_dir(tmp_path: Path) -> None:
    new_dir = tmp_path / "test_dir"
    _ensure_local_directory(new_dir)
    assert new_dir.is_dir()


def test_ensure_local_directory_exists_ok(tmp_path: Path) -> None:
    existing_dir = tmp_path / "existing_dir"
    existing_dir.mkdir()
    _ensure_local_directory(existing_dir)
    assert existing_dir.is_dir()


def test_ensure_local_directory_creates_nested_dirs(tmp_path: Path) -> None:
    nested_dir = tmp_path / "parent" / "child"
    _ensure_local_directory(nested_dir)
    assert nested_dir.is_dir()


# --- Tests for _is_document_file ---
@pytest.mark.parametrize(
    ("filename", "document_suffix", "expected"),
    [
        ("config.collection-doc.json", ".collection-doc.json", True),
        ("data.collection-doc.json", ".collection-doc.json", True),
        ("settings.doc.json", ".doc.json", True),
        ("data.txt", ".collection-doc.json", False),
        ("config.json", ".collection-doc.json", False),
        ("no_extension", ".collection-doc.json", False),
        ("config.collection-doc.json", ".different-suffix", False),
    ],
)
def test_is_document_file(
    filename: str, document_suffix: str, *, expected: bool
) -> None:  # FBT001
    assert _is_document_file(Path(filename), document_suffix) == expected


# --- Tests for _get_document_key_from_filename ---
@pytest.mark.parametrize(
    ("filename", "document_suffix", "expected"),
    [
        ("config.collection-doc.json", ".collection-doc.json", "config"),
        ("settings.doc.json", ".doc.json", "settings"),
        ("data.txt", ".collection-doc.json", "data.txt"),
        ("config.json", ".collection-doc.json", "config.json"),
        ("file.collection-doc.json", ".different-suffix", "file.collection-doc.json"),
    ],
)
def test_get_document_key_from_filename(
    filename: str, document_suffix: str, expected: str
) -> None:
    assert _get_document_key_from_filename(filename, document_suffix) == expected


# --- Tests for _save_document_as_json ---
def test_save_document_as_json(tmp_path: Path) -> None:
    doc_data = {"key": "value", "number": 123}
    file_path = tmp_path / "output.json"

    _save_document_as_json(doc_data, file_path)

    assert file_path.is_file()
    with file_path.open(encoding="utf-8") as f:  # PTH123
        loaded_data = json.load(f)
    assert loaded_data == doc_data


def test_save_document_as_json_creates_parent_dirs(tmp_path: Path) -> None:
    doc_data = {"test": "data"}
    file_path = tmp_path / "parent_dir" / "output.json"

    _save_document_as_json(doc_data, file_path)

    assert file_path.is_file()
    assert file_path.parent.is_dir()


# --- Tests for _is_valid_collection_key ---
@pytest.mark.parametrize(
    ("key_name", "expected"),
    [
        ("validKey", True),
        ("Valid-Key_123", True),
        ("key with spaces", False),
        (" key_starts_space", False),
        ("key_ends_space ", False),
        ("", False),
        (".", False),
        ("..", False),
        ("key/with/slash", False),
        ("key\\with\\backslash", False),
        ("key<invalid", False),
        ("key>invalid", False),
        ("key:invalid", False),
        ('key"invalid', False),
        ("key|invalid", False),
        ("key?invalid", False),
        ("key*invalid", False),
        ("key\ncontrolchar", False),
        ("a" * 256, True),
    ],
)
def test_is_valid_collection_key(key_name: str, *, expected: bool) -> None:  # FBT001
    assert _is_valid_collection_key(key_name) == expected


# --- Tests for _warn_invalid_key ---
def test_warn_invalid_key(caplog: pytest.LogCaptureFixture) -> None:
    item_path = Path("some/invalid/item.txt")
    reason = "Contains forbidden characters"

    with caplog.at_level(logging.WARNING):
        _warn_invalid_key(item_path, reason)

    assert len(caplog.records) == 1
    assert caplog.records[0].levelname == "WARNING"
    assert f"Skipping '{item_path}'. Reason: {reason}" in caplog.records[0].message


# --- Tests for _load_ignore_patterns ---
def test_load_ignore_patterns_file_exists_and_not_empty(tmp_path: Path) -> None:
    ignore_content = "*.log\nbuild/\n!important.log"
    ignore_file = tmp_path / ".numerous-ignore"
    ignore_file.write_text(ignore_content)

    patterns = _load_ignore_patterns(ignore_file)

    assert len(patterns) == EXPECTED_PATTERNS_COUNT
    assert all(isinstance(p, pathspec_pattern.Pattern) for p in patterns)


def test_load_ignore_patterns_filters_comments_and_empty_lines(tmp_path: Path) -> None:
    ignore_content = (
        "# This is a comment\n"
        "*.tmp\n"
        "\n"
        "cache/\n"
        "   # Another comment with leading spaces\n"
        "   !specific.file  # Comment after pattern with leading spaces"
    )
    ignore_file = tmp_path / ".myignore"
    ignore_file.write_text(ignore_content)

    patterns = _load_ignore_patterns(ignore_file)

    assert len(patterns) == EXPECTED_PATTERNS_COUNT


def test_load_ignore_patterns_file_not_exist(tmp_path: Path) -> None:
    ignore_file = tmp_path / "non_existent_ignore_file"
    patterns = _load_ignore_patterns(ignore_file)
    assert len(patterns) == 0


def test_load_ignore_patterns_file_is_empty(tmp_path: Path) -> None:
    ignore_file = tmp_path / ".empty-ignore"
    ignore_file.write_text("")
    patterns = _load_ignore_patterns(ignore_file)
    assert len(patterns) == 0


# --- Tests for _should_ignore_path ---
@pytest.fixture
def sample_ignore_patterns(tmp_path: Path) -> list[pathspec_pattern.Pattern]:
    """Set up ignore patterns for testing."""
    ignore_content = (
        "*.log\n"
        "temp/\n"
        "*.tmp\n"
        "!important.log\n"
        "node_modules/\n"
        "coverage.xml\n"
        "docs/**/*.md\n"
        "!docs/important_doc.md"
    )
    ignore_file = tmp_path / ".numerous-ignore"
    ignore_file.write_text(ignore_content)
    return _load_ignore_patterns(ignore_file)


@pytest.mark.parametrize(
    ("path_to_check_str", "expected_ignore"),
    [
        ("file.log", True),
        ("important.log", False),
        ("another.txt", False),
        ("temp/file.txt", True),
        ("temp/subdir/another.tmp", True),
        ("my.tmp", True),
        ("sub/my.tmp", True),
        ("node_modules/somelib/file.js", True),
        ("src/app.py", False),
        ("coverage.xml", True),
        ("project/coverage.xml", True),
        ("docs/guide/getting_started.md", True),
        ("docs/index.html", False),
        ("docs/important_doc.md", False),
        ("other/docs/nested/file.md", False),
    ],
)
def test_should_ignore_path(
    sample_ignore_patterns: list[pathspec_pattern.Pattern],
    tmp_path: Path,
    path_to_check_str: str,
    *,
    expected_ignore: bool,
) -> None:
    full_path_to_check = tmp_path / path_to_check_str

    full_path_to_check.parent.mkdir(parents=True, exist_ok=True)
    if path_to_check_str.endswith("/"):
        full_path_to_check.mkdir(exist_ok=True)
    else:
        full_path_to_check.touch(exist_ok=True)

    assert (
        _should_ignore_path(full_path_to_check, sample_ignore_patterns, tmp_path)
        == expected_ignore
    )


def test_should_ignore_path_no_patterns(tmp_path: Path) -> None:
    path_to_check = tmp_path / "some_dir" / "file.txt"
    path_to_check.parent.mkdir(parents=True, exist_ok=True)
    path_to_check.touch(exist_ok=True)
    assert not _should_ignore_path(path_to_check, [], tmp_path)


def test_should_ignore_path_path_not_relative_to_base(
    sample_ignore_patterns: list[pathspec_pattern.Pattern], tmp_path: Path
) -> None:
    if tmp_path.parent == tmp_path:
        pytest.skip(
            "Cannot reliably create a path outside the root directory for this test."
        )
        return

    alt_root = tmp_path.parent
    different_project_path = alt_root / "another_project_area" / "data.log"
    different_project_path.parent.mkdir(parents=True, exist_ok=True)
    different_project_path.touch(exist_ok=True)

    assert not _should_ignore_path(
        different_project_path, sample_ignore_patterns, tmp_path
    )


# --- Mocks for SDK objects ---
@pytest.fixture
def mock_file_ref() -> MagicMock:
    """Provide a MagicMock for a FileReference."""
    mock = MagicMock(spec=FileReference)
    mock.key = "test_file.txt"
    mock.read_bytes.return_value = b"file content"
    return mock


@pytest.fixture
def mock_doc_ref() -> MagicMock:
    """Provide a MagicMock for a DocumentReference."""
    mock = MagicMock(spec=DocumentReference)
    mock.key = "test_doc"
    mock.get.return_value = {"data": "document content"}
    return mock


@pytest.fixture
def mock_collection_ref() -> MagicMock:
    """Provide a MagicMock for a CollectionReference."""
    mock = MagicMock(spec=CollectionReference)
    mock.id = "root_collection_id"
    mock.key = "root_collection_key"
    mock.files.return_value = iter([])
    mock.documents.return_value = iter([])
    mock.collections.return_value = iter([])
    return mock


# --- Tests for bulk_download ---
def test_bulk_download_empty_collection(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    with caplog.at_level(logging.INFO):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    download_path = tmp_path / mock_collection_ref.key
    assert download_path.is_dir()
    assert len(list(download_path.iterdir())) == 0
    assert (
        f"Starting bulk download for collection '{mock_collection_ref.key}'"
        in caplog.text
    )
    assert f"Processing collection '{mock_collection_ref.key}'" in caplog.text
    completed_msg = (
        f"Bulk download for collection '{mock_collection_ref.key}' "
        f"completed successfully."
    )
    assert completed_msg in caplog.text


def test_bulk_download_with_files_and_documents(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    mock_file_ref: MagicMock,
    mock_doc_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    mock_collection_ref.files.return_value = iter([mock_file_ref])
    mock_collection_ref.documents.return_value = iter([mock_doc_ref])

    with caplog.at_level(logging.DEBUG):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    root_download_dir = tmp_path / mock_collection_ref.key
    assert root_download_dir.is_dir()

    local_file = root_download_dir / mock_file_ref.key
    assert local_file.is_file()
    assert local_file.read_bytes() == mock_file_ref.read_bytes.return_value
    mock_file_ref.read_bytes.assert_called_once()

    local_doc = root_download_dir / f"{mock_doc_ref.key}.collection-doc.json"
    assert local_doc.is_file()
    assert json.loads(local_doc.read_text()) == mock_doc_ref.get.return_value
    mock_doc_ref.get.assert_called_once()

    assert (
        f"Downloading file: '{mock_collection_ref.key}/{mock_file_ref.key}'"
        in caplog.text
    )
    assert (
        f"Downloading document: '{mock_collection_ref.key}/{mock_doc_ref.key}'"
        in caplog.text
    )


def test_bulk_download_nested_collections(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    root_file = MagicMock(
        spec=FileReference,
        key="root_file.dat",
        read_bytes=MagicMock(return_value=b"root data"),
    )
    mock_collection_ref.files.return_value = iter([root_file])

    sub_col = MagicMock(spec=CollectionReference)
    sub_col.id = "sub_col_id"
    sub_col.key = "sub_collection"
    sub_col.files.return_value = iter([])
    sub_col.documents.return_value = iter([])

    sub_sub_col = MagicMock(spec=CollectionReference)
    sub_sub_col.id = "sub_sub_col_id"
    sub_sub_col.key = "sub_sub_collection"
    sub_sub_col.files.return_value = iter([])
    sub_sub_col.documents.return_value = iter([])
    sub_sub_col.collections.return_value = iter([])

    sub_col.collections.return_value = iter([sub_sub_col])
    mock_collection_ref.collections.return_value = iter([sub_col])

    with caplog.at_level(logging.INFO):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    root_download_dir = tmp_path / mock_collection_ref.key
    assert (root_download_dir / root_file.key).is_file()
    assert (root_download_dir / root_file.key).read_bytes() == b"root data"

    sub_col_dir = root_download_dir / sub_col.key
    assert sub_col_dir.is_dir()

    sub_sub_col_dir = sub_col_dir / sub_sub_col.key
    assert sub_sub_col_dir.is_dir()
    assert len(list(sub_sub_col_dir.iterdir())) == 0

    assert f"Processing collection '{mock_collection_ref.key}'" in caplog.text
    assert f"Processing collection '{sub_col.key}'" in caplog.text
    assert f"Processing collection '{sub_sub_col.key}'" in caplog.text


def test_bulk_download_file_read_error_continues(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    mock_file_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    error_file_key = "error_file.txt"
    mock_file_ref.key = error_file_key
    mock_file_ref.read_bytes.side_effect = Exception("Simulated read error!")

    good_file_key = "good_file.txt"
    good_file_mock = MagicMock(
        spec=FileReference,
        key=good_file_key,
        read_bytes=MagicMock(return_value=b"good data"),
    )
    mock_collection_ref.files.return_value = iter([mock_file_ref, good_file_mock])

    with caplog.at_level(logging.ERROR):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    root_download_dir = tmp_path / mock_collection_ref.key
    assert not (root_download_dir / error_file_key).exists()
    assert (root_download_dir / good_file_key).is_file()
    assert (root_download_dir / good_file_key).read_bytes() == b"good data"
    assert f"Error downloading file '{error_file_key}'" in caplog.text
    assert "Simulated read error!" in caplog.text


def test_bulk_download_document_get_error_continues(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    mock_doc_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    error_doc_key = "error_doc"
    mock_doc_ref.key = error_doc_key
    mock_doc_ref.get.side_effect = Exception("Simulated get error!")

    good_doc_key = "good_doc"
    good_doc_data = {"status": "ok"}
    good_doc_mock = MagicMock(
        spec=DocumentReference,
        key=good_doc_key,
        get=MagicMock(return_value=good_doc_data),
    )
    mock_collection_ref.documents.return_value = iter([mock_doc_ref, good_doc_mock])

    with caplog.at_level(logging.ERROR):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    root_download_dir = tmp_path / mock_collection_ref.key
    assert not (root_download_dir / f"{error_doc_key}.collection-doc.json").exists()
    good_doc_file = root_download_dir / f"{good_doc_key}.collection-doc.json"
    assert good_doc_file.is_file()
    assert json.loads(good_doc_file.read_text()) == good_doc_data
    assert f"Error downloading document '{error_doc_key}'" in caplog.text
    assert "Simulated get error!" in caplog.text


def test_bulk_download_listing_files_error(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    mock_collection_ref.files.side_effect = Exception("Simulated list files error!")

    with caplog.at_level(logging.ERROR):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    assert (
        f"Error listing files for collection '{mock_collection_ref.key}'" in caplog.text
    )
    assert "Simulated list files error!" in caplog.text


def test_bulk_download_listing_subcollections_error(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    mock_collection_ref.collections.side_effect = Exception(
        "Simulated list sub-collections error!"
    )
    with caplog.at_level(logging.ERROR):
        bulk_download(mock_collection_ref, local_base_path=tmp_path)

    assert (
        f"Error listing sub-collections for collection '{mock_collection_ref.key}'"
        in caplog.text
    )
    assert "Simulated list sub-collections error!" in caplog.text


def test_bulk_download_overwrites_existing_local_files(
    tmp_path: Path, mock_collection_ref: MagicMock, mock_file_ref: MagicMock
) -> None:
    root_download_dir = tmp_path / mock_collection_ref.key
    root_download_dir.mkdir(parents=True, exist_ok=True)

    local_file_path = root_download_dir / mock_file_ref.key
    local_file_path.write_text("old content")

    mock_file_ref.read_bytes.return_value = b"new content"
    mock_collection_ref.files.return_value = iter([mock_file_ref])

    bulk_download(mock_collection_ref, local_base_path=tmp_path)
    assert local_file_path.read_bytes() == b"new content"


# --- Tests for bulk_upload ---
def test_bulk_upload_empty_directory(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    source_base.mkdir()
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir()

    with caplog.at_level(logging.INFO):
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_collection_ref.file.assert_not_called()
    mock_collection_ref.document.assert_not_called()
    mock_collection_ref.collection.assert_not_called()
    assert f"Starting bulk upload from '{collection_upload_dir}'" in caplog.text
    completed_msg = (
        f"Bulk upload from '{collection_upload_dir}' to collection "
        f"'{mock_collection_ref.key}' completed."
    )
    assert completed_msg in caplog.text


def test_bulk_upload_source_directory_not_found(
    tmp_path: Path, mock_collection_ref: MagicMock
) -> None:  # Removed caplog
    non_existent_base = tmp_path / "non_existent_source"
    with pytest.raises(FileNotFoundError) as excinfo:
        bulk_upload(mock_collection_ref, local_base_path=non_existent_base)

    expected_source_dir = non_existent_base / mock_collection_ref.key
    assert str(expected_source_dir) in str(excinfo.value)
    assert (
        f"Source directory '{expected_source_dir}' not found or not a directory."
        in str(excinfo.value)
    )


def test_bulk_upload_with_files_and_json_documents(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir(parents=True, exist_ok=True)

    local_file_content = b"upload data"
    local_file = collection_upload_dir / "upload_me.txt"
    local_file.write_bytes(local_file_content)

    local_doc_content = {"key": "upload_value"}
    local_doc_file = collection_upload_dir / "upload_doc.collection-doc.json"
    local_doc_file.write_text(json.dumps(local_doc_content))

    created_file_ref_mock = MagicMock(spec=FileReference)
    mock_collection_ref.file.return_value = created_file_ref_mock
    created_doc_ref_mock = MagicMock(spec=DocumentReference)
    mock_collection_ref.document.return_value = created_doc_ref_mock

    with caplog.at_level(logging.INFO):
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_collection_ref.file.assert_called_once_with(local_file.name)
    created_file_ref_mock.save.assert_called_once_with(local_file_content)
    mock_collection_ref.document.assert_called_once_with("upload_doc")
    created_doc_ref_mock.set.assert_called_once_with(local_doc_content)
    assert f"Uploaded file '{local_file.name}'" in caplog.text
    assert f"Uploaded document '{local_doc_file.name}'" in caplog.text


def test_bulk_upload_nested_directories_create_collections(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key

    sub_dir_name = "my_sub_folder"
    file_in_sub_name = "data.txt"
    local_sub_dir = collection_upload_dir / sub_dir_name
    local_sub_dir.mkdir(parents=True, exist_ok=True)
    (local_sub_dir / file_in_sub_name).write_text("sub data")

    mock_sub_collection = MagicMock(spec=CollectionReference)
    mock_sub_collection.key = sub_dir_name
    mock_collection_ref.collection.return_value = mock_sub_collection
    mock_file_in_sub_ref = MagicMock(spec=FileReference)
    mock_sub_collection.file.return_value = mock_file_in_sub_ref

    with caplog.at_level(logging.DEBUG):
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_collection_ref.collection.assert_called_once_with(sub_dir_name)
    mock_sub_collection.file.assert_called_once_with(file_in_sub_name)
    mock_file_in_sub_ref.save.assert_called_once_with(b"sub data")
    assert f"Found directory: '{local_sub_dir}'" in caplog.text
    assert f"Creating/getting sub-collection '{sub_dir_name}'" in caplog.text
    assert (
        f"Uploaded file '{file_in_sub_name}' to collection '{sub_dir_name}'"
        in caplog.text
    )


def test_bulk_upload_with_ignore_file(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir(parents=True, exist_ok=True)

    (collection_upload_dir / "file_to_upload.txt").write_text("upload this")
    (collection_upload_dir / "file_to_ignore.log").write_text("ignore this")
    (collection_upload_dir / "another_to_upload.dat").write_text("upload this too")

    ignore_file = collection_upload_dir / ".numerous-ignore"
    ignore_file.write_text("*.log")

    mock_file_upload_ref1 = MagicMock(spec=FileReference)
    mock_file_upload_ref2 = MagicMock(spec=FileReference)
    mock_collection_ref.file.side_effect = [
        mock_file_upload_ref1,
        mock_file_upload_ref2,
    ]

    with caplog.at_level(logging.INFO):
        bulk_upload(
            mock_collection_ref,
            local_base_path=source_base,
            ignore_file_name=".numerous-ignore",
        )

    assert mock_collection_ref.file.call_count == EXPECTED_FILE_CALL_COUNT
    mock_collection_ref.file.assert_any_call("file_to_upload.txt")
    mock_collection_ref.file.assert_any_call("another_to_upload.dat")
    assert f"Loaded 1 ignore patterns from '{ignore_file}'." in caplog.text
    ignore_msg = (
        f"Ignoring '{collection_upload_dir / 'file_to_ignore.log'}' "
        f"due to ignore rules."
    )
    assert ignore_msg in caplog.text
    assert "Uploaded file 'file_to_upload.txt'" in caplog.text
    assert "Uploaded file 'another_to_upload.dat'" in caplog.text
    assert "Uploaded file 'file_to_ignore.log'" not in caplog.text


def test_bulk_upload_invalid_key_name_skipped(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir(parents=True, exist_ok=True)

    problematic_filename = " leading_space.txt"
    (collection_upload_dir / problematic_filename).write_text("problem data")
    (collection_upload_dir / "good_file.txt").write_text("good data")

    mock_good_file_ref = MagicMock(spec=FileReference)
    mock_collection_ref.file.return_value = mock_good_file_ref

    with caplog.at_level(logging.INFO), patch.object(
        Path, "iterdir"
    ) as mock_iterdir:  # SIM117
        mock_iterdir.return_value = iter(
            [
                collection_upload_dir / problematic_filename,
                collection_upload_dir / "good_file.txt",
            ]
        )
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_collection_ref.file.assert_called_once_with("good_file.txt")
    assert "Uploaded file 'good_file.txt'" in caplog.text


def test_bulk_upload_file_write_error_continues(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir(parents=True, exist_ok=True)

    error_file_path = collection_upload_dir / "error_file.txt"
    error_file_path.write_text("error data")
    good_file_path = collection_upload_dir / "good_file.txt"
    good_file_path.write_text("good data")

    mock_error_file_ref = MagicMock(spec=FileReference)
    mock_error_file_ref.save.side_effect = Exception("Simulated API save error!")
    mock_good_file_ref = MagicMock(spec=FileReference)

    def file_side_effect(key: str) -> MagicMock:
        if key == error_file_path.name:
            return mock_error_file_ref
        if key == good_file_path.name:
            return mock_good_file_ref
        pytest.fail(f"Unexpected call to collection.file with key: {key}")
        return MagicMock()

    mock_collection_ref.file.side_effect = file_side_effect

    with caplog.at_level(logging.ERROR):
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_error_file_ref.save.assert_called_once()
    mock_good_file_ref.save.assert_called_once()
    assert f"Error uploading file/document '{error_file_path.name}'" in caplog.text
    assert "Simulated API save error!" in caplog.text


def test_bulk_upload_json_decode_error_skips_file(
    tmp_path: Path,
    mock_collection_ref: MagicMock,
    caplog: pytest.LogCaptureFixture,  # Changed LogCaptureFixture
) -> None:
    source_base = tmp_path / "upload_src"
    collection_upload_dir = source_base / mock_collection_ref.key
    collection_upload_dir.mkdir(parents=True, exist_ok=True)

    bad_json_file = collection_upload_dir / "malformed.collection-doc.json"
    bad_json_file.write_text("{not json")

    good_json_file = collection_upload_dir / "good.collection-doc.json"
    good_json_file.write_text(json.dumps({"status": "fine"}))

    mock_good_doc_ref = MagicMock(spec=DocumentReference)
    mock_collection_ref.document.return_value = mock_good_doc_ref

    with caplog.at_level(logging.ERROR):
        bulk_upload(mock_collection_ref, local_base_path=source_base)

    mock_collection_ref.document.assert_called_once_with("good")
    mock_good_doc_ref.set.assert_called_once_with({"status": "fine"})
    assert f"Error decoding JSON from '{bad_json_file}'" in caplog.text
