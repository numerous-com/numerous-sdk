import pytest

from numerous._client import Client, _get_client


def test_open_client_returns_new_client(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", "organization-id")

    client = _get_client()

    assert isinstance(client, Client)
