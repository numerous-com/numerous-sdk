package create

import (
	"testing"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestOrganizationCreate(t *testing.T) {
	t.Run("does not create an organization if user is not signed in", func(t *testing.T) {
		var u *auth.User
		m := new(auth.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		c, transportMock := test.CreateMockGQLClient()

		err := organizationCreate(m, c)

		assert.NoError(t, err)
		m.AssertExpectations(t)
		transportMock.AssertExpectations(t)
	})
}
