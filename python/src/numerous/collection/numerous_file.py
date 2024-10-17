"""Class for working with numerous files."""

from io import BufferedReader, TextIOWrapper
from typing import Any, List, Optional

import requests

from numerous.collection._client import Client
from numerous.generated.graphql.fragments import CollectionFileReference
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import base64_to_dict, dict_to_base64


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
        self.local_path: str = None
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
        """Check if the file is not empty."""
        if self.is_deleted:
            return False
        try:
            response = requests.get(self.download_url)
            response.raise_for_status()

        except requests.exceptions.RequestException as e:
            return False

        return len(response.content) > 0

    @property
    def tags(self) -> dict[str, str]:
        """Get the tags for the files."""
        if self.file_id is not None:
            return self._tags

        msg = "Cannot get tags from a non-existent files."
        raise ValueError(msg)

    def download(self, local_path) -> None:

        self.local_path = local_path

        self._update_file()

    def _update_file(self) -> None:
        try:
            response = requests.get(self.download_url)
            response.raise_for_status()

            with open(self.local_path, "wb") as local_file:
                local_file.write(response.content)

        except requests.exceptions.RequestException as e:
            raise e
        except IOError as e:
            raise e

    def read_text(self) -> List[str]:
        try:
            with open(self.local_path, "r") as file:
                return file.readlines()
        except IOError as e:
            return []

    def read_bytes(self) -> bytes:
        try:
            with open(self.local_path, "rb") as file:
                return file.read()
        except IOError as e:
            return b""

    def open(self) -> BufferedReader:
        try:
            return open(self.local_path, "rb")
        except IOError:
            return None

    def save(self, data: bytes) -> None:
        try:
            response = requests.post(self.upload_url, files={"file": data})
            response.raise_for_status()
        except requests.exceptions.RequestException as e:
            raise e

    def save_file(self, data: TextIOWrapper) -> None:
        try:
            data.seek(0)
            file_content = data.read().encode("utf-8")

            response = requests.post(self.upload_url, files={"file": file_content})
            response.raise_for_status()
        except requests.exceptions.RequestException as e:
            raise e

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
