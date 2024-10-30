from unittest.mock import Mock

import pytest

from numerous import local


@pytest.fixture(autouse=True)
def _ensure_local_mode(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("NUMEROUS_API_URL", raising=False)


def test_marimo_get_session_in_local_mode_returns_expected_local_user(
    tmp_path_factory: pytest.TempPathFactory,
) -> None:
    # patch cookies
    from numerous.experimental.marimo._cookies.cookies import use_cookie_storage
    from numerous.experimental.marimo._cookies.files import FileCookieStorage

    path = tmp_path_factory.mktemp("marimo-cookies")
    use_cookie_storage(FileCookieStorage(path, lambda: "test-ident"))

    from numerous.frameworks.marimo import get_session

    session = get_session()

    assert session.user == local.local_user


def test_streamlit_get_session_in_local_mode_returns_expected_local_user() -> None:
    from numerous.frameworks.streamlit import get_session

    session = get_session()

    assert session.user == local.local_user


def test_fastapi_get_session_in_local_mode_returns_expected_local_user() -> None:
    from fastapi import Request

    from numerous.frameworks.fastapi import get_session

    session = get_session(Request(scope={"type": "http", "headers": {}}))

    assert session.user == local.local_user


def test_flask_get_session_in_local_mode_returns_expected_local_user() -> None:
    from flask import Flask

    from numerous.frameworks.flask import get_session

    with Flask("test_app").test_request_context():
        session = get_session()

        assert session.user == local.local_user


def test_dash_get_session_in_local_mode_returns_expected_local_user() -> None:
    import dash

    from numerous.frameworks.dash import get_session

    app = dash.Dash()
    with app.server.test_request_context():
        session = get_session()

        assert session.user == local.local_user


def test_panel_get_session_in_local_mode_returns_expected_local_user() -> None:
    from bokeh.document import Document
    from panel.io.state import set_curdoc

    from numerous.frameworks.panel import get_session

    mock_doc = Mock(Document)
    mock_doc.session_context.request.cookies = {}

    with set_curdoc(mock_doc):
        session = get_session()

        assert session.user == local.local_user
