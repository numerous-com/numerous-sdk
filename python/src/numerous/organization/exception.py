"""Exceptions related to organization."""


class OrganizationIDMissingError(Exception):
    _msg = "NUMEROUS_ORGANIZATION_ID environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class OrganizationNotFoundError(Exception):
    _msg = "Organization not found"

    def __init__(self, organization_id: str) -> None:
        super().__init__(f"{self._msg}: {organization_id}")
