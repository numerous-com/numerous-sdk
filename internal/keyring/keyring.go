package keyring

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	secretRefreshToken                = "Numerous CLI Refresh Token"
	secretAccessToken                 = "Numerous CLI Access Token"
	secretAccessTokenChunkSizeInBytes = 2048

	// Access tokens have no size limit, but should be smaller than (50*2048) bytes.
	// The max number of loops safeguards against infinite loops, however unlikely.
	secretAccessTokenMaxChunks = 50
)

// ErrTokenSize is thrown when the token is invalid is larger than (50*2048) bytes.
var ErrTokenSize = errors.New("token is invalid")

// File-based keyring configuration
var (
	useFileBasedKeyring    bool
	credentialsFile        string
	keyringFallbackEnabled bool
	hasNotifiedFallback    bool
)

func init() {
	// Check if file-based keyring is explicitly enabled via environment variable
	if os.Getenv("NUMEROUS_LOGIN_USE_KEYRING") == "false" {
		useFileBasedKeyring = true
		setupFileBasedCredentials()
		return
	}

	// Enable automatic fallback for known problematic environments
	keyringFallbackEnabled = true

	// Set up file-based credentials path for potential fallback
	home, err := os.UserHomeDir()
	if err == nil {
		numerousDir := filepath.Join(home, ".numerous")
		credentialsFile = filepath.Join(numerousDir, "credentials.json")
	}
}

// setupFileBasedCredentials initializes the file-based credentials system
func setupFileBasedCredentials() {
	home, err := os.UserHomeDir()
	if err == nil {
		// Create ~/.numerous directory if it doesn't exist
		numerousDir := filepath.Join(home, ".numerous")
		os.MkdirAll(numerousDir, 0700)
		credentialsFile = filepath.Join(numerousDir, "credentials.json")
		if !hasNotifiedFallback {
			fmt.Printf("Using file-based credentials storage: %s\n", credentialsFile)
		}
	}
}

// tryKeyringFallback attempts to enable file-based keyring as a fallback
func tryKeyringFallback() {
	if !keyringFallbackEnabled || useFileBasedKeyring {
		return
	}

	useFileBasedKeyring = true
	setupFileBasedCredentials()

	if !hasNotifiedFallback {
		hasNotifiedFallback = true
		fmt.Printf("⚠️  Keyring access failed, switching to file-based credential storage\n")
		fmt.Printf("✓  Credentials will be stored securely at: %s\n", credentialsFile)
		if runtime.GOOS == "linux" {
			fmt.Printf("💡 Tip: Set NUMEROUS_LOGIN_USE_KEYRING=false to skip keyring attempts\n")
		}
		fmt.Println()
	}
}

// isKeyringError checks if an error is likely due to keyring unavailability
func isKeyringError(err error) bool {
	if err == nil {
		return false
	}

	// Common keyring errors that indicate the keyring is not available
	errStr := err.Error()
	keyringUnavailableErrors := []string{
		"secret service is not available",
		"no keyring available",
		"keyring not available",
		"dbus: session bus is not available",
		"Cannot autolaunch D-Bus without X11",
		"unknown collection",
		"prompt dismissed",
		"failed to unlock correct collection",
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

	for _, keyringErr := range keyringUnavailableErrors {
		if contains(errStr, keyringErr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

// indexOf finds the index of substr in s, returns -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// CredentialsData represents the structure of our credentials file
type CredentialsData struct {
	RefreshTokens map[string]string `json:"refresh_tokens"`
	AccessTokens  map[string]string `json:"access_tokens"`
}

// loadCredentials loads the credentials from the file
func loadCredentials() (CredentialsData, error) {
	creds := CredentialsData{
		RefreshTokens: make(map[string]string),
		AccessTokens:  make(map[string]string),
	}

	// If the file doesn't exist yet, return empty credentials
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		return creds, nil
	}

	data, err := os.ReadFile(credentialsFile)
	if err != nil {
		return creds, err
	}

	err = json.Unmarshal(data, &creds)
	return creds, err
}

// saveCredentials saves the credentials to the file
func saveCredentials(creds CredentialsData) error {
	// Ensure directory exists
	if credentialsFile != "" {
		dir := filepath.Dir(credentialsFile)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsFile, data, 0600)
}

// StoreRefreshToken stores a tenant's refresh token in the system keyring or file.
func StoreRefreshToken(tenant, value string) error {
	if useFileBasedKeyring {
		creds, err := loadCredentials()
		if err != nil {
			return err
		}
		creds.RefreshTokens[tenant] = value
		return saveCredentials(creds)
	}

	// Try keyring first
	err := keyring.Set(secretRefreshToken, tenant, value)
	if err != nil && isKeyringError(err) {
		// Fall back to file-based storage
		tryKeyringFallback()
		creds, loadErr := loadCredentials()
		if loadErr != nil {
			return loadErr
		}
		creds.RefreshTokens[tenant] = value
		return saveCredentials(creds)
	}
	return err
}

// GetRefreshToken retrieves a tenant's refresh token from the system keyring or file.
func GetRefreshToken(tenant string) (string, error) {
	if useFileBasedKeyring {
		creds, err := loadCredentials()
		if err != nil {
			return "", err
		}
		token, exists := creds.RefreshTokens[tenant]
		if !exists {
			return "", keyring.ErrNotFound
		}
		return token, nil
	}

	// Try keyring first
	token, err := keyring.Get(secretRefreshToken, tenant)
	if err != nil && isKeyringError(err) {
		// Fall back to file-based storage
		tryKeyringFallback()
		creds, loadErr := loadCredentials()
		if loadErr != nil {
			return "", loadErr
		}
		fileToken, exists := creds.RefreshTokens[tenant]
		if !exists {
			return "", keyring.ErrNotFound
		}
		return fileToken, nil
	}
	return token, err
}

// StoreAccessToken stores a tenant's access token in the system keyring or file.
func StoreAccessToken(tenant, value string) error {
	if useFileBasedKeyring {
		creds, err := loadCredentials()
		if err != nil {
			return err
		}
		creds.AccessTokens[tenant] = value
		return saveCredentials(creds)
	}

	chunks := chunk(value, secretAccessTokenChunkSizeInBytes)

	if len(chunks) > secretAccessTokenMaxChunks {
		return ErrTokenSize
	}

	// Try keyring first
	var firstError error
	for i, chunk := range chunks {
		err := keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, i), tenant, chunk)
		if err != nil {
			if firstError == nil {
				firstError = err
			}
			if isKeyringError(err) {
				// Fall back to file-based storage
				tryKeyringFallback()
				creds, loadErr := loadCredentials()
				if loadErr != nil {
					return loadErr
				}
				creds.AccessTokens[tenant] = value
				return saveCredentials(creds)
			}
		}
	}

	return firstError
}

// DeleteTokens deletes a tenant's tokens from the system keyring or file.
func DeleteTokens(tenant string) error {
	if useFileBasedKeyring {
		creds, err := loadCredentials()
		if err != nil {
			return err
		}
		delete(creds.RefreshTokens, tenant)
		delete(creds.AccessTokens, tenant)
		return saveCredentials(creds)
	}

	// Try keyring operations first
	deleteAccessTokenErr := deleteAccessToken(tenant)
	deleteRefreshTokenErr := keyring.Delete(secretRefreshToken, tenant)

	// Check if errors are due to keyring unavailability
	if (deleteAccessTokenErr != nil && isKeyringError(deleteAccessTokenErr)) ||
		(deleteRefreshTokenErr != nil && isKeyringError(deleteRefreshTokenErr)) {
		// Fall back to file-based storage
		tryKeyringFallback()
		creds, err := loadCredentials()
		if err != nil {
			return err
		}
		delete(creds.RefreshTokens, tenant)
		delete(creds.AccessTokens, tenant)
		return saveCredentials(creds)
	}

	if deleteAccessTokenErr != nil {
		return deleteAccessTokenErr
	}
	if deleteRefreshTokenErr != nil {
		return deleteRefreshTokenErr
	}

	return nil
}

func deleteAccessToken(tenant string) error {
	if useFileBasedKeyring {
		// This is handled in DeleteTokens for file-based storage
		return nil
	}

	for i := 0; i < secretAccessTokenMaxChunks; i++ {
		err := keyring.Delete(fmt.Sprintf("%s %d", secretAccessToken, i), tenant)
		// Only return if we have pulled more than 1 item from the keyring, otherwise this will be
		// a valid "secret not found in keyring".
		if err == keyring.ErrNotFound && i > 0 {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return keyring.ErrNotFound
}

// GetAccessToken retrieves a tenant's access token from the system keyring or file.
func GetAccessToken(tenant string) (string, error) {
	if useFileBasedKeyring {
		creds, err := loadCredentials()
		if err != nil {
			return "", err
		}
		token, exists := creds.AccessTokens[tenant]
		if !exists {
			return "", keyring.ErrNotFound
		}
		return token, nil
	}

	var accessToken string

	// Try keyring first
	for i := 0; i < secretAccessTokenMaxChunks; i++ {
		a, err := keyring.Get(fmt.Sprintf("%s %d", secretAccessToken, i), tenant)
		// Only return if we have pulled more than 1 item from the keyring, otherwise this will be
		// a valid "secret not found in keyring".
		if err == keyring.ErrNotFound && i > 0 {
			return accessToken, nil
		}
		if err != nil {
			if isKeyringError(err) {
				// Fall back to file-based storage
				tryKeyringFallback()
				creds, loadErr := loadCredentials()
				if loadErr != nil {
					return "", loadErr
				}
				fileToken, exists := creds.AccessTokens[tenant]
				if !exists {
					return "", keyring.ErrNotFound
				}
				return fileToken, nil
			}
			return "", err
		}
		accessToken += a
	}

	return accessToken, nil
}

// Splits string into chunks of chunkSize
func chunk(slice string, chunkSize int) []string {
	var chunks []string
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		// Necessary check to avoid slicing beyond
		// slice capacity.
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
