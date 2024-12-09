package auth

import (
	"fmt"
	"net/http"
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

	t.Run("CheckAuthenticationStatus", func(t *testing.T) {
		t.Run("when user is logged in with wrong issuer returns error", func(t *testing.T) {
			goKeyring.MockInit()

			accessToken := test.GenerateJWT(t, "https://bad-issuer.com/", time.Now().Add(time.Hour))
			storeTestTokens(t, testTenant, accessToken, "refresh-token")

			user := getLoggedInUserFromKeyring(testTenant)
			actualError := user.CheckAuthenticationStatus()

			require.EqualError(t, actualError, "\"iss\" not satisfied: values do not match")
		})

		t.Run("with expired accessToken returns error", func(t *testing.T) {
			goKeyring.MockInit()

			accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(-time.Hour))
			storeTestTokens(t, testTenant, accessToken, "refresh-token")

			user := getLoggedInUserFromKeyring(testTenant)
			actualError := user.CheckAuthenticationStatus()

			require.EqualError(t, actualError, ErrExpiredToken.Error())
		})

		t.Run("when user is logged in returns nil", func(t *testing.T) {
			goKeyring.MockInit()

			accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))
			storeTestTokens(t, testTenant, accessToken, "refresh-token")

			user := getLoggedInUserFromKeyring(testTenant)
			actualError := user.CheckAuthenticationStatus()

			require.NoError(t, actualError)
		})
	})

	t.Run("HasExpired", func(t *testing.T) {
		type testCase struct {
			name               string
			expiration         time.Time
			expectedHasExpired bool
		}

		for _, testCase := range []testCase{
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
		} {
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
	})

	t.Run("TestRefreshAccessToken", func(t *testing.T) {
		t.Run("Refreshes access token if it has expired", func(t *testing.T) {
			refreshToken := "refresh-token"
			accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(-time.Hour))
			expectedNewAccessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))

			user := &User{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				Tenant:       testTenant,
			}

			client := &http.Client{}
			m := new(MockAuthenticator)
			m.On("RegenerateAccessToken", client, refreshToken).Return(expectedNewAccessToken, nil)
			m.On("StoreAccessToken", expectedNewAccessToken).Return(nil)

			err := user.RefreshAccessToken(client, m)
			require.NoError(t, err)

			assert.Equal(t, expectedNewAccessToken, user.AccessToken)
			m.AssertExpectations(t)
		})

		t.Run("Does not refresh if user not logged in", func(t *testing.T) {
			var user *User
			m := new(MockAuthenticator)
			client := &http.Client{}
			err := user.RefreshAccessToken(client, m)

			assert.ErrorIs(t, err, ErrUserNotLoggedIn)
			m.AssertExpectations(t)
		})

		t.Run("Does not refresh if access token has not expired", func(t *testing.T) {
			// Create test tokens
			refreshToken := "refresh-token"
			accessToken := test.GenerateJWT(t, testIssuer, time.Now().Add(time.Hour))

			user := &User{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				Tenant:       testTenant,
			}

			// Mock function from authenticator
			client := &http.Client{}
			m := new(MockAuthenticator)

			// Execute function we test
			err := user.RefreshAccessToken(client, m)
			require.NoError(t, err)

			assert.Equal(t, user.AccessToken, accessToken)
			m.AssertExpectations(t)
		})
	})
}

func storeTestTokens(t *testing.T, tenant, accessToken, refreshToken string) {
	t.Helper()
	err := keyring.StoreAccessToken(tenant, accessToken)
	require.NoError(t, err)
	err = keyring.StoreRefreshToken(tenant, refreshToken)
	require.NoError(t, err)
}
