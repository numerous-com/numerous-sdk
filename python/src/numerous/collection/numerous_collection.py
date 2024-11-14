"""Class for working with numerous collections."""

from dataclasses import dataclass
from typing import Generator, Iterator, Optional

from numerous.collection._client import Client
from numerous.collection.numerous_document import NumerousDocument
from numerous.collection.numerous_file import NumerousFile
from numerous.generated.graphql.fragments import CollectionReference
from numerous.generated.graphql.input_types import TagInput


@dataclass
class CollectionNotFoundError(Exception):
    parent_id: Optional[str]
    key: str


class NumerousCollection:
    def __init__(self, collection_ref: CollectionReference, _client: Client) -> None:
        self.key = collection_ref.key
        self.id = collection_ref.id
        self._client = _client

    def collection(self, collection_key: str) -> "NumerousCollection":
        """
        Get or create a child collection of this collection by key.

        Args:
            collection_key: Key of the nested collection. A key uniquely identifies a
                collection within the parent collection. Keys are case sensitive, and
                can be used as human-readable identifiers for collections

        Returns:
            NumerousCollection: The child collection identified by the given key.

        """
        collection_ref = self._client.get_collection_reference(
            collection_key=collection_key, parent_collection_id=self.id
        )

        if collection_ref is None:
            raise CollectionNotFoundError(parent_id=self.id, key=collection_key)

        return NumerousCollection(collection_ref, self._client)

    def document(self, key: str) -> NumerousDocument:
        """
        Get or create a document by key.

        Args:
            key: Key of the document. A key uniquely identifies a document within its
                collection. Keys are case sensitive.

        Returns:
            The document in the collection with the given key.

        """
        numerous_doc_ref = self._client.get_collection_document(self.id, key)
        if numerous_doc_ref is not None:
            numerous_document = NumerousDocument(
                self._client,
                numerous_doc_ref.key,
                (self.id, self.key),
                numerous_doc_ref,
            )
        else:
            numerous_document = NumerousDocument(self._client, key, (self.id, self.key))

        return numerous_document

    def file(self, key: str) -> NumerousFile:
        """
        Get or create a file by key.

        Attributes
        ----------
        key (str): The key of the file.

        """
        numerous_file = self._client.get_collection_file(self.id, key)
        if numerous_file is None:
            msg = "Failed to retrieve or create the file."
            raise ValueError(msg)

        return numerous_file

    def save_file(self, file_key: str, file_data: str) -> None:
        """
        Save data to a file in the collection.

        If the file with the specified key already exists,
        it will be overwritten with the new data.

        Args:
        ----
            file_key (str): The key of the file to save or update.
            file_data (str): The data to be written to the file.

        Raises:
        ------
            ValueError: If the file cannot be created or saved.

        """
        numerous_file = self.file(file_key)
        numerous_file.save(file_data)

    def files(
        self, tag_key: Optional[str] = None, tag_value: Optional[str] = None
    ) -> Iterator[NumerousFile]:
        """
        Retrieve files from the collection, filtered by a tag key and value.

        Parameters
        ----------
        tag_key : Optional[str]
            The key of the tag used to filter files (optional).
        tag_value : Optional[str]
            The value of the tag used to filter files (optional).

        Yields
        ------
        NumerousFile
            Yields NumerousFile objects from the collection.

        """
        end_cursor = ""
        tag_input = None
        if tag_key is not None and tag_value is not None:
            tag_input = TagInput(key=tag_key, value=tag_value)
        has_next_page = True
        while has_next_page:
            result = self._client.get_collection_files(self.key, end_cursor, tag_input)
            if result is None:
                break
            numerous_files, has_next_page, end_cursor = result
            if numerous_files is None:
                break
            for numerous_file in numerous_files:
                if numerous_file is None:
                    continue
                yield numerous_file

    def documents(
        self, tag_key: Optional[str] = None, tag_value: Optional[str] = None
    ) -> Iterator[NumerousDocument]:
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
            tag_input = TagInput(key=tag_key, value=tag_value)
        has_next_page = True
        while has_next_page:
            result = self._client.get_collection_documents(
                self.id, end_cursor, tag_input
            )
            if result is None:
                break
            numerous_doc_refs, has_next_page, end_cursor = result
            if numerous_doc_refs is None:
                break
            for numerous_doc_ref in numerous_doc_refs:
                if numerous_doc_ref is None:
                    continue
                yield NumerousDocument(
                    self._client,
                    numerous_doc_ref.key,
                    (self.id, self.key),
                    numerous_doc_ref,
                )

    def collections(self) -> Generator["NumerousCollection", None, None]:
        """
        Retrieve nested collections from the collection.

        Yields:
            Nested collections of this collection.

        """
        end_cursor = ""
        has_next_page = True
        while has_next_page:
            result = self._client.get_collection_collections(self.id, end_cursor)
            if result is None:
                break
            collection_refs, has_next_page, end_cursor = result
            if collection_refs is None:
                break
            for collection_ref in collection_refs:
                yield NumerousCollection(collection_ref, self._client)
