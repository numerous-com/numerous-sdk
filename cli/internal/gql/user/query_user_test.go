package user

import (
	"testing"

	"numerous/cli/internal/gql/organization"
	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestQueryUser(t *testing.T) {
	t.Run("can return user on user query", func(t *testing.T) {
		membership := organization.OrganizationMembership{
			Role: organization.Admin,
			Organization: organization.Organization{
				ID:   "1",
				Name: "Test org",
				Slug: "test-org-slug",
			},
		}
		expectedUser := User{
			FullName:    "Test User",
			Memberships: []organization.OrganizationMembership{membership},
		}
		membershipAsString := test.OrganizationMembershipToResponse(struct {
			Role         test.Role
			Organization struct {
				ID   string
				Name string
				Slug string
			}
		}{
			Role:         test.Role(membership.Role),
			Organization: membership.Organization,
		})

		response := `{
			"data": {
				"me": {
					"fullName": "` + expectedUser.FullName + `",
					"memberships": [` + membershipAsString + `]
				}
			}
		}`
		c := test.CreateTestGqlClient(response)

		actualUser, err := QueryUser(c)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, actualUser)
	})

	t.Run("can return permission denied error", func(t *testing.T) {
		userNotFoundResponse := `{"errors":[{"message":"permission denied","path":["me"]}],"data":null}`
		c := test.CreateTestGqlClient(userNotFoundResponse)

		actualUser, err := QueryUser(c)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "permission denied")
		assert.Equal(t, User{}, actualUser)
	})
}
