from importlib import import_module
from typing import Any, Dict, Optional
from numerous.user import User
from abc import ABC, abstractmethod


class BaseSession(ABC):
    def __init__(self):
        self._cookies: Dict[str, Any] = {}
        self._user: User = User(self._get_cookies)

    @abstractmethod
    def _get_cookies(self) -> Dict[str, Any]:
        raise NotImplementedError("Cookies getter is not implemented")

    @property
    def cookies(self) -> Dict[str, Any]:
        return self._get_cookies()

    @property
    def user(self) -> Optional[Dict[str, Any]]:
        return self._user