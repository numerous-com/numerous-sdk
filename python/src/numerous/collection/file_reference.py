"""Class for working with numerous files."""

from __future__ import annotations

from typing import TYPE_CHECKING, BinaryIO

from numerous.generated.graphql.input_types import TagInput


if TYPE_CHECKING:
    from io import TextIOWrapper

    from numerous.collection._client import Client

NO_FILE_ERROR_MSG = "File does not exist."


class FileReference:
    """
    Represents a file in a collection.

    Attributes:
        key: The key of the file.
        file_id: The unique identifier of the file.

    """

    def __init__(
        self,
        *,
        client: Client,
        key: str,
        file_id: str,
    ) -> None:
        """
        Initialize a file reference.

        Args:
            client: The client used to interact with the Numerous collection.
            key: The key of the file.
            file_id: The unique identifier of the file.
            tags: An optional list of tags associated with the file.

        """
        self.key: str = key
        self.file_id: str = file_id
        self._client: Client = client

    @property
    def exists(self) -> bool:
        """
        Indicate whether the file exists.

        Returns:
            True if the file exists; False otherwise.

        """
        return self._client.file_exists(self.file_id)

    @property
    def tags(self) -> dict[str, str]:
        """
        Return the tags associated with the file.

        Returns:
            A dictionary of tag key-value pairs.

        """
        tags = self._client.collection_file_tags(self.file_id)
        if tags is None:
            raise ValueError(NO_FILE_ERROR_MSG)
        return tags

    def read_text(self) -> str:
        """
        Read the file's content as text.

        Returns:
            The text content of the file.

        """
        return self._client.read_text(self.file_id)

    def read_bytes(self) -> bytes:
        """
        Read the file's content as bytes.

        Returns:
            The byte content of the file.

        """
        return self._client.read_bytes(self.file_id)

    def open(self) -> BinaryIO:
        """
        Open the file for reading in binary mode.

        Returns:
            A binary file-like object for reading the file.

        """
        return self._client.open_file(self.file_id)

    def save(self, data: bytes | str) -> None:
        """
        Upload and saves data to the file on the server.

        Args:
            data: The content to save to the file, either as bytes or string.

        """
        self._client.save_file(self.file_id, data)

    def save_file(self, data: TextIOWrapper) -> None:
        """
        Upload and saves a text file to the server.

        Args:
            data: A file-like object containing the text content to upload.

        """
        self._client.save_file(self.file_id, data.read())

    def delete(self) -> None:
        """Delete the file from the server."""
        self._client.delete_collection_file(self.file_id)

    def tag(self, key: str, value: str) -> None:
        """
        Add a tag to the file.

        Args:
            key: The tag key.
            value: The tag value.

        """
        self._client.add_collection_file_tag(
            self.file_id, TagInput(key=key, value=value)
        )

    def tag_delete(self, tag_key: str) -> None:
        """
        Delete a tag from the file.

        Args:
            tag_key: The key of the tag to delete.

        Raises:
            ValueError: If the file does not exist.

        """
        self._client.delete_collection_file_tag(self.file_id, tag_key)
