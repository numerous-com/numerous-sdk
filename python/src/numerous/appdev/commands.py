"""App development CLI command implementations."""

import json
import logging
import sys
import traceback
from dataclasses import asdict, dataclass
from pathlib import Path
from textwrap import indent
from typing import Any, Optional, TextIO, Type

from numerous.data_model import dump_data_model
from numerous.generated.graphql import Client
from numerous.session import Session
from numerous.utils import AppT


log = logging.getLogger(__name__)


@dataclass
class AppNotFoundError(Exception):
    app: str
    existing_apps: list[str]


@dataclass
class AppLoadRaisedError(Exception):
    traceback: str
    typename: str


@dataclass
class AppSyntaxError(Exception):
    msg: str
    context: str
    lineno: int
    offset: int


def read_app(app_module: Path, app_class: str, output: TextIO = sys.stdout) -> None:
    try:
        app_cls = _read_app_definition(app_module, app_class)
    except Exception as e:  # noqa: BLE001
        print_error(output, e)
    else:
        print_app(output, app_cls)


def _transform_lineno(lineno: Optional[int]) -> int:
    if lineno is None:
        return 0
    return lineno - 2


def _transform_offset(offset: Optional[int]) -> int:
    if offset is None:
        return 0
    return offset - 4


def _read_app_definition(app_module: Path, app_class: str) -> Any:  # noqa: ANN401
    scope: dict[str, Any] = {}
    module_text = app_module.read_text()

    # Check for syntax errors with raw code text
    try:
        compile(module_text, str(app_module), "exec")
    except SyntaxError as e:
        text = e.text or ""
        error_pointer = " " * ((e.offset or 0) - 1) + "^"

        # ensure newline between text and pointer
        if not text.endswith("\n"):
            text += "\n"

        raise AppSyntaxError(
            msg=e.msg,
            context=text + error_pointer,
            lineno=e.lineno or 0,
            offset=e.offset or 0,
        ) from e

    indented_module_text = indent(module_text, "    ")
    exception_handled_module_text = (
        "try:\n"
        f"{indented_module_text}\n"
        "except ModuleNotFoundError:\n"
        "    raise\n"
        "except BaseException as e:\n"
        "    __numerous_read_error__ = e\n"
    )

    code = compile(exception_handled_module_text, str(app_module), "exec")

    exec(code, scope)  # noqa: S102

    unknown_error = scope.get("__numerous_read_error__")
    if isinstance(unknown_error, BaseException):
        tb = traceback.TracebackException.from_exception(unknown_error)

        # handle inserted exception handler offsetting position
        for frame in tb.stack:
            if frame.filename == str(app_module):
                frame.lineno = _transform_lineno(frame.lineno)
                if sys.version_info >= (3, 11):
                    frame.colno = _transform_offset(frame.colno)
                    frame.end_lineno = _transform_lineno(frame.end_lineno)
                    frame.end_colno = _transform_offset(frame.end_colno)

        raise AppLoadRaisedError(
            typename=type(unknown_error).__name__,
            traceback="".join(tb.format()),
        )

    try:
        return scope[app_class]
    except KeyError as err:
        raise AppNotFoundError(
            app_class,
            [
                app.__name__
                for app in scope.values()
                if getattr(app, "__numerous_app__", False)
            ],
        ) from err


def print_app(output: TextIO, cls: Type[AppT]) -> None:
    data_model = dump_data_model(cls)
    output.write(json.dumps({"app": asdict(data_model)}))
    output.flush()


def print_error(output: TextIO, error: Exception) -> None:
    if isinstance(error, AppNotFoundError):
        output.write(
            json.dumps(
                {
                    "error": {
                        "appnotfound": {
                            "app": error.app,
                            "found_apps": error.existing_apps,
                        },
                    },
                },
            ),
        )
    elif isinstance(error, AppSyntaxError):
        output.write(
            json.dumps(
                {
                    "error": {
                        "appsyntax": {
                            "msg": error.msg,
                            "context": error.context,
                            "pos": {
                                "line": error.lineno,
                                "offset": error.offset,
                            },
                        },
                    },
                },
            ),
        )
        output.flush()
    elif isinstance(error, ModuleNotFoundError):
        output.write(
            json.dumps(
                {
                    "error": {
                        "modulenotfound": {
                            "module": error.name,
                        },
                    },
                },
            ),
        )
        output.flush()
    elif isinstance(error, AppLoadRaisedError):
        output.write(
            json.dumps(
                {
                    "error": {
                        "unknown": {
                            "typename": error.typename,
                            "traceback": error.traceback,
                        },
                    },
                },
            ),
        )
        output.flush()


async def run_app_session(
    graphql_url: str,
    graphql_ws_url: str,
    session_id: str,
    app_module: Path,
    app_class: str,
) -> None:
    log.info("running %s:%s in session %s", app_module, app_class, session_id)
    gql = Client(graphql_url, ws_url=graphql_ws_url)
    app_cls = _read_app_definition(app_module, app_class)
    session = await Session.initialize(session_id, gql, app_cls)
    await session.run()
    log.info("app session stopped")
