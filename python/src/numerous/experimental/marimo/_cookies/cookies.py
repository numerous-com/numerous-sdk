"""Access cookies in a marimo notebook."""

from typing import Optional, Protocol


class CookiesNotPatchedError(Exception): ...


class CookieStorage(Protocol):
    def set(self, c: dict[str, str]) -> None:
        """Set cookies for the current session."""

    def get(self) -> dict[str, str]:
        """Get cookies for the current session."""


_cookies: Optional[CookieStorage] = None


def set_cookies_impl(impl: CookieStorage) -> None:
    global _cookies  # noqa: PLW0603
    _cookies = impl


def cookies() -> dict[str, str]:
    if _cookies is None:
        raise CookiesNotPatchedError
    return _cookies.get()
