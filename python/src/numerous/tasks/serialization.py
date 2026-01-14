"""Serialization utilities for task inputs and outputs."""

from __future__ import annotations

import base64
import binascii
import json
from typing import Any, Optional


def serialize_task_inputs(inputs: dict[str, Any]) -> str:
    return _serialize_task_data(inputs)


def deserialize_task_inputs(input_json: Optional[str]) -> dict[str, Any]:
    return _deserialize_task_data(input_json)


def serialize_task_output(output: dict[str, Any]) -> str:
    return _serialize_task_data(output)


def deserialize_task_output(output_json: Optional[str]) -> dict[str, Any]:
    return _deserialize_task_data(output_json)


def _serialize_task_data(data: dict[str, Any]) -> str:
    json_str = json.dumps(data)
    return base64.b64encode(json_str.encode("utf-8")).decode("utf-8")


def _deserialize_task_data(json_str: Optional[str]) -> dict[str, Any]:
    if not json_str:
        return {}

    try:
        decoded = base64.b64decode(json_str).decode("utf-8")
    except (ValueError, binascii.Error, UnicodeDecodeError):
        return {"_raw": json_str}

    try:
        data = json.loads(decoded)
        if isinstance(data, dict):
            return data
        return {"_raw": decoded}  # noqa: TRY300
    except (ValueError, json.JSONDecodeError):
        return {"_raw": decoded}


def serialize_task_progress(
    value: float, message: Optional[str] = None
) -> dict[str, Any]:
    if not 0 <= value <= 1:
        msg = f"Progress value must be between 0 and 1, got {value}"
        raise ValueError(msg)

    # Backend expects percent (0-100), convert from fraction (0-1)
    progress_percent = value * 100

    progress: dict[str, Any] = {"value": progress_percent}
    if message is not None:
        progress["message"] = message
    return progress
