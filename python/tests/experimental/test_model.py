import pytest
from numerous.experimental.model import BaseModel, Field


# Create a field with an annotation but no default value.
field = Field(annotation=int)


def test_field_value() -> None:
    # Since no default value is provided, getting the value should raise an error.
    with pytest.raises(ValueError):  # noqa: PT011
        field.value  # noqa: B018


def test_assignment_before_parent_model() -> None:
    # The field has no parent model, so setting the value should raise an error,
    # since validation occurs in the model.
    with pytest.raises(ValueError):  # noqa: PT011
        field.value = 5


def test_check_name_before_assignment() -> None:
    # The field has not been named, so getting the name should raise an error.
    with pytest.raises(ValueError):  # noqa: PT011
        field.name  # noqa: B018


def test_rename() -> None:
    # Naming the field should work.
    field.name = "field"

    # The field has been named, so naming again should raise an error.
    with pytest.raises(ValueError):  # noqa: PT011
        field.name = "field2"

    # Check that the name is correct.
    assert field.name == "field", "Field name is incorrect."


def test_field_default_value() -> None:
    # Create a field with a default value.
    val = 5
    field = Field(default=val)

    # The field was provided a default value, so getting the value should work.
    assert field.value == val, "Field value is incorrect."


a_val = 5
a_val2 = 10


# Define a model with a field.
class DataModel(BaseModel):
    a = Field(default=a_val)


def test_model_field_value() -> None:
    model = DataModel()

    assert model.a.value == a_val


def test_model_field_name() -> None:
    model = DataModel()

    assert model.a.name == "a"


def test_parent_model_correct() -> None:
    model = DataModel()

    assert model.a._parent_model == model  # noqa: SLF001


def test_set_value() -> None:
    model = DataModel()

    model.a.value = a_val2

    assert model.a.value == a_val2


def test_set_wrong_type_raises_valueerror() -> None:
    model = DataModel()

    with pytest.raises(ValueError):  # noqa: PT011
        model.a.name = "Text"


def test_submodel_() -> None:
    number_value = 10

    class SubModel(BaseModel):
        number_field = Field(default=number_value)

    class DataModel(BaseModel):
        number_field = Field(default=5)
        sub_model = SubModel()

    model = DataModel()

    assert model.sub_model.number_field.value == number_value
    assert model.pydantic_model.sub_model.number_field == number_value  # type: ignore[attr-defined]


def test_non_existing_field_raise_error() -> None:
    class DataModel(BaseModel):
        a = Field(default=5)

    # Trying to set a non-existing field should raise an error.
    with pytest.raises(AttributeError):
        DataModel(b=5)  # type: ignore[arg-type]


def test_no_default_provided_and_no_value_raises_value_error() -> None:
    class DataModel(BaseModel):
        a = Field(annotation=int)

    # Trying to instanciate the model without a value should raise an error.
    with pytest.raises(ValueError):  # noqa: PT011
        DataModel()


def test_annotation_is_used_to_validate_value() -> None:
    class DataModel(BaseModel):
        a = Field(annotation=int)

    DataModel(a=5)  # type: ignore[arg-type]

    with pytest.raises(ValueError):  # noqa: PT011
        DataModel(a="text")  # type: ignore[arg-type]


def test_defining_field_without_annotation_or_default_raises_value_error() -> None:
    """Test that you cannot define a model without an annotation or a default value."""
    with pytest.raises(ValueError):  # noqa: PT011

        class DataModel(BaseModel):
            a = Field()
