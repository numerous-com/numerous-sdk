"""Launches the correct binary for the users OS."""

import os
import platform
import subprocess
import sys
from pathlib import Path

from .upgrade import check_for_updates


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
        f"Sorry but architecture {arch} is not currently supported :(",
    )
    sys.exit(1)


def get_platform_executable_name() -> str:
    return f"{get_executable_os_name()}_{get_executable_arch_name()}"


def main() -> int:
    check_for_updates()
    bin_dir = Path(__file__).parent / "bin"
    exe_path = bin_dir / "cli"
    if not exe_path.exists():
        exe_name = get_platform_executable_name()
        exe_path = bin_dir / exe_name

    if not exe_path.exists():
        print(  # noqa: T201
            "Numerous CLI executable compatible with your system was not found. :("
        )
        sys.exit(1)

    try:
        process = subprocess.Popen(args=[str(exe_path)] + sys.argv[1:])
        exit_code = process.wait()
    except KeyboardInterrupt:
        process.kill()
        sys.exit(1)
    else:
        sys.exit(exit_code)


if __name__ == "__main__":
    main()
