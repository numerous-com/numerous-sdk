"""Module for handling user-related functionality."""

from numerous.collection import collection, NumerousCollection
from numerous.experimental.framework_detection import FrameworkDetector


class User:
    def __init__(self) -> None:
        self.framework_detection = FrameworkDetector()

    @property
    def id(self) -> str:
        """Get the user ID from cookies or return a default value."""
        cookies = self.framework_detection.get_cookies()
        return cookies.get("numerous_user_id") or "local-user"

    def collection(self, collection_key: str) -> NumerousCollection:
        """Get a user-specific collection."""
        user_collection = collection("users").collection(self.id)
        if user_collection is None:
            error_message = "User collection not found"
            raise ValueError(error_message)
        return user_collection.collection(collection_key)
