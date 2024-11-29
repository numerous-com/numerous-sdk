from __future__ import annotations

from unittest.mock import Mock, call

import pytest

from numerous.collections._client import Client, CollectionIdentifier
from numerous.session.user import User


TEST_ORGANIZATION_ID = "test-organization-id"
TEST_USER_COLLECTION_KEY = "users"
TEST_USER_COLLECTION_ID = "test-user-collection-id"
TEST_USER_ID = "test-collection-id"
TEST_COLLECTION_KEY = "test-collection-key"


@pytest.fixture(autouse=True)
def _set_env_vars(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("NUMEROUS_API_URL", "url_value")
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", TEST_ORGANIZATION_ID)
    monkeypatch.setenv("NUMEROUS_API_ACCESS_TOKEN", "token")


@pytest.fixture
def client() -> Mock:
    def mock_collection_reference(
        collection_key: str, parent_collection_id: str | None = None
    ) -> CollectionIdentifier:
        ref = (collection_key, parent_collection_id)
        if ref == (TEST_USER_COLLECTION_KEY, None):
            return CollectionIdentifier(
                id=TEST_USER_COLLECTION_ID, key=TEST_USER_COLLECTION_KEY
            )
        if ref == (TEST_USER_ID, TEST_USER_COLLECTION_ID):
            return CollectionIdentifier(id=TEST_USER_ID, key=TEST_COLLECTION_KEY)
        pytest.fail("unexpected mock call")

    client = Mock(Client)
    client.collection_reference.side_effect = mock_collection_reference

    return client


def test_user_collection_property_returns_expected_collection(client: Mock) -> None:
    user = User(id=TEST_USER_ID, name="John Doe", _client=client)

    assert user.collection is not None
    assert user.collection.id == TEST_USER_ID
    assert user.collection.key == TEST_COLLECTION_KEY


def test_user_collection_property_makes_expected_calls(client: Mock) -> None:
    user = User(id=TEST_USER_ID, name="John Doe", _client=client)

    user.collection  # noqa: B018

    client.collection_reference.assert_has_calls(
        [
            call("users"),
            call(
                collection_key=TEST_USER_ID,
                parent_collection_id=TEST_USER_COLLECTION_ID,
            ),
        ]
    )


def test_from_user_info_returns_user_with_correct_attributes(client: Mock) -> None:
    user_info = {"user_id": TEST_USER_ID, "user_full_name": "Jane Smith"}

    user = User.from_user_info(user_info, _client=client)

    assert user.id == TEST_USER_ID
    assert user.name == "Jane Smith"
