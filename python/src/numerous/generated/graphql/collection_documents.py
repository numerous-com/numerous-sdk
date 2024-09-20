# Generated by ariadne-codegen
# Source: queries.gql

from typing import List, Literal, Optional, Union

from pydantic import Field

from .base_model import BaseModel
from .fragments import CollectionDocumentReference


class CollectionDocuments(BaseModel):
    collection_create: Union[
        "CollectionDocumentsCollectionCreateCollection",
        "CollectionDocumentsCollectionCreateCollectionNotFound",
    ] = Field(alias="collectionCreate", discriminator="typename__")


class CollectionDocumentsCollectionCreateCollection(BaseModel):
    typename__: Literal["Collection"] = Field(alias="__typename")
    id: str
    key: str
    documents: "CollectionDocumentsCollectionCreateCollectionDocuments"


class CollectionDocumentsCollectionCreateCollectionDocuments(BaseModel):
    edges: List["CollectionDocumentsCollectionCreateCollectionDocumentsEdges"]
    page_info: "CollectionDocumentsCollectionCreateCollectionDocumentsPageInfo" = Field(
        alias="pageInfo"
    )


class CollectionDocumentsCollectionCreateCollectionDocumentsEdges(BaseModel):
    node: "CollectionDocumentsCollectionCreateCollectionDocumentsEdgesNode"


class CollectionDocumentsCollectionCreateCollectionDocumentsEdgesNode(
    CollectionDocumentReference
):
    typename__: Literal["CollectionDocument"] = Field(alias="__typename")


class CollectionDocumentsCollectionCreateCollectionDocumentsPageInfo(BaseModel):
    has_next_page: bool = Field(alias="hasNextPage")
    end_cursor: Optional[str] = Field(alias="endCursor")


class CollectionDocumentsCollectionCreateCollectionNotFound(BaseModel):
    typename__: Literal["CollectionNotFound"] = Field(alias="__typename")


CollectionDocuments.model_rebuild()
CollectionDocumentsCollectionCreateCollection.model_rebuild()
CollectionDocumentsCollectionCreateCollectionDocuments.model_rebuild()
CollectionDocumentsCollectionCreateCollectionDocumentsEdges.model_rebuild()
