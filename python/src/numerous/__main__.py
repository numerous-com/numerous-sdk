"""An entrypoint for creating and running app sessions."""

import argparse
import asyncio
import logging
import sys
from pathlib import Path

from numerous.appdev.commands import (
    read_app,
    run_app_session,
)

log = logging.getLogger(__name__)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    shared_parser = argparse.ArgumentParser()
    shared_parser.add_argument("module_path", type=Path)
    shared_parser.add_argument("class_name")

    cmd_parsers = parser.add_subparsers(title="Command", dest="cmd")
    read_parser = cmd_parsers.add_parser(
        "read",
        parents=[shared_parser],
        add_help=False,
    )
    run_parser = cmd_parsers.add_parser("run", parents=[shared_parser], add_help=False)
    run_parser.add_argument("--graphql-url", required=True)
    run_parser.add_argument("--graphql-ws-url", required=True)
    run_parser.add_argument("session_id")

    ns = parser.parse_args()

    if ns.cmd == "read":
        read_app(ns.module_path, ns.class_name)
    elif ns.cmd == "run":
        coroutine = run_app_session(
            ns.graphql_url,
            ns.graphql_ws_url,
            ns.session_id,
            ns.module_path,
            ns.class_name,
        )
        asyncio.run(coroutine)
    else:
        print(f"Unsupported command {ns.cmd!r}")  # noqa: T201
        sys.exit(1)
