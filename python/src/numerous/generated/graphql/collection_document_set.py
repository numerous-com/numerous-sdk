# Generated by ariadne-codegen
# Source: queries.gql

from typing import Literal, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionDocumentReference


class CollectionDocumentSet(BaseModel):
    collection_document_set: Union[
        "CollectionDocumentSetCollectionDocumentSetCollectionDocument",
        "CollectionDocumentSetCollectionDocumentSetCollectionNotFound",
    ] = Field(alias="collectionDocumentSet", discriminator="typename__")


class CollectionDocumentSetCollectionDocumentSetCollectionDocument(
    CollectionDocumentReference
):
    typename__: Literal["CollectionDocument"] = Field(alias="__typename")


class CollectionDocumentSetCollectionDocumentSetCollectionNotFound(BaseModel):
    typename__: Literal["CollectionNotFound"] = Field(alias="__typename")


CollectionDocumentSet.model_rebuild()
