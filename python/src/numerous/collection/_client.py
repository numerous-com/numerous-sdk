"""
Collections client for managing collections, files and documents.

This module defines a Client protocol which specifies all required methods needed to
manage collections, documents and files.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, BinaryIO, Protocol


if TYPE_CHECKING:
    from numerous.collection.file_reference import FileReference
    from numerous.generated.graphql.fragments import (
        CollectionDocumentReference,
        CollectionReference,
    )
    from numerous.generated.graphql.input_types import TagInput


class Client(Protocol):
    def get_collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionReference: ...

    def get_collection_document(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentReference | None: ...

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> CollectionDocumentReference | None: ...

    def delete_collection_document(
        self, document_id: str
    ) -> CollectionDocumentReference | None: ...

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> CollectionDocumentReference | None: ...

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> CollectionDocumentReference | None: ...

    def get_collection_documents(
        self, collection_id: str, end_cursor: str, tag_input: TagInput | None
    ) -> tuple[list[CollectionDocumentReference | None] | None, bool, str]: ...

    def get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[list[CollectionReference] | None, bool, str]: ...

    def get_collection_files(
        self, collection_id: str, end_cursor: str, tag_input: TagInput | None
    ) -> tuple[list[FileReference], bool, str]: ...

    def create_collection_file_reference(
        self, collection_id: str, file_key: str
    ) -> FileReference | None: ...

    def collection_file_tags(self, file_id: str) -> dict[str, str] | None: ...

    def delete_collection_file(self, file_id: str) -> None: ...

    def add_collection_file_tag(self, file_id: str, tag: TagInput) -> None: ...

    def delete_collection_file_tag(self, file_id: str, tag_key: str) -> None: ...

    def read_text(self, file_id: str) -> str: ...

    def read_bytes(self, file_id: str) -> bytes: ...

    def save_file(self, file_id: str, data: bytes | str) -> None: ...

    def open_file(self, file_id: str) -> BinaryIO: ...

    def file_exists(self, file_id: str) -> bool: ...
