from __future__ import annotations

import contextlib
import json
import urllib.request
from dataclasses import asdict, dataclass
from datetime import datetime
from importlib.metadata import version as metadata_version
from pathlib import Path
from typing import Any

from packaging import version

from .color import blue, green


RESPONSE_STATUS_OK = 200
SECONDS_IN_5_MIN = 300


@dataclass
class VersionState:
    latest_version: str | None = None
    last_checked_at: str | None = None


def _fetch_data_from_pipy() -> dict[str, Any]:
    pipy_url = "https://pypi.org/pypi/numerous/json"
    with urllib.request.urlopen(pipy_url) as response:  # noqa: S310
        if response.status == RESPONSE_STATUS_OK:
            data = response.read()
            encoding = response.info().get_content_charset("utf-8")
            return json.loads(data.decode(encoding))  # type: ignore[no-any-return]

        message = f"HTTP request failed with status code {response.status}"
        raise Exception(  # noqa: TRY002
            message,
        )


def get_version_from_pipy() -> str | None:
    try:
        response = _fetch_data_from_pipy()
        return response["info"]["version"]  # type: ignore[no-any-return]
    except:  # noqa: E722
        return None


def check_for_updates() -> None:
    try:
        _check_for_updates_internal()
    except Exception:  # noqa: BLE001
        contextlib.suppress(Exception)


def _check_for_updates_internal() -> None:
    # Find the state file
    state_file = Path(".numerous/state.json")
    home_directory = Path.home()
    file_path = home_directory / state_file

    # Load the state file or create a default state if it doesn't exist
    state: VersionState
    try:
        with Path.open(file_path) as file:
            data = json.load(file)
            state = VersionState(**data)
    except FileNotFoundError:
        state = VersionState()

    # Maybe update the state with the latest version
    state = _update_with_latest_version(state)

    if not state.latest_version:
        # We couldn't get the latest version, so do nothing
        return

    current_version = metadata_version("numerous")
    if current_version and version.parse(state.latest_version) > version.parse(
        current_version,
    ):
        upg_str = f"{current_version} → {state.latest_version}"
        message = f"Heads up - A newer version of Numerous is available {upg_str}"
        print(blue(message))  # noqa: T201
        install_cmd = green("pip install --upgrade numerous")
        guide_cmd = f"Please upgrade by running {install_cmd} for the best experience."
        print(guide_cmd)  # noqa: T201
        print()  # noqa: T201

    _maybe_create_directory(file_path)
    with Path.open(file_path, "w") as file:
        json.dump(asdict(state), file)


def _maybe_create_directory(file_path: Path) -> None:
    state_directory = Path(file_path).parent
    if not Path.exists(state_directory):
        Path.mkdir(state_directory, parents=True)


def _update_with_latest_version(state: VersionState) -> VersionState:
    if state.last_checked_at:
        last_checked_date = datetime.strptime(
            state.last_checked_at,
            "%Y-%m-%d %H:%M:%S",
        ).astimezone()
        if (datetime.now().astimezone() - last_checked_date).seconds < SECONDS_IN_5_MIN:
            return state

    try:
        version = get_version_from_pipy()
        state.latest_version = version
        state.last_checked_at = (
            datetime.now().astimezone().strftime("%Y-%m-%d %H:%M:%S")
        )
        return state  # noqa: TRY300
    except Exception:  # noqa: BLE001
        return state
