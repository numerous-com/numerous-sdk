import pytest
from unittest.mock import patch, Mock
from numerous.collection.collection import collection

ORGANIZATION_ID = "test_org"
COLLECTION_NAME = "test_collection"
NESTED_COLLECTION_NAME = "nested_test_collection"
COLLECTION_KEY = Mock(key="test_key", id="test_id")
NESTED_COLLECTION_KEY = Mock(key="nested_test_key", id="nested_test_id")


def test_collection():
    mock_client = Mock()
    mock_client.get_collection_key.return_value = COLLECTION_KEY
    result = collection(COLLECTION_NAME,mock_client)
    
    mock_client.get_collection_key.assert_called_once()
    mock_client.get_collection_key.assert_called_once_with(COLLECTION_NAME)
    assert result.key == COLLECTION_KEY.key
    assert result.id == COLLECTION_KEY.id

def test_numerous_collection():
    mock_client = Mock()
    mock_client.get_collection_key.return_value = COLLECTION_KEY
    mock_client.get_collection_key_with_parent.return_value = NESTED_COLLECTION_KEY
    result = collection(COLLECTION_NAME,mock_client)

    nested_result =  result.collection(NESTED_COLLECTION_NAME)
    

    assert nested_result.key == NESTED_COLLECTION_KEY.key
    assert nested_result.id == NESTED_COLLECTION_KEY.id
