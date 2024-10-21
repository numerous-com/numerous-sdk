"""Module for handling user-related functionality."""

from typing import Any, Callable, Dict
from numerous.collection import NumerousCollection, collection


class User:
    def __init__(self, cookie_getter: Callable[[], Dict[str, Any]]) -> None:
        self._cookie_getter = cookie_getter

    @property
    def id(self) -> str:
        """Get the user ID from cookies or return a default value."""
        cookies = self._cookie_getter()
        return cookies.get("numerous_user_id") or "local-user"
    
    @property
    def name(self) -> str:
        """Get the user name from cookies or return a default value."""
        raise NotImplementedError("User name is not supported yet")

    def collection(self, collection_key: str) -> NumerousCollection|None:
        """Get a user-specific collection."""
        user_collection = collection("users").collection(self.id)
        if user_collection is None:
            error_message = "User collection not found"
            raise ValueError(error_message)
        return user_collection.collection(collection_key)
