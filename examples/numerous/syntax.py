# mypy: ignore-errors

from numerous import app


@app
class SyntaxErrorApp:
    my_syntax_error str  # noqa: E999 ]
    my_field: str
