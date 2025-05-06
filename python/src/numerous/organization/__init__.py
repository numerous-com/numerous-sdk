"""Package for managing organizations."""

__all__ = ("organization_from_env", "Organization")

from .factory import organization_from_env
from .organization import Organization
