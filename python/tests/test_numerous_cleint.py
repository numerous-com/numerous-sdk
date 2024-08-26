import pytest
from numerous._client import Client, _open_client


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch:pytest.MonkeyPatch)->None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


def test_open_client()->None:
    """Testing client."""
    client = _open_client()

    assert isinstance(client, Client)
