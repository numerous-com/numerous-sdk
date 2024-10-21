from typing import Any, Dict
from numerous.user_session import BaseSession
import streamlit as st

class StreamlitSession(BaseSession):
    
    def _get_cookies(self) -> Dict[str, Any]:
        return st.context.cookies
    
session = StreamlitSession()