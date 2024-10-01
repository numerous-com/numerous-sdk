"""Exceptions related to collections."""


class ParentCollectionNotFoundError(Exception):
    _msg = "Parent collection not found"

    def __init__(self, collection_id: str) -> None:
        self.collection_id = collection_id
        super().__init__(self._msg)
