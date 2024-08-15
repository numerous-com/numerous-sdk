"""GraphQL client wrapper for numerous."""

import asyncio
import os
from typing import Optional

from numerous.generated.graphql.client import Client
from numerous.generated.graphql.fragments import CollectionKey

API_URL_NOT_SET="NUMEROUS_API_URL environment variable is not set"
MESSAGE_NOT_SET = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"
class NumerousClient:
    def __init__(self, client: Client)->None:
        self.client = client
        self.organization_id = ""
        auth_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")
        if not auth_token:
            raise ValueError(MESSAGE_NOT_SET)

        self.kwargs = {"headers": {"Authorization": f"Bearer {auth_token}"}}

    async def _create_collection(
        self, collection_key: str, parent_collection_key: Optional[str] = None
    ) -> CollectionKey:
        response = await self.client.collection_create(
            self.organization_id,
            collection_key,
            parent_collection_key,
            kwargs=self.kwargs,
        )
        return CollectionKey(id = response.collection_create,
                             key = response.collection_create.key)

    def _get_collection_key(self, collection_key: str) -> CollectionKey:
        return asyncio.run(self._create_collection(collection_key))

    def _get_collection_key_with_parent(
        self, collection_key: str, parent_collection_key: str
    ) -> CollectionKey:
        return asyncio.run(
            self._create_collection(collection_key, parent_collection_key)
        )


def _open_client() -> NumerousClient:
    url = os.getenv("NUMEROUS_API_URL")
    if not url:
        raise ValueError(API_URL_NOT_SET)

    client = Client(url=url)
    return NumerousClient(client)
