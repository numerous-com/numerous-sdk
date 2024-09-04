"""Get or create a collection by name."""

from typing import Optional

from numerous._client import Client, _open_client
from numerous.collection.numerous_collection import NumerousCollection


def collection(
    collection_key: str, numerous_open_client: Optional[Client] = None
) -> NumerousCollection:
    """Get or create a collection by name."""
    if numerous_open_client is None:
        numerous_open_client = _open_client()
    collection_ref_key = numerous_open_client.get_collection_reference(collection_key)
    return NumerousCollection(collection_ref_key, numerous_open_client)
