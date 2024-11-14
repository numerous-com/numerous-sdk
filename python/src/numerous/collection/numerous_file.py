"""Class for working with numerous files."""
from __future__ import annotations

from typing import TYPE_CHECKING, BinaryIO

from numerous.generated.graphql.input_types import TagInput


if TYPE_CHECKING:
    from io import TextIOWrapper

    from numerous.collection._client import Client
    from numerous.generated.graphql.fragments import CollectionFileReferenceTags



_NO_FILE_MSG_ = "File dont exists."


class NumerousFile:
    """
    Represents a file in a Numerous collection.

    Attributes
    ----------
        key (str): The key of the file.
        collection_info tuple[str, str]: The id
            and key of collection file belongs to.
        data (Optional[dict[str, Any]]): The data of the file.
        id (Optional[str]): The unique identifier of the file.
        client (Client): The client to connect.
        numerous_file_tags (dict[str, str]): The tags associated with the file.

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
        self.key: str = key
        self.file_id = file_id
        self._client: Client = client
        self._exists = exists
        self._tags: dict[str, str]  = {}

        if numerous_file_tags is not None:
            dict_of_tags = {tag.key: tag.value for tag in numerous_file_tags}
            self._tags= dict_of_tags if dict_of_tags else {}

    @property
    def exists(self) -> bool:
        """
        Check if the file exists.

        Returns
        -------
            bool: True if the file exists,
            False otherwise.

        """
        return self._exists

    @property
    def tags(self) -> dict[str, str]:
        """
        Get the tags associated with the file.

        Returns
        -------
            dict[str, str]: Dictionary of tag key-value pairs.

        """
        return self._tags

    def read_text(self) -> str:
        """
        Read the file's content as text.

        Returns
        -------
            List[str]: The lines of text from the file.

        Raises
        ------
            OSError: If an error occurs while reading the file.
            ValueError: If there is no local path for the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)
        return self._client.read_text(self.file_id)


    def read_bytes(self) -> bytes:
        """
        Read the file's content as bytes.

        Returns
        -------
            bytes: The byte content of the file.

        Raises
        ------
            ValueError: If there is no local path for the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)
        return self._client.read_bytes(self.file_id)


    def open(self) -> BinaryIO:
        """
        Open the file for reading in binary mode.

        Returns
        -------
            Optional[BufferedReader]: The file reader,
            or None if the file cannot be opened.

        Raises
        ------
            ValueError: If there is no local path for the file.
            OSError: If an error occurs while reading the file.

        """
        if not self.exists:
            raise ValueError(_NO_FILE_MSG_)

        return self._client.open_file(self.file_id)

    def save(self, data: bytes | str) -> None:
        """
        Upload and save the file to the server.

        Args:
        ----
            data (bytes): The binary content to save to the file.

        Raises:
        ------
            HTTPError: If an error occurs during the upload.

        """
        return self._client.save_data_file(self.file_id,data)


    def save_file(self, data: TextIOWrapper) -> None:
        """
        Upload and save a text file to the server.

        Args:
        ----
            data (TextIOWrapper): The text content to upload.

        Raises:
        ------
            HTTPError: If an error occurs during the upload.

        """
        return self._client.save_file(self.file_id,data)


    def delete(self) -> None:
        """
        Delete the file.

        Raises
        ------
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
        Add a tag to the files.

        Args:
        ----
            key (str): The tag key.
            value (str): The tag value.

        Raises:
        ------
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
        ----
            tag_key (str): The key of the tag to delete.

        Raises:
        ------
            ValueError: If the file does not exist.

        """
        if not self.exists:
            msg = "Cannot delete tag from a non-existent file."
            raise ValueError(msg)

        tagged_file = self._client.delete_collection_file_tag(self.file_id, tag_key)

        if tagged_file is not None:
            self._tags = tagged_file.tags
