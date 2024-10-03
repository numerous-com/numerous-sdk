from __future__ import annotations

import typing as t
from multiprocessing import process

from .cookies import set_cookies_impl
from .tempfile import TempFileCookieStorage


if t.TYPE_CHECKING:
    from fastapi import FastAPI, Request, Response


def add_marimo_cookie_middleware(
    app: FastAPI, session_ident: t.Callable[[], str] | None = None
) -> None:
    """
    Add a middleware that enables accessing cookies in marimo apps.

    Parameters
    ----------
    app : FastAPI
        The FastAPI app to add the middleware on.

    session_ident : Optional[Callable[[], str]]
        The identity function which must return a unique value for each session.

    """
    cookies = TempFileCookieStorage(session_ident or _ident)
    set_cookies_impl(cookies)

    @app.middleware("http")  # type: ignore[misc]
    async def middleware(
        request: Request,
        call_next: t.Callable[[Request], t.Awaitable[Response]],
    ) -> Response:
        cookies.set(request.cookies)
        return await call_next(request)


def _ident() -> str:
    pid = process.current_process().ident
    if pid is None:
        return "no-process-id"
    return str(pid)
