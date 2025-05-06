import os

from numerous._client.exceptions import (
    APIAccessTokenMissingError,
    OrganizationIDMissingError,
)
from numerous._client.graphql_client import GraphQLClient

from .graphql.client import Client as GQLClient


_DEFAULT_NUMEROUS_API_URL = "https://api.numerous.com/query"


def graphql_client_from_env() -> GraphQLClient:
    api_url = os.getenv("NUMEROUS_API_URL", _DEFAULT_NUMEROUS_API_URL)
    organization_id = os.getenv("NUMEROUS_ORGANIZATION_ID")
    access_token = os.getenv("NUMEROUS_API_ACCESS_TOKEN")

    if organization_id is None:
        raise OrganizationIDMissingError
    if access_token is None:
        raise APIAccessTokenMissingError

    gql = GQLClient(url=api_url)

    return GraphQLClient(gql, organization_id, access_token)
