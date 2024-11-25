"""Local mode utilities."""

from os import getenv

from numerous.session.user import User


local_user = User(id="local_user", name="Local User")


def is_local_mode() -> bool:
    url = getenv("NUMEROUS_API_URL")
    return url is None
