# Generated by ariadne-codegen
# Source: queries.gql

from typing import Annotated, Literal, Optional, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionDocumentReference


class CollectionDocumentInCollection(BaseModel):
    collection: Optional[
        Annotated[
            Union[
                "CollectionDocumentInCollectionCollectionCollection",
                "CollectionDocumentInCollectionCollectionCollectionNotFound",
            ],
            Field(discriminator="typename__"),
        ]
    ]


class CollectionDocumentInCollectionCollectionCollection(BaseModel):
    typename__: Literal["Collection"] = Field(alias="__typename")
    document: Optional["CollectionDocumentInCollectionCollectionCollectionDocument"]


class CollectionDocumentInCollectionCollectionCollectionDocument(
    CollectionDocumentReference
):
    typename__: Literal["CollectionDocument"] = Field(alias="__typename")


class CollectionDocumentInCollectionCollectionCollectionNotFound(BaseModel):
    typename__: Literal["CollectionNotFound"] = Field(alias="__typename")


CollectionDocumentInCollection.model_rebuild()
CollectionDocumentInCollectionCollectionCollection.model_rebuild()
