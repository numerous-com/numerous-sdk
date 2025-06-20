"""Helper functions for bulk collection operations."""

from __future__ import annotations

import json
import logging
from pathlib import Path  # noqa: TCH003
from typing import Any


# Constants
ASCII_CONTROL_CHAR_THRESHOLD = 32


logger = logging.getLogger(__name__)


def _ensure_local_directory(path: Path) -> None:
    """
    Ensure the local directory structure for the given path exists.

    If the directory structure does not exist, it will be created.

    Args:
        path: The path to the directory to create.

    """
    path.mkdir(parents=True, exist_ok=True)


def _is_document_file(file_path: Path, document_suffix: str) -> bool:
    """
    Check if a file has the specified document suffix.

    Args:
        file_path: The path to the file.
        document_suffix: The suffix that identifies document files.

    Returns:
        True if the file has the document suffix, False otherwise.

    """
    return file_path.name.endswith(document_suffix)


def _get_document_key_from_filename(filename: str, document_suffix: str) -> str:
    """
    Extract the document key from a filename by removing the document suffix.

    Args:
        filename: The filename to process.
        document_suffix: The suffix to remove from the filename.

    Returns:
        The document key (filename without the suffix).

    """
    if filename.endswith(document_suffix):
        return filename[: -len(document_suffix)]
    return filename


def _save_document_as_json(doc_data: dict[str, Any], file_path: Path) -> None:
    """
    Save document data as a JSON file.

    The JSON file will be pretty-printed with an indent of 2.
    Parent directories for file_path will be created if they don't exist.

    Args:
        doc_data: The dictionary data to save.
        file_path: The path where the JSON file will be saved.

    """
    _ensure_local_directory(file_path.parent)
    with file_path.open("w", encoding="utf-8") as f:
        json.dump(doc_data, f, indent=2)


def _is_valid_collection_key(key_name: str) -> bool:
    r"""
    Validate if a string is a permissible key for a collection or entity.

    Validation criteria:
    - Not empty and does not contain only whitespace.
    - Not '.' or '..'.
    - Does not start or end with whitespace.
    - Does not contain control characters (ASCII < 32).
    - Does not contain any of the following characters: < > : " | ? * / \ or spaces.

    Args:
        key_name: The key string to validate.

    Returns:
        True if the key is valid, False otherwise.

    """
    if not key_name or key_name.strip() != key_name:
        return False

    if key_name in {".", ".."}:
        return False

    # Check for invalid characters (spaces, slashes, control chars, etc.).
    invalid_chars = r'<>:"|?*/\ '
    return not any(
        (char in invalid_chars) or (ord(char) < ASCII_CONTROL_CHAR_THRESHOLD)
        for char in key_name
    )


def _warn_invalid_key(item_path: Path, reason: str) -> None:
    """
    Log a warning message about an invalid key.

    Args:
        item_path: The path of the item with the invalid key.
        reason: The reason why the key is invalid.

    """
    logger.warning("Skipping '%s'. Reason: %s", item_path, reason)
