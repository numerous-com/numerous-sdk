"""Module for handling user-related functionality."""

from typing import Any, Callable, Dict
from numerous.collection import NumerousCollection, collection
import base64
import json

class User:
    def __init__(self, cookie_getter: Callable[[], Dict[str, Any]]) -> None:
        self._cookie_getter = cookie_getter

    def _get_user_info(self) -> Dict[str, Any]:
        """Decode and return the user info from the cookie."""
        cookies = self._cookie_getter()
        user_info_b64 = cookies.get("numerous_user_info")
        if not user_info_b64:
            return {}
        try:
            user_info_json = base64.b64decode(user_info_b64).decode('utf-8')
            return json.loads(user_info_json)
        except (ValueError, json.JSONDecodeError):
            return {}

    @property
    def id(self) -> str:
        """Get the user ID from the decoded cookie or return a default value."""
        user_info = self._get_user_info()
        return user_info.get("user_id") or "local-user"
    
    @property
    def name(self) -> str:
        """Get the user name from the decoded cookie or return a default value."""
        user_info = self._get_user_info()
        return user_info.get("user_full_name") or "Local User"

    def collection(self, collection_key: str) -> NumerousCollection|None:
        """Get a user-specific collection."""
        user_collection = collection("users").collection(self.id)
        if user_collection is None:
            error_message = "User collection not found"
            raise ValueError(error_message)
        return user_collection.collection(collection_key)
