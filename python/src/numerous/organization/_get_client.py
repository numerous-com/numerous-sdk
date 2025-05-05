from __future__ import annotations

from typing import TYPE_CHECKING


if TYPE_CHECKING:
    from ._client import Client

_client: Client | None = None


def get_client() -> Client:
    global _client  # noqa: PLW0603

    if _client is not None:
        return _client

    from numerous._client.factory import graphql_client_from_env

    _client = graphql_client_from_env()

    return _client
