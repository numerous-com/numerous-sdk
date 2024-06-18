package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	goKeyring "github.com/zalando/go-keyring"

	"numerous.com/cli/internal/keyring"
	"numerous.com/cli/internal/test"
)

func TestUser(t *testing.T) {
	const testTenant string = "numerous-test.eu.com"
	testIssuer := fmt.Sprintf("https://%s/", testTenant)
	t.Run("Can get logged in user, based on keyring", func(t *testing.T) {
		goKeyring.MockInit()
		expectedUser := User{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			Tenant:       testTenant,
		}
		storeTestTokens(t, testTenant, expectedUser.AccessToken, expectedUser.RefreshToken)

		actualUser := getLoggedInUserFromKeyring(testTenant)

		assert.Equal(t, expectedUser, *actualUser)
	})

	t.Run("Cannot get logged in user, if nothing is in keyring", func(t *testing.T) {
		goKeyring.MockInit()

		actualUser := getLoggedInUserFromKeyring(testTenant)

		assert.Nil(t, actualUser)
	})

	t.Run("CheckAuthenticationStatus when user is logged in with wrong issuer returns error", func(t *testing.T) {
		goKeyring.MockInit()

		accessToken := test.GenerateJWT(t, "https://bad-issuer.com/", time.Now().Add(time.Hour))
		storeTestTokens(t, testTenant, accessToken, "refresh-token")

		user := getLoggedInUserFromKeyring(testTenant)
		actualError := user.CheckAuthenticationStatus()

		require.EqualError(t, actualError, "\"iss\" not satisfied: values do not match")
	})

	t.Run("CheckAuthenticationStatus with expired accessToken returns error", func(t *testing.T) {
		goKeyring.MockInit()

		accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(-time.Hour))
		storeTestTokens(t, testTenant, accessToken, "refresh-token")

		user := getLoggedInUserFromKeyring(testTenant)
		actualError := user.CheckAuthenticationStatus()

		require.EqualError(t, actualError, ErrExpiredToken.Error())
	})

	t.Run("CheckAuthenticationStatus when user is logged in returns nil", func(t *testing.T) {
		goKeyring.MockInit()

		accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))
		storeTestTokens(t, testTenant, accessToken, "refresh-token")

		user := getLoggedInUserFromKeyring(testTenant)
		actualError := user.CheckAuthenticationStatus()

		require.NoError(t, actualError)
	})

	hasExpiredTestCases := []struct {
		name               string
		expiration         time.Time
		expectedHasExpired bool
	}{
		{
			name:               "User.HasExpired returns false when expiration of token is after current time",
			expiration:         time.Now().Add(2 * time.Hour),
			expectedHasExpired: false,
		},
		{
			name:               "User.HasExpired returns true when expiration of token is before current time",
			expiration:         time.Now().Add(-2 * time.Hour),
			expectedHasExpired: true,
		},
	}
	for _, testCase := range hasExpiredTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			goKeyring.MockInit()

			accessToken := test.GenerateJWT(t, testIssuer, testCase.expiration)
			user := User{
				AccessToken:  accessToken,
				RefreshToken: "refresh-token",
			}

			assert.Equal(t, testCase.expectedHasExpired, user.HasExpiredToken())
		})
	}
}

func storeTestTokens(t *testing.T, tenant, accessToken, refreshToken string) {
	t.Helper()
	err := keyring.StoreAccessToken(tenant, accessToken)
	require.NoError(t, err)
	err = keyring.StoreRefreshToken(tenant, refreshToken)
	require.NoError(t, err)
}
