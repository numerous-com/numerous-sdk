"""Get or create a collection by name."""

from typing import Optional

from numerous._client import Client, _open_client
from numerous.collection.numerous_collection import NumerousCollection


def collection(
    collection_key: str, _client: Optional[Client] = None
) -> NumerousCollection:
    """Get or create a collection by name."""
    if _client is None:
        _client = _open_client()
    collection_ref_key = _client.get_collection_reference(collection_key)
    return NumerousCollection(collection_ref_key, _client)
