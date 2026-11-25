import base64
import json
from types import SimpleNamespace
from unittest import mock

import pytest

from numerous._utils.local import local_user
from numerous.mcp import (
    UISessionStore,
    caller_from_fastmcp_context,
    caller_from_headers,
)
from numerous.organization._client import Client as OrgClient


def test_caller_from_headers_user_token(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "https://api.numerous.com/query")

    caller = caller_from_headers(
        {
            "X-Numerous-Auth-Type": "user",
            "X-Numerous-User-Id": "user-uuid-1",
            "X-Numerous-User-Name": "Alice Example",
            "X-Numerous-User-Email": "alice@example.com",
            "X-Numerous-Organization-Id": "org-api-1",
            "X-Numerous-Organization-Slug": "acme",
            "X-Numerous-Organization-Name": "Acme Corp",
            "X-Numerous-UI-Session-Id": "ui-sess-42",
        }
    )

    assert caller.is_user is True
    assert caller.is_api_key is False
    assert caller.auth_type == "user"
    assert caller.user is not None
    assert caller.user.id == "user-uuid-1"
    assert caller.user.name == "Alice Example"
    assert caller.user.email == "alice@example.com"
    assert caller.organization_id == "org-api-1"
    assert caller.organization_slug == "acme"
    assert caller.organization_name == "Acme Corp"
    assert caller.ui_session_id == "ui-sess-42"

    org_client = mock.Mock(spec=OrgClient)
    org = caller.organization(_client=org_client)
    assert org.id == "org-api-1"


def test_caller_from_headers_api_key_never_populates_user_or_ui_session(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "https://api.numerous.com/query")

    caller = caller_from_headers(
        {
            "X-Numerous-Auth-Type": "api_key",
            "X-Numerous-User-Id": "should-be-ignored",
            "X-Numerous-UI-Session-Id": "should-be-ignored",
            "X-Numerous-Organization-Id": "org-api-1",
            "X-Numerous-Organization-Slug": "acme",
            "X-Numerous-API-Key-Id": "key-uuid-1",
            "X-Numerous-API-Key-Name": "CI Key",
        }
    )

    assert caller.is_api_key is True
    assert caller.is_user is False
    assert caller.user is None
    assert caller.ui_session_id is None
    assert caller.organization_id == "org-api-1"
    assert caller.api_key_id == "key-uuid-1"
    assert caller.api_key_name == "CI Key"


def test_caller_from_headers_cookie_fallback(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "https://api.numerous.com/query")
    encoded = base64.b64encode(
        json.dumps(
            {
                "user_id": "cookie-user-1",
                "user_full_name": "Cookie User",
                "user_email": "cookie@example.com",
            }
        ).encode()
    ).decode()

    caller = caller_from_headers({}, cookies={"numerous_user_info": encoded})

    assert caller.is_user is True
    assert caller.user is not None
    assert caller.user.id == "cookie-user-1"
    assert caller.user.name == "Cookie User"
    assert caller.user.email == "cookie@example.com"


def test_caller_from_headers_local_fallback_and_missing_error(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("NUMEROUS_API_URL", raising=False)

    local_caller = caller_from_headers({})
    assert local_caller.user == local_user

    monkeypatch.setenv("NUMEROUS_API_URL", "https://api.numerous.com/query")
    with pytest.raises(ValueError, match="Verified Numerous caller identity"):
        caller_from_headers({})


def test_caller_from_fastmcp_context(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "https://api.numerous.com/query")
    ctx = SimpleNamespace(
        request_context=SimpleNamespace(
            request=SimpleNamespace(
                headers={
                    "x-numerous-auth-type": "user",
                    "x-numerous-user-id": "fastmcp-user",
                    "x-numerous-user-name": "FastMCP User",
                    "x-numerous-organization-id": "org-1",
                    "x-numerous-ui-session-id": "sess-99",
                },
                cookies={},
            )
        )
    )

    caller = caller_from_fastmcp_context(ctx)

    assert caller.is_user is True
    assert caller.user is not None
    assert caller.user.id == "fastmcp-user"
    assert caller.ui_session_id == "sess-99"


def test_ui_session_store_crud() -> None:
    store = UISessionStore()

    assert store.get(None) is None
    assert store.get("missing") is None

    store.set("sess-1", {"quarter": "2026-Q3"})
    assert store.get("sess-1") == {"quarter": "2026-Q3"}

    updated = store.update("sess-1", region="EMEA")
    assert updated == {"quarter": "2026-Q3", "region": "EMEA"}
    assert store.get("sess-1") == {"quarter": "2026-Q3", "region": "EMEA"}

    store.delete("sess-1")
    assert store.get("sess-1") is None
