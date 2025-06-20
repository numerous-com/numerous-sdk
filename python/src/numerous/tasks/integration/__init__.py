"""
Integration module for remote API communication.

This module provides handlers and clients for communicating with the Numerous API
during integration testing and remote task execution.
"""

from .remote_handler import RemoteTaskControlHandler

__all__ = ["RemoteTaskControlHandler"] 