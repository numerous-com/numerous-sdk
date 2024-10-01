"""Class for working with numerous documents."""

from typing import Any, Optional

from numerous.collection._client import Client
from numerous.generated.graphql.fragments import CollectionDocumentReference
from numerous.generated.graphql.input_types import TagInput
from numerous.jsonbase64 import base64_to_dict, dict_to_base64


class NumerousDocument:
    """
    Represents a document in a Numerous collection.

    Attributes
    ----------
        key (str): The key of the document.
        collection_info tuple[str, str]: The id
            and key of collection document belongs to.
        data (Optional[dict[str, Any]]): The data of the document.
        id (Optional[str]): The unique identifier of the document.
        client (Client): The client to connect.
        tags (dict[str, str]): The tags associated with the document.

    """

    def __init__(
        self,
        client: Client,
        key: str,
        collection_info: tuple[str, str],
        numerous_doc_ref: Optional[CollectionDocumentReference] = None,
    ) -> None:
        self.key: str = key
        self.collection_id: str = collection_info[0]
        self.collection_key: str = collection_info[1]
        self._client: Client = client
        self.document_id: Optional[str] = None
        self.data: Optional[dict[str, Any]] = None

        if numerous_doc_ref is not None:
            dict_of_tags = {tag.key: tag.value for tag in numerous_doc_ref.tags}
            self.data = base64_to_dict(numerous_doc_ref.data)
            self.document_id = numerous_doc_ref.id
            self._tags: dict[str, str] = (
                dict_of_tags if dict_of_tags is not None else {}
            )

    @property
    def exists(self) -> bool:
        """Check if the document exists."""
        return self.document_id is not None

    @property
    def tags(self) -> dict[str, str]:
        """Get the tags for the document."""
        if self.document_id is not None:
            return self._tags

        msg = "Cannot get tags from a non-existent document."
        raise ValueError(msg)

    def set(self, data: dict[str, Any]) -> None:
        """
        Set the data for the document.

        Args:
        ----
            data (dict[str, Any]): The data to set for the document.

        Raises:
        ------
            ValueError: If the document data setting fails.

        """
        base64_data = dict_to_base64(data)
        document = self._client.set_collection_document(
            self.collection_id, self.key, base64_data
        )
        if document is not None:
            self.document_id = document.id
        else:
            msg = "Failed to delete the document."
            raise ValueError(msg)
        self.data = data

    def get(self) -> Optional[dict[str, Any]]:
        """
        Get the data of the document.

        Returns
        -------
            dict[str, Any]: The data of the document.

        Raises
        ------
            ValueError: If the document does not exist.

        """
        if not self.exists:
            msg = "Document does not exist."
            raise ValueError(msg)
        self._fetch_data(self.key)
        return self.data

    def delete(self) -> None:
        """
        Delete the document.

        Raises
        ------
            ValueError: If the document does not exist or deletion failed.

        """
        if self.document_id is not None:
            deleted_document = self._client.delete_collection_document(self.document_id)

            if deleted_document is not None and deleted_document.id == self.document_id:
                self.document_id = None
                self.data = None
                self._tags = {}
            else:
                msg = "Failed to delete the document."
                raise ValueError(msg)
        else:
            msg = "Cannot delete a non-existent document."
            raise ValueError(msg)

    def tag(self, key: str, value: str) -> None:
        """
        Add a tag to the document.

        Args:
        ----
            key (str): The tag key.
            value (str): The tag value.

        Raises:
        ------
            ValueError: If the document does not exist.

        """
        if self.document_id is not None:
            tagged_document = self._client.add_collection_document_tag(
                self.document_id, TagInput(key=key, value=value)
            )
        else:
            msg = "Cannot tag a non-existent document."
            raise ValueError(msg)

        if tagged_document is not None:
            self.tags.update({tag.key: tag.value for tag in tagged_document.tags})

    def tag_delete(self, tag_key: str) -> None:
        """
        Delete a tag from the document.

        Args:
        ----
            tag_key (str): The key of the tag to delete.

        Raises:
        ------
            ValueError: If the document does not exist.

        """
        if self.document_id is not None:
            tagged_document = self._client.delete_collection_document_tag(
                self.document_id, tag_key
            )
        else:
            msg = "Cannot delete tag from a non-existent document."
            raise ValueError(msg)

        if tagged_document is not None:
            self._tags = {tag.key: tag.value for tag in tagged_document.tags}

    def _fetch_data(self, document_key: str) -> None:
        """Fetch the data from the server."""
        if self.document_id is not None:
            document = self._client.get_collection_document(
                self.collection_key, document_key
            )
        else:
            msg = "Cannot fetch data from a non-existent document."
            raise ValueError(msg)

        if document is not None:
            self.data = base64_to_dict(document.data)
