"""GraphQL client wrapper for numerous."""

import asyncio
import os
from typing import Optional, Union

from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_document import (
    CollectionDocumentCollectionCreateCollectionDocument,
    CollectionDocumentCollectionCreateCollectionNotFound,
)
from numerous.generated.graphql.collection_document_delete import (
    CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument,
    CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocumentNotFound,
)
from numerous.generated.graphql.collection_document_set import (
    CollectionDocumentSetCollectionDocumentSetCollectionDocument,
    CollectionDocumentSetCollectionDocumentSetCollectionNotFound,
)
from numerous.generated.graphql.collection_document_tag_add import (
    CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument,
    CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound,
)
from numerous.generated.graphql.collection_document_tag_delete import (
    CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument,
    CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound,
)
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionNotFound,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput



API_URL_NOT_SET = "NUMEROUS_API_URL environment variable is not set"
MESSAGE_NOT_SET = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"


class Client:
    def __init__(self, client: GQLClient) -> None:
        self.client = client
        self.organization_id = os.getenv("ORGANIZATION_ID", "default_organization")

        auth_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")
        if not auth_token:
            raise ValueError(MESSAGE_NOT_SET)

        self.kwargs = {"headers": {"Authorization": f"Bearer {auth_token}"}}

    def _create_collection_ref(
        self, collection_response: Union[CollectionReference, CollectionNotFound]
    ) -> Optional[CollectionReference]:
        if isinstance(collection_response, CollectionReference):
            return CollectionReference(
                id=collection_response.id, key=collection_response.key
            )
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

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: Optional[str] = None
    ) -> Optional[CollectionReference]:
        """
        Retrieve a collection by its key and parent key.

        This method retrieves a collection based on its key and parent key,
        or creates it if it doesn't exist.
        """
        return asyncio.run(
            self._create_collection(collection_key, parent_collection_id)
        )


    def _create_collection_document_ref(
        self,
        collection_response: Optional[
            Union[
                CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument,
                CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument,
                CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument,
                CollectionDocumentSetCollectionDocumentSetCollectionDocument,
                CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound,
                CollectionDocumentSetCollectionDocumentSetCollectionNotFound,
                CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocumentNotFound,
                CollectionDocumentCollectionCreateCollectionDocument,
                CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound,
            ]
        ],
    ) -> Optional[CollectionDocumentReference]:
        if isinstance(
            collection_response,
            (
                CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument,
                CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument,
                CollectionDocumentSetCollectionDocumentSetCollectionDocument,
                CollectionDocumentCollectionCreateCollectionDocument,
                CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument,
            ),
        ):
            return CollectionDocumentReference(
                id=collection_response.id,
                key=collection_response.key,
                data=collection_response.data,
                tags=collection_response.tags,
            )
        return None

    async def _get_collection_document(
        self, collection_key: str, document_key: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document(
            self.organization_id, collection_key, document_key
        )
        if isinstance(
            response.collection_create,
            CollectionDocumentCollectionCreateCollectionNotFound,
        ):
            return None
        return self._create_collection_document_ref(response.collection_create.document)

    def get_collection_document(
        self, collection_key: str, document_key: str
    ) -> Optional[CollectionDocumentReference]:

        return asyncio.run(self._get_collection_document(collection_key, document_key))

    async def _set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_set(
            collection_id, document_key, document_data
        )
        return self._create_collection_document_ref(response.collection_document_set)

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:

        return asyncio.run(
            self._set_collection_document(collection_id, document_key, document_data)
        )

    async def _delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_delete(document_id)
        return self._create_collection_document_ref(response.collection_document_delete)

    def delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:

        return asyncio.run(self._delete_collection_document(document_id))

    async def _add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_tag_add(document_id, tag)
        return self._create_collection_document_ref(
            response.collection_document_tag_add
        )

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:

        return asyncio.run(self._add_collection_document_tag(document_id, tag))

    async def _delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_tag_delete(
            document_id, tag_key
        )
        return self._create_collection_document_ref(
            response.collection_document_tag_delete
        )

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:

        return asyncio.run(self._delete_collection_document_tag(document_id, tag_key))


def _open_client() -> Client:
    url = os.getenv("NUMEROUS_API_URL")
    if not url:
        raise ValueError(API_URL_NOT_SET)
    client = GQLClient(url=url)
    return Client(client)
