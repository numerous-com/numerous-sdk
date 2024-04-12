package organization

import (
	"fmt"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	t.Run("can return organization on OrganizationCreate mutation", func(t *testing.T) {
		expectedOrganization := Organization{
			ID:   "id",
			Name: "name",
			Slug: "slug",
		}
		response := organizationToQueryResult("organizationCreate", expectedOrganization)
		c := test.CreateTestGqlClient(response)

		actualOrganization, err := Create(expectedOrganization.Name, c)

		assert.NoError(t, err)
		assert.Equal(t, expectedOrganization, actualOrganization)
	})

	t.Run("can return error on OrganizationCreate mutation if name contains invalid character", func(t *testing.T) {
		organizationInvalidCharactersResponse := `{"errors":[{"message":"organization name contains invalid characters","path":["organizationCreate"]}],"data":null}`
		c := test.CreateTestGqlClient(organizationInvalidCharactersResponse)

		actualOrganization, err := Create("!", c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "organization name contains invalid characters")
		assert.Equal(t, Organization{}, actualOrganization)
	})

	t.Run("can return error on OrganizationCreate mutation if unknown error", func(t *testing.T) {
		organizationErrorResponse := `{"errors":[{"message":"unknown error","path":["organizationCreate"]}],"data":null}`
		c := test.CreateTestGqlClient(organizationErrorResponse)

		actualOrganization, err := Create("", c)

		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrOrganizationNameInvalidCharacter)
		assert.Equal(t, Organization{}, actualOrganization)
	})
}

func organizationToQueryResult(queryName string, o Organization) string {
	return fmt.Sprintf(`{
		"data": {
			"%s": %s
		}
	}`, queryName, organizationToResponse(o))
}

func organizationToResponse(o Organization) string {
	return fmt.Sprintf(`{
		"id": "%s",
		"name": "%s",
		"slug": "%s"
	}`, o.ID, o.Name, o.Slug)
}
