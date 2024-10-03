import threading
from multiprocessing import process

from marimo._server.router import APIRouter
from starlette.requests import Request
from starlette.routing import Route

from .cookies import set_cookies_impl
from .tempfile import TempFileCookieStorage


def patch_cookies(marimo_args: list[str]) -> None:
    from marimo._server.api.deps import AppState
    from marimo._server.api.endpoints import execution
    from marimo._server.api.utils import parse_request
    from marimo._server.models.models import (
        BaseResponse,
        InstantiateRequest,
        SuccessResponse,
    )

    ident = _thread_ident if marimo_args and marimo_args[0] == "edit" else _proc_ident
    cookies = TempFileCookieStorage(ident)
    set_cookies_impl(cookies)

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


def _thread_ident() -> str:
    return f"thread-{threading.get_ident()}"


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
