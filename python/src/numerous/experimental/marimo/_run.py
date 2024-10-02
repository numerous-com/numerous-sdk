"""Run patched marimo."""

from ._cookies.patch import patch_cookies


def run_marimo(args: list[str]) -> None:
    from marimo._cli.cli import main

    patch_cookies()

    main(args=args, prog_name="python -m numerous marimo")
