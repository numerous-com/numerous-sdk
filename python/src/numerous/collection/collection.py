"""Get or create a collection by name."""

from typing import Optional

from numerous._client._get_client import get_client
from numerous.collection.numerous_collection import NumerousCollection

from ._client import Client


def collection(
    collection_key: str, _client: Optional[Client] = None
) -> NumerousCollection:
    """
    Get or create a root collection by key.

    Use this function as an entry point to interact with collections.
    If the collection does not exist, it is created.

    Args:
        collection_key: Key of the collection. A key is a string that uniquely
            identifies a collection.

    Returns:
        The collection identified by the given key.

    """
    if _client is None:
        _client = get_client()
    collection_ref = _client.get_collection_reference(collection_key)
    return NumerousCollection(collection_ref, _client)
