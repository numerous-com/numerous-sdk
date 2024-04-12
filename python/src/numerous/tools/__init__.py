"""
A deprecated package.

Only raises an error to help users fix their usage of this package.
"""

msg = (
    "You are trying to import from 'numerous.tools', which is deprecated. Use "
    "'numerous.apps', and the @app decorator instead."
)
raise RuntimeError(msg)
