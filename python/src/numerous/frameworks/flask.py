from typing import Any, Dict
from numerous.user_session import BaseSession
from flask import request

class FlaskSession(BaseSession):
    
    def _get_cookies(self) -> Dict[str, Any]:
        return request.cookies
    
session = FlaskSession()