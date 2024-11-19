"""Module for handling user-related functionality."""

from dataclasses import dataclass
from typing import Any, Optional

from numerous.collection import CollectionReference, collection
from numerous.collection._client import Client


@dataclass
class User:
    """
    Represents a Numerous platform user.

    Attributes:
        id (str): The unique identifier for the user.
        name (str): The full name of the user.

    """

    id: str
    name: str
    _client: Optional[Client] = None

    @property
    def collection(self) -> CollectionReference:
        """
        A user's collection.

        A collection scoped for the current user. It is a child-collection of the
        root collection with the key "users", and has the user's ID as key.

        Equivalent to:
        >>> collection("users").collection(user.id)

        Returns:
            NumerousCollection: The collection for this user.

        """
        return collection("users", self._client).collection(self.id)

    @staticmethod
    def from_user_info(
        user_info: dict[str, Any], _client: Optional[Client] = None
    ) -> "User":
        """
        Create a User instance from a dictionary of user information.

        Args:
            user_info (dict[str, Any]): A dictionary containing user information.

        Returns:
            User: A new User instance created from the provided information.

        """
        return User(
            id=user_info["user_id"],
            name=user_info["user_full_name"],
            _client=_client,
        )
