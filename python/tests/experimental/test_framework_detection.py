import pytest
from unittest.mock import patch, MagicMock
from numerous.experimental.framework_detection import FrameworkDetector

@pytest.fixture
def mock_import_module():
    with patch('numerous.experimental.framework_detection.import_module') as mock:
        yield mock

def test_detect_framework_streamlit(mock_import_module):
    with patch.object(FrameworkDetector, '_detect_framework', return_value='Streamlit'):
        detector = FrameworkDetector()
        assert detector.framework == "Streamlit"

def test_detect_framework_marimo(mock_import_module):
    with patch.object(FrameworkDetector, '_detect_framework', return_value='Marimo'):
        detector = FrameworkDetector()
        assert detector.framework == "Marimo"

def test_detect_framework_not_found(mock_import_module):
    mock_import_module.side_effect = ImportError
    with pytest.raises(ValueError, match="No framework detected"):
        FrameworkDetector()

def test_import_framework_success(mock_import_module):
    mock_module = MagicMock()
    mock_import_module.return_value = mock_module
    detector = FrameworkDetector()
    module = detector._import_framework("streamlit")
    assert module == mock_module

def test_import_framework_failure(mock_import_module):
    mock_import_module.side_effect = ImportError
    with pytest.raises(ValueError, match="No framework detected"):
        detector = FrameworkDetector()
        detector._import_framework("streamlit")
        
@pytest.mark.parametrize("framework,expected_cookies", [
    ("Streamlit", {"key": "value"}),
    ("Marimo", {"key": "value"}),
    ("Dash", {"key": "value"}),
    ("Panel", {"key": "value"}),
    ("Flask", {"key": "value"}),
    ("Fastapi", {"key": "value"}),
])
def test_get_cookies(framework, expected_cookies, mock_import_module):
    mock_module = MagicMock()
    mock_import_module.return_value = mock_module

    if framework == "Streamlit":
        mock_module.context.cookies = expected_cookies
    elif framework == "Marimo":
        with patch('numerous.experimental.framework_detection.FrameworkDetector._import_framework') as mock_import, \
             patch('numerous.experimental.marimo._cookies.cookies._cookies') as mock_cookies:
            mock_cookie_storage = MagicMock()
            mock_cookie_storage.get.return_value = expected_cookies
            mock_cookies.get.return_value = expected_cookies
            #mock_cookies.return_value = mock_cookie_storage
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
    detector._framework_module = mock_module

    assert detector.get_cookies() == expected_cookies

def test_get_cookies_unknown_framework():
    detector = FrameworkDetector()
    detector.framework = "Unknown"
    detector._framework_module = MagicMock()

    assert detector.get_cookies() == {}

def test_get_cookies_framework_not_imported():
    detector = FrameworkDetector()
    detector._framework_module = None
    
    with pytest.raises(RuntimeError, match="Framework module not imported"):
        detector.get_cookies()
