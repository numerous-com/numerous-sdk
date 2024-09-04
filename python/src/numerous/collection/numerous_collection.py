"""Class for working with numerous collections."""

from typing import Optional

from numerous._client import Client
from numerous.generated.graphql.fragments import CollectionReference


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
