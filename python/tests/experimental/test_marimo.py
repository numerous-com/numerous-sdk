import html
import json

import pytest
from numerous.experimental.marimo import Field
from numerous.experimental.model import BaseModel

number_value = 5
number_min = 0
number_max = 10
text_value = "text"


class DataModel(BaseModel):
    number_field = Field(default=number_value, ge=number_min, le=number_max)  # type: ignore[arg-type]
    text_field = Field(default=text_value)


def test_number_field_has_default_value() -> None:
    model = DataModel()

    assert model.number_field.value == number_value


def test_slider_values_match_field_defintion() -> None:
    model = DataModel()

    marimo_slider = model.number_field.slider(label="Test")

    assert marimo_slider.start == number_min
    assert marimo_slider.stop == number_max
    assert marimo_slider.value == number_value


def test_number_values_match_field_defintion() -> None:
    model = DataModel()

    marimo_number = model.number_field.number(label="Test")

    assert marimo_number.start == number_min
    assert marimo_number.stop == number_max
    assert marimo_number.value == number_value


def test_number_out_of_range_raises_value_error() -> None:
    model = DataModel()

    with pytest.raises(ValueError):  # noqa: PT011
        model.number_field.value = number_max + 1

    with pytest.raises(ValueError):  # noqa: PT011
        model.number_field.value = number_min - 1


def test_text_values_match_field_defintion() -> None:
    model = DataModel()

    marimo_text = model.text_field.text(label="Test")

    assert marimo_text.value == text_value


def test_label_none_uses_field_name() -> None:
    class DataModel(BaseModel):
        field_name = Field(default=5, ge=0, le=10)  # type: ignore[arg-type]

    model = DataModel()

    marimo_slider = model.field_name.slider(label=None)
    expected_data_label_value = marimo_escape_html(
        '<span class="markdown"><span class="paragraph">field_name</span></span>',
    )
    assert expected_data_label_value in marimo_slider.text


def marimo_escape_html(value: str) -> str:
    processed = html.escape(json.dumps(value))
    return processed.replace("\\", "&#92;").replace("$", "&#36;")
