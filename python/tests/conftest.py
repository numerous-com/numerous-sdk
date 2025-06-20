"""
Pytest configuration for Numerous SDK tests.

This module configures pytest markers and fixtures for running different
types of tests including unit tests and optional integration tests.
"""

import pytest


def pytest_addoption(parser):
    """Add command-line options for test configuration."""
    parser.addoption(
        "--integration",
        action="store_true",
        default=False,
        help="Run integration tests that require API at localhost:8080"
    )


def pytest_configure(config):
    """Configure pytest markers."""
    config.addinivalue_line(
        "markers", 
        "integration: marks tests as integration tests (require API at localhost:8080)"
    )
    config.addinivalue_line(
        "markers",
        "unit: marks tests as unit tests (default, no external dependencies)"
    )


def pytest_collection_modifyitems(config, items):
    """
    Modify test collection based on command-line options.
    
    If --integration is not passed, skip integration tests.
    """
    if not config.getoption("--integration"):
        skip_integration = pytest.mark.skip(
            reason="Integration tests require --integration flag and API at localhost:8080"
        )
        for item in items:
            if "integration" in item.keywords:
                item.add_marker(skip_integration)


@pytest.fixture(scope="session")
def api_base_url():
    """Base URL for API integration tests."""
    return "http://localhost:8080"


@pytest.fixture(scope="session")
def integration_config():
    """Configuration for integration tests."""
    return {
        "api_url": "http://localhost:8080",
        "timeout": 30,
        "retry_attempts": 3,
        "retry_delay": 1.0
    }


@pytest.fixture
def skip_if_no_integration(request):
    """
    Skip test if integration marker is present but --integration flag not used.
    
    This is a failsafe fixture for tests that should only run with integration flag.
    """
    if "integration" in request.keywords and not request.config.getoption("--integration"):
        pytest.skip("Test requires --integration flag") 