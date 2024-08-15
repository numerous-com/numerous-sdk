import os

from numerous.numerous_client import NumerousClient, _open_client


os.environ["NUMEROUS_API_URL"] = "url_value"
os.environ["NUMEROUS_API_ACCESS_TOKEN"] = "token"


def test_open_client()->None:
    """Testing client."""
    client = _open_client()

    assert isinstance(client, NumerousClient)
