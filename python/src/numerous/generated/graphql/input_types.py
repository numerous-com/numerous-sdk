# Generated by ariadne-codegen
# Source: ../shared/schema.gql

from typing import List, Optional

from pydantic import Field

from .base_model import BaseModel
from .enums import Role


class NewOrganization(BaseModel):
    name: str
    slug: Optional[str] = None


class OrganizationInvitationInput(BaseModel):
    role: Role
    email: str


class NewTool(BaseModel):
    user_id: str = Field(alias="userId")
    manifest: str


class AppCreateInfo(BaseModel):
    name: str
    display_name: str = Field(alias="displayName")
    description: str


class AppDeployInput(BaseModel):
    app_relative_path: Optional[str] = Field(alias="appRelativePath", default=None)
    secrets: Optional[List["AppSecret"]] = None


class SubscriptionOfferInput(BaseModel):
    email: str
    app_name: str = Field(alias="appName")


class ElementInput(BaseModel):
    element_id: str = Field(alias="elementID")
    text_value: Optional[str] = Field(alias="textValue", default=None)
    number_value: Optional[float] = Field(alias="numberValue", default=None)
    html_value: Optional[str] = Field(alias="htmlValue", default=None)
    slider_value: Optional[float] = Field(alias="sliderValue", default=None)


class ElementSelectInput(BaseModel):
    select_element_id: str = Field(alias="selectElementID")
    selected_option_id: str = Field(alias="selectedOptionID")


class ListElementInput(BaseModel):
    list_element_id: str = Field(alias="listElementID")


class AppSecret(BaseModel):
    name: str
    base_64_value: str = Field(alias="base64Value")


class BuildPushInput(BaseModel):
    secrets: Optional[List["AppSecret"]] = None
