"""Getting deployment info."""

import os

from numerous.deployment.exception import DeploymentIDMissingError


def get_deployment_id_from_env() -> str:
    """
    Get deployment ID from environment variables.

    Returns:
        The deployment ID from NUMEROUS_DEPLOYMENT_ID environment variable.

    Raises:
        DeploymentIDMissingError: If NUMEROUS_DEPLOYMENT_ID is not set.

    """
    deployment_id = os.getenv("NUMEROUS_DEPLOYMENT_ID")
    if deployment_id is None:
        raise DeploymentIDMissingError

    return deployment_id
