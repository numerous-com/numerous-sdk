import json
from pathlib import Path
from typing import Callable


class FileCookieStorage:
    def __init__(self, cookie_path: Path, ident: Callable[[], str]) -> None:
        self._ident = ident
        self._cookie_path = cookie_path

    def get(self) -> dict[str, str]:
        try:
            with self._cookie_file().open("r") as f:
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
        with self._cookie_file().open("w+") as f:
            json.dump(c, f)

    def _cookie_file(self) -> Path:
        return self._cookie_path / self._ident()
