"""Parse proxy-verified caller identity for Numerous App MCP and API requests."""

from __future__ import annotations

import base64
import json
import os
from dataclasses import dataclass
from typing import TYPE_CHECKING, Any, Literal

from numerous._utils.local import is_local_mode, local_user
from numerous.organization import Organization, organization_from_env
from numerous.organization._get_client import get_client as get_org_client
from numerous.session.user import User


if TYPE_CHECKING:
    from collections.abc import Mapping

    from numerous.organization._client import Client as OrgClient


HEADER_AUTH_TYPE = "x-numerous-auth-type"
HEADER_USER_ID = "x-numerous-user-id"
HEADER_USER_EMAIL = "x-numerous-user-email"
HEADER_USER_NAME = "x-numerous-user-name"
HEADER_ORGANIZATION_ID = "x-numerous-organization-id"
HEADER_ORGANIZATION_SLUG = "x-numerous-organization-slug"
HEADER_ORGANIZATION_NAME = "x-numerous-organization-name"
HEADER_API_KEY_ID = "x-numerous-api-key-id"
HEADER_API_KEY_NAME = "x-numerous-api-key-name"
HEADER_UI_SESSION_ID = "x-numerous-ui-session-id"

USER_INFO_COOKIE_NAME = "numerous_user_info"


@dataclass(frozen=True)
class CallerIdentity:
    """Verified caller identity forwarded by the Numerous proxy."""

    auth_type: Literal["user", "api_key"]
    user: User | None = None
    organization_id: str | None = None
    organization_slug: str | None = None
    organization_name: str | None = None
    api_key_id: str | None = None
    api_key_name: str | None = None
    ui_session_id: str | None = None

    @property
    def is_user(self) -> bool:
        """Return True if the request authenticated as an individual user."""
        return self.auth_type == "user" and self.user is not None

    @property
    def is_api_key(self) -> bool:
        """Return True if the request authenticated with an App API key."""
        return self.auth_type == "api_key"

    def organization(self, _client: OrgClient | None = None) -> Organization:
        """Return a Numerous Organization object for this caller's organization."""
        if self.organization_id:
            client = _client if _client is not None else get_org_client()
            return Organization(organization_id=self.organization_id, _client=client)
        return organization_from_env(_client=_client)


def _decode_user_info_cookie(raw_cookie: str) -> User | None:
    try:
        decoded = base64.b64decode(raw_cookie).decode("utf-8")
        data = json.loads(decoded)
    except (ValueError, UnicodeDecodeError):
        return None
    if not isinstance(data, dict) or "user_id" not in data:
        return None
    return User(
        id=str(data["user_id"]),
        name=str(data.get("user_full_name", "")),
        email=str(data["user_email"]) if data.get("user_email") is not None else None,
    )


def _non_empty(value: str | None) -> str | None:
    if value is None:
        return None
    stripped = value.strip()
    return stripped or None


def caller_from_headers(
    headers: Mapping[str, str],
    cookies: Mapping[str, str] | None = None,
    *,
    allow_local_fallback: bool = True,
) -> CallerIdentity:
    """
    Build a verified CallerIdentity from HTTP headers (and optional cookies).

    On Numerous, the reverse proxy strips any client-supplied ``X-Numerous-*``
    headers and ``numerous_user_info`` cookies before injecting verified values.
    """
    lowered = {str(k).lower(): str(v) for k, v in headers.items()}
    auth_type = _non_empty(lowered.get(HEADER_AUTH_TYPE))
    org_id = _non_empty(lowered.get(HEADER_ORGANIZATION_ID))
    org_slug = _non_empty(lowered.get(HEADER_ORGANIZATION_SLUG))
    org_name = _non_empty(lowered.get(HEADER_ORGANIZATION_NAME))
    ui_session_id = _non_empty(lowered.get(HEADER_UI_SESSION_ID))

    if auth_type == "api_key":
        return CallerIdentity(
            auth_type="api_key",
            user=None,
            organization_id=org_id,
            organization_slug=org_slug,
            organization_name=org_name,
            api_key_id=_non_empty(lowered.get(HEADER_API_KEY_ID)),
            api_key_name=_non_empty(lowered.get(HEADER_API_KEY_NAME)),
            ui_session_id=None,
        )

    user_id = _non_empty(lowered.get(HEADER_USER_ID))
    if auth_type == "user" or user_id is not None:
        if user_id is None:
            msg = "Missing X-Numerous-User-Id header for user-authenticated request"
            raise ValueError(msg)
        user = User(
            id=user_id,
            name=_non_empty(lowered.get(HEADER_USER_NAME)) or "",
            email=_non_empty(lowered.get(HEADER_USER_EMAIL)),
        )
        return CallerIdentity(
            auth_type="user",
            user=user,
            organization_id=org_id,
            organization_slug=org_slug,
            organization_name=org_name,
            ui_session_id=ui_session_id,
        )

    if cookies and (raw_cookie := cookies.get(USER_INFO_COOKIE_NAME)):
        cookie_user = _decode_user_info_cookie(raw_cookie)
        if cookie_user is not None:
            return CallerIdentity(
                auth_type="user",
                user=cookie_user,
                organization_id=org_id,
                organization_slug=org_slug,
                organization_name=org_name,
                ui_session_id=ui_session_id,
            )

    if allow_local_fallback and is_local_mode():
        return CallerIdentity(
            auth_type="user",
            user=local_user,
            organization_id=org_id or _non_empty(os.getenv("NUMEROUS_ORGANIZATION_ID")),
            organization_slug=org_slug,
            organization_name=org_name,
            ui_session_id=ui_session_id,
        )

    msg = "Verified Numerous caller identity headers are missing"
    raise ValueError(msg)


def caller_from_fastmcp_context(
    ctx: Any,  # noqa: ANN401
    *,
    allow_local_fallback: bool = True,
) -> CallerIdentity:
    """Extract verified CallerIdentity from a FastMCP tool Context."""
    request_context = getattr(ctx, "request_context", None)
    request = getattr(request_context, "request", None)
    headers: Mapping[str, str] = getattr(request, "headers", None) or {}
    cookies: Mapping[str, str] | None = getattr(request, "cookies", None)
    return caller_from_headers(
        headers,
        cookies,
        allow_local_fallback=allow_local_fallback,
    )
