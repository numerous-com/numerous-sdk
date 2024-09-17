"""Class for working with numerous collections."""

from typing import Dict, Optional

from numerous._client import Client
from numerous.collection.numerous_document import NumerousDocument
from numerous.generated.graphql.fragments import CollectionReference


class NumerousCollection:
    def __init__(
        self, collection_ref: Optional[CollectionReference], _client: Client
    ) -> None:
        if collection_ref is not None:
            self.key = collection_ref.key
            self.id = collection_ref.id
        self._client = _client
        self.documents: Dict[str, NumerousDocument] = {}

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
                self._client, numerous_doc_ref.key, self.id, numerous_doc_ref
            )
        else:
            numerous_document = NumerousDocument(self._client, key, self.id)

        self.documents.update({numerous_document.key: numerous_document})
        return numerous_document
