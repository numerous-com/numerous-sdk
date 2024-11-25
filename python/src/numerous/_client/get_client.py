from __future__ import annotations

import os
import typing
from pathlib import Path

from .fs_client import FileSystemClient
from .graphql.client import Client as GQLClient
from .graphql_client import GraphQLClient


if typing.TYPE_CHECKING:
    from numerous.collections._client import Client

_client: Client | None = None


def get_client() -> Client:
    global _client  # noqa: PLW0603

    if _client is not None:
        return _client

    url = os.getenv("NUMEROUS_API_URL")
    if url:
        _client = GraphQLClient(GQLClient(url=url))
    else:
        _client = FileSystemClient(
            Path(os.getenv("NUMEROUS_COLLECTIONS_BASE_PATH", "collections"))
        )

    return _client
