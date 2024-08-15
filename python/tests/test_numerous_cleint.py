import unittest
from unittest.mock import Mock

from numerous.numerous_client import NumerousClient, _open_client


class TestNumerousClient(unittest.TestCase):

    def test_open_client(self, mock_client:Mock)->None:
        """Testing clinet."""
        mock_client_instance = mock_client.return_value
        client = _open_client("org_id")

        assert isinstance(client, NumerousClient)
        assert client.client == mock_client_instance
