"""GraphQL client wrapper for numerous."""

from __future__ import annotations

import io
from typing import TYPE_CHECKING, BinaryIO

import requests

from numerous._client.graphql.collection_document import (
    CollectionDocumentCollectionDocumentCollectionDocument,
    CollectionDocumentCollectionDocumentCollectionDocumentNotFound,
)
from numerous.collections._client import (
    CollectionDocumentIdentifier,
    CollectionFileIdentifier,
    CollectionIdentifier,
    Tag,
)
from numerous.collections.exceptions import ParentCollectionNotFoundError

from .graphql import fragments
from .graphql.collection_collections import (
    CollectionCollectionsCollectionCollection,
)
from .graphql.collection_document_in_collection import (
    CollectionDocumentInCollectionCollectionCollectionDocument,
    CollectionDocumentInCollectionCollectionCollectionNotFound,
)
from .graphql.collection_documents import (
    CollectionDocumentsCollectionCollection,
    CollectionDocumentsCollectionCollectionDocumentsEdgesNode,
)
from .graphql.collection_file import (
    CollectionFileCollectionFileCollectionFile,
    CollectionFileCollectionFileCollectionFileNotFound,
)
from .graphql.collection_files import (
    CollectionFilesCollectionCollection,
    CollectionFilesCollectionCollectionFilesEdgesNode,
)
from .graphql.input_types import TagInput
from .threaded_event_loop import ThreadedEventLoop


if TYPE_CHECKING:
    from .graphql.client import Client as GQLClient
    from .graphql.collection_document_delete import (
        CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument,
        CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocumentNotFound,
    )
    from .graphql.collection_document_set import (
        CollectionDocumentSetCollectionDocumentSetCollectionDocument,
        CollectionDocumentSetCollectionDocumentSetCollectionNotFound,
    )
    from .graphql.collection_document_tag_add import (
        CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument,
        CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound,
    )
    from .graphql.collection_document_tag_delete import (
        CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument,
        CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound,
    )
    from .graphql.collection_file_create import (
        CollectionFileCreateCollectionFileCreateCollectionFile,
        CollectionFileCreateCollectionFileCreateCollectionNotFound,
    )
    from .graphql.collection_file_delete import (
        CollectionFileDeleteCollectionFileDeleteCollectionFile,
        CollectionFileDeleteCollectionFileDeleteCollectionFileNotFound,
    )
    from .graphql.collection_file_tag_add import (
        CollectionFileTagAddCollectionFileTagAddCollectionFile,
        CollectionFileTagAddCollectionFileTagAddCollectionFileNotFound,
    )
    from .graphql.collection_file_tag_delete import (
        CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile,
        CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound,
    )


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
    def __init__(self, gql: GQLClient, organization_id: str, access_token: str) -> None:
        self._gql = gql
        self._loop = ThreadedEventLoop()
        self._loop.start()
        self._organization_id = organization_id
        self._headers = {"Authorization": f"Bearer {access_token}"}

    async def _collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionIdentifier:
        response = await self._gql.collection_create(
            self._organization_id,
            collection_key,
            parent_collection_id,
            headers=self._headers,
        )
        if isinstance(response.collection_create, fragments.CollectionNotFound):
            raise ParentCollectionNotFoundError(
                collection_id=response.collection_create.id
            )

        return CollectionIdentifier(
            id=response.collection_create.id, key=response.collection_create.key
        )

    def collection_reference(
        self, collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionIdentifier:
        return self._loop.await_coro(
            self._collection_reference(collection_key, parent_collection_id)
        )

    def collection_collections(
        self, collection_key: str, end_cursor: str
    ) -> tuple[list[CollectionIdentifier] | None, bool, str]:
        return self._loop.await_coro(
            self._collection_collections(collection_key, end_cursor)
        )

    async def _collection_collections(
        self, collection_id: str, end_cursor: str
    ) -> tuple[list[CollectionIdentifier] | None, bool, str]:
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
            CollectionIdentifier(id=edge.node.id, key=edge.node.key) for edge in edges
        ]

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def collection_documents(
        self, collection_id: str, end_cursor: str, tag: Tag | None
    ) -> tuple[list[CollectionDocumentIdentifier] | None, bool, str]:
        return self._loop.await_coro(
            self._collection_documents(collection_id, end_cursor, _tag_input(tag))
        )

    async def _collection_documents(
        self,
        collection_id: str,
        end_cursor: str,
        tag_input: TagInput | None,
    ) -> tuple[list[CollectionDocumentIdentifier] | None, bool, str]:
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

        results: list[CollectionDocumentIdentifier] = []
        for edge in edges:
            result = self._document_identifier(edge.node)
            if result is not None:
                results.append(result)

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return results, has_next_page, end_cursor

    def collection_files(
        self, collection_id: str, end_cursor: str, tag: Tag | None
    ) -> tuple[list[CollectionFileIdentifier], bool, str]:
        return self._loop.await_coro(
            self._collection_files(collection_id, end_cursor, _tag_input(tag))
        )

    async def _collection_files(
        self,
        collection_id: str,
        end_cursor: str,
        tag: TagInput | None,
    ) -> tuple[list[CollectionFileIdentifier], bool, str]:
        response = await self._gql.collection_files(
            collection_id,
            tag,
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

        result: list[CollectionFileIdentifier] = []
        for edge in edges:
            if ref := self._file_identifier(edge.node):
                result.append(ref)  # noqa: PERF401

        end_cursor = page_info.end_cursor or ""
        has_next_page = page_info.has_next_page

        return result, has_next_page, end_cursor

    def _document_identifier(
        self,
        resp: CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocument
        | CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument
        | CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocument
        | CollectionDocumentSetCollectionDocumentSetCollectionDocument
        | CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound
        | CollectionDocumentSetCollectionDocumentSetCollectionNotFound
        | CollectionDocumentDeleteCollectionDocumentDeleteCollectionDocumentNotFound
        | CollectionDocumentInCollectionCollectionCollectionDocument
        | CollectionDocumentsCollectionCollectionDocumentsEdgesNode
        | CollectionDocumentTagDeleteCollectionDocumentTagDeleteCollectionDocumentNotFound  # noqa: E501
        | None,
    ) -> CollectionDocumentIdentifier | None:
        if isinstance(resp, fragments.CollectionDocumentReference):
            return CollectionDocumentIdentifier(id=resp.id, key=resp.key)
        return None

    def document_reference(
        self, collection_id: str, document_key: str
    ) -> CollectionDocumentIdentifier | None:
        response = self._loop.await_coro(
            self._gql.collection_document_in_collection(
                collection_id,
                document_key,
                headers=self._headers,
            )
        )
        if isinstance(
            response.collection,
            CollectionDocumentInCollectionCollectionCollectionNotFound,
        ):
            raise ParentCollectionNotFoundError(collection_id=response.collection.id)

        if response.collection is None:
            raise ParentCollectionNotFoundError(collection_id=collection_id)

        if response.collection.document is None:
            return None

        return self._document_identifier(response.collection.document)

    def document_exists(self, document_id: str) -> bool:
        return self._document_query(document_id) is not None

    def document_tags(self, document_id: str) -> dict[str, str] | None:
        doc = self._document_query(document_id)
        if doc is None:
            return None

        return {tag.key: tag.value for tag in doc.tags}

    def _document_query(
        self, document_id: str
    ) -> CollectionDocumentCollectionDocumentCollectionDocument | None:
        doc = self._loop.await_coro(
            self._gql.collection_document(document_id, headers=self._headers)
        )
        if isinstance(
            doc.collection_document,
            CollectionDocumentCollectionDocumentCollectionDocumentNotFound,
        ):
            return None

        return doc.collection_document

    def document_set(
        self, collection_id: str, document_key: str, document_data: str
    ) -> None:
        self._loop.await_coro(
            self._gql.collection_document_set(
                collection_id,
                document_key,
                document_data,
                headers=self._headers,
            )
        )

    def document_get(self, document_id: str) -> str | None:
        doc = self._document_query(document_id)
        return None if doc is None else doc.data

    def document_delete(self, document_id: str) -> None:
        self._loop.await_coro(
            self._gql.collection_document_delete(document_id, headers=self._headers)
        )

    def document_tag_add(self, document_id: str, tag: Tag) -> None:
        self._loop.await_coro(
            self._gql.collection_document_tag_add(
                document_id, _tag_input_strict(tag), headers=self._headers
            )
        )

    def document_tag_delete(self, document_id: str, tag_key: str) -> None:
        self._loop.await_coro(
            self._gql.collection_document_tag_delete(
                document_id, tag_key, headers=self._headers
            )
        )

    def file_reference(
        self, collection_id: str, file_key: str
    ) -> CollectionFileIdentifier | None:
        return self._loop.await_coro(
            self._file_reference_create(collection_id, file_key)
        )

    async def _file_reference_create(
        self, collection_id: str, file_key: str
    ) -> CollectionFileIdentifier | None:
        response = await self._gql.collection_file_create(
            collection_id,
            file_key,
            headers=self._headers,
        )
        return self._file_identifier(response.collection_file_create)

    def _file_identifier(
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
    ) -> CollectionFileIdentifier | None:
        if not isinstance(resp, fragments.CollectionFileReference):
            return None

        return CollectionFileIdentifier(key=resp.key, id=resp.id)

    def file_tags(self, file_id: str) -> dict[str, str] | None:
        file = self._file_query(file_id)

        if not isinstance(file, CollectionFileCollectionFileCollectionFile):
            return None

        return {tag.key: tag.value for tag in file.tags}

    async def _file_delete(self, file_id: str) -> None:
        await self._gql.collection_file_delete(file_id, headers=self._headers)

    def file_delete(self, file_id: str) -> None:
        self._loop.await_coro(self._file_delete(file_id))

    async def _file_tag_add(self, file_id: str, tag: TagInput) -> None:
        await self._gql.collection_file_tag_add(file_id, tag, headers=self._headers)

    def file_tag_add(self, file_id: str, tag: Tag) -> None:
        self._loop.await_coro(self._file_tag_add(file_id, _tag_input_strict(tag)))

    async def _file_delete_tag(self, file_id: str, tag_key: str) -> None:
        await self._gql.collection_file_tag_delete(
            file_id, tag_key, headers=self._headers
        )

    def file_delete_tag(self, file_id: str, tag_key: str) -> None:
        return self._loop.await_coro(self._file_delete_tag(file_id, tag_key))

    def file_save(self, file_id: str, data: bytes | str) -> None:
        file = self._file_query(file_id)
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

    def file_read_text(self, file_id: str) -> str:
        return self._file_download(file_id).text

    def file_read_bytes(self, file_id: str) -> bytes:
        return self._file_download(file_id).content

    def file_open(self, file_id: str) -> BinaryIO:
        return io.BytesIO(self._file_download(file_id).content)

    def _file_query(
        self, file_id: str
    ) -> (
        CollectionFileCollectionFileCollectionFileNotFound
        | CollectionFileCollectionFileCollectionFile
        | None
    ):
        return self._loop.await_coro(
            self._gql.collection_file(file_id, headers=self._headers)
        ).collection_file

    def _file_download(self, file_id: str) -> requests.Response:
        file = self._file_query(file_id)

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
        file = self._file_query(file_id)

        if file is None or isinstance(
            file, CollectionFileCollectionFileCollectionFileNotFound
        ):
            return False

        return file.download_url is not None and file.download_url.strip() != ""


def _tag_input_strict(tag: Tag) -> TagInput:
    return TagInput(key=tag.key, value=tag.value)


def _tag_input(tag: Tag | None) -> TagInput | None:
    return _tag_input_strict(tag) if tag else None
