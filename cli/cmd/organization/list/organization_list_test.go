package list

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"numerous/cli/auth"
	"numerous/cli/internal/gql/organization"
	"numerous/cli/test"
	"numerous/cli/test/mocks"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	testUser := &auth.User{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Tenant:       "numerous-testing.com",
	}

	t.Run("can list organizations if user is signed in", func(t *testing.T) {
		m := new(mocks.MockAuthenticator)
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
			organizationsAsStrings = append(organizationsAsStrings, test.OrganizationMembershipToResponse(struct {
				Role         test.Role
				Organization struct {
					ID   string
					Name string
					Slug string
				}
			}{
				Role:         test.Role(organization.Role),
				Organization: organization.Organization,
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
		c, transportMock := test.CreateMockGqlClient(response)

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})

	t.Run("does not query list if user is not signed in", func(t *testing.T) {
		var u *auth.User
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		c, transportMock := test.CreateMockGqlClient()

		err := list(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})
}
