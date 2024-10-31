# Generated by ariadne-codegen
# Source: ../shared/schema.gql

from typing import Any, List, Optional

from pydantic import Field

from .base_model import BaseModel
from .enums import Role


class PersonalAccessTokenCreateInput(BaseModel):
    name: str
    description: str
    expires_at: Optional[Any] = Field(alias="expiresAt", default=None)


class NewOrganization(BaseModel):
    name: str
    slug: Optional[str] = None


class OrganizationMemberEditRoleInput(BaseModel):
    user_id: str = Field(alias="userId")
    role: Role


class Auth0WhiteLabelInvitationInput(BaseModel):
    email: str
    organization_id: str = Field(alias="organizationID")


class OrganizationInvitationInput(BaseModel):
    role: Role
    email: str


class NewTool(BaseModel):
    user_id: str = Field(alias="userId")
    manifest: str


class AppVersionInput(BaseModel):
    version: Optional[str] = None
    message: Optional[str] = None


class AppCreateInfo(BaseModel):
    app_slug: str = Field(alias="appSlug")
    display_name: str = Field(alias="displayName")
    description: str


class AppDeployInput(BaseModel):
    app_relative_path: Optional[str] = Field(alias="appRelativePath", default=None)
    secrets: Optional[List["AppSecret"]] = None
    skip_metadata_update: Optional[bool] = Field(
        alias="skipMetadataUpdate", default=None
    )


class AppDeployLogsInput(BaseModel):
    organization_slug: str = Field(alias="organizationSlug")
    app_slug: str = Field(alias="appSlug")


class AppRenameInput(BaseModel):
    app_id: str = Field(alias="appID")
    app_name: str = Field(alias="appName")


class AppDescriptionUpdateInput(BaseModel):
    app_id: str = Field(alias="appID")
    app_description: Optional[str] = Field(alias="appDescription", default=None)


class AppDeleteInput(BaseModel):
    app_slug: str = Field(alias="appSlug")
    organization_slug: str = Field(alias="organizationSlug")


class AppVersionCreateGitHubInput(BaseModel):
    owner: str
    repo: str


class PaymentConfigurationInput(BaseModel):
    monthly_price_usd: Any = Field(alias="monthlyPriceUSD")
    trial_days: Optional[int] = Field(alias="trialDays", default=None)


class SubscriptionOfferInput(BaseModel):
    email: str
    app_slug: str = Field(alias="appSlug")
    message: Optional[str] = None
    payment: Optional["PaymentConfigurationInput"] = None


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


class TagInput(BaseModel):
    key: str
    value: str


AppDeployInput.model_rebuild()
SubscriptionOfferInput.model_rebuild()
BuildPushInput.model_rebuild()
