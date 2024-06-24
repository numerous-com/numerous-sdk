package keyring

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

const testTenantName = "numerous-cli-test.eu.auth0.com"

func TestSecrets(t *testing.T) {
	t.Run("fails to retrieve an nonexistent refresh token", func(t *testing.T) {
		keyring.MockInit()

		_, actualError := GetRefreshToken(testTenantName)
		assert.EqualError(t, actualError, keyring.ErrNotFound.Error())
	})

	t.Run("successfully retrieves an existent refresh token", func(t *testing.T) {
		keyring.MockInit()

		expectedRefreshToken := "fake-refresh-token"
		err := keyring.Set(secretRefreshToken, testTenantName, expectedRefreshToken)
		require.NoError(t, err)

		actualRefreshToken, err := GetRefreshToken(testTenantName)
		require.NoError(t, err)
		assert.Equal(t, expectedRefreshToken, actualRefreshToken)
	})

	t.Run("successfully stores a refresh token", func(t *testing.T) {
		keyring.MockInit()

		expectedRefreshToken := "fake-refresh-token"
		err := StoreRefreshToken(testTenantName, expectedRefreshToken)
		require.NoError(t, err)

		actualRefreshToken, err := GetRefreshToken(testTenantName)
		require.NoError(t, err)
		assert.Equal(t, expectedRefreshToken, actualRefreshToken)
	})

	t.Run("fails to retrieve an nonexistent access token", func(t *testing.T) {
		keyring.MockInit()

		_, actualError := GetAccessToken(testTenantName)
		require.EqualError(t, actualError, keyring.ErrNotFound.Error())
	})

	t.Run("successfully stores an access token", func(t *testing.T) {
		keyring.MockInit()

		expectedAccessToken := randomStringOfLength((2048 * 5) + 1) // Some arbitrarily long random string.
		err := StoreAccessToken(testTenantName, expectedAccessToken)
		require.NoError(t, err)

		actualAccessToken, err := GetAccessToken(testTenantName)
		require.NoError(t, err)
		assert.Equal(t, expectedAccessToken, actualAccessToken)
	})

	t.Run("successfully retrieves an access token split up into multiple chunks", func(t *testing.T) {
		keyring.MockInit()

		err := keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 0), testTenantName, "chunk0")
		require.NoError(t, err)
		err = keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 1), testTenantName, "chunk1")
		require.NoError(t, err)
		err = keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 2), testTenantName, "chunk2")
		require.NoError(t, err)

		actualAccessToken, err := GetAccessToken(testTenantName)
		require.NoError(t, err)
		assert.Equal(t, "chunk0chunk1chunk2", actualAccessToken)
	})

	t.Run("successfully deletes an access token split up into multiple chunks", func(t *testing.T) {
		keyring.MockInit()

		err := keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 0), testTenantName, "chunk0")
		require.NoError(t, err)
		err = keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 1), testTenantName, "chunk1")
		require.NoError(t, err)
		err = keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 2), testTenantName, "chunk2")
		require.NoError(t, err)
		err = keyring.Set(secretRefreshToken, testTenantName, "fake-refresh-token")
		require.NoError(t, err)

		// Ensure setup is correct
		actualAccessToken, err := GetAccessToken(testTenantName)
		require.NoError(t, err)
		assert.Equal(t, "chunk0chunk1chunk2", actualAccessToken)

		err = DeleteTokens(testTenantName)
		require.NoError(t, err)

		_, actualError := GetAccessToken(testTenantName)
		require.EqualError(t, actualError, keyring.ErrNotFound.Error())
	})

	t.Run("successfully deletes an refresh token", func(t *testing.T) {
		keyring.MockInit()

		err := keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 0), testTenantName, "access-token")
		require.NoError(t, err)
		err = keyring.Set(secretRefreshToken, testTenantName, "fake-refresh-token")
		require.NoError(t, err)

		err = DeleteTokens(testTenantName)
		require.NoError(t, err)

		_, actualError := GetRefreshToken(testTenantName)
		require.EqualError(t, actualError, keyring.ErrNotFound.Error())
	})

	testCases := []struct {
		missingTokenType string
		addToken         func() error
	}{
		{
			missingTokenType: "accessToken",
			addToken: func() error {
				return keyring.Set(secretRefreshToken, testTenantName, "fake-refresh-token")
			},
		},
		{
			missingTokenType: "refreshToken",
			addToken: func() error {
				return keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, 0), testTenantName, "access-token")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("deleteTokens returns error %s does not exists", testCase.missingTokenType), func(t *testing.T) {
			keyring.MockInit()

			err := testCase.addToken()
			require.NoError(t, err)

			actualError := DeleteTokens(testTenantName)
			require.EqualError(t, actualError, keyring.ErrNotFound.Error())
		})
	}

	t.Run("returns error if access token is greater than 50 * 2048 bytes", func(t *testing.T) {
		keyring.MockInit()

		accessToken := randomStringOfLength(50*2048 + 1)

		err := StoreAccessToken(testTenantName, accessToken)

		require.EqualError(t, err, ErrTokenSize.Error())
	})

	t.Run("can store and read a token of size 50 * 2048 bytes", func(t *testing.T) {
		keyring.MockInit()

		expectedAccesToken := randomStringOfLength(50 * 2048)

		err := StoreAccessToken(testTenantName, expectedAccesToken)
		require.NoError(t, err)
		actualAccessToken, err := GetAccessToken(testTenantName)
		require.NoError(t, err)

		assert.Equal(t, expectedAccesToken, actualAccessToken)
	})
}

func randomStringOfLength(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(b)
}
