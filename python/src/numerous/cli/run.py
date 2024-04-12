"""Launches the correct binary for the users OS."""

import os
import platform
import subprocess
import sys
from pathlib import Path

from numerous.cli._upgrade import check_for_updates


def get_executable_os_name() -> str:
    system = platform.system()
    if system == "Linux":
        return "linux"
    if system == "Darwin":
        return "darwin"
    if os.name == "nt":
        return "windows"

    print(  # noqa: T201
        f"Sorry but operating system {system} is not currently supported :(",
    )
    sys.exit(1)


def get_executable_arch_name() -> str:
    arch = platform.machine().lower()
    if arch in ("x86_64", "amd64"):
        return "amd64"
    if arch == "arm64":
        return "arm64"
    print(  # noqa: T201
        f"Sorry but architecutre {arch} is not currently supported :(",
    )
    sys.exit(1)


def get_executable_name() -> str:
    return f"build/{get_executable_os_name()}_{get_executable_arch_name()}"


def main() -> None:
    check_for_updates()
    exe_name = get_executable_name()
    exe_path = Path(__file__).parent / exe_name
    try:
        process = subprocess.Popen(args=[str(exe_path)] + sys.argv[1:])
        process.wait()
    except KeyboardInterrupt:
        process.kill()


if __name__ == "__main__":
    main()
