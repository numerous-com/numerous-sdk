"""
Mock implementations for Task 1.0: Mock Backend Unit Testing.

This package provides comprehensive mock implementations for:
- MockExecutionBackend: Full mock backend for task execution
- MockTaskControlHandler: Simulated progress/status/logging with state tracking
- Mock session management with in-memory state persistence
- Pytest fixtures for reusable mock backend setup
"""

from .backend import MockExecutionBackend
from .handler import MockTaskControlHandler
from .session import MockSessionManager
from .fixtures import mock_backend, mock_handler, mock_session_manager

__all__ = [
    'MockExecutionBackend',
    'MockTaskControlHandler', 
    'MockSessionManager',
    'mock_backend',
    'mock_handler',
    'mock_session_manager'
] 