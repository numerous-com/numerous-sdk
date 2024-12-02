"""Class for working with numerous files."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING, BinaryIO

from numerous.collections._client import Tag


if TYPE_CHECKING:
    from io import TextIOWrapper

    from numerous.collections._client import Client

NO_FILE_ERROR_MSG = "File does not exist."


@dataclass
class FileReference:
    """Represents a file in a collection."""

    id: str
    key: str
    _client: Client

    @property
    def exists(self) -> bool:
        """
        Indicate whether the file exists.

        Returns:
            True if the file exists; False otherwise.

        """
        return self._client.file_exists(self.id)

    @property
    def tags(self) -> dict[str, str]:
        """
        Return the tags associated with the file.

        Returns:
            A dictionary of tag key-value pairs.

        """
        tags = self._client.file_tags(self.id)
        if tags is None:
            raise ValueError(NO_FILE_ERROR_MSG)
        return tags

    def read_text(self) -> str:
        """
        Read the file's content as text.

        Returns:
            The text content of the file.

        """
        return self._client.file_read_text(self.id)

    def read_bytes(self) -> bytes:
        """
        Read the file's content as bytes.

        Returns:
            The byte content of the file.

        """
        return self._client.file_read_bytes(self.id)

    def open(self) -> BinaryIO:
        """
        Open the file for reading in binary mode.

        Returns:
            A binary file-like object for reading the file.

        """
        return self._client.file_open(self.id)

    def save(self, data: bytes | str) -> None:
        """
        Upload and saves data to the file on the server.

        Args:
            data: The content to save to the file, either as bytes or string.

        """
        self._client.file_save(self.id, data)

    def save_file(self, data: TextIOWrapper) -> None:
        """
        Upload and saves a text file to the server.

        Args:
            data: A file-like object containing the text content to upload.

        """
        self._client.file_save(self.id, data.read())

    def delete(self) -> None:
        """Delete the file from the server."""
        self._client.file_delete(self.id)

    def tag(self, key: str, value: str) -> None:
        """
        Add a tag to the file.

        Args:
            key: The tag key.
            value: The tag value.

        """
        self._client.file_tag_add(self.id, Tag(key=key, value=value))

    def tag_delete(self, tag_key: str) -> None:
        """
        Delete a tag from the file.

        Args:
            tag_key: The key of the tag to delete.

        Raises:
            ValueError: If the file does not exist.

        """
        self._client.file_delete_tag(self.id, tag_key)
