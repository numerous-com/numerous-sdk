"""Miscellaneous utilities."""

from os import getenv
from typing import TypeVar


class _MissingType:
    pass


MISSING = _MissingType()

AppT = TypeVar("AppT")


def is_local_mode() -> bool:
    url = getenv("NUMEROUS_API_URL")
    return url is None
