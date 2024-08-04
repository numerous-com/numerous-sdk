"""Get or create a collection by name."""

from numerous.collection.numerous_collection import NumerousCollection
from numerous.numerous_client import NumerousClient, open_client


def collection(
    collection_name: str, numerous_open_client: NumerousClient = None
) -> NumerousCollection:
    """Get or create a collection by name."""
    if numerous_open_client is None:
        numerous_open_client = open_client("")
    collection_key = numerous_open_client.get_collection_key(collection_name)
    return NumerousCollection(collection_key, numerous_open_client)
