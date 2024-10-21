from typing import Any, Dict
from numerous.user_session import BaseSession
import panel as pn

class PanelSession(BaseSession):
    
    def _get_cookies(self) -> Dict[str, Any]:
        return pn.request.cookies
    
session = PanelSession()