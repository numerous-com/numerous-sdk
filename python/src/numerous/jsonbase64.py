"""dict to base64 conversion."""

import base64
import json
from typing import Any, cast


DECODED_JSON_NOT_DICT = "Decoded JSON is not a dictionary"


def dict_to_base64(input_dict: dict[str, Any]) -> str:
    json_str = json.dumps(input_dict)
    json_bytes = json_str.encode("utf-8")
    base64_bytes = base64.b64encode(json_bytes)
    return base64_bytes.decode("utf-8")


def base64_to_dict(base64_str: str) -> dict[str, Any]:
    json_str = base64.b64decode(base64_str).decode("utf-8")
    result = json.loads(json_str)
    if not isinstance(result, dict):
        raise TypeError(DECODED_JSON_NOT_DICT)
    return cast(dict[str, str], result)
