import os
from pathlib import Path
from typing import Optional

from numerous.collection._client import Client
from numerous.generated.graphql.client import Client as GQLClient

from ._fs_client import FileSystemClient
from ._graphql_client import GraphQLClient


_client: Optional[Client] = None


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
