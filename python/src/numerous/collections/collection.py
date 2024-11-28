"""Get or create a collection by name."""

from __future__ import annotations

import typing

import numerous._client.exceptions
from numerous._client.get_client import get_client
from numerous.collections.collection_reference import CollectionReference
from numerous.collections.exceptions import ParentCollectionNotFoundError


if typing.TYPE_CHECKING:
    from ._client import Client


def collection(
    collection_key: str, _client: Client | None = None
) -> CollectionReference:
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
    try:
        collection_ref = _client.get_collection_reference(collection_key)
    except numerous._client.exceptions.ParentCollectionNotFoundError as error:  # noqa: SLF001
        raise ParentCollectionNotFoundError(error.collection_id) from error

    return CollectionReference(collection_ref.id, collection_ref.key, _client)