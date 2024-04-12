from numerous import app, field


@app
class FieldApp:
    field_text_default: str = field()
    field_number_default: float = field()

    my_text_field: str = field(label="Text Field Label")
    my_number_field: float = field(label="Number Field Label")

    my_text_field_no_label: str = field(default="My text field")
    my_number_field_no_label: float = field(default=42.0)

    my_text_field_with_default_value: str = field(
        label="Text Field Label",
        default="My text field",
    )
    my_number_field_with_default_value: float = field(
        label="Number Field Label",
        default=42.0,
    )
