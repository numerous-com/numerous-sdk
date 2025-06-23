"""Helper functions for handling .numerous-ignore files for bulk operations."""

from __future__ import annotations

from pathlib import Path  # noqa: TCH003

import pathspec
from pathspec import pattern as pathspec_pattern


def _load_ignore_patterns(ignore_file_path: Path) -> list[pathspec_pattern.Pattern]:
    """
    Load ignore patterns from a gitignore-style file.

    Args:
        ignore_file_path: Path to the ignore file.

    Returns:
        A list of compiled pathspec.Pattern objects.
        Returns an empty list if the ignore file does not exist or is empty.

    """
    patterns: list[pathspec_pattern.Pattern] = []
    if ignore_file_path.is_file():
        with ignore_file_path.open("r", encoding="utf-8") as f:
            lines = [
                line
                for line in f.read().splitlines()
                if line.strip() and not line.strip().startswith("#")
            ]
            if lines:
                spec = pathspec.PathSpec.from_lines("gitwildmatch", lines)
                patterns = list(spec.patterns)
    return patterns


def _should_ignore_path(
    path_to_check: Path, patterns: list[pathspec_pattern.Pattern], base_path: Path
) -> bool:
    """
    Check if a given path should be ignored based on the loaded patterns.

    The path_to_check is made relative to base_path before matching.

    Args:
        path_to_check: The Path object of the file or directory to check.
        patterns: A list of compiled pathspec.Pattern objects.
        base_path: The root path from which ignore patterns are defined
                   (e.g., collection root).

    Returns:
        True if the path should be ignored, False otherwise.

    """
    if not patterns:
        return False

    spec = pathspec.PathSpec(patterns)

    try:
        relative_path = path_to_check.relative_to(base_path)
    except ValueError:
        return False

    path_str = str(relative_path)
    if path_to_check.is_dir() and not path_str.endswith("/"):
        path_str += "/"

    return bool(spec.match_file(path_str))
