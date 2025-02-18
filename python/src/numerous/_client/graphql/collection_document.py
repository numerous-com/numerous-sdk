# Generated by ariadne-codegen
# Source: queries.gql

from typing import Annotated, Literal, Optional, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionDocumentWithData


class CollectionDocument(BaseModel):
    collection_document: Optional[
        Annotated[
            Union[
                "CollectionDocumentCollectionDocumentCollectionDocument",
                "CollectionDocumentCollectionDocumentCollectionDocumentNotFound",
            ],
            Field(discriminator="typename__"),
        ]
    ] = Field(alias="collectionDocument")


class CollectionDocumentCollectionDocumentCollectionDocument(
    CollectionDocumentWithData
):
    typename__: Literal["CollectionDocument"] = Field(alias="__typename")


class CollectionDocumentCollectionDocumentCollectionDocumentNotFound(BaseModel):
    typename__: Literal["CollectionDocumentNotFound"] = Field(alias="__typename")


CollectionDocument.model_rebuild()
