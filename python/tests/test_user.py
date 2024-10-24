from numerous.collection import NumerousCollection
from numerous.user import User


def test_user_collection_property_returns_numerous_collection():
    user = User(id="123", name="John Doe")
    assert isinstance(user.collection, NumerousCollection)

def test_user_collection_property_uses_user_id():
    user = User(id="123", name="John Doe")
    assert user.collection.key == "123"

def test_from_user_info_creates_user_with_correct_attributes():
    user_info = {"user_id": "456", "name": "Jane Smith"}
    user = User.from_user_info(user_info)
    assert user.id == "456" and user.name == "Jane Smith"

def test_from_user_info_returns_user_instance():
    user_info = {"user_id": "789", "name": "Alice Johnson"}
    user = User.from_user_info(user_info)
    assert isinstance(user, User)
