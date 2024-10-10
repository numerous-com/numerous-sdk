import pytest
from unittest.mock import Mock, patch
from numerous.experimental.user import User
from numerous.collection import NumerousCollection

@pytest.fixture
def mock_framework_detector():
    return Mock()

@pytest.fixture
def user(mock_framework_detector):
    user = User()
    user.framework_detection = mock_framework_detector
    return user

def test_user_id_with_cookie(user, mock_framework_detector):
    mock_framework_detector.get_cookies.return_value = {"numerous_user_id": "test-user-123"}
    assert user.id == "test-user-123"

def test_user_id_without_cookie(user, mock_framework_detector):
    mock_framework_detector.get_cookies.return_value = {}
    assert user.id == "local-user"

@patch('numerous.experimental.user.collection')
def test_collection_success(mock_collection, user):
    mock_user_collection = Mock()
    mock_collection.return_value.collection.return_value = mock_user_collection
    mock_subcollection = Mock(spec=NumerousCollection)
    mock_user_collection.collection.return_value = mock_subcollection

    result = user.collection("test-collection")
    
    assert result == mock_subcollection
    mock_collection.assert_called_once_with("users")
    mock_collection.return_value.collection.assert_called_once_with(user.id)
    mock_user_collection.collection.assert_called_once_with("test-collection")

@patch('numerous.experimental.user.collection')
def test_collection_user_not_found(mock_collection, user):
    mock_collection.return_value.collection.return_value = None

    with pytest.raises(ValueError, match="User collection not found"):
        user.collection("test-collection")