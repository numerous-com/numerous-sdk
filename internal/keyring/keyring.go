package keyring

import (
	"errors"
	"fmt"

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

// ErrNotFound is re-exported from the keyring library for external use
var ErrNotFound = keyring.ErrNotFound

// StoreRefreshToken stores a tenant's refresh token in the system keyring.
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
