"""Get or create a collection by name."""

from typing import Optional

from numerous._client._get_client import get_client
from numerous.collection.numerous_collection import NumerousCollection

from ._client import Client


def collection(
    collection_key: str, _client: Optional[Client] = None
) -> NumerousCollection:
    """Get or create a collection by name."""
    if _client is None:
        _client = get_client()
    collection_ref = _client.get_collection_reference(collection_key)
    return NumerousCollection(collection_ref, _client)
