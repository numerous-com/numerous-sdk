package list

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql/organization"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	testUser := &auth.User{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Tenant:       "numerous-testing.com",
	}

	t.Run("can list organizations if user is signed in", func(t *testing.T) {
		m := new(auth.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(testUser)

		organizationsAsStrings := []string{}
		for i := 0; i < 3; i++ {
			id := strconv.Itoa(i)
			organization := organization.OrganizationMembership{
				Role: organization.Admin,
				Organization: organization.Organization{
					ID:   id,
					Name: "Name " + id,
					Slug: fmt.Sprintf("name-%s-slug", id),
				},
			}
			organizationsAsStrings = append(organizationsAsStrings, test.OrganizationMembershipToResponse(test.Membership{
				Role: test.Role(organization.Role),
				Organization: test.Organization{
					ID:   organization.Organization.ID,
					Name: organization.Organization.Name,
					Slug: organization.Organization.Slug,
				},
			}))
		}

		response := `{
			"data": {
				"me": {
					"fullName": "",
					"memberships": [` + strings.Join(organizationsAsStrings, ", ") + `]
				}
			}
		}`
		c, transportMock := test.CreateMockGQLClient(response)

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})

	t.Run("does not query list if user is not signed in", func(t *testing.T) {
		var u *auth.User
		m := new(auth.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		c, transportMock := test.CreateMockGQLClient()

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})
}
