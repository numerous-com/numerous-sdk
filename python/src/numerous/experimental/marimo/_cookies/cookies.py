"""Access cookies in a marimo notebook."""

import json
import tempfile
from multiprocessing import process
from pathlib import Path


_tmp_cookies = Path(tempfile.mkdtemp("_cookies"))


def cookies() -> dict[str, str]:
    try:
        pid = process.current_process().ident
        with (_tmp_cookies / str(pid)).open("r") as f:
            c = json.load(f)
    except (json.decoder.JSONDecodeError, TypeError):
        return {}

    if not isinstance(c, dict):
        msg = f"unexpected cookies data stored expected dict, got {type(c).__name__}"
        raise TypeError(msg)
    return c


def set_cookies(c: dict[str, str]) -> None:
    pid = process.current_process().ident
    with (_tmp_cookies / str(pid)).open("w+") as f:
        json.dump(c, f)
