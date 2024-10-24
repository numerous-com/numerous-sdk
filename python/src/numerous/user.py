"""Module for handling user-related functionality."""

from dataclasses import dataclass
from typing import Any

from numerous.collection import NumerousCollection, collection


@dataclass
class User:
    id: str
    name: str

    @property
    def collection(self) -> NumerousCollection | None:
        """
        Get the NumerousCollection associated with this user.

        Returns:
            NumerousCollection | None: The collection for this user,
            or None if not found.

        """
        return collection("users").collection(self.id)

    @staticmethod
    def from_user_info(user_info: dict[str, Any]) -> "User":
        """
        Create a User instance from a dictionary of user information.

        Args:
            user_info (dict[str, Any]): A dictionary containing user information.

        Returns:
            User: A new User instance created from the provided information.

        """
        return User(id=user_info["user_id"], name=user_info["name"])
