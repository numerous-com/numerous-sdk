"""
Organization module for managing organization data.

This module defines the Organization class which represents a user's organization
and provides access to organization-specific data.
"""

from __future__ import annotations

import time
from typing import TYPE_CHECKING

from numerous.organization.exception import OrganizationNotFoundError


if TYPE_CHECKING:
    from numerous.organization._client import Client


_UPDATE_INTERVAL_SECONDS = 300


class Organization:
    """
    Represents a Numerous platform organization.

    Attributes:
        id (str): The unique identifier for the organization.
        slug (str): The slug of the organization. Can be updated and cached to
            minimize API calls.

    """

    id: str
    _client: Client
    _slug: str | None = None
    _last_update_time: float = 0.0

    def __init__(self, organization_id: str, _client: Client) -> None:
        self.id = organization_id
        self._client = _client

    @property
    def slug(self) -> str:
        """
        Get the organization slug.

        The slug is cached and updated at most once every 5 minutes.

        Returns:
            The organization slug.

        """
        current_time = time.time()
        if (
            self._slug is None
            or current_time - self._last_update_time > _UPDATE_INTERVAL_SECONDS
        ):
            org_data = self._client.get_organization(self.id)
            if org_data is None:
                raise OrganizationNotFoundError(organization_id=self.id)

            self._slug = org_data.slug
            self._last_update_time = current_time

        return self._slug
