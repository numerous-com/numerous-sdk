from typing import Any, Dict
from numerous.user_session import BaseSession
from numerous.experimental import marimo

class MarimoSession(BaseSession):
    
    def _get_cookies(self) -> Dict[str, Any]:
        return marimo.cookies()
    
session = MarimoSession()