"""Access cookies in a marimo notebook."""

from typing import Optional, Protocol


class CookiesNotPatchedError(Exception): ...


class Cookies(Protocol):
    def set(self, c: dict[str, str]) -> None:
        pass

    def get(self) -> dict[str, str]:
        pass


_cookies: Optional[Cookies] = None


def set_cookies_impl(impl: Cookies) -> None:
    global _cookies  # noqa: PLW0603
    _cookies = impl


def cookies() -> dict[str, str]:
    if _cookies is None:
        raise CookiesNotPatchedError
    return _cookies.get()
