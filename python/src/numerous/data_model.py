"""Tool data model for communication."""

import inspect
from dataclasses import dataclass
from typing import Any, Optional, Type

import plotly.graph_objects as go

from numerous._plotly import plotly_html
from numerous.apps import HTML, Field, Slider
from numerous.utils import MISSING, AppT


@dataclass(frozen=True)
class ElementDataModel:
    name: str
    label: str


@dataclass(frozen=True)
class HTMLElementDataModel(ElementDataModel):
    default: str
    type: str = "html"


@dataclass(frozen=True)
class TextFieldDataModel(ElementDataModel):
    default: str
    type: str = "string"


@dataclass(frozen=True)
class NumberFieldDataModel(ElementDataModel):
    default: float
    type: str = "number"


@dataclass(frozen=True)
class PlotlyElementDataModel(ElementDataModel):
    default: str
    type: str = "html"


@dataclass(frozen=True)
class SliderElementDataModel(ElementDataModel):
    default: float
    slider_min_value: float
    slider_max_value: float
    type: str = "slider"


@dataclass(frozen=True)
class ActionDataModel(ElementDataModel):
    type: str = "action"


@dataclass(frozen=True)
class ContainerDataModel(ElementDataModel):
    elements: list[ElementDataModel]
    type: str = "container"


@dataclass(frozen=True)
class AppDataModel:
    name: str
    title: str
    elements: list[ElementDataModel]


class ToolDataModelError(Exception):
    def __init__(self, cls: Type[AppT], name: str, annotation: str) -> None:
        super().__init__(
            f"Invalid Tool. Unsupport field: {cls.__name__}.{name}: {annotation}",
        )


class ToolDefinitionError(Exception):
    def __init__(self, cls: Type[AppT]) -> None:
        self.cls = cls
        super().__init__(
            f"Tool {self.cls.__name__} not defined with the @app decorator",
        )


def dump_data_model(cls: Type[AppT]) -> AppDataModel:
    if not getattr(cls, "__numerous_app__", False):
        raise ToolDefinitionError(cls)

    title = getattr(cls, "__title__", cls.__name__)

    return AppDataModel(
        name=cls.__name__,
        title=title,
        elements=_dump_element_data_models(cls),
    )


def _dump_element_from_annotation(
    name: str,
    annotation: Any,  # noqa: ANN401
    default: Any,  # noqa: ANN401
) -> Optional[ElementDataModel]:
    if annotation is str and (
        data_model := _dump_element_from_str_annotation(name, default)
    ):
        return data_model

    if annotation is float and (
        data_model := _dump_element_from_float_annotation(name, default)
    ):
        return data_model

    if annotation is go.Figure:
        label = default.label if isinstance(default, Field) else name
        default = (
            plotly_html(default.default)
            if isinstance(default, Field) and isinstance(default.default, go.Figure)
            else ""
        )
        return PlotlyElementDataModel(name=name, label=label or name, default=default)

    if getattr(annotation, "__container__", False):
        return dump_container_data_model(name, annotation)

    return None


def _dump_element_from_float_annotation(
    name: str,
    default: Any,  # noqa: ANN401
) -> Optional[ElementDataModel]:
    if isinstance(default, (float, int)) or default is MISSING:
        return NumberFieldDataModel(
            name=name,
            label=name,
            default=float(0.0 if default is MISSING else default),
        )
    if isinstance(default, Slider):
        default_value = (
            default.min_value
            if default.default is MISSING  # type: ignore[comparison-overlap]
            else default.default
        )

        return SliderElementDataModel(
            name=name,
            label=default.label if default.label else name,
            default=default_value,
            slider_min_value=default.min_value,
            slider_max_value=default.max_value,
        )
    if isinstance(default, Field):
        return NumberFieldDataModel(
            name=name,
            label=default.label if default.label else name,
            default=float(0.0 if default.default is MISSING else default.default),  # type: ignore[comparison-overlap]
        )
    return None


def _dump_element_from_str_annotation(
    name: str,
    default: Any,  # noqa: ANN401
) -> Optional[ElementDataModel]:
    if isinstance(default, str) or default is MISSING:
        default = "" if default is MISSING else default
        return TextFieldDataModel(name=name, label=name, default=default)
    if isinstance(default, HTML):
        return HTMLElementDataModel(name=name, label=name, default=default.default)
    if isinstance(default, Field):
        return TextFieldDataModel(
            name=name,
            label=default.label if default.label else name,
            default=str("" if default.default is MISSING else default.default),  # type: ignore[comparison-overlap]
        )
    return None


def _dump_element_data_models(cls: Type[AppT]) -> list[ElementDataModel]:
    elements: list[ElementDataModel] = []
    annotations = cls.__annotations__ if hasattr(cls, "__annotations__") else {}
    for name, annotation in annotations.items():
        default = getattr(cls, name, MISSING)
        if elem := _dump_element_from_annotation(name, annotation, default):
            elements.append(elem)
        else:
            raise ToolDataModelError(cls, name, annotation)

    for name, func in inspect.getmembers(cls, inspect.isfunction):
        if getattr(func, "__action__", False):
            elements.append(ActionDataModel(name, name))

    return elements


class ContainerDefinitionError(Exception):
    def __init__(self, cls: Type[AppT]) -> None:
        self.cls = cls
        super().__init__(
            f"Container {self.cls.__name__} not defined with the @container decorator",
        )


def dump_container_data_model(name: str, cls: Type[AppT]) -> ContainerDataModel:
    if not getattr(cls, "__container__", False):
        raise ContainerDefinitionError(cls)

    elements = _dump_element_data_models(cls)

    return ContainerDataModel(name=name, label=name, elements=elements)
