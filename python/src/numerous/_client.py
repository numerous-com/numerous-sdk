"""GraphQL client wrapper for numerous."""

import os
from typing import Optional, Union

from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_collections import (
    CollectionCollectionsCollectionCreateCollection,
    CollectionCollectionsCollectionCreateCollectionCollectionsEdgesNode,
)
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
from numerous.generated.graphql.collection_documents import (
    CollectionDocumentsCollectionCreateCollection,
    CollectionDocumentsCollectionCreateCollectionDocumentsEdgesNode,
)
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionNotFound,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput
from numerous.threaded_event_loop import ThreadedEventLoop


API_URL_NOT_SET = "NUMEROUS_API_URL environment variable is not set"
MESSAGE_NOT_SET = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"
COLLECTED_OBJECTS_NUMBER = 100


class Client:
    def __init__(self, client: GQLClient) -> None:
        self.client = client
        self._threaded_event_loop = ThreadedEventLoop()
        self._threaded_event_loop.start()
        self.organization_id = os.getenv("ORGANIZATION_ID", "default_organization")

        auth_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")
        if not auth_token:
            raise ValueError(MESSAGE_NOT_SET)

        self.kwargs = {"headers": {"Authorization": f"Bearer {auth_token}"}}

    def _create_collection_ref(
        self,
        collection_response: Union[
            CollectionReference,
            CollectionCollectionsCollectionCreateCollectionCollectionsEdgesNode,
            CollectionNotFound,
        ],
    ) -> Optional[CollectionReference]:
        if isinstance(
            collection_response,
            (
                CollectionReference,
                CollectionCollectionsCollectionCreateCollectionCollectionsEdgesNode,
            ),
        ):
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
            **self.kwargs,
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
        return self._threaded_event_loop.await_coro(
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
                CollectionDocumentsCollectionCreateCollectionDocumentsEdgesNode,
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
                CollectionDocumentsCollectionCreateCollectionDocumentsEdgesNode,
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
            self.organization_id,
            collection_key,
            document_key,
            **self.kwargs,
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
        return self._threaded_event_loop.await_coro(
            self._get_collection_document(collection_key, document_key)
        )

    async def _set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_set(
            collection_id,
            document_key,
            document_data,
            **self.kwargs,
        )
        return self._create_collection_document_ref(response.collection_document_set)

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:
        return self._threaded_event_loop.await_coro(
            self._set_collection_document(collection_id, document_key, document_data)
        )

    async def _delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_delete(
            document_id, **self.kwargs
        )
        return self._create_collection_document_ref(response.collection_document_delete)

    def delete_collection_document(
        self, document_id: str
    ) -> Optional[CollectionDocumentReference]:
        return self._threaded_event_loop.await_coro(
            self._delete_collection_document(document_id)
        )

    async def _add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_tag_add(
            document_id, tag, **self.kwargs
        )
        return self._create_collection_document_ref(
            response.collection_document_tag_add
        )

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> Optional[CollectionDocumentReference]:
        return self._threaded_event_loop.await_coro(
            self._add_collection_document_tag(document_id, tag)
        )

    async def _delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self.client.collection_document_tag_delete(
            document_id, tag_key, **self.kwargs
        )
        return self._create_collection_document_ref(
            response.collection_document_tag_delete
        )

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> Optional[CollectionDocumentReference]:
        return self._threaded_event_loop.await_coro(
            self._delete_collection_document_tag(document_id, tag_key)
        )

    async def _get_collection_documents(
        self,
        collection_key: str,
        end_cursor: str,
        tag_input: Optional[TagInput],
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]:
        response = await self.client.collection_documents(
            self.organization_id,
            collection_key,
            tag_input,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            **self.kwargs,
        )

        collection = response.collection_create
        if not isinstance(collection, CollectionDocumentsCollectionCreateCollection):
            return [], False, ""

        documents = collection.documents
        edges = documents.edges
        page_info = documents.page_info

        result = [self._create_collection_document_ref(edge.node) for edge in edges]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_documents(
        self, collection_key: str, end_cursor: str, tag_input: Optional[TagInput]
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_documents(collection_key, end_cursor, tag_input)
        )

    async def _get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[Optional[list[Optional[CollectionReference]]], bool, str]:
        response = await self.client.collection_collections(
            self.organization_id,
            collection_key,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            **self.kwargs,
        )

        collection = response.collection_create
        if not isinstance(collection, CollectionCollectionsCollectionCreateCollection):
            return [], False, ""

        collections = collection.collections
        edges = collections.edges
        page_info = collections.page_info

        result = [self._create_collection_ref(edge.node) for edge in edges]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[Optional[list[Optional[CollectionReference]]], bool, str]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_collections(collection_key, end_cursor)
        )


def _open_client() -> Client:
    url = os.getenv("NUMEROUS_API_URL")
    if not url:
        raise ValueError(API_URL_NOT_SET)
    client = GQLClient(url=url)
    return Client(client)
