"""Access cookies in a marimo notebook."""

from __future__ import annotations

import warnings
from pathlib import Path

from numerous.experimental.marimo._cookies.files import FileCookieStorage


_cookie_storage: FileCookieStorage | None = None


def use_cookie_storage(cs: FileCookieStorage) -> None:
    global _cookie_storage  # noqa: PLW0603
    _cookie_storage = cs


def use_fallback_cookie_storage() -> FileCookieStorage:
    fallback = _fallback_cookie_storage()
    use_cookie_storage(fallback)
    return fallback


def cookies() -> dict[str, str]:
    if _cookie_storage is None:
        return use_fallback_cookie_storage().get()
    return _cookie_storage.get()


def _fallback_cookie_storage() -> FileCookieStorage:
    warnings.warn(
        "marimo has not been patched for cookie support, or is running in edit "
        "mode. Using fallback cookie storage.",
        RuntimeWarning,
        stacklevel=2,
    )
    return FileCookieStorage(Path(), lambda: "numerous-marimo-cookie")
