from numerous.collection import NumerousCollection
from numerous.user import User


def test_user_collection_property_returns_numerous_collection() -> None:
    user = User(id="123", name="John Doe")
    assert isinstance(user.collection, NumerousCollection)

def test_user_collection_property_uses_user_id() -> None:
    user = User(id="123", name="John Doe")
    if user.collection is None:
        msg = "Collection is None"
        raise ValueError(msg)
    assert user.collection.key == "123"

def test_from_user_info_creates_user_with_correct_attributes() -> None:
    user_info = {"user_id": "456", "name": "Jane Smith"}
    user = User.from_user_info(user_info)

    assert user.id == "456"
    assert user.name == "Jane Smith"

def test_from_user_info_returns_user_instance() -> None:
    user_info = {"user_id": "789", "name": "Alice Johnson"}
    user = User.from_user_info(user_info)
    assert isinstance(user, User)
