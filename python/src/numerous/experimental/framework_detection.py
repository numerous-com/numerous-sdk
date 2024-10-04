import sys
from importlib import import_module

class FrameworkDetector:
    def __init__(self):
        self.framework = self._detect_framework()
        self._framework_module = None

    def _detect_framework(self):
        frameworks = ['streamlit', 'marimo', 'dash', 'panel', 'flask', 'fastapi']
        for fw in frameworks:
            if fw in sys.modules:
                self._import_framework(fw)
                
                return fw.capitalize()
        raise ValueError("No framework detected")

    def _import_framework(self, framework):
        try:
            self._framework_module = import_module(framework)
            if framework == 'dash':
                import flask
                globals()['flask'] = flask
        except ImportError:
            raise ImportError(f"Framework {framework} not found")

    def get_framework(self):
        return self.framework

    def get_cookies(self):
        if self.framework == 'Streamlit':
            return self._framework_module.context.cookies
        elif self.framework == 'Marimo':
            return {}
        elif self.framework == 'Dash':
            return flask.request.cookies
        elif self.framework == 'Panel':
            return self._framework_module.state.cookies
        elif self.framework == 'Flask':
            return self._framework_module.request.cookies
        elif self.framework == 'Fastapi':
            return lambda request: self._framework_module.Request.cookies
        else:
            return {}
