from typing import Generator
from unittest.mock import MagicMock, patch

import pytest

from numerous.experimental.framework_detection import FrameworkDetector


@pytest.fixture
def mock_import_module() -> Generator[MagicMock, None, None]:
    with patch("numerous.experimental.framework_detection.import_module") as mock:
        yield mock

def test_detect_framework_streamlit() -> None:
    with patch.object(FrameworkDetector, "_detect_framework", return_value="Streamlit"):
        detector = FrameworkDetector()
        assert detector.framework == "Streamlit"

def test_detect_framework_marimo() -> None:
    with patch.object(FrameworkDetector, "_detect_framework", return_value="Marimo"):
        detector = FrameworkDetector()
        assert detector.framework == "Marimo"

def test_detect_framework_not_found(mock_import_module: MagicMock) -> None:
    mock_import_module.side_effect = ImportError
    with pytest.raises(ValueError, match="No framework detected"):
        FrameworkDetector()

def test_import_framework_success(mock_import_module: MagicMock) -> None:
    mock_module = MagicMock()
    mock_import_module.return_value = mock_module
    detector = FrameworkDetector()
    module = detector._import_framework("streamlit")
    assert module == mock_module

def test_import_framework_failure(mock_import_module: MagicMock) -> None:
    mock_import_module.side_effect = ImportError
    detector = FrameworkDetector()
    with pytest.raises(ValueError, match="No framework detected"):
        detector._import_framework("streamlit")


@pytest.mark.parametrize(("framework", "expected_cookies"), [
    ("Streamlit", {"key": "value"}),
    ("Marimo", {"key": "value"}),
    ("Dash", {"key": "value"}),
    ("Panel", {"key": "value"}),
    ("Flask", {"key": "value"}),
    ("Fastapi", {"key": "value"}),
])
def test_get_cookies(
    framework: str,
    expected_cookies: dict[str, str],
    mock_import_module: MagicMock
) -> None:
    mock_module = MagicMock()
    mock_import_module.return_value = mock_module

    if framework == "Streamlit":
        mock_module.context.cookies = expected_cookies
    elif framework == "Marimo":
        with patch("numerous.experimental.framework_detection.\
                   FrameworkDetector._import_framework") as mock_import, \
             patch("numerous.experimental.marimo._cookies.cookies._cookies")\
                  as mock_cookies:
            mock_cookies.get.return_value = expected_cookies
            mock_import.return_value = MagicMock()
            detector = FrameworkDetector()
            detector.framework = framework
            assert detector.get_cookies() == expected_cookies
            return
    elif framework in ["Dash", "Flask"]:
        mock_module.request.cookies = expected_cookies
    elif framework == "Panel":
        mock_module.state.cookies = expected_cookies
    elif framework == "Fastapi":
        mock_module.Request.cookies = expected_cookies

    detector = FrameworkDetector()
    detector.framework = framework
    # ruff: noqa: SLF001
    detector._framework_module = mock_module  # Consider making this attribute public

    assert detector.get_cookies() == expected_cookies

def test_get_cookies_unknown_framework() -> None:
    detector = FrameworkDetector()
    detector.framework = "Unknown"
    detector._framework_module = MagicMock()  # Consider making this attribute public

    assert detector.get_cookies() == {}

def test_get_cookies_framework_not_imported() -> None:
    detector = FrameworkDetector()
    detector._framework_module = None  # Consider making this attribute public

    with pytest.raises(RuntimeError, match="Framework module not imported"):
        detector.get_cookies()
