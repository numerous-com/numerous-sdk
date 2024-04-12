package test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Generates a JWT

issued 1 hour before passed in expiration
*/
func GenerateJWT(t *testing.T, issuer string, expiration time.Time) string {
	t.Helper()
	// Generate a private key for RS256
	amountOfBits := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, amountOfBits)
	require.NoError(t, err)

	// Create a new token
	token := jwt.New()

	// Set standard claims
	require.NoError(t, token.Set(jwt.SubjectKey, "123456"))
	require.NoError(t, token.Set(jwt.IssuerKey, issuer))
	require.NoError(t, token.Set(jwt.IssuedAtKey, expiration.Add(-time.Hour)))
	require.NoError(t, token.Set(jwt.ExpirationKey, expiration))

	// Sign the token with RS256 algorithm using the private key
	signedToken, err := jwt.Sign(token, jwa.RS256, privateKey)
	require.NoError(t, err)
	tokenString := string(signedToken)
	assert.NotEmpty(t, tokenString)

	return tokenString
}
