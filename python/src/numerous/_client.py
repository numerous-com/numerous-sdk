"""GraphQL client wrapper for numerous."""

import asyncio
import os
from typing import Optional, Union

from numerous.generated.graphql.client import Client as GQLCleint
from numerous.generated.graphql.fragments import CollectionNotFound, CollectionReference


API_URL_NOT_SET="NUMEROUS_API_URL environment variable is not set"
MESSAGE_NOT_SET = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"
class Client:
    def __init__(self, client: GQLCleint)->None:
        self.client = client
        self.organization_id = ""
        auth_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")
        if not auth_token:
            raise ValueError(MESSAGE_NOT_SET)

        self.kwargs = {"headers": {"Authorization": f"Bearer {auth_token}"}}

    def _create_collection_ref(self,
                               collection_response:Union[CollectionReference,
                                                         CollectionNotFound])->Optional[CollectionReference]:
        if isinstance(collection_response, CollectionReference):
            return  CollectionReference(id=collection_response.id,
                                        key=collection_response.key)
        return None


    async def _create_collection(
        self, collection_key: str, parent_collection_key: Optional[str] = None
    ) -> Optional[CollectionReference]:
        response = await self.client.collection_create(
            self.organization_id,
            collection_key,
            parent_collection_key,
            kwargs=self.kwargs,
        )
        return self._create_collection_ref(response.collection_create)


    def get_collection_reference(self,
                                  collection_key: str) -> Optional[CollectionReference]:
        """Retrieve a collection by key or create it if it doesn't exist."""
        return asyncio.run(self._create_collection(collection_key))

    def get_collection_reference_with_parent(
        self, collection_ref: str, parent_collection_id: str
    ) -> Optional[CollectionReference]:
        """
        Retrieve a collection by its key and parent key.

        This method retrieves a collection based on its key and parent key,
        or creates it if it doesn't exist.
        """
        return asyncio.run(
            self._create_collection(collection_ref, parent_collection_id)
        )


def _open_client() -> Client:
    url = os.getenv("NUMEROUS_API_URL")
    if not url:
        raise ValueError(API_URL_NOT_SET)
    client = GQLCleint(url=url)
    return Client(client)
