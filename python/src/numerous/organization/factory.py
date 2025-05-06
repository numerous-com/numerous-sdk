"""Creating Organization instances."""

import os
from typing import Optional

from numerous.organization._client import Client
from numerous.organization._get_client import get_client
from numerous.organization.exception import OrganizationIDMissingError
from numerous.organization.organization import Organization


def organization_from_env(_client: Optional[Client] = None) -> Organization:
    """
    Create a Organization instance from environment variables.

    Args:
        _client: A client instance.

    Returns:
        A new Organization instance created environment variables.

    """
    if _client is None:
        _client = get_client()

    organization_id = os.getenv("NUMEROUS_ORGANIZATION_ID")
    if organization_id is None:
        raise OrganizationIDMissingError

    return Organization(organization_id=organization_id, _client=_client)
