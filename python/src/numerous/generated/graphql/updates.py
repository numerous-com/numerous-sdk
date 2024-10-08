# Generated by ariadne-codegen
# Source: queries.gql

from typing import Literal, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import (
    ButtonValue,
    GraphContext,
    HTMLValue,
    NumberFieldValue,
    SliderValue,
    TextFieldValue,
)


class Updates(BaseModel):
    tool_session_event: Union[
        "UpdatesToolSessionEventToolSessionElementAdded",
        "UpdatesToolSessionEventToolSessionElementRemoved",
        "UpdatesToolSessionEventToolSessionElementUpdated",
        "UpdatesToolSessionEventToolSessionActionTriggered",
    ] = Field(alias="toolSessionEvent", discriminator="typename__")


class UpdatesToolSessionEventToolSessionElementAdded(BaseModel):
    typename__: Literal["ToolSessionElementAdded"] = Field(alias="__typename")


class UpdatesToolSessionEventToolSessionElementRemoved(BaseModel):
    typename__: Literal["ToolSessionElementRemoved"] = Field(alias="__typename")


class UpdatesToolSessionEventToolSessionElementUpdated(BaseModel):
    typename__: Literal["ToolSessionElementUpdated"] = Field(alias="__typename")
    element: Union[
        "UpdatesToolSessionEventToolSessionElementUpdatedElementElement",
        "UpdatesToolSessionEventToolSessionElementUpdatedElementButton",
        "UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement",
        "UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField",
        "UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement",
        "UpdatesToolSessionEventToolSessionElementUpdatedElementTextField",
    ] = Field(discriminator="typename__")


class UpdatesToolSessionEventToolSessionElementUpdatedElementElement(BaseModel):
    typename__: Literal["Container", "Element", "ElementList", "ElementSelect"] = Field(
        alias="__typename"
    )
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementElementGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementElementGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionElementUpdatedElementButton(ButtonValue):
    typename__: Literal["Button"] = Field(alias="__typename")
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementButtonGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementButtonGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement(HTMLValue):
    typename__: Literal["HTMLElement"] = Field(alias="__typename")
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElementGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElementGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField(
    NumberFieldValue
):
    typename__: Literal["NumberField"] = Field(alias="__typename")
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementNumberFieldGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementNumberFieldGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement(SliderValue):
    typename__: Literal["SliderElement"] = Field(alias="__typename")
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElementGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElementGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionElementUpdatedElementTextField(TextFieldValue):
    typename__: Literal["TextField"] = Field(alias="__typename")
    id: str
    name: str
    graph_context: (
        "UpdatesToolSessionEventToolSessionElementUpdatedElementTextFieldGraphContext"
    ) = Field(alias="graphContext")


class UpdatesToolSessionEventToolSessionElementUpdatedElementTextFieldGraphContext(
    GraphContext
):
    pass


class UpdatesToolSessionEventToolSessionActionTriggered(BaseModel):
    typename__: Literal["ToolSessionActionTriggered"] = Field(alias="__typename")
    element: "UpdatesToolSessionEventToolSessionActionTriggeredElement"


class UpdatesToolSessionEventToolSessionActionTriggeredElement(BaseModel):
    id: str
    name: str


Updates.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdated.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementElement.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementButton.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementHTMLElement.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementNumberField.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementSliderElement.model_rebuild()
UpdatesToolSessionEventToolSessionElementUpdatedElementTextField.model_rebuild()
UpdatesToolSessionEventToolSessionActionTriggered.model_rebuild()
