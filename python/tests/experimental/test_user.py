from unittest.mock import Mock, patch

import pytest

from numerous.collection import NumerousCollection
from numerous.experimental.user import User


@pytest.fixture
def mock_framework_detector() -> Mock:
    return Mock()

@pytest.fixture
def user(mock_framework_detector: Mock) -> User:
    user = User()
    user.framework_detection = mock_framework_detector
    return user

def test_user_id_with_cookie(user: User, mock_framework_detector: Mock) -> None:
    mock_framework_detector.get_cookies.return_value =\
          {"numerous_user_id": "test-user-123"}
    assert user.id == "test-user-123"

def test_user_id_without_cookie(user: User, mock_framework_detector: Mock) -> None:
    mock_framework_detector.get_cookies.return_value = {}
    assert user.id == "local-user"

@patch("numerous.experimental.user.collection")
def test_collection_success(mock_collection: Mock, user: User) -> None:
    mock_user_collection = Mock()
    mock_collection.return_value.collection.return_value = mock_user_collection
    mock_subcollection = Mock(spec=NumerousCollection)
    mock_user_collection.collection.return_value = mock_subcollection

    result = user.collection("test-collection")

    assert result == mock_subcollection
    mock_collection.assert_called_once_with("users")
    mock_collection.return_value.collection.assert_called_once_with(user.id)
    mock_user_collection.collection.assert_called_once_with("test-collection")

@patch("numerous.experimental.user.collection")
def test_collection_user_not_found(mock_collection: Mock, user: User) -> None:
    mock_collection.return_value.collection.return_value = None

    with pytest.raises(ValueError, match="User collection not found"):
        user.collection("test-collection")
