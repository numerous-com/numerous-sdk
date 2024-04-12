package logout

import (
	"errors"
	"net/http"
	"testing"

	"numerous/cli/auth"
	"numerous/cli/test/mocks"

	"github.com/stretchr/testify/assert"
)

func TestLogout(t *testing.T) {
	testUser := &auth.User{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Tenant:       "numerous-testing.com",
	}

	t.Run("Can logout if user is signed in", func(t *testing.T) {
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(testUser)
		m.On("RevokeRefreshToken", http.DefaultClient, testUser.RefreshToken).Return(nil)
		m.On("RemoveLoggedInUserFromKeyring").Return(nil)

		err := logout(m)

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("Can logout, even if refresh token could not be revoked", func(t *testing.T) {
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(testUser)
		m.On("RevokeRefreshToken", http.DefaultClient, testUser.RefreshToken).Return(auth.ErrInvalidClient)
		m.On("RemoveLoggedInUserFromKeyring").Return(nil)

		err := logout(m)

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("Does not call any methods on Authenticator if user is not logged in", func(t *testing.T) {
		var u *auth.User
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(u)

		err := logout(m)

		assert.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("Returns error if it cannot remove from keyring", func(t *testing.T) {
		errCouldNotRemove := errors.New("could not remove")
		m := new(mocks.MockAuthenticator)
		m.On("GetLoggedInUserFromKeyring").Return(testUser)
		m.On("RevokeRefreshToken", http.DefaultClient, testUser.RefreshToken).Return(nil)
		m.On("RemoveLoggedInUserFromKeyring").Return(errCouldNotRemove)

		err := logout(m)

		assert.ErrorIs(t, err, errCouldNotRemove)
		m.AssertExpectations(t)
	})
}
