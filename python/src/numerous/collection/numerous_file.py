"""Class for working with numerous files."""

from __future__ import annotations

from typing import TYPE_CHECKING, BinaryIO

from numerous.generated.graphql.input_types import TagInput


if TYPE_CHECKING:
    from io import TextIOWrapper

    from numerous.collection._client import Client
    from numerous.generated.graphql.fragments import CollectionFileReferenceTags

_NO_FILE_MSG_ = "File does not exist."


class NumerousFile:
    """
    Represents a file in a Numerous collection.

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
        exists: bool = False,
        numerous_file_tags: list[CollectionFileReferenceTags] | None = None,
    ) -> None:
        """
        Initialize a NumerousFile instance.

        Args:
            client: The client used to interact with the Numerous collection.
            key: The key of the file.
            file_id: The unique identifier of the file.
            exists: Indicates whether the file exists.
            numerous_file_tags: An optional list of tags associated with the file.

        """
        self.key: str = key
        self.file_id: str = file_id
        self._client: Client = client
        self._exists: bool = exists
        self._tags: dict[str, str] = {}

        if numerous_file_tags is not None:
            dict_of_tags = {tag.key: tag.value for tag in numerous_file_tags}
            self._tags = dict_of_tags if dict_of_tags else {}

    @property
    def exists(self) -> bool:
        """
        Indicate whether the file exists.

        Returns:
            True if the file exists; False otherwise.

        """
        return self._exists

    @property
    def tags(self) -> dict[str, str]:
        """
        Return the tags associated with the file.

        Returns:
            A dictionary of tag key-value pairs.

        """
        return self._tags

    def read_text(self) -> str:
        """
        Read the file's content as text.

        Returns:
            The text content of the file.

        Raises:
            ValueError: If the file does not exist.
            OSError: If an error occurs while reading the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)
        return self._client.read_text(self.file_id)

    def read_bytes(self) -> bytes:
        """
        Read the file's content as bytes.

        Returns:
            The byte content of the file.

        Raises:
            ValueError: If the file does not exist.
            OSError: If an error occurs while reading the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)
        return self._client.read_bytes(self.file_id)

    def open(self) -> BinaryIO:
        """
        Open the file for reading in binary mode.

        Returns:
            A binary file-like object for reading the file.

        Raises:
            ValueError: If the file does not exist.
            OSError: If an error occurs while opening the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)
        return self._client.open_file(self.file_id)

    def save(self, data: bytes | str) -> None:
        """
        Upload and saves data to the file on the server.

        Args:
            data: The content to save to the file, either as bytes or string.

        Raises:
            HTTPError: If an error occurs during the upload.

        """
        self._client.save_data_file(self.file_id, data)

    def save_file(self, data: TextIOWrapper) -> None:
        """
        Upload and saves a text file to the server.

        Args:
            data: A file-like object containing the text content to upload.

        Raises:
            HTTPError: If an error occurs during the upload.

        """
        self._client.save_file(self.file_id, data)

    def delete(self) -> None:
        """
        Delete the file from the server.

        Raises:
            ValueError: If the file does not exist or deletion failed.

        """
        deleted_file = self._client.delete_collection_file(self.file_id)
        if deleted_file is None:
            msg = "Failed to delete the file."
            raise ValueError(msg)

        self.key = ""
        self._tags = {}
        self._exists = False

    def tag(self, key: str, value: str) -> None:
        """
        Add a tag to the file.

        Args:
            key: The tag key.
            value: The tag value.

        Raises:
            ValueError: If the file does not exist.

        """
        if not self.exists:
            msg = "Cannot tag a non-existent file."
            raise ValueError(msg)

        tagged_file = self._client.add_collection_file_tag(
            self.file_id, TagInput(key=key, value=value)
        )

        if tagged_file is not None:
            self._tags = tagged_file.tags

    def tag_delete(self, tag_key: str) -> None:
        """
        Delete a tag from the file.

        Args:
            tag_key: The key of the tag to delete.

        Raises:
            ValueError: If the file does not exist.

        """
        if not self.exists:
            msg = "Cannot delete tag from a non-existent file."
            raise ValueError(msg)

        tagged_file = self._client.delete_collection_file_tag(self.file_id, tag_key)

        if tagged_file is not None:
            self._tags = tagged_file.tags
