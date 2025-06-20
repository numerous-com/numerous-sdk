"""
Core implementation for bulk download and upload operations for Numerous collections.

This module provides the main bulk_download and bulk_upload functions for
efficiently transferring entire collection hierarchies.
"""

from __future__ import annotations

import json
import logging
from pathlib import Path
from typing import TYPE_CHECKING


if TYPE_CHECKING:
    from numerous.collections import CollectionReference

from .helpers import (
    _ensure_local_directory,
    _get_document_key_from_filename,
    _is_document_file,
    _is_valid_collection_key,
    _save_document_as_json,
    _warn_invalid_key,
)
from .ignore import _load_ignore_patterns, _should_ignore_path


logger = logging.getLogger(__name__)

DEFAULT_IGNORE_FILE = ".numerous-ignore"
DEFAULT_DOCUMENT_SUFFIX = ".collection-doc.json"


def bulk_download(  # noqa: C901
    collection_ref: CollectionReference,
    local_base_path: Path = Path("collections"),
    document_suffix: str = DEFAULT_DOCUMENT_SUFFIX,
) -> None:
    """
    Download an entire collection hierarchy recursively to the local filesystem.

    Downloads all files and documents from the given collection and all its
    nested sub-collections. Files are saved as binary, and documents are
    saved as JSON files with the specified suffix.

    The local directory structure will mirror the collection hierarchy, using
    collection keys as folder names. Existing local files will be overwritten.

    Args:
        collection_ref: The `CollectionReference` of the root collection to download.
        local_base_path: Local base directory where collection data will be saved.
                         Defaults to "collections" in the current working directory.
                         The actual download will be inside a subdirectory named
                         after the root collection's key,
                         e.g., `local_base_path/collection_key/...`
        document_suffix: Suffix to append to document filenames when saving locally.
                         Defaults to ".collection-doc.json".

    """
    root_collection_download_path = local_base_path / collection_ref.key

    _ensure_local_directory(root_collection_download_path)
    logger.info(
        "Starting bulk download for collection '%s' into '%s'",
        collection_ref.key,
        root_collection_download_path,
    )

    def _download_recursive(
        current_col: CollectionReference, target_local_dir: Path
    ) -> None:
        """Download a given collection recursively to the target local directory."""
        logger.info(
            "Processing collection '%s' into '%s'",
            current_col.key,
            target_local_dir,
        )
        _ensure_local_directory(target_local_dir)

        try:
            for file_item in current_col.files():
                try:
                    local_file_path = target_local_dir / file_item.key
                    logger.debug(
                        "Downloading file: '%s/%s' to '%s'",
                        current_col.key,
                        file_item.key,
                        local_file_path,
                    )
                    file_content = file_item.read_bytes()
                    with local_file_path.open("wb") as f_out:
                        f_out.write(file_content)
                except Exception:  # noqa: PERF203
                    logger.exception(
                        "Error downloading file '%s' from collection '%s'",
                        file_item.key,
                        current_col.key,
                    )
        except Exception:
            logger.exception("Error listing files for collection '%s'", current_col.key)

        try:
            for doc_item in current_col.documents():
                try:
                    local_doc_path = (
                        target_local_dir / f"{doc_item.key}{document_suffix}"
                    )
                    logger.debug(
                        "Downloading document: '%s/%s' to '%s'",
                        current_col.key,
                        doc_item.key,
                        local_doc_path,
                    )
                    document_data = doc_item.get()
                    if document_data is not None:
                        _save_document_as_json(document_data, local_doc_path)
                    else:
                        logger.warning(
                            "Document '%s' in collection '%s' is empty, skipping save.",
                            doc_item.key,
                            current_col.key,
                        )
                except Exception:  # noqa: PERF203
                    logger.exception(
                        "Error downloading document '%s' from collection '%s'",
                        doc_item.key,
                        current_col.key,
                    )
        except Exception:
            logger.exception(
                "Error listing documents for collection '%s'", current_col.key
            )

        try:
            for sub_col in current_col.collections():
                sub_collection_local_path = target_local_dir / sub_col.key
                _download_recursive(sub_col, sub_collection_local_path)
        except Exception:
            logger.exception(
                "Error listing sub-collections for collection '%s'",
                current_col.key,
            )

    try:
        _download_recursive(collection_ref, root_collection_download_path)
        logger.info(
            "Bulk download for collection '%s' completed successfully.",
            collection_ref.key,
        )
    except Exception as e:
        logger.critical(
            "Bulk download for collection '%s' failed critically: %s",
            collection_ref.key,
            e,
        )
        raise


def bulk_upload(  # noqa: C901, PLR0915
    collection_ref: CollectionReference,
    local_base_path: Path = Path("collections"),
    ignore_file_name: str = DEFAULT_IGNORE_FILE,
    document_suffix: str = DEFAULT_DOCUMENT_SUFFIX,
) -> None:
    """
    Upload a local directory structure recursively to a Numerous collection.

    Scans the specified local directory (expected to be
    `local_base_path/collection_ref.key`) and uploads files and documents.
    It creates nested collections as needed.

    Files with names matching patterns in an ignore file
    (default: .numerous-ignore) will be skipped. Files with the specified
    document suffix will be uploaded as documents. Existing remote files and
    documents will be overwritten.

    Args:
        collection_ref: The `CollectionReference` of the root collection to upload into.
        local_base_path: Local base directory from which to upload. The function
                         expects content to be uploaded to be inside a subdirectory
                         named after the root collection's key,
                         e.g., `local_base_path/collection_key/...`
        ignore_file_name: Name of the ignore file (e.g., ".numerous-ignore")
                          located in the root of the directory being uploaded
                          (i.e., inside `local_base_path/collection_key/`).
        document_suffix: Suffix that identifies files to be uploaded as documents.
                         These files will have the suffix removed from their key.
                         Defaults to ".collection-doc.json".

    """
    source_directory_root = local_base_path / collection_ref.key
    if not source_directory_root.is_dir():
        msg = (
            f"Source directory '{source_directory_root}' not found or not a directory."
        )
        logger.error(
            "Source directory '%s' does not exist or is not a directory. "
            "Aborting upload.",
            source_directory_root,
        )
        raise FileNotFoundError(msg)

    ignore_file_path = source_directory_root / ignore_file_name
    ignore_patterns = _load_ignore_patterns(ignore_file_path)
    if ignore_patterns:
        logger.info(
            "Loaded %d ignore patterns from '%s'.",
            len(ignore_patterns),
            ignore_file_path,
        )
    else:
        logger.info(
            "No ignore patterns loaded (file '%s' not found or empty).",
            ignore_file_path,
        )

    logger.info(
        "Starting bulk upload from '%s' to collection '%s'.",
        source_directory_root,
        collection_ref.key,
    )

    def _upload_recursive(  # noqa: C901
        current_local_dir: Path,
        target_col: CollectionReference,
        effective_ignore_file_path: Path,
    ) -> None:
        """Upload contents of current_local_dir recursively to target_col."""
        logger.info(
            "Processing directory '%s' for upload to collection '%s'.",
            current_local_dir,
            target_col.key,
        )

        for item_path in current_local_dir.iterdir():
            if item_path == effective_ignore_file_path:
                logger.debug("Skipping the ignore file itself: %s", item_path)
                continue

            if _should_ignore_path(item_path, ignore_patterns, source_directory_root):
                logger.info("Ignoring '%s' due to ignore rules.", item_path)
                continue

            item_key = item_path.name
            if not _is_valid_collection_key(item_key):
                _warn_invalid_key(
                    item_path, "Name contains invalid characters or is a reserved name."
                )
                continue

            if item_path.is_dir():
                logger.debug(
                    "Found directory: '%s'. Creating/getting sub-collection '%s'.",
                    item_path,
                    item_key,
                )
                try:
                    sub_collection = target_col.collection(item_key)
                    _upload_recursive(
                        item_path, sub_collection, effective_ignore_file_path
                    )
                except Exception:
                    logger.exception(
                        "Error creating or accessing sub-collection '%s' from '%s'",
                        item_key,
                        target_col.key,
                    )

            elif item_path.is_file():
                logger.debug(
                    "Found file: '%s'. Preparing for upload as '%s'.",
                    item_path,
                    item_key,
                )
                try:
                    if _is_document_file(item_path, document_suffix):
                        logger.debug("Treating '%s' as a document.", item_key)
                        with item_path.open("r", encoding="utf-8") as f_in:
                            try:
                                doc_data = json.load(f_in)
                            except json.JSONDecodeError:
                                logger.exception(
                                    "Error decoding JSON from '%s'. Skipping.",
                                    item_path,
                                )
                                continue

                        doc_key = _get_document_key_from_filename(
                            item_key, document_suffix
                        )
                        doc_ref = target_col.document(doc_key)
                        doc_ref.set(doc_data)
                        logger.info(
                            "Uploaded document '%s' (key: '%s') to collection '%s'.",
                            item_key,
                            doc_key,
                            target_col.key,
                        )
                    else:
                        file_ref = target_col.file(item_key)
                        with item_path.open("rb") as f_in:
                            file_content = f_in.read()
                        file_ref.save(file_content)
                        logger.info(
                            "Uploaded file '%s' to collection '%s'.",
                            item_key,
                            target_col.key,
                        )
                except Exception:
                    logger.exception(
                        "Error uploading file/document '%s' to collection '%s'",
                        item_key,
                        target_col.key,
                    )
            else:
                logger.warning(
                    "Skipping '%s' as it is not a file or directory.", item_path
                )

    try:
        _upload_recursive(source_directory_root, collection_ref, ignore_file_path)
        logger.info(
            "Bulk upload from '%s' to collection '%s' completed.",
            source_directory_root,
            collection_ref.key,
        )
    except Exception as e:
        logger.critical(
            "Bulk upload from '%s' failed critically: %s",
            source_directory_root,
            e,
        )
        raise
