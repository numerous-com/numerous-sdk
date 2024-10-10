"""Module for framework detection functionality."""

from importlib import import_module
from types import ModuleType
from typing import Any, Callable, Dict

import numerous.experimental.marimo


class FrameworkDetector:
    def __init__(self) -> None:
        self.framework = self._detect_framework()
        self._framework_module: Any = self._import_framework(self.framework.lower())

    def _detect_framework(self) -> str:
        frameworks = ["streamlit", "marimo", "dash", "panel", "flask", "fastapi"]
        for fw in frameworks:
            if self._try_import(fw):
                # Found the framework
                return fw.capitalize()
        no_framework_error = "No framework detected"
        raise ValueError(no_framework_error)

    def _try_import(self, module: str) -> bool:
        try:
            import_module(module)
        except ImportError:
            return False
        return True

    def _import_framework(self, framework: str) -> ModuleType:
        try:
            module = import_module(framework)
            if framework == "flask":
                import flask
                globals()["flask"] = flask
        except ImportError as err:
            framework_not_found = f"Framework {framework} not found"
            raise ImportError(framework_not_found) from err
        else:
            return module

    def get_framework(self) -> str:
        """Get the detected framework."""
        return self.framework

    def get_cookies(self) -> Dict[str, Any]:
        """Get cookies based on the detected framework."""
        if self._framework_module is None:
            framework_not_imported_error = "Framework module not imported"
            raise RuntimeError(framework_not_imported_error)

        cookie_getters: Dict[str, Callable[[], Dict[str, Any]]] = {
            "Streamlit": lambda: self._framework_module.context.cookies,
            "Marimo": lambda: numerous.experimental.marimo.cookies(),
            "Dash": lambda: self._framework_module.request.cookies,
            "Panel": lambda: self._framework_module.state.cookies,
            "Flask": lambda: self._framework_module.request.cookies,
            "Fastapi": lambda: self._framework_module.Request.cookies,
        }
        getter = cookie_getters.get(self.framework)
        if getter is None:
            return {}
        return getter()
