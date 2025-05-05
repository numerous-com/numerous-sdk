class APIURLMissingError(Exception):
    _msg = "NUMEROUS_API_URL environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class APIAccessTokenMissingError(Exception):
    _msg = "NUMEROUS_API_ACCESS_TOKEN environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)


class OrganizationIDMissingError(Exception):
    _msg = "NUMEROUS_ORGANIZATION_ID environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)
