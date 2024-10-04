import json
import tempfile
from pathlib import Path
from typing import Callable


class TempFileCookieStorage:
    def __init__(self, ident: Callable[[], str]) -> None:
        self._ident = ident
        self._tempdir = Path(tempfile.mkdtemp("_cookies"))

    def get(self) -> dict[str, str]:
        try:
            with (self._tempdir / self._ident()).open("r") as f:
                c = json.load(f)
        except (json.decoder.JSONDecodeError, TypeError, FileNotFoundError):
            return {}

        if not isinstance(c, dict):
            msg = (
                f"unexpected cookies data stored expected dict, got {type(c).__name__}"
            )
            raise TypeError(msg)
        return c

    def set(self, c: dict[str, str]) -> None:
        with (self._tempdir / self._ident()).open("w+") as f:
            json.dump(c, f)
