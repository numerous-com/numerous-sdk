"""Module for integrating Numerous with Streamlit."""

from typing import Any, Dict

import streamlit as st

from numerous import user_session


class StreamlitCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        cookies = {key: str(val) for key, val in st.context.cookies.items()}
        
        return cookies

session = user_session.Session(cg=StreamlitCookieGetter())
