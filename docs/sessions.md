# Sessions

The `Session` object represents a session of a user interacting with an app.
From a session, you can access user information and the cookies set for the
session. You obtain a session by calling the `get_session` method from one of
the [supported frameworks](frameworks.md).

Due to each framework having a slightly different way to access the session
through cookies, the `Session` object is framework-specific. Make sure to use
the correct import for the framework you are using.

## Example: Streamlit

Below is an example of how to use the session to access cookies and active user
information. See the [user page](user.md) for more information about how to
work with user information.

```py
import numerous.frameworks.streamlit

session = numerous.frameworks.streamlit.get_session()

# Access cookies of the session
for name, value in session.cookies.items():
    st.write(f"Cookie '{name}' has value '{value}'")

# Access user information
st.write(f"Your name is {session.user.name}")
```
