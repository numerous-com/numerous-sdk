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
	return keyring.Set(secretRefreshToken, tenant, value)
}

// GetRefreshToken retrieves a tenant's refresh token from the system keyring.
func GetRefreshToken(tenant string) (string, error) {
	return keyring.Get(secretRefreshToken, tenant)
}

func StoreAccessToken(tenant, value string) error {
	chunks := chunk(value, secretAccessTokenChunkSizeInBytes)

	if len(chunks) > secretAccessTokenMaxChunks {
		return ErrTokenSize
	}

	for i, chunk := range chunks {
		err := keyring.Set(fmt.Sprintf("%s %d", secretAccessToken, i), tenant, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteTokens(tenant string) error {
	deleteAccessTokenErr := deleteAccessToken(tenant)
	deleteRefreshTokenErr := keyring.Delete(secretRefreshToken, tenant)

	if deleteAccessTokenErr != nil {
		return deleteAccessTokenErr
	}
	if deleteRefreshTokenErr != nil {
		return deleteRefreshTokenErr
	}

	return nil
}

func deleteAccessToken(tenant string) error {
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

func GetAccessToken(tenant string) (string, error) {
	var accessToken string

	for i := 0; i < secretAccessTokenMaxChunks; i++ {
		a, err := keyring.Get(fmt.Sprintf("%s %d", secretAccessToken, i), tenant)
		// Only return if we have pulled more than 1 item from the keyring, otherwise this will be
		// a valid "secret not found in keyring".
		if err == keyring.ErrNotFound && i > 0 {
			return accessToken, nil
		}
		if err != nil {
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
