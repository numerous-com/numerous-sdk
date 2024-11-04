from __future__ import annotations

import tempfile
import typing as t
from multiprocessing import process
from pathlib import Path

from .cookies import use_cookie_storage
from .files import FileCookieStorage


if t.TYPE_CHECKING:
    from fastapi import FastAPI, Request, Response


def add_marimo_cookie_middleware(
    app: FastAPI,
    session_ident: t.Callable[[], str] | None = None,
    cookies_dir: Path | None = None,
) -> None:
    """
    Add a middleware that enables accessing cookies in marimo apps.

    Args:
        app: The FastAPI app to add the middleware on.
        session_ident: The identity function which must return a unique value for each
            session.
        cookies_dir: Path to the directory where cookies are stored.

    """
    cookies_dir = cookies_dir or Path(
        tempfile.mkdtemp(prefix="numerous_marimo_cookies")
    )
    cookies = FileCookieStorage(cookies_dir, session_ident or _ident)
    use_cookie_storage(cookies)

    @app.middleware("http")  # type: ignore[misc, unused-ignore]
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
