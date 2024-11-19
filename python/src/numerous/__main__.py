"""An entrypoint for creating and running app sessions."""

import argparse
import logging
import sys

from numerous.experimental.marimo._run import run_marimo


log = logging.getLogger(__name__)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    cmd_parsers = parser.add_subparsers(title="Command", dest="cmd")
    marimo_parser = cmd_parsers.add_parser("marimo")
    marimo_parser.add_argument("marimo_args", nargs=argparse.REMAINDER)
    ns = parser.parse_args()

    if ns.cmd == "marimo":
        run_marimo(ns.marimo_args)
    else:
        print(f"Unsupported command {ns.cmd!r}")  # noqa: T201
        sys.exit(1)
