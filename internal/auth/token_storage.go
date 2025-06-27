package auth

import (
	"errors"
	"fmt"
	"os"

	"numerous.com/cli/internal/keyring"
	"numerous.com/cli/internal/output"
)

// TokenStorage defines the interface for token storage backends
type TokenStorage interface {
	StoreAccessToken(tenant, token string) error
	StoreRefreshToken(tenant, token string) error
	StoreBothTokens(tenant, accessToken, refreshToken string) error
	GetLoggedInUser(tenant string) *User
	RemoveTokens(tenant string) error
}

// KeyringStorage implements TokenStorage using the system keyring
type KeyringStorage struct{}

// NewKeyrringStorage creates a new keyring storage instance
func NewKeyringStorage() *KeyringStorage {
	return &KeyringStorage{}
}

func (ks *KeyringStorage) StoreAccessToken(tenant, token string) error {
	return keyring.StoreAccessToken(tenant, token)
}

func (ks *KeyringStorage) StoreRefreshToken(tenant, token string) error {
	return keyring.StoreRefreshToken(tenant, token)
}

func (ks *KeyringStorage) StoreBothTokens(tenant, accessToken, refreshToken string) error {
	if err := ks.StoreAccessToken(tenant, accessToken); err != nil {
		return err
	}
	if err := ks.StoreRefreshToken(tenant, refreshToken); err != nil {
		// Clean up access token if refresh token storage fails
		_ = keyring.DeleteTokens(tenant)
		return err
	}

	return nil
}

func (ks *KeyringStorage) GetLoggedInUser(tenant string) *User {
	return getLoggedInUserFromKeyring(tenant)
}

func (ks *KeyringStorage) RemoveTokens(tenant string) error {
	return keyring.DeleteTokens(tenant)
}

// FileTokenStorage implements TokenStorage using file-based storage
type FileTokenStorage struct {
	fileStorage *FileStorage
}

// NewFileTokenStorage creates a new file-based token storage instance
func NewFileTokenStorage() *FileTokenStorage {
	return &FileTokenStorage{
		fileStorage: NewFileStorage(),
	}
}

func (fts *FileTokenStorage) StoreAccessToken(tenant, token string) error {
	// For file storage, we need both tokens to store properly
	// Try to get existing tokens first
	existingData, err := fts.fileStorage.RetrieveToken()
	if err != nil && err != ErrFileNotFound {
		return fmt.Errorf("failed to read existing tokens: %v", err)
	}

	var refreshToken string
	if existingData != nil && existingData.Tenant == tenant {
		refreshToken = existingData.RefreshToken
	}

	if refreshToken == "" {
		return errors.New("cannot store access token without refresh token in file storage")
	}

	return fts.fileStorage.StoreToken(token, refreshToken, tenant)
}

func (fts *FileTokenStorage) StoreRefreshToken(tenant, token string) error {
	// For file storage, we need both tokens to store properly
	// Try to get existing tokens first
	existingData, err := fts.fileStorage.RetrieveToken()
	if err != nil && err != ErrFileNotFound {
		return fmt.Errorf("failed to read existing tokens: %v", err)
	}

	var accessToken string
	if existingData != nil && existingData.Tenant == tenant {
		accessToken = existingData.AccessToken
	}

	if accessToken == "" {
		return errors.New("cannot store refresh token without access token in file storage")
	}

	return fts.fileStorage.StoreToken(accessToken, token, tenant)
}

func (fts *FileTokenStorage) StoreBothTokens(tenant, accessToken, refreshToken string) error {
	// Check if we should ask for user consent
	if !hasUserConsentedToFileStorage() {
		consented, err := requestFileStorageConsent(fts.fileStorage.GetTokenDirectory())
		if err != nil {
			return fmt.Errorf("failed to get user consent: %v", err)
		}
		if !consented {
			return ErrUserDeclinedConsent
		}
		setUserConsentedToFileStorage(true)
	}

	return fts.fileStorage.StoreToken(accessToken, refreshToken, tenant)
}

func (fts *FileTokenStorage) GetLoggedInUser(tenant string) *User {
	tokenData, err := fts.fileStorage.RetrieveToken()
	if err != nil {
		return nil
	}

	// Verify tenant matches
	if tokenData.Tenant != tenant {
		return nil
	}

	return &User{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		Tenant:       tenant,
	}
}

func (fts *FileTokenStorage) RemoveTokens(tenant string) error {
	return fts.fileStorage.DeleteToken()
}

// TokenStorageMode represents the mode of token storage
type TokenStorageMode int

const (
	KeyringMode TokenStorageMode = iota
	FileMode
)

// CreateTokenStorage creates the appropriate token storage based on availability and environment
func CreateTokenStorage() (TokenStorage, TokenStorageMode) {
	// Check if file mode is forced via environment variable
	if os.Getenv("NUMEROUS_FORCE_FILE_STORAGE") != "" {
		return NewFileTokenStorage(), FileMode
	}

	// Check if keyring is available
	if isKeychainAvailable() {
		return NewKeyringStorage(), KeyringMode
	}

	// Display warning and fallback to file storage
	output.PrintWarning("Keyring service unavailable", "Token will be stored in a local file with reduced security.")

	return NewFileTokenStorage(), FileMode
}

// isKeychainAvailable checks if keyring is available
func isKeychainAvailable() bool {
	// Try a simple keyring operation to test availability
	_, err := keyring.GetRefreshToken("test-availability-check")
	// If error is "not found", keyring is available but no token exists
	// Any other error indicates keyring unavailability
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}

// Simple consent tracking - session-based only
var userConsentedToFileStorage bool

func hasUserConsentedToFileStorage() bool {
	return userConsentedToFileStorage
}

func setUserConsentedToFileStorage(consented bool) {
	userConsentedToFileStorage = consented
}

func requestFileStorageConsent(tokenDir string) (bool, error) {
	// Display security warning using consistent output functions
	output.PrintWarning("Keyring service unavailable", fmt.Sprintf("Token will be stored in a local file with reduced security.\nFile location: %s/.token (permissions: 600)", tokenDir))
	fmt.Print("Continue with file storage? [y/N]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, fmt.Errorf("failed to read user input: %v", err)
	}

	return response == "y" || response == "Y" || response == "yes" || response == "Yes", nil
}
