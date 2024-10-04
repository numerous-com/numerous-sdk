"""Run patched marimo."""

from marimo._cli.cli import main

from ._cookies.patch import patch_cookies


def run_marimo(args: list[str]) -> None:
    patch_cookies(args)

    main(args=args, prog_name="python -m numerous marimo")
