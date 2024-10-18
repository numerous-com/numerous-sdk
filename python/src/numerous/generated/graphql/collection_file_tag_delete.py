# Generated by ariadne-codegen
# Source: queries.gql

from typing import Literal, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionFileReference


class CollectionFileTagDelete(BaseModel):
    collection_file_tag_delete: Union[
        "CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile",
        "CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound",
    ] = Field(alias="collectionFileTagDelete", discriminator="typename__")


class CollectionFileTagDeleteCollectionFileTagDeleteCollectionFile(
    CollectionFileReference
):
    typename__: Literal["CollectionFile"] = Field(alias="__typename")


class CollectionFileTagDeleteCollectionFileTagDeleteCollectionFileNotFound(BaseModel):
    typename__: Literal["CollectionFileNotFound"] = Field(alias="__typename")


CollectionFileTagDelete.model_rebuild()