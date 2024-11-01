# Sessions

The `Session` object represents a session of a user interacting with an app. From a session, you can access user information and the cookies set for the session. You obtain a session by calling the `get_session` method from one of the [supported frameworks](frameworks.md).

Due to each framework having a slightly different way to access the session through cookies, the `Session` object is framework-specific. Make sure to use the correct import for the framework you are using.

## ::: numerous.user_session.Session
    options:
        show_root_heading: true
