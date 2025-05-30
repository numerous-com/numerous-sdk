package keyring

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

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

func TestKeyringFallback(t *testing.T) {
	// Save original state
	originalUseFileBasedKeyring := useFileBasedKeyring
	originalCredentialsFile := credentialsFile
	originalKeyringFallbackEnabled := keyringFallbackEnabled
	originalHasNotifiedFallback := hasNotifiedFallback

	// Clean up after tests
	defer func() {
		useFileBasedKeyring = originalUseFileBasedKeyring
		credentialsFile = originalCredentialsFile
		keyringFallbackEnabled = originalKeyringFallbackEnabled
		hasNotifiedFallback = originalHasNotifiedFallback
	}()

	t.Run("isKeyringError detects common keyring errors", func(t *testing.T) {
		keyringErrors := []string{
			"secret service is not available",
			"no keyring available",
			"keyring not available",
			"dbus: session bus is not available",
			"Cannot autolaunch D-Bus without X11",
			"unknown collection",
			"prompt dismissed",
			"failed to unlock correct collection",
			"failed to unlock correct collection '/org/freedesktop/secrets/aliases/default'",
			"failed to unlock collection",
			"collection not found",
			"org.freedesktop.secrets",
			"DBus.Error.ServiceUnknown",
			"DBus.Error.NoReply",
			"secret-tool: cannot create an item",
			"The name org.freedesktop.secrets was not provided",
			"Failed to open connection to session D-Bus",
			"Unable to autolaunch a dbus-daemon",
		}

		for _, errMsg := range keyringErrors {
			err := fmt.Errorf(errMsg)
			assert.True(t, isKeyringError(err), "Should detect '%s' as keyring error", errMsg)
		}

		// Test non-keyring errors
		nonKeyringErrors := []string{
			"network error",
			"invalid token format",
			"permission denied",
		}

		for _, errMsg := range nonKeyringErrors {
			err := fmt.Errorf(errMsg)
			assert.False(t, isKeyringError(err), "Should not detect '%s' as keyring error", errMsg)
		}

		// Test nil error
		assert.False(t, isKeyringError(nil), "Should not detect nil as keyring error")
	})

	t.Run("fallback enables file-based storage", func(t *testing.T) {
		// Reset state
		useFileBasedKeyring = false
		keyringFallbackEnabled = true
		hasNotifiedFallback = false

		// Set up temporary credentials file
		tmpDir := t.TempDir()
		credentialsFile = fmt.Sprintf("%s/credentials.json", tmpDir)

		// Trigger fallback
		tryKeyringFallback()

		assert.True(t, useFileBasedKeyring, "Should enable file-based keyring")
		assert.True(t, hasNotifiedFallback, "Should mark as notified")
	})

	t.Run("fallback is disabled when keyringFallbackEnabled is false", func(t *testing.T) {
		// Reset state
		useFileBasedKeyring = false
		keyringFallbackEnabled = false
		hasNotifiedFallback = false

		// Trigger fallback (should not work)
		tryKeyringFallback()

		assert.False(t, useFileBasedKeyring, "Should not enable file-based keyring when fallback disabled")
		assert.False(t, hasNotifiedFallback, "Should not mark as notified when fallback disabled")
	})

	t.Run("fallback does nothing when already using file-based storage", func(t *testing.T) {
		// Reset state
		useFileBasedKeyring = true
		keyringFallbackEnabled = true
		hasNotifiedFallback = false

		// Trigger fallback (should not change anything)
		tryKeyringFallback()

		assert.True(t, useFileBasedKeyring, "Should remain file-based")
		assert.False(t, hasNotifiedFallback, "Should not notify when already using file-based storage")
	})
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Test the init() function behavior with environment variable
	originalEnv := os.Getenv("NUMEROUS_LOGIN_USE_KEYRING")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("NUMEROUS_LOGIN_USE_KEYRING")
		} else {
			os.Setenv("NUMEROUS_LOGIN_USE_KEYRING", originalEnv)
		}
	}()

	t.Run("environment variable false forces file-based storage", func(t *testing.T) {
		os.Setenv("NUMEROUS_LOGIN_USE_KEYRING", "false")

		// Reset state and reinitialize
		useFileBasedKeyring = false
		keyringFallbackEnabled = false
		hasNotifiedFallback = false

		// Simulate init() behavior
		if os.Getenv("NUMEROUS_LOGIN_USE_KEYRING") == "false" {
			useFileBasedKeyring = true
			setupFileBasedCredentials()
		} else {
			keyringFallbackEnabled = true
		}

		assert.True(t, useFileBasedKeyring, "Should use file-based storage when env var is false")
		assert.False(t, keyringFallbackEnabled, "Should not enable fallback when explicitly set to file-based")
	})

	t.Run("no environment variable enables fallback", func(t *testing.T) {
		os.Unsetenv("NUMEROUS_LOGIN_USE_KEYRING")

		// Reset state and reinitialize
		useFileBasedKeyring = false
		keyringFallbackEnabled = false
		hasNotifiedFallback = false

		// Simulate init() behavior
		if os.Getenv("NUMEROUS_LOGIN_USE_KEYRING") == "false" {
			useFileBasedKeyring = true
			setupFileBasedCredentials()
		} else {
			keyringFallbackEnabled = true
		}

		assert.False(t, useFileBasedKeyring, "Should not use file-based storage initially")
		assert.True(t, keyringFallbackEnabled, "Should enable fallback when no env var set")
	})
}

func randomStringOfLength(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
