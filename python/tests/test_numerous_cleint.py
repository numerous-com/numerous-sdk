from typing import Generator

import pytest
from numerous.numerous_client import NumerousClient, _open_client


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch:Generator)->None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


def test_open_client()->None:
    """Testing client."""
    client = _open_client()

    assert isinstance(client, NumerousClient)
