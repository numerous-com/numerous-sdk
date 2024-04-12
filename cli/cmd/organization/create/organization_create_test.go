package create

import (
	"testing"

	"numerous/cli/auth"
	"numerous/cli/test"
	"numerous/cli/test/mocks"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationCreate(t *testing.T) {
	t.Run("does not create an organization if user is not signed in", func(t *testing.T) {
		var u *auth.User
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		c, transportMock := test.CreateMockGqlClient()

		err := organizationCreate(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})
}
