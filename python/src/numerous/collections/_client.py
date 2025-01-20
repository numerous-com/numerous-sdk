"""
Collections client for managing collections, files, and documents.

This module defines a client protocol, which specifies all required methods needed to
manage collections, documents, and files.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import BinaryIO, Protocol


@dataclass
class Tag:
    key: str
    value: str


@dataclass
class CollectionDocumentIdentifier:
    id: str
    key: str


@dataclass
class CollectionIdentifier:
    id: str
    key: str


@dataclass
class CollectionFileIdentifier:
    id: str
    key: str


class Client(Protocol):
    def collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionIdentifier: ...

    def collection_documents(
        self, collection_id: str, end_cursor: str, tag: Tag | None
    ) -> tuple[list[CollectionDocumentIdentifier] | None, bool, str]: ...

    def collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[list[CollectionIdentifier] | None, bool, str]: ...

    def collection_files(
        self, collection_id: str, end_cursor: str, tag: Tag | None
    ) -> tuple[list[CollectionFileIdentifier], bool, str]: ...

    def document_reference(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentIdentifier | None: ...

    def document_get(self, document_id: str) -> str | None: ...

    def document_exists(self, document_id: str) -> bool: ...

    def document_tags(self, document_id: str) -> dict[str, str] | None: ...

    def document_set(
        self, collection_id: str, document_key: str, document_data: str
    ) -> None: ...

    def document_delete(self, document_id: str) -> None: ...

    def document_tag_add(self, document_id: str, tag: Tag) -> None: ...

    def document_tag_delete(self, document_id: str, tag_key: str) -> None: ...

    def file_reference(
        self, collection_id: str, file_key: str
    ) -> CollectionFileIdentifier | None: ...

    def file_tags(self, file_id: str) -> dict[str, str] | None: ...

    def file_delete(self, file_id: str) -> None: ...

    def file_tag_add(self, file_id: str, tag: Tag) -> None: ...

    def file_delete_tag(self, file_id: str, tag_key: str) -> None: ...

    def file_read_text(self, file_id: str) -> str: ...

    def file_read_bytes(self, file_id: str) -> bytes: ...

    def file_save(self, file_id: str, data: bytes | str) -> None: ...

    def file_open(self, file_id: str) -> BinaryIO: ...

    def file_exists(self, file_id: str) -> bool: ...
