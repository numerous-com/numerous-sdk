"""
Bulk operations for Numerous Collections.

This package provides high-level functions to download and upload
entire collection hierarchies between the Numerous platform and the local filesystem.
"""

from __future__ import annotations


__all__ = ["bulk_download", "bulk_upload"]

from .main import bulk_download, bulk_upload
