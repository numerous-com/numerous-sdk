from marimo._server.router import APIRouter
from starlette.requests import Request
from starlette.routing import Route

from .cookies import set_cookies


def patch_cookies() -> None:
    from marimo._server.api.deps import AppState
    from marimo._server.api.endpoints import execution
    from marimo._server.api.utils import parse_request
    from marimo._server.models.models import (
        BaseResponse,
        InstantiateRequest,
        SuccessResponse,
    )

    router: APIRouter = execution.router

    remove_old_route(router)

    @router.post("/instantiate")
    async def instantiate(
        *,
        request: Request,
    ) -> BaseResponse:
        set_cookies(request.cookies)
        app_state = AppState(request)
        body = await parse_request(request, cls=InstantiateRequest)
        app_state.require_current_session().instantiate(body)

        return SuccessResponse()


def remove_old_route(router: APIRouter) -> None:
    to_remove = None
    for r in router.routes:
        if isinstance(r, Route) and r.path == "/instantiate":
            to_remove = r

    if to_remove is None:
        msg = "could not find route"
        raise RuntimeError(msg)

    router.routes.remove(to_remove)
