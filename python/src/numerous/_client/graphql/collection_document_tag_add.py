# Generated by ariadne-codegen
# Source: queries.gql

from typing import Literal, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionDocumentReference


class CollectionDocumentTagAdd(BaseModel):
    collection_document_tag_add: Union[
        "CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument",
        "CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound",
    ] = Field(alias="collectionDocumentTagAdd", discriminator="typename__")


class CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocument(
    CollectionDocumentReference
):
    typename__: Literal["CollectionDocument"] = Field(alias="__typename")


class CollectionDocumentTagAddCollectionDocumentTagAddCollectionDocumentNotFound(
    BaseModel
):
    typename__: Literal["CollectionDocumentNotFound"] = Field(alias="__typename")


CollectionDocumentTagAdd.model_rebuild()
