# Generated by ariadne-codegen
# Source: queries.gql

from typing import List, Literal, Optional

from pydantic import Field

from .base_model import BaseModel


class ButtonValue(BaseModel):
    button_value: str = Field(alias="buttonValue")


class CollectionReference(BaseModel):
    id: str
    key: str


class CollectionNotFound(BaseModel):
    id: str


class GraphContext(BaseModel):
    parent: Optional["GraphContextParent"]
    affected_by: List[Optional["GraphContextAffectedBy"]] = Field(alias="affectedBy")
    affects: List[Optional["GraphContextAffects"]]


class GraphContextParent(BaseModel):
    typename__: Literal[
        "Container", "ElementGraphParent", "ElementList", "ElementSelect"
    ] = Field(alias="__typename")
    id: str


class GraphContextAffectedBy(BaseModel):
    typename__: Literal[
        "Button",
        "Container",
        "Element",
        "ElementList",
        "ElementSelect",
        "HTMLElement",
        "NumberField",
        "SliderElement",
        "TextField",
    ] = Field(alias="__typename")
    id: str


class GraphContextAffects(BaseModel):
    typename__: Literal[
        "Button",
        "Container",
        "Element",
        "ElementList",
        "ElementSelect",
        "HTMLElement",
        "NumberField",
        "SliderElement",
        "TextField",
    ] = Field(alias="__typename")
    id: str


class HTMLValue(BaseModel):
    html: str


class NumberFieldValue(BaseModel):
    number_value: float = Field(alias="numberValue")


class SliderValue(BaseModel):
    slider_value: float = Field(alias="sliderValue")
    min_value: float = Field(alias="minValue")
    max_value: float = Field(alias="maxValue")


class TextFieldValue(BaseModel):
    text_value: str = Field(alias="textValue")


ButtonValue.model_rebuild()
CollectionReference.model_rebuild()
CollectionNotFound.model_rebuild()
GraphContext.model_rebuild()
HTMLValue.model_rebuild()
NumberFieldValue.model_rebuild()
SliderValue.model_rebuild()
TextFieldValue.model_rebuild()
