from __future__ import annotations

from datetime import timedelta
from typing import TYPE_CHECKING
from unittest.mock import Mock

import pytest
from freezegun import freeze_time

from numerous._client.graphql.client import Client as GQLClient
from numerous._client.graphql.organization_by_id import (
    OrganizationByID,
    OrganizationByIDOrganizationByIdOrganization,
)


if TYPE_CHECKING:
    from numerous.organization._client import Client

from numerous.organization import Organization, organization_from_env
from numerous.organization.exception import OrganizationIDMissingError


ORGANIZATION_ID = "test-organization-id"
ORGANIZATION_SLUG = "test-organization-slug"
HEADERS_WITH_AUTHORIZATION = {"headers": {"Authorization": "Bearer token"}}
UPDATE_INTERVAL = timedelta(seconds=300)


@pytest.fixture
def gql() -> Mock:
    return Mock(GQLClient)


@pytest.fixture
def client(gql: Mock) -> Client:
    from numerous._client.graphql_client import GraphQLClient

    return GraphQLClient(gql, ORGANIZATION_ID, "token")


def test_organization_reads_id_from_env(
    monkeypatch: pytest.MonkeyPatch, client: Client
) -> None:
    monkeypatch.setenv("NUMEROUS_ORGANIZATION_ID", ORGANIZATION_ID)
    org = organization_from_env(client)

    assert org.id == ORGANIZATION_ID


def test_organization_from_env_raises_error_if_id_is_not_set(
    monkeypatch: pytest.MonkeyPatch, client: Client
) -> None:
    monkeypatch.delenv("NUMEROUS_ORGANIZATION_ID", raising=False)
    with pytest.raises(OrganizationIDMissingError):
        organization_from_env(client)


def test_organization_slug_access_calls_client(gql: Mock, client: Client) -> None:
    org = Organization(ORGANIZATION_ID, client)

    gql.organization_by_id.return_value = OrganizationByID(
        organizationById=OrganizationByIDOrganizationByIdOrganization(
            __typename="Organization",
            id=ORGANIZATION_ID,
            slug=ORGANIZATION_SLUG,
        )
    )

    assert org.slug == ORGANIZATION_SLUG

    gql.organization_by_id.assert_called_once_with(
        ORGANIZATION_ID, **HEADERS_WITH_AUTHORIZATION
    )


def test_organization_slug_not_call_client_if_recently_updated(
    gql: Mock, client: Client
) -> None:
    org = Organization(ORGANIZATION_ID, client)

    gql.organization_by_id.return_value = OrganizationByID(
        organizationById=OrganizationByIDOrganizationByIdOrganization(
            __typename="Organization",
            id=ORGANIZATION_ID,
            slug=ORGANIZATION_SLUG,
        )
    )

    with freeze_time("2025-05-02") as frozen_datetime:
        assert org.slug == ORGANIZATION_SLUG
        gql.organization_by_id.assert_called_once_with(
            ORGANIZATION_ID, **HEADERS_WITH_AUTHORIZATION
        )

        frozen_datetime.tick(UPDATE_INTERVAL)
        gql.organization_by_id.reset_mock()

        assert org.slug == ORGANIZATION_SLUG
        gql.organization_by_id.assert_not_called()


def test_organization_slug_calls_client_if_updated_more_than_update_interval_ago(
    gql: Mock, client: Client
) -> None:
    org = Organization(ORGANIZATION_ID, client)

    gql.organization_by_id.return_value = OrganizationByID(
        organizationById=OrganizationByIDOrganizationByIdOrganization(
            __typename="Organization",
            id=ORGANIZATION_ID,
            slug=ORGANIZATION_SLUG,
        )
    )

    with freeze_time("2025-05-02") as frozen_datetime:
        assert org.slug == ORGANIZATION_SLUG
        gql.organization_by_id.assert_called_once_with(
            ORGANIZATION_ID, **HEADERS_WITH_AUTHORIZATION
        )

        frozen_datetime.tick(UPDATE_INTERVAL + timedelta(seconds=1))
        gql.organization_by_id.reset_mock()

        new_slug = "new-slug"
        gql.organization_by_id.return_value = OrganizationByID(
            organizationById=OrganizationByIDOrganizationByIdOrganization(
                __typename="Organization",
                id=ORGANIZATION_ID,
                slug=new_slug,
            )
        )

        assert org.slug == new_slug
        gql.organization_by_id.assert_called_once_with(
            ORGANIZATION_ID, **HEADERS_WITH_AUTHORIZATION
        )
