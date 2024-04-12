import json
from io import StringIO
from pathlib import Path
from textwrap import dedent

import pytest
from numerous.appdev.commands import (
    read_app,
)


@pytest.fixture()
def default_app_file(tmp_path: Path) -> Path:
    appfile = tmp_path / "app.py"
    # fmt: off
    appfile.write_text(dedent("""
        from numerous.apps import app, field

        @app
        class MyApp:
            field: str = field(label="My Field", default="my default")
    """))
    return appfile


def test_read_prints_expected_data_model(
    default_app_file: Path,
) -> None:
    output = StringIO()
    read_app(default_app_file, "MyApp", output=output)

    expected_data_model = json.dumps(
        {
            "app": {
                "name": "MyApp",
                "title": "MyApp",
                "elements": [
                    {
                        "name": "field",
                        "label": "My Field",
                        "default": "my default",
                        "type": "string",
                    },
                ],
            },
        },
    )
    assert output.getvalue() == expected_data_model


def test_read_app_with_invalid_app_class_raises_appnotfound_error(
    default_app_file: Path,
) -> None:
    output = StringIO()

    read_app(default_app_file, "NonExistingApp", output)

    expected_error = json.dumps(
        {
            "error": {
                "appnotfound": {
                    "app": "NonExistingApp",
                    "found_apps": ["MyApp"],
                },
            },
        },
    )
    assert output.getvalue() == expected_error


def test_read_app_with_syntax_error_raises_syntaxerror(
    tmp_path: Path,
) -> None:
    appfile = tmp_path / "app.py"
    # fmt: off
    src = dedent("""
        from numerous.apps import app, field

        @app
        class MyApp:
            field: str = field():some syntax error:
    """)
    appfile.write_text(src)
    output = StringIO()

    read_app(appfile, "MyApp", output=output)

    expected_context = "\n".join([  # noqa: FLY002
        "    field: str = field():some syntax error:",
        "                        ^",
    ])
    expected_error = json.dumps({
        "error": {
            "appsyntax": {
                "msg": "invalid syntax",
                "context": expected_context,
                "pos": {"line": 6, "offset": 25}},
            },
        },
    )
    actual = output.getvalue()
    assert actual == expected_error


def test_read_app_with_import_error_raises_modulenotfounderror(
    tmp_path: Path,
) -> None:
    appfile = tmp_path / "app.py"
    # fmt: off
    src = dedent("""
        from numerous.apps import app, field
        from weirdlibrary import stuff

        @app
        class MyApp:
            field: str = field()
    """)
    appfile.write_text(src)
    output = StringIO()

    read_app(appfile, "MyApp", output=output)


    expected_error = json.dumps({
        "error": {
            "modulenotfound": {
                "module": "weirdlibrary",
            },
        },
    })
    assert output.getvalue() == expected_error


def test_read_app_raising_exception_at_module_level_returns_unknown_error(
    tmp_path: Path,
) -> None:
    appfile = tmp_path / "app.py"
    # fmt: off
    src = dedent("""
        from numerous.apps import app, field

        raise Exception("raising at module level")

        @app
        class MyApp:
            field: str = field()
    """)
    appfile.write_text(src)
    output = StringIO()

    read_app(appfile, "MyApp", output=output)

    # fmt: off
    expected_traceback = dedent(f"""
        Traceback (most recent call last):
          File "{appfile}", line 3, in <module>
        Exception: raising at module level
    """).lstrip("\n")
    expected_error = json.dumps({
        "error": {
            "unknown": {
                "typename": "Exception",
                "traceback": expected_traceback,
            },
        },
    })
    actual = output.getvalue()
    assert actual == expected_error


def test_read_deprecated_tool_app_prints_expected_error(tmp_path: Path) -> None:
    output = StringIO()
    appfile = tmp_path / "app.py"
    # fmt: off
    appfile.write_text(dedent("""
        from numerous.tools import tool, field

        @tool
        class MyApp:
            field: str = field(label="My Field", default="my default")
    """))
    read_app(appfile, "MyApp", output=output)

    import numerous
    deprecated_module_file = Path(numerous.__file__).parent / "tools" / "__init__.py"
    # fmt: off
    expected_traceback = dedent(f"""
    Traceback (most recent call last):
      File "{appfile}", line 1, in <module>
      File "{deprecated_module_file}", line 11, in <module>
        raise RuntimeError(msg)
    RuntimeError: You are trying to import from 'numerous.tools', which is deprecated. Use 'numerous.apps', and the @app decorator instead.
    """).lstrip("\n")  # noqa: E501
    expected_output = json.dumps(
        {
            "error": {
                "unknown": {
                    "typename": "RuntimeError",
                    "traceback": expected_traceback,
                },
            },
        },
    )
    actual_output = output.getvalue()
    assert actual_output == expected_output


def test_read_app_without_empty_line_prints_expected_data_model(tmp_path: Path) -> None:
    appfile = tmp_path / "app.py"
    # fmt: off
    appfile.write_text(dedent("""
        from numerous.apps import app, field

        @app
        class MyApp:
            field: str = field(label="My Field", default="my default")"""))

    output = StringIO()
    read_app(appfile, "MyApp", output=output)

    expected_data_model = json.dumps(
        {
            "app": {
                "name": "MyApp",
                "title": "MyApp",
                "elements": [
                    {
                        "name": "field",
                        "label": "My Field",
                        "default": "my default",
                        "type": "string",
                    },
                ],
            },
        },
    )
    assert output.getvalue() == expected_data_model
