import tempfile
from multiprocessing import process
from pathlib import Path

from marimo._server.router import APIRouter
from starlette.requests import Request
from starlette.routing import Route

from .cookies import set_cookie_storage
from .tempfile import FileCookieStorage


def patch_cookies(marimo_args: list[str]) -> None:
    from marimo._server.api.deps import AppState
    from marimo._server.api.endpoints import execution
    from marimo._server.api.utils import parse_request
    from marimo._server.models.models import (
        BaseResponse,
        InstantiateRequest,
        SuccessResponse,
    )

    if marimo_args and marimo_args[0] == "edit":
        cookies = FileCookieStorage(Path(), _edit_ident)
    else:
        cookie_path = tempfile.mkdtemp(prefix="marimo_cookies")
        cookies = FileCookieStorage(Path(cookie_path), _proc_ident)
    set_cookie_storage(cookies)

    router: APIRouter = execution.router
    _remove_old_route(router)

    @router.post("/instantiate")
    async def instantiate(
        *,
        request: Request,
    ) -> BaseResponse:
        cookies.set(request.cookies)
        app_state = AppState(request)
        body = await parse_request(request, cls=InstantiateRequest)
        app_state.require_current_session().instantiate(body)

        return SuccessResponse()


def _edit_ident() -> str:
    """When running in edit mode, we use a constant session identifier."""
    return "edit-session-cookie"


def _proc_ident() -> str:
    return f"proc-{process.current_process().ident or 'none'}"


def _remove_old_route(router: APIRouter) -> None:
    to_remove = None
    for r in router.routes:
        if isinstance(r, Route) and r.path == "/instantiate":
            to_remove = r

    if to_remove is None:
        msg = "could not find route"
        raise RuntimeError(msg)

    router.routes.remove(to_remove)
