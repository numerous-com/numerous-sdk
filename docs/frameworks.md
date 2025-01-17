# Frameworks

The Numerous SDK supports a number of popular Python web frameworks.

The `frameworks` package provides framework-specific implementations of Numerous features. Currently, the main framework-specific function is `get_session()`, which handles user authentication by accessing session information through cookies in a way that's compatible with each framework's request handling.

You need to add the frameworks to your project's requirements directly since they are not part of or installed with the Numerous SDK.

## Dash

[Dash](https://dash.plotly.com/) is supported by the `numerous.frameworks.dash` module.

### ::: numerous.frameworks.dash.get_session

    options:
        show_root_heading: true

###

## FastAPI

[FastAPI](https://fastapi.tiangolo.com/) is supported by the `numerous.frameworks.fastapi` module.

### ::: numerous.frameworks.fastapi.get_session

    options:
        show_root_heading: true

###

## Flask

[Flask](https://flask.palletsprojects.com/) is supported by the `numerous.frameworks.flask` module.

### ::: numerous.frameworks.flask.get_session

    options:
        show_root_heading: true

###

## Marimo

[Marimo](https://marimo.io/) is supported by the `numerous.frameworks.marimo` module.

### ::: numerous.frameworks.marimo.get_session

    options:
        show_root_heading: true

###

## Panel

[Panel](https://panel.holoviz.org/) is supported by the `numerous.frameworks.panel` module.

### ::: numerous.frameworks.panel.get_session

    options:
        show_root_heading: true

###

## Streamlit

[Streamlit](https://streamlit.io/) is supported by the `numerous.frameworks.streamlit` module.

### ::: numerous.frameworks.streamlit.get_session

    options:
        show_root_heading: true
