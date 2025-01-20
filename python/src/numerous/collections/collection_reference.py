"""Class for working with Numerous collections."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING, Generator

from numerous.collections._client import CollectionDocumentIdentifier, Tag
from numerous.collections.document_reference import DocumentReference
from numerous.collections.file_reference import FileReference


if TYPE_CHECKING:
    from numerous.collections._client import Client, CollectionFileIdentifier


@dataclass
class CollectionReference:
    id: str
    key: str
    _client: Client

    def collection(self, collection_key: str) -> CollectionReference:
        """
        Get or create a child collection of this collection by key.

        Args:
            collection_key: Key of the nested collection. A key uniquely identifies a
                collection within the parent collection. Keys are case sensitive, and
                can be used as human-readable identifiers for collections

        Returns:
            NumerousCollection: The child collection identified by the given key.

        """
        ref = self._client.collection_reference(
            collection_key=collection_key, parent_collection_id=self.id
        )

        return CollectionReference(ref.id, ref.key, self._client)

    def document(self, key: str) -> DocumentReference:
        """
        Get or create a document by key.

        Args:
            key: Key of the document. A key uniquely identifies a document within its
                collection. Keys are case sensitive.

        Returns:
            The document in the collection with the given key.

        Raises:
            ParentCollectionNotFoundError: If the collection does not exist, for example
                if it was deleted.

        """
        doc_ref = self._client.document_reference(self.id, key)
        if doc_ref is None:
            return self._document_reference(doc_id=None, doc_key=key)
        return self._document_reference_from_identifier(doc_ref)

    def _document_reference_from_identifier(
        self, doc_ref: CollectionDocumentIdentifier
    ) -> DocumentReference:
        return self._document_reference(doc_ref.id, doc_ref.key)

    def _document_reference(
        self,
        doc_id: str | None,
        doc_key: str,
    ) -> DocumentReference:
        return DocumentReference(
            id=doc_id,
            key=doc_key,
            collection_id=self.id,
            collection_key=self.key,
            _client=self._client,
        )

    def file(self, key: str) -> FileReference:
        """
        Get or create a file by key.

        Args:
            key: The key of the file.

        """
        file_identifier = self._client.file_reference(self.id, key)
        if file_identifier is None:
            msg = "Failed to retrieve or create the file."
            raise ValueError(msg)

        return self._file_reference_from_identifier(file_identifier)

    def save_file(self, file_key: str, file_data: str) -> None:
        """
        Save data to a file in the collection.

        If the file with the specified key already exists,
        it will be overwritten with the new data.

        Args:
            file_key: The key of the file to save or update.
            file_data: The data to be written to the file.

        Raises:
            ValueError: If the file cannot be created or saved.

        """
        file = self.file(file_key)
        file.save(file_data)

    def files(
        self, tag_key: str | None = None, tag_value: str | None = None
    ) -> Generator[FileReference, None, None]:
        """
        Retrieve files from the collection, filtered by a tag key and value.

        Args:
            tag_key: The key of the tag used to filter files.
            tag_value: The value of the tag used to filter files.

        Yields:
            File references from the collection.

        """
        end_cursor = ""
        tag = None
        if tag_key is not None and tag_value is not None:
            tag = Tag(key=tag_key, value=tag_value)
        has_next_page = True
        while has_next_page:
            result = self._client.collection_files(self.id, end_cursor, tag)
            if result is None:
                break
            file_identifiers, has_next_page, end_cursor = result
            for file_identifier in file_identifiers:
                if file_identifier is None:
                    continue
                yield self._file_reference_from_identifier(file_identifier)

    def documents(
        self, tag_key: str | None = None, tag_value: str | None = None
    ) -> Generator[DocumentReference, None, None]:
        """
        Retrieve documents from the collection, filtered by a tag key and value.

        Args:
            tag_key: If this and `tag_value` is specified, filter documents with this
                tag.
            tag_value: If this and `tag_key` is specified, filter documents with this
                tag.

        Yields:
            Documents from the collection.

        """
        end_cursor = ""
        tag_input = None
        if tag_key is not None and tag_value is not None:
            tag_input = Tag(key=tag_key, value=tag_value)
        has_next_page = True
        while has_next_page:
            result = self._client.collection_documents(self.id, end_cursor, tag_input)
            doc_refs, has_next_page, end_cursor = result
            if doc_refs is None:
                break
            for doc_ref in doc_refs:
                if doc_ref is None:
                    continue
                yield self._document_reference_from_identifier(doc_ref)

    def collections(self) -> Generator[CollectionReference, None, None]:
        """
        Retrieve nested collections from the collection.

        Yields:
            Nested collections of this collection.

        """
        end_cursor = ""
        has_next_page = True
        while has_next_page:
            result = self._client.collection_collections(self.id, end_cursor)
            if result is None:
                break
            refs, has_next_page, end_cursor = result
            if refs is None:
                break
            for ref in refs:
                yield CollectionReference(ref.id, ref.key, self._client)

    def _file_reference_from_identifier(
        self, identifier: CollectionFileIdentifier
    ) -> FileReference:
        return FileReference(id=identifier.id, key=identifier.key, _client=self._client)
