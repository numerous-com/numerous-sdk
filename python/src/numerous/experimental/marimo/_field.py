"""
Marimo Fields.

The marimo module provides classes and functions for creating fields
specifically useful in marimo apps.

Marimo is a Python library for building interactive web applications.

This module contains the `Field` class, which is a field with a state that
can be used in a Marimo app. It also provides utility functions
for creating different types of UI elements such as sliders,
numbers, and text fields.

"""

from typing import Any, Dict, Type, TypeVar, Union

import marimo as mo
from marimo._runtime.state import State as MoState

from numerous.experimental.model import Field as BaseField


def _auto_label(key: str, label: Union[str, None]) -> str:
    """
    Automatically assigns a label to a key if the label is None.

    Args:
    ----
        key (str): The key to be labeled.
        label (str): The label to assign to the key. If None,
            the key will be used as the label.

    Returns:
    -------
        str: The assigned label.

    """
    if label is None:
        label = key

    return label


T = TypeVar("T", mo.ui.number, mo.ui.slider)


class Field(BaseField, MoState[Union[str, float, int]]):
    def __init__(
        self,
        default: Union[str, float, None] = None,
        annotation: Union[type, None] = None,
        **kwargs: Dict[str, Any],
    ) -> None:
        """
        Field with a state that can be used in a Marimo app.

        Args:
        ----
            default (str | float, optional): The default value for the field.
            Defaults to `...`.
            annotation (type | None, optional): The type annotation for the field.
            Defaults to None.
            **kwargs (dict): Additional keyword arguments for the field.

        """
        BaseField.__init__(self, default=default, annotation=annotation, **kwargs)
        MoState.__init__(self, self.value)

    def _number_ui(
        self,
        ui_cls: Type[T],
        step: float = 1,
        label: Union[str, None] = None,
    ) -> T:
        if not hasattr(self, "field_info"):
            error_msg = "The field_info attribute is not defined."
            raise AttributeError(error_msg)
        _min = self.field_info.metadata[0].ge
        _max = self.field_info.metadata[1].le

        return ui_cls(
            _min,
            _max,
            value=float(self.get()),
            on_change=self.set,
            label=_auto_label(self.name, label),
            step=step,
        )

    def slider(
        self,
        step: float = 1,
        label: Union[str, None] = None,
    ) -> mo.ui.slider:
        """
        Create a slider UI element.

        Args:
        ----
            step (float, optional): The step size for the slider. Defaults to 1.
            label (str | None, optional): The label for the slider. Defaults to None.
            Defaults to None.

        Returns:
        -------
            mo.ui.slider: The created slider UI element.

        """
        return self._number_ui(mo.ui.slider, step, label)

    def number(
        self,
        step: float = 1,
        label: Union[str, None] = None,
    ) -> mo.ui.number:
        """
        Create a number UI element.

        Args:
        ----
            step (float, optional): The step value for the number UI element.
                Defaults to 1.
            label (str, optional): The label for the number UI element.
                Defaults to None.

        Returns:
        -------
            mo.ui.number: The created number UI element.

        """
        number_ui = self._number_ui(mo.ui.number, step, label)

        if isinstance(number_ui, mo.ui.number):
            return number_ui

        error_msg = "The number UI element is not an instance of mo.ui.number."
        raise TypeError(error_msg)

    def text(self, label: Union[str, None] = None) -> mo.ui.text:
        """
        Return a text field widget.

        Args:
        ----
            label (str, optional): The label for the text field. Defaults to None.

        Returns:
        -------
            mo.ui.text: The text field widget.

        """
        return mo.ui.text(
            value=str(self.get()),
            on_change=self.set,
            label=_auto_label(self.name, label),
        )

    def set(self, value: Any) -> None:  # noqa: ANN401
        """
        Set the value of the Marimo object.

        Args:
        ----
            value (int | str): The value to be set. It can be an integer or a string.

        Overrides the set method to call the mo state set_value and then calls the super
        class's set method.

        """
        self._set_value(value)
        super().set(value)
