"""Class for working with numerous collections."""

from numerous.generated.graphql.fragments import CollectionKey
from numerous.numerous_client import NumerousClient


class NumerousCollection:
    def __init__(
        self, collection_key: CollectionKey, numerous_open_client: NumerousClient = None
    )->None:
        self.key = collection_key.key
        self.id = collection_key.id
        self.numerous_open_client = numerous_open_client

    def collection(self, collection_name: str) -> "NumerousCollection":
        """Get or create a collection by name."""
        collection_key = self.numerous_open_client.get_collection_key_with_parent(
            collection_key=collection_name, parent_collection_key=self.id
        )
        return NumerousCollection(collection_key, self.numerous_open_client)
