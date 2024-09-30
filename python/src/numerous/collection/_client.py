"""
Collections client for managing collections, files and documents.

This module defines a Client protocol which specifies all required methods needed to
manage collections, documents and files.
"""

from typing import Optional, Protocol

from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput


class Client(Protocol):
    def get_collection_reference(
        self, collection_key: str, parent_collection_id: Optional[str] = None
    ) -> CollectionReference: ...

    def get_collection_document(
        self, collection_key: str, document_key: str
    ) -> Optional[CollectionDocumentReference]: ...

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]: ...

    def delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]: ...

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]: ...

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]: ...

    def get_collection_documents(
        self, collection_key: str, end_cursor: str, tag_input: Optional[TagInput]
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]: ...

    def get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[Optional[list[CollectionReference]], bool, str]: ...
