from typing import Any, Dict
from numerous.user_session import BaseSession
from fastapi import Request

class FastapiSession(BaseSession):
    
    def _get_cookies(self) -> Dict[str, Any]:
        return Request.cookies
    
session = FastapiSession()