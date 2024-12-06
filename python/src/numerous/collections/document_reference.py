"""Class for working with numerous documents."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING, Any

from numerous._utils.jsonbase64 import base64_to_dict, dict_to_base64
from numerous.collections._client import Tag


if TYPE_CHECKING:
    from numerous.collections._client import Client


@dataclass
class DocumentDoesNotExistError(Exception):
    collection_id: str
    key: str


@dataclass
class DocumentReference:
    id: str | None
    key: str
    collection_id: str
    collection_key: str
    _client: Client

    @property
    def exists(self) -> bool:
        """True if the document exists, False otherwise."""
        self._set_id_if_reference_exists()
        if self.id is None:
            return False

        return self._client.document_exists(self.id)

    def _set_id_if_reference_exists(self) -> None:
        if self.id is not None:
            return

        ref = self._client.document_reference(self.collection_id, self.key)
        if ref is not None:
            self.id = ref.id

    @property
    def tags(self) -> dict[str, str]:
        """Get the tags for the document."""
        self._set_id_if_reference_exists()
        if self.id is None:
            return {}

        return self._client.document_tags(self.id) or {}

    def set(self, data: dict[str, Any]) -> None:
        """
        Set the data for the document.

        Args:
            data: The data to set for the document.

        Raises:
            TypeError: If the data is not JSON serializable.

        """
        base64_data = dict_to_base64(data)
        self._client.document_set(self.collection_id, self.key, base64_data)

    def get(self) -> dict[str, Any] | None:
        """
        Get the data of the document.

        Returns:
            The data of the document if it is set.

        """
        self._set_id_if_reference_exists()
        if self.id is None:
            return None

        data = self._client.document_get(self.id)
        if data is None:
            return None

        return base64_to_dict(data)

    def delete(self) -> None:
        """Delete the referenced document."""
        self._set_id_if_reference_exists()
        if self.id is None:
            raise DocumentDoesNotExistError(
                collection_id=self.collection_id, key=self.key
            )

        self._client.document_delete(self.id)

    def tag(self, key: str, value: str) -> None:
        """
        Add a tag to the document.

        Args:
            key: The tag key.
            value: The tag value.

        """
        self._set_id_if_reference_exists()
        if self.id is None:
            raise DocumentDoesNotExistError(
                collection_id=self.collection_id, key=self.key
            )

        self._client.document_tag_add(self.id, Tag(key=key, value=value))

    def tag_delete(self, tag_key: str) -> None:
        """
        Delete a tag from the document.

        Args:
            tag_key: The key of the tag to delete.

        """
        self._set_id_if_reference_exists()
        if self.id is None:
            raise DocumentDoesNotExistError(
                collection_id=self.collection_id, key=self.key
            )

        self._client.document_tag_delete(self.id, tag_key)
