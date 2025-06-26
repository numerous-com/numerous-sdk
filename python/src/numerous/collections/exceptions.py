"""Exceptions related to collections."""


class ParentCollectionNotFoundError(Exception):
    _msg = "Parent collection not found"

    def __init__(self, collection_id: str) -> None:
        self.collection_id = collection_id
        super().__init__(self._msg)


class OrganizationMismatchError(Exception):
    _msg = "Organization mismatch"

    def __init__(
        self,
        parent_id: str,
        parent_organization_id: str,
        requested_organization_id: str,
    ) -> None:
        self.parent_id = parent_id
        self.parent_organization_id = parent_organization_id
        self.requested_organization_id = requested_organization_id
        super().__init__(self._msg)
