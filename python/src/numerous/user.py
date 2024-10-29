"""Module for handling user-related functionality."""

from dataclasses import dataclass
from typing import Any, Optional

from numerous._client._graphql_client import GraphQLClient
from numerous.collection import NumerousCollection, collection


@dataclass
class User:
    id: str
    name: str
    _client: Optional[GraphQLClient] = None

    @property
    def collection(self) -> Optional["NumerousCollection"]:
        """
        Get the NumerousCollection associated with this user.

        Returns:
            NumerousCollection | None: The collection for this user,
            or None if not found.

        """
        return collection("users", self._client).collection(self.id)

    @staticmethod
    def from_user_info(
        user_info: dict[str, Any], _client: Optional[GraphQLClient] = None
    ) -> "User":
        """
        Create a User instance from a dictionary of user information.

        Args:
            user_info (dict[str, Any]): A dictionary containing user information.
            _client (GraphQLClient | None): A GraphQL client instance.

        Returns:
            User: A new User instance created from the provided information.

        """
        return User(id=user_info["user_id"], name=user_info["name"], _client=_client)
