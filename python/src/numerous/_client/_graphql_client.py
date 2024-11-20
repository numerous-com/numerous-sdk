"""GraphQL client wrapper for numerous."""

from __future__ import annotations

import io
import os
from typing import TYPE_CHECKING, BinaryIO

import requests

from numerous.collection.exceptions import ParentCollectionNotFoundError
from numerous.collection.file_reference import FileReference
from numerous.generated.graphql.collection_collections import (
    CollectionCollectionsCollectionCollection,
    CollectionCollectionsCollectionCollectionCollectionsEdgesNode,
)
from numerous.generated.graphql.collection_document import (
    CollectionDocumentCollectionCollectionDocument,
    CollectionDocumentCollectionCollectionNotFound,
)
from numerous.generated.graphql.collection_documents import (
    CollectionDocumentsCollectionCollection,
    CollectionDocumentsCollectionCollectionDocumentsEdgesNode,
)
from numerous.generated.graphql.collection_file import (
    CollectionFileCollectionFileCollectionFile,
    CollectionFileCollectionFileCollectionFileNotFound,
)
from numerous.generated.graphql.collection_files import (
    CollectionFilesCollectionCollection,
    CollectionFilesCollectionCollectionFilesEdgesNode,
)
from numerous.generated.graphql.fragments import (
    CollectionDocumentReference,
    CollectionFileReference,
    CollectionNotFound,
    CollectionReference,
)
from numerous.threaded_event_loop import ThreadedEventLoop


if TYPE_CHECKING:
    from numerous.generated.graphql.client import Client as GQLClient
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
    from numerous.generated.graphql.collection_file_create import (
        CollectionFileCreateCollectionFileCreateCollectionFile,
        CollectionFileCreateCollectionFileCreateCollectionNotFound,
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
    from numerous.generated.graphql.input_types import TagInput

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
        self._loop = ThreadedEventLoop()
        self._loop.start()

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
        collection_response: CollectionReference
        | CollectionCollectionsCollectionCollectionCollectionsEdgesNode
        | CollectionNotFound,
    ) -> CollectionReference:
        if isinstance(collection_response, CollectionNotFound):
            raise ParentCollectionNotFoundError(collection_id=collection_response.id)

        return CollectionReference(
            id=collection_response.id, key=collection_response.key
        )

    async def _create_collection(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionReference:
        response = await self._gql.collection_create(
            self._organization_id,
            collection_key,
            parent_collection_id,
            headers=self._headers,
        )
        return self._create_collection_ref(response.collection_create)

    def get_collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionReference:
        """
        Retrieve a collection by its key and parent key.

        This method retrieves a collection based on its key and parent key,
        or creates it if it doesn't exist.
        """
        return self._loop.await_coro(
            self._create_collection(collection_key, parent_collection_id)
        )

    def _create_collection_document_ref(
        self,
        resp: CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument
        | CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument
        | CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument
        | CollectionDocumentSetCollectionDocumentSetCollectionDocument
        | CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound
        | CollectionDocumentSetCollectionDocumentSetCollectionNotFound
        | CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocumentNotFound
        | CollectionDocumentCollectionCollectionDocument
        | CollectionDocumentsCollectionCollectionDocumentsEdgesNode
        | CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound  # noqa: E501
        | None,
    ) -> CollectionDocumentReference | None:
        if isinstance(resp, CollectionDocumentReference):
            return CollectionDocumentReference(
                id=resp.id,
                key=resp.key,
                data=resp.data,
                tags=resp.tags,
            )
        return None

    async def _get_collection_document(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentReference | None:
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
    ) -> CollectionDocumentReference | None:
        return self._loop.await_coro(
            self._get_collection_document(collection_id, document_key)
        )

    async def _set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> CollectionDocumentReference | None:
        response = await self._gql.collection_document_set(
            collection_id,
            document_key,
            document_data,
            headers=self._headers,
        )
        return self._create_collection_document_ref(response.collection_document_set)

    def set_collection_document(
        self, collection_id: str, document_key: str, document_data: str
    ) -> CollectionDocumentReference | None:
        return self._loop.await_coro(
            self._set_collection_document(collection_id, document_key, document_data)
        )

    async def _delete_collection_document(
        self, document_id: str
    ) -> CollectionDocumentReference | None:
        response = await self._gql.collection_document_delete(
            document_id, headers=self._headers
        )
        return self._create_collection_document_ref(response.collection_document_delete)

    def delete_collection_document(
        self, document_id: str
    ) -> CollectionDocumentReference | None:
        return self._loop.await_coro(self._delete_collection_document(document_id))

    async def _add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> CollectionDocumentReference | None:
        response = await self._gql.collection_document_tag_add(
            document_id, tag, headers=self._headers
        )
        return self._create_collection_document_ref(
            response.collection_document_tag_add
        )

    def add_collection_document_tag(
        self, document_id: str, tag: TagInput
    ) -> CollectionDocumentReference | None:
        return self._loop.await_coro(
            self._add_collection_document_tag(document_id, tag)
        )

    async def _delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> CollectionDocumentReference | None:
        response = await self._gql.collection_document_tag_delete(
            document_id, tag_key, headers=self._headers
        )
        return self._create_collection_document_ref(
            response.collection_document_tag_delete
        )

    def delete_collection_document_tag(
        self, document_id: str, tag_key: str
    ) -> CollectionDocumentReference | None:
        return self._loop.await_coro(
            self._delete_collection_document_tag(document_id, tag_key)
        )

    async def _get_collection_documents(
        self,
        collection_id: str,
        end_cursor: str,
        tag_input: TagInput | None,
    ) -> tuple[list[CollectionDocumentReference | None] | None, bool, str]:
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
        self, collection_id: str, end_cursor: str, tag_input: TagInput | None
    ) -> tuple[list[CollectionDocumentReference | None] | None, bool, str]:
        return self._loop.await_coro(
            self._get_collection_documents(collection_id, end_cursor, tag_input)
        )

    def _create_collection_files_ref(
        self,
        resp: (
            CollectionFileCreateCollectionFileCreateCollectionFile
            | CollectionFileCreateCollectionFileCreateCollectionNotFound
            | CollectionFileDeleteCollectionFileDeleteCollectionFile
            | CollectionFileDeleteCollectionFileDeleteCollectionFileNotFound
            | CollectionFilesCollectionCollectionFilesEdgesNode
            | CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile
            | CollectionFileTagAddCollectionFileTagAddCollectionFile
            | CollectionFileTagAddCollectionFileTagAddCollectionFileNotFound
            | CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound
            | None
        ),
    ) -> FileReference | None:
        if not isinstance(resp, CollectionFileReference):
            return None

        return FileReference(client=self, key=resp.key, file_id=resp.id)

    async def _create_collection_file_reference(
        self, collection_id: str, file_key: str
    ) -> FileReference | None:
        response = await self._gql.collection_file_create(
            collection_id,
            file_key,
            headers=self._headers,
        )
        return self._create_collection_files_ref(response.collection_file_create)

    def create_collection_file_reference(
        self, collection_id: str, file_key: str
    ) -> FileReference | None:
        return self._loop.await_coro(
            self._create_collection_file_reference(collection_id, file_key)
        )

    def collection_file_tags(self, file_id: str) -> dict[str, str] | None:
        file = self._collection_file(file_id)

        if not isinstance(file, CollectionFileCollectionFileCollectionFile):
            return None

        return {tag.key: tag.value for tag in file.tags}

    async def _delete_collection_file(self, file_id: str) -> None:
        await self._gql.collection_file_delete(file_id, headers=self._headers)

    def delete_collection_file(self, file_id: str) -> None:
        self._loop.await_coro(self._delete_collection_file(file_id))

    async def _get_collection_files(
        self,
        collection_id: str,
        end_cursor: str,
        tag_input: TagInput | None,
    ) -> tuple[list[FileReference], bool, str]:
        response = await self._gql.collection_files(
            collection_id,
            tag_input,
            after=end_cursor,
            first=COLLECTED_OBJECTS_NUMBER,
            headers=self._headers,
        )

        collection = response.collection
        if not isinstance(collection, CollectionFilesCollectionCollection):
            return [], False, ""

        files = collection.files
        edges = files.edges
        page_info = files.page_info

        result: list[FileReference] = []
        for edge in edges:
            if ref := self._create_collection_files_ref(edge.node):
                result.append(ref)  # noqa: PERF401

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def get_collection_files(
        self, collection_id: str, end_cursor: str, tag_input: TagInput | None
    ) -> tuple[list[FileReference], bool, str]:
        return self._loop.await_coro(
            self._get_collection_files(collection_id, end_cursor, tag_input)
        )

    async def _add_collection_file_tag(self, file_id: str, tag: TagInput) -> None:
        await self._gql.collection_file_tag_add(file_id, tag, headers=self._headers)

    def add_collection_file_tag(self, file_id: str, tag: TagInput) -> None:
        self._loop.await_coro(self._add_collection_file_tag(file_id, tag))

    async def _delete_collection_file_tag(self, file_id: str, tag_key: str) -> None:
        await self._gql.collection_file_tag_delete(
            file_id, tag_key, headers=self._headers
        )

    def delete_collection_file_tag(self, file_id: str, tag_key: str) -> None:
        return self._loop.await_coro(self._delete_collection_file_tag(file_id, tag_key))

    async def _get_collection_collections(
        self, collection_id: str, end_cursor: str
    ) -> tuple[list[CollectionReference] | None, bool, str]:
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
    ) -> tuple[list[CollectionReference] | None, bool, str]:
        return self._loop.await_coro(
            self._get_collection_collections(collection_key, end_cursor)
        )

    def save_file(self, file_id: str, data: bytes | str) -> None:
        file = self._collection_file(file_id)
        if file is None or isinstance(
            file, CollectionFileCollectionFileCollectionFileNotFound
        ):
            return

        if file.upload_url is None:
            msg = "No upload URL for this file."
            raise ValueError(msg)

        content_type = "application/octet-stream"
        if isinstance(data, str):
            content_type = "text/plain"
            data = data.encode()  # Convert string to bytes

        response = requests.put(
            file.upload_url,
            timeout=_REQUEST_TIMEOUT_SECONDS_,
            headers={"Content-Type": content_type, "Content-Length": str(len(data))},
            data=data,
        )
        response.raise_for_status()

    def read_text(self, file_id: str) -> str:
        return self._request_file(file_id).text

    def read_bytes(self, file_id: str) -> bytes:
        return self._request_file(file_id).content

    def open_file(self, file_id: str) -> BinaryIO:
        return io.BytesIO(self._request_file(file_id).content)

    def _collection_file(
        self, file_id: str
    ) -> (
        CollectionFileCollectionFileCollectionFileNotFound
        | CollectionFileCollectionFileCollectionFile
        | None
    ):
        return self._loop.await_coro(
            self._gql.collection_file(file_id, headers=self._headers)
        ).collection_file

    def _request_file(self, file_id: str) -> requests.Response:
        file = self._collection_file(file_id)

        if file is None or isinstance(
            file, CollectionFileCollectionFileCollectionFileNotFound
        ):
            msg = "Collection file not found"
            raise ValueError(msg)

        if file.download_url is None:
            msg = "No download URL for this file."
            raise ValueError(msg)

        response = requests.get(file.download_url, timeout=_REQUEST_TIMEOUT_SECONDS_)
        response.raise_for_status()

        return response

    def file_exists(self, file_id: str) -> bool:
        file = self._collection_file(file_id)

        if file is None or isinstance(
            file, CollectionFileCollectionFileCollectionFileNotFound
        ):
            return False

        return file.download_url is not None and file.download_url.strip() != ""
