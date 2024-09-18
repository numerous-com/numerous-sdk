"""Miscellaneous utilities."""

from typing import TypeVar


class _MissingType:
    pass


MISSING = _MissingType()

AppT = TypeVar("AppT")
