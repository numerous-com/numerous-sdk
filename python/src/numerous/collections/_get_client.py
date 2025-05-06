from __future__ import annotations

import os
from pathlib import Path
from typing import TYPE_CHECKING


if TYPE_CHECKING:
    from ._client import Client

_client: Client | None = None


def get_client() -> Client:
    global _client  # noqa: PLW0603

    if _client is not None:
        return _client

    from numerous._client.exceptions import (
        APIAccessTokenMissingError,
        OrganizationIDMissingError,
    )
    from numerous._client.factory import graphql_client_from_env

    try:
        _client = graphql_client_from_env()
    except (OrganizationIDMissingError, APIAccessTokenMissingError):
        from numerous._client.fs_client import FileSystemClient

        base_path = Path(os.getenv("NUMEROUS_COLLECTIONS_BASE_PATH", "collections"))
        _client = FileSystemClient(base_path)

    return _client
