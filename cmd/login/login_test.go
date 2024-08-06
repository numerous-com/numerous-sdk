package login

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testTenant = "test.domain.com"
	testIssuer = fmt.Sprintf("https://%s/", testTenant)
)

func TestLogin(t *testing.T) {
	state := auth.DeviceCodeState{
		DeviceCode:      "some-code",
		UserCode:        "some-long-user-code",
		VerificationURI: "https://test.domain.com/device/code/some-code",
		ExpiresIn:       8400,
		Interval:        5,
	}
	result := auth.Result{
		IDToken:      "some-id-token",
		AccessToken:  "some-access-token",
		RefreshToken: "some-refresh-token",
		ExpiresAt:    time.Now().Add(time.Second * time.Duration(state.ExpiresIn)),
	}
	expectedUser := &auth.User{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Tenant:       testTenant,
	}

	m := new(auth.MockAuthenticator)
	m.On("GetDeviceCode", mock.Anything, mock.Anything).Return(state, nil)
	m.On("OpenURL", state.VerificationURI).Return(nil)
	m.On("WaitUntilUserLogsIn", mock.Anything, mock.Anything, state).Return(result, nil)
	m.On("StoreAccessToken", result.AccessToken).Return(nil)
	m.On("StoreRefreshToken", result.RefreshToken).Return(nil)
	m.On("GetLoggedInUserFromKeyring").Return(&auth.User{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		Tenant:       testTenant,
	})
	acutalUser, _ := Login(m, context.Background())

	m.AssertExpectations(t)
	assert.Equal(t, expectedUser, acutalUser)
}

func TestRefreshAccessToken(t *testing.T) {
	t.Run("Refreshes access token if it has expired", func(t *testing.T) {
		// Create test tokens
		refreshToken := "refresh-token"
		accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(-time.Hour))
		expectedNewAccessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))

		user := &auth.User{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Tenant:       testTenant,
		}

		// Mock function from authenticator
		client := &http.Client{}
		m := new(auth.MockAuthenticator)
		m.On("RegenerateAccessToken", client, refreshToken).Return(expectedNewAccessToken, nil)
		m.On("StoreAccessToken", expectedNewAccessToken).Return(nil)

		// Execute function we test
		err := RefreshAccessToken(user, client, m)
		require.NoError(t, err)

		assert.Equal(t, expectedNewAccessToken, user.AccessToken)
		m.AssertExpectations(t)
	})

	t.Run("Does not refresh if user not logged in", func(t *testing.T) {
		var user *auth.User
		m := new(auth.MockAuthenticator)
		client := &http.Client{}
		err := RefreshAccessToken(user, client, m)

		require.EqualError(t, err, auth.ErrUserNotLoggedIn.Error())
		m.AssertExpectations(t)
	})

	t.Run("Does not refresh if access token has not expired", func(t *testing.T) {
		// Create test tokens
		refreshToken := "refresh-token"
		accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))

		user := &auth.User{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Tenant:       testTenant,
		}

		// Mock function from authenticator
		client := &http.Client{}
		m := new(auth.MockAuthenticator)

		// Execute function we test
		err := RefreshAccessToken(user, client, m)
		require.NoError(t, err)

		assert.Equal(t, user.AccessToken, accessToken)
		m.AssertExpectations(t)
	})
}
