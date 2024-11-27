from __future__ import annotations

import os
from pathlib import Path

from .fs_client import FileSystemClient
from .graphql.client import Client as GQLClient
from .graphql_client import GraphQLClient


_client: FileSystemClient | GraphQLClient | None = None

_DEFAULT_NUMEROUS_API_URL = "https://api.numerous.com/query"


def get_client() -> FileSystemClient | GraphQLClient:
    global _client  # noqa: PLW0603

    if _client is not None:
        return _client

    api_url = os.getenv("NUMEROUS_API_URL", _DEFAULT_NUMEROUS_API_URL)
    organization_id = os.getenv("NUMEROUS_ORGANIZATION_ID")
    access_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")

    if organization_id and access_token:
        gql = GQLClient(url=api_url)
        _client = GraphQLClient(gql, organization_id, access_token)
    else:
        base_path = Path(os.getenv("NUMEROUS_COLLECTIONS_BASE_PATH", "collections"))
        _client = FileSystemClient(base_path)

    return _client
