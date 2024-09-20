"""Class for working with numerous collections."""

from typing import Iterator, Optional

from numerous._client import Client
from numerous.collection.numerous_document import NumerousDocument
from numerous.generated.graphql.fragments import CollectionReference
from numerous.generated.graphql.input_types import TagInput


class NumerousCollection:
    def __init__(
        self, collection_ref: Optional[CollectionReference], _client: Client
    ) -> None:
        if collection_ref is not None:
            self.key = collection_ref.key
            self.id = collection_ref.id
        self._client = _client

    def collection(self, collection_name: str) -> Optional["NumerousCollection"]:
        """Get or create a collection by name."""
        collection_ref = self._client.get_collection_reference(
            collection_key=collection_name, parent_collection_id=self.id
        )
        if collection_ref is not None:
            return NumerousCollection(collection_ref, self._client)
        return None

    def document(self, key: str) -> NumerousDocument:
        """
        Get or create a document by key.

        Attributes
        ----------
        key (str): The key of the document.

        """
        numerous_doc_ref = self._client.get_collection_document(self.key, key)
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

    def documents(
        self, tag_key: Optional[str] = None, tag_value: Optional[str] = None
    ) -> Iterator[NumerousDocument]:
        """
        Retrieve documents from the collection, filtered by a tag key and value.

        Parameters
        ----------
        tag_key : Optional[str]
            The key of the tag used to filter documents (optional).
        tag_value : Optional[str]
            The value of the tag used to filter documents (optional).

        Yields
        ------
        NumerousDocument
            Yields NumerousDocument objects from the collection.

        """
        end_cursor = ""
        while True:
            tag_input = None
            if tag_key is not None and tag_value is not None:
                tag_input = TagInput(key=tag_key, value=tag_value)
            result = self._client.get_collection_documents(
                self.key, end_cursor, tag_input
            )
            if result is None:
                break
            numerous_doc_refs, has_next_page, end_cursor = result
            if numerous_doc_refs is None:
                break
            for numerous_doc_ref in numerous_doc_refs:
                if numerous_doc_ref is None:
                    return
                yield NumerousDocument(
                    self._client,
                    numerous_doc_ref.key,
                    (self.id, self.key),
                    numerous_doc_ref,
                )
            if not has_next_page:
                break

    def collections(self) -> Iterator["NumerousCollection"]:
        """
        Retrieve nested collections from the collection.

        Yields
        ------
        NumerousCollection
            Yields NumerousCollection objects.

        """
        end_cursor = ""
        while True:
            result = self._client.get_collection_collections(self.key, end_cursor)
            if result is None:
                break
            collection_ref_keys, has_next_page, end_cursor = result
            if collection_ref_keys is None:
                break
            for collection_ref_key in collection_ref_keys:
                if collection_ref_key is None:
                    return
                yield NumerousCollection(collection_ref_key, self._client)
            if not has_next_page:
                break
