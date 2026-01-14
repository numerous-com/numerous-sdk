"""Exceptions related to deployment."""


class DeploymentIDMissingError(Exception):
    _msg = "NUMEROUS_DEPLOYMENT_ID environment variable is not set"

    def __init__(self) -> None:
        super().__init__(self._msg)
