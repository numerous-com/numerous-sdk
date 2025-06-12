"""
Pytest fixtures for Task 1.0: Mock Backend Unit Testing.

Provides reusable pytest fixtures for:
- MockExecutionBackend setup and cleanup
- MockTaskControlHandler configuration  
- MockSessionManager initialization
- Combined mock environment setup
- Automatic state cleanup between tests
"""

import pytest
from typing import Generator, Optional
import logging

from .backend import MockExecutionBackend, MockExecutionMode
from .handler import MockTaskControlHandler
from .session import MockSessionManager

# Import for handler switching
from numerous.tasks.control import set_task_control_handler, get_task_control_handler

logger = logging.getLogger(__name__)


@pytest.fixture
def mock_backend() -> Generator[MockExecutionBackend, None, None]:
    """
    Pytest fixture providing a clean MockExecutionBackend.
    
    Features:
    - Automatically starts up the backend
    - Cleans up state after test completion
    - Provides fresh backend instance for each test
    """
    backend = MockExecutionBackend()
    backend.startup()
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture  
def mock_handler(request) -> Generator[MockTaskControlHandler, None, None]:
    """
    Pytest fixture providing a clean MockTaskControlHandler.
    
    Features:
    - Automatically sets as the global task control handler
    - Restores original handler after test completion
    - Clears handler history after test
    - Supports parameterization for different configurations
    """
    # Check for simulate_delays parameter
    simulate_delays = getattr(request, 'param', {}).get('simulate_delays', False)
    
    # Store original handler
    original_handler = get_task_control_handler()
    
    # Create and set mock handler
    handler = MockTaskControlHandler(simulate_delays=simulate_delays)
    set_task_control_handler(handler)
    
    try:
        yield handler
    finally:
        # Restore original handler
        set_task_control_handler(original_handler)
        handler.clear_history()


@pytest.fixture
def mock_session_manager() -> Generator[MockSessionManager, None, None]:
    """
    Pytest fixture providing a clean MockSessionManager.
    
    Features:
    - Provides fresh session manager for each test
    - Automatically cleans up session state after test
    """
    manager = MockSessionManager()
    
    try:
        yield manager
    finally:
        manager.clear_state()


@pytest.fixture
def mock_environment(mock_backend, mock_handler, mock_session_manager):
    """
    Pytest fixture providing a complete mock environment.
    
    Combines:
    - MockExecutionBackend (started and ready)
    - MockTaskControlHandler (set as global handler)
    - MockSessionManager (clean state)
    
    Returns a tuple of (backend, handler, session_manager) for convenience.
    """
    return mock_backend, mock_handler, mock_session_manager


@pytest.fixture
def mock_backend_immediate() -> Generator[MockExecutionBackend, None, None]:
    """
    Pytest fixture providing a MockExecutionBackend configured for immediate completion.
    """
    backend = MockExecutionBackend()
    backend.startup()
    backend.set_default_execution_mode(MockExecutionMode.IMMEDIATE)
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture
def mock_backend_delayed() -> Generator[MockExecutionBackend, None, None]:
    """
    Pytest fixture providing a MockExecutionBackend configured for delayed completion.
    """
    backend = MockExecutionBackend()
    backend.startup()
    backend.set_default_execution_mode(MockExecutionMode.DELAYED, delay=0.1)
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture
def mock_backend_manual() -> Generator[MockExecutionBackend, None, None]:
    """
    Pytest fixture providing a MockExecutionBackend configured for manual completion.
    """
    backend = MockExecutionBackend()
    backend.startup()
    backend.set_default_execution_mode(MockExecutionMode.MANUAL)
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture
def mock_backend_failure() -> Generator[MockExecutionBackend, None, None]:
    """
    Pytest fixture providing a MockExecutionBackend configured for failure mode.
    """
    backend = MockExecutionBackend()
    backend.startup()
    backend.set_default_execution_mode(MockExecutionMode.FAILURE)
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture
def mock_handler_with_delays() -> Generator[MockTaskControlHandler, None, None]:
    """
    Pytest fixture providing a MockTaskControlHandler with simulated delays.
    """
    # Store original handler
    original_handler = get_task_control_handler()
    
    # Create and set mock handler with delays
    handler = MockTaskControlHandler(simulate_delays=True)
    set_task_control_handler(handler)
    
    try:
        yield handler
    finally:
        # Restore original handler
        set_task_control_handler(original_handler)
        handler.clear_history()


# Parameterized fixtures for different configurations

@pytest.fixture(params=[
    {"mode": MockExecutionMode.IMMEDIATE},
    {"mode": MockExecutionMode.DELAYED, "delay": 0.05},
    {"mode": MockExecutionMode.MANUAL},
])
def mock_backend_parameterized(request) -> Generator[MockExecutionBackend, None, None]:
    """
    Parameterized fixture that provides backends with different execution modes.
    Useful for testing the same logic across different execution scenarios.
    """
    backend = MockExecutionBackend()
    backend.startup()
    
    config = request.param
    mode = config["mode"]
    delay = config.get("delay", 0.0)
    
    backend.set_default_execution_mode(mode, delay)
    
    try:
        yield backend
    finally:
        backend.shutdown()
        backend.clear_state()


@pytest.fixture(params=[False, True])
def mock_handler_parameterized(request) -> Generator[MockTaskControlHandler, None, None]:
    """
    Parameterized fixture that provides handlers with and without delays.
    """
    simulate_delays = request.param
    
    # Store original handler
    original_handler = get_task_control_handler()
    
    # Create and set mock handler
    handler = MockTaskControlHandler(simulate_delays=simulate_delays)
    set_task_control_handler(handler)
    
    try:
        yield handler
    finally:
        # Restore original handler
        set_task_control_handler(original_handler)
        handler.clear_history()


# Utility fixtures for test setup

@pytest.fixture
def cleanup_handlers():
    """
    Fixture to ensure handler cleanup in case of test failures.
    Use this when manually setting up handlers in tests.
    """
    original_handler = get_task_control_handler()
    
    yield
    
    # Always restore original handler
    set_task_control_handler(original_handler)


@pytest.fixture(scope="session")
def mock_logging_setup():
    """
    Session-scoped fixture to configure logging for mock components.
    """
    # Configure logging for mock components
    mock_logger = logging.getLogger("tests.mocks")
    mock_logger.setLevel(logging.INFO)
    
    # Add handler if none exists
    if not mock_logger.handlers:
        handler = logging.StreamHandler()
        formatter = logging.Formatter(
            '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )
        handler.setFormatter(formatter)
        mock_logger.addHandler(handler)
    
    yield mock_logger 