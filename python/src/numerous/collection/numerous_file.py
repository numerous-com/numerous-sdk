"""Class for working with numerous files."""

from io import BufferedReader, TextIOWrapper
from pathlib import Path
from typing import List, Optional, Union

import requests

from numerous.collection._client import Client
from numerous.generated.graphql.fragments import CollectionFileReference
from numerous.generated.graphql.input_types import TagInput


_REQUEST_TIMEOUT_SECONDS_ = 1.5


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
        tags (dict[str, str]): The tags associated with the file.

    """

    def __init__(
        self,
        client: Client,
        key: str,
        collection_info: tuple[str, str],
        numerous_file_ref: Optional[CollectionFileReference] = None,
    ) -> None:
        self.key: str = key
        self.collection_id: str = collection_info[0]
        self.collection_key: str = collection_info[1]
        self._client: Client = client
        self.file_id: Optional[str] = None
        self.local_path: Optional[Path] = None
        self.is_deleted = False

        if numerous_file_ref is not None:
            dict_of_tags = {tag.key: tag.value for tag in numerous_file_ref.tags}
            self.download_url = numerous_file_ref.download_url
            self.upload_url = numerous_file_ref.upload_url
            self.file_id = numerous_file_ref.id
            self._tags: dict[str, str] = (
                dict_of_tags if dict_of_tags is not None else {}
            )

    @property
    def exists(self) -> bool:
        """
        Check if the file exists by verifying its contents.

        Returns
        -------
            bool: True if the file exists and has content, False otherwise.

        """
        if self.is_deleted or self.download_url is None:
            return False
        try:
            response = requests.get(
                self.download_url, timeout=_REQUEST_TIMEOUT_SECONDS_
            )
            response.raise_for_status()

        except requests.exceptions.RequestException:
            return False

        return len(response.content) > 0

    @property
    def tags(self) -> dict[str, str]:
        """
        Get the tags associated with the file.

        Returns
        -------
            dict[str, str]: Dictionary of tag key-value pairs.

        Raises
        ------
            ValueError: If the file does not exist.

        """
        if self.file_id is not None:
            return self._tags

        msg = "Cannot get tags from a non-existent files."
        raise ValueError(msg)

    def download(self, local_path: str) -> None:
        """
        Download the file to the specified local path.

        Args:
        ----
            local_path (str): The local path to save the downloaded file.

        """
        self.local_path = Path(local_path)

        self._update_file()

    def _update_file(self) -> None:
        """
        Update the local file by downloading it from the server.

        Raises
        ------
            ValueError: If there is no download URL or local path for the file.
            RequestException: If an error occurs during the request.
            OSError: If an error occurs while saving the file locally.

        """
        if self.download_url is None:
            msg = "No download URL for this file."
            raise ValueError(msg)
        if self.local_path is None:
            msg = "No local path for this file."
            raise ValueError(msg)
        response = requests.get(self.download_url, timeout=_REQUEST_TIMEOUT_SECONDS_)
        response.raise_for_status()

        with Path.open(self.local_path, "wb") as local_file:
            local_file.write(response.content)

    def read_text(self) -> List[str]:
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
        if self.local_path is None:
            msg = "No local path for this file."
            raise ValueError(msg)
        try:
            with Path.open(self.local_path) as file:
                return file.readlines()
        except OSError:
            return []

    def read_bytes(self) -> bytes:
        """
        Read the file's content as bytes.

        Returns
        -------
            bytes: The byte content of the file.

        Raises
        ------
            OSError: If an error occurs while reading the file.
            ValueError: If there is no local path for the file.


        """
        if self.local_path is None:
            msg = "No local path for this file."
            raise ValueError(msg)
        try:
            with Path.open(self.local_path, "rb") as file:
                return file.read()
        except OSError:
            return b""

    def open(self) -> BufferedReader:
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
        if self.local_path is None:
            msg = "No local path for this file."
            raise ValueError(msg)

        return Path.open(self.local_path, "rb")


    def save(self, data: Union[bytes, str]) -> None:
        """
        Upload and save the file to the server.

        Args:
        ----
            data (bytes): The binary content to save to the file.

        Raises:
        ------
            HTTPError: If an error occurs during the upload.
            ValueError: If there is no upload URL for the file.


        """
        if self.upload_url is None:
            msg = "No upload URL for this file."
            raise ValueError(msg)
        response = requests.post(
            self.upload_url, files={"file": data}, timeout=_REQUEST_TIMEOUT_SECONDS_
        )
        response.raise_for_status()

    def save_file(self, data: TextIOWrapper) -> None:
        """
        Upload and save a text file to the server.

        Args:
        ----
            data (TextIOWrapper): The text content to upload.

        Raises:
        ------
            HTTPError: If an error occurs during the upload.
            ValueError: If there is no upload URL for the file.


        """
        data.seek(0)
        file_content = data.read().encode("utf-8")
        if self.upload_url is None:
            msg = "No upload URL for this file."
            raise ValueError(msg)
        response = requests.post(
            self.upload_url,
            files={"file": file_content},
            timeout=_REQUEST_TIMEOUT_SECONDS_,
        )
        response.raise_for_status()

    def delete(self) -> None:
        """
        Delete the file.

        Raises
        ------
            ValueError: If the file does not exist or deletion failed.

        """
        if self.file_id is not None:
            deleted_file = self._client.delete_collection_file(self.file_id)

            if deleted_file is not None and deleted_file.id == self.file_id:
                self.file_id = None
                self.download_url = None
                self.upload_url = None
                self.local_path = None
                self._tags = {}
                self.is_deleted = True
            else:
                msg = "Failed to delete the file."
                raise ValueError(msg)
        else:
            msg = "Cannot delete a non-existent file."
            raise ValueError(msg)

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
        if self.file_id is not None:
            tagged_file = self._client.add_collection_file_tag(
                self.file_id, TagInput(key=key, value=value)
            )
        else:
            msg = "Cannot tag a non-existent file."
            raise ValueError(msg)

        if tagged_file is not None:
            self.tags.update({tag.key: tag.value for tag in tagged_file.tags})

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
        if self.file_id is not None:
            tagged_file = self._client.delete_collection_file_tag(self.file_id, tag_key)
        else:
            msg = "Cannot delete tag from a non-existent file."
            raise ValueError(msg)

        if tagged_file is not None:
            self._tags = {tag.key: tag.value for tag in tagged_file.tags}
