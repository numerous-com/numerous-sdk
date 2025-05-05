"""
Organization client for managing organization data.

This module defines a client protocol, which specifies all required methods needed to
manage organization data.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import Protocol


@dataclass
class OrganizationData:
    id: str
    slug: str


class Client(Protocol):
    def get_organization(self, organization_id: str) -> OrganizationData | None: ...
