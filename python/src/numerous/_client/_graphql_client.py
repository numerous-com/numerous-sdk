"""GraphQL client wrapper for numerous."""

import io
import os
from io import TextIOWrapper
from typing import BinaryIO, Optional, Union

import requests

from numerous.collection.exceptions import ParentCollectionNotFoundError
from numerous.collection.numerous_file import NumerousFile
from numerous.generated.graphql.client import Client as GQLClient
from numerous.generated.graphql.collection_collections import (
    CollectionCollectionsCollectionCollection,
    CollectionCollectionsCollectionCollectionCollectionsEdgesNode,
)
from numerous.generated.graphql.collection_document import (
    CollectionDocumentCollectionCollectionDocument,
    CollectionDocumentCollectionCollectionNotFound,
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
    CollectionDocumentsCollectionCollection,
    CollectionDocumentsCollectionCollectionDocumentsEdgesNode,
)
from numerous.generated.graphql.collection_file import (
    CollectionFileCollectionFileCreateCollectionFile,
    CollectionFileCollectionFileCreateCollectionNotFound,
)
from numerous.generated.graphql.collection_file_delete import (
    CollectionFileDeleteCollectionFileDeleteCollectionFile,
    CollectionFileDeleteCollectionFileDeleteCollectionFileNotFound,
)
from numerous.generated.graphql.collection_file_tag_add import (
    CollectionFileTagAddCollectionFileTagAddCollectionFile,
    CollectionFileTagAddCollectionFileTagAddCollectionFileNotFound,
)
from numerous.generated.graphql.collection_file_tag_delete import (
    CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile,
    CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound,
)
from numerous.generated.graphql.collection_files import (
    CollectionFilesCollectionCreateCollection,
    CollectionFilesCollectionCreateCollectionFilesEdgesNode,
)
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionFileReference,
    CollectionNotFound,
    CollectionReference,
)
from numerous.generated.graphql.input_types import TagInput
from numerous.threaded_event_loop import ThreadedEventLoop


COLLECTED_OBJECTS_NUMBER = 100
_REQUEST_TIMEOUT_SECONDS_ = 1.5


class APIURLMissingError(Exception):
    _msg = "NUMEROUS_API_URL environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class APIAccessTokenMissingError(Exception):
    _msg = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class OrganizationIDMissingError(Exception):
    _msg = "NUMEROUS_ORGANIZATION_ID environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class GraphQLClient:
    def __init__(self, gql: GQLClient) -> None:
        self._gql = gql
        self._threaded_event_loop = ThreadedEventLoop()
        self._threaded_event_loop.start()
        self._files_references: dict[str, CollectionFileReference] = {}

        organization_id = os.getenv("NUMEROUS_ORGANIZATION_ID")
        if not organization_id:
            raise OrganizationIDMissingError
        self._organization_id = organization_id

        auth_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")
        if not auth_token:
            raise APIAccessTokenMissingError

        self._headers = {"Authorization": f"Bearer {auth_token}"}

    def _create_collection_ref(
        self,
        collection_response: Union[
            CollectionReference,
            CollectionCollectionsCollectionCollectionCollectionsEdgesNode,
            CollectionNotFound,
        ],
    ) -> CollectionReference:
        if isinstance(collection_response, CollectionNotFound):
            raise ParentCollectionNotFoundError(collection_id=collection_response.id)

        return CollectionReference(
            id=collection_response.id, key=collection_response.key
        )

    async def _create_collection(
        self, collection_key: str, parent_collection_id: Optional[str] = None
    ) -> CollectionReference:
        response = await self._gql.collection_create(
            self._organization_id,
            collection_key,
            parent_collection_id,
            headers=self._headers,
        )
        return self._create_collection_ref(response.collection_create)

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: Optional[str] = None
    ) -> CollectionReference:
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
                CollectionDocumentCollectionCollectionDocument,
                CollectionDocumentsCollectionCollectionDocumentsEdgesNode,
                CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound,
            ]
        ],
    ) -> Optional[CollectionDocumentReference]:
        if isinstance(collection_response, CollectionDocumentReference):
            return CollectionDocumentReference(
                id=collection_response.id,
                key=collection_response.key,
                data=collection_response.data,
                tags=collection_response.tags,
            )
        return None

    async def _get_collection_document(
        self, collection_id: str, document_key: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self._gql.collection_document(
            collection_id,
            document_key,
            headers=self._headers,
        )
        if isinstance(
            response.collection,
            CollectionDocumentCollectionCollectionNotFound,
        ):
            return None
        if response.collection is None:
            return None
        return self._create_collection_document_ref(response.collection.document)

    def get_collection_document(
        self, collection_id: str, document_key: str
    ) -> Optional[CollectionDocumentReference]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_document(collection_id, document_key)
        )

    async def _set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> Optional[CollectionDocumentReference]:
        response = await self._gql.collection_document_set(
            collection_id,
            document_key,
            document_data,
            headers=self._headers,
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
        response = await self._gql.collection_document_delete(
            document_id, headers=self._headers
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
        response = await self._gql.collection_document_tag_add(
            document_id, tag, headers=self._headers
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
        response = await self._gql.collection_document_tag_delete(
            document_id, tag_key, headers=self._headers
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
        collection_id: str,
        end_cursor: str,
        tag_input: Optional[TagInput],
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]:
        response = await self._gql.collection_documents(
            collection_id,
            tag_input,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            headers=self._headers,
        )

        collection = response.collection
        if not isinstance(collection, CollectionDocumentsCollectionCollection):
            return [], False, ""

        documents = collection.documents
        edges = documents.edges
        page_info = documents.page_info

        result = [self._create_collection_document_ref(edge.node) for edge in edges]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_documents(
        self, collection_id: str, end_cursor: str, tag_input: Optional[TagInput]
    ) -> tuple[Optional[list[Optional[CollectionDocumentReference]]], bool, str]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_documents(collection_id, end_cursor, tag_input)
        )

    def _create_collection_files_ref(
        self,
        collection_response: Optional[
            Union[
                CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound,
                CollectionFileCollectionFileCreateCollectionFile,
                CollectionFileCollectionFileCreateCollectionNotFound,
                CollectionFileDeleteCollectionFileDeleteCollectionFile,
                CollectionFileDeleteCollectionFileDeleteCollectionFileNotFound,
                CollectionFilesCollectionCreateCollectionFilesEdgesNode,
                CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile,
                CollectionFileTagAddCollectionFileTagAddCollectionFile,
                CollectionFileTagAddCollectionFileTagAddCollectionFileNotFound,
                CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound,
            ]
        ],
    ) -> Optional[NumerousFile]:
        if isinstance(collection_response, CollectionFileReference):
            self._files_references[collection_response.id] = collection_response
            exists = False
            if (
                collection_response.download_url
                and collection_response.download_url.strip() != ""
            ):
                exists = True
            return NumerousFile(
                client=self,
                key=collection_response.key,
                file_id=collection_response.id,
                exists=exists,
                numerous_file_tags=collection_response.tags,
            )

        return None

    async def _get_collection_file(
        self, collection_id: str, file_key: str
    ) -> Optional[NumerousFile]:
        response = await self._gql.collection_file(
            collection_id,
            file_key,
            headers=self._headers,
        )
        return self._create_collection_files_ref(response.collection_file_create)

    def get_collection_file(
        self, collection_id: str, file_key: str
    ) -> Optional[NumerousFile]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_file(collection_id, file_key)
        )

    async def _delete_collection_file(self, file_id: str) -> Optional[NumerousFile]:
        response = await self._gql.collection_file_delete(
            file_id,
            headers=self._headers,
        )
        return self._create_collection_files_ref(response.collection_file_delete)

    def delete_collection_file(self, file_id: str) -> Optional[NumerousFile]:
        return self._threaded_event_loop.await_coro(
            self._delete_collection_file(file_id)
        )

    async def _get_collection_files(
        self,
        collection_key: str,
        end_cursor: str,
        tag_input: Optional[TagInput],
    ) -> tuple[Optional[list[Optional[NumerousFile]]], bool, str]:
        response = await self._gql.collection_files(
            self._organization_id,
            collection_key,
            tag_input,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            headers=self._headers,
        )

        collection = response.collection_create
        if not isinstance(collection, CollectionFilesCollectionCreateCollection):
            return [], False, ""

        files = collection.files
        edges = files.edges
        page_info = files.page_info

        result = [self._create_collection_files_ref(edge.node) for edge in edges]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_files(
        self, collection_key: str, end_cursor: str, tag_input: Optional[TagInput]
    ) -> tuple[Optional[list[Optional[NumerousFile]]], bool, str]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_files(collection_key, end_cursor, tag_input)
        )

    async def _add_collection_file_tag(
        self, file_id: str, tag: TagInput
    ) -> Optional[NumerousFile]:
        response = await self._gql.collection_file_tag_add(
            file_id, tag, headers=self._headers
        )
        return self._create_collection_files_ref(response.collection_file_tag_add)

    def add_collection_file_tag(
        self, file_id: str, tag: TagInput
    ) -> Optional[NumerousFile]:
        return self._threaded_event_loop.await_coro(
            self._add_collection_file_tag(file_id, tag)
        )

    async def _delete_collection_file_tag(
        self, file_id: str, tag_key: str
    ) -> Optional[NumerousFile]:
        response = await self._gql.collection_file_tag_delete(
            file_id, tag_key, headers=self._headers
        )
        return self._create_collection_files_ref(response.collection_file_tag_delete)

    def delete_collection_file_tag(
        self, file_id: str, tag_key: str
    ) -> Optional[NumerousFile]:
        return self._threaded_event_loop.await_coro(
            self._delete_collection_file_tag(file_id, tag_key)
        )

    async def _get_collection_collections(
        self, collection_id: str, end_cursor: str
    ) -> tuple[Optional[list[CollectionReference]], bool, str]:
        response = await self._gql.collection_collections(
            collection_id,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            headers=self._headers,
        )

        collection = response.collection
        if not isinstance(collection, CollectionCollectionsCollectionCollection):
            return [], False, ""

        collections = collection.collections
        edges = collections.edges
        page_info = collections.page_info

        result = [
            CollectionReference(id=edge.node.id, key=edge.node.key) for edge in edges
        ]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[Optional[list[CollectionReference]], bool, str]:
        return self._threaded_event_loop.await_coro(
            self._get_collection_collections(collection_key, end_cursor)
        )

    def read_text(self, file_id: str) -> str:
        download_url = self._files_references[file_id].download_url
        if download_url is None:
            msg = "No download URL for this file."
            raise ValueError(msg)
        response = requests.get(download_url, timeout=_REQUEST_TIMEOUT_SECONDS_)
        response.raise_for_status()

        return response.text

    def read_bytes(self, file_id: str) -> bytes:
        download_url = self._files_references[file_id].download_url
        if download_url is None:
            msg = "No download URL for this file."
            raise ValueError(msg)
        response = requests.get(download_url, timeout=_REQUEST_TIMEOUT_SECONDS_)
        response.raise_for_status()

        return response.content

    def save_data_file(self, file_id: str, data: Union[bytes, str]) -> None:
        upload_url = self._files_references[file_id].upload_url
        if upload_url is None:
            msg = "No upload URL for this file."
            raise ValueError(msg)

        if isinstance(data, str):
            data = data.encode("utf-8")  # Convert string to bytes

        response = requests.put(
            upload_url, files={"file": data}, timeout=_REQUEST_TIMEOUT_SECONDS_
        )
        response.raise_for_status()

    def save_file(self, file_id: str, data: TextIOWrapper) -> None:
        upload_url = self._files_references[file_id].upload_url
        if upload_url is None:
            msg = "No upload URL for this file."
            raise ValueError(msg)

        data.seek(0)
        file_content = data.read().encode("utf-8")

        response = requests.put(
            upload_url,
            files={"file": file_content},
            timeout=_REQUEST_TIMEOUT_SECONDS_,
        )
        response.raise_for_status()

    def open_file(self, file_id: str) -> BinaryIO:
        download_url = self._files_references[file_id].download_url
        if download_url is None:
            msg = "No download URL for this file."
            raise ValueError(msg)

        response = requests.get(download_url, timeout=_REQUEST_TIMEOUT_SECONDS_)
        response.raise_for_status()
        return io.BytesIO(response.content)
