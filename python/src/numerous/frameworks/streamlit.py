"""Module for integrating Numerous with Streamlit."""

from typing import Dict

import streamlit as st

from numerous import user_session
from numerous.utils import is_local_mode


class StreamlitCookieGetter:
    def cookies(self) -> Dict[str, str]:
        """Get the cookies associated with the current request."""
        local_cookies = {}
        if is_local_mode():
            # Update the cookies on the streamlit server
            local_cookies = user_session.get_encoded_user_info()
        return {
            **local_cookies,
            **{key: str(val) for key, val in st.context.cookies.items()},
        }


session = user_session.Session(cg=StreamlitCookieGetter())


def get_session() -> user_session.Session:
    return user_session.Session(cg=StreamlitCookieGetter())
