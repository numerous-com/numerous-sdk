package gql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPURL(t *testing.T) {
	t.Run("uses environment variable if set", func(t *testing.T) {
		t.Setenv("NUMEROUS_GRAPHQL_HTTP_URL", "https://test-graphql-url")

		assert.Equal(t, "https://test-graphql-url", GetHTTPURL())
	})

	t.Run("uses default variable if environment is not set", func(t *testing.T) {
		t.Setenv("NUMEROUS_GRAPHQL_HTTP_URL", "")

		assert.Equal(t, httpURL, GetHTTPURL())
	})
}

func TestGetWSURL(t *testing.T) {
	t.Run("uses environment variable if set", func(t *testing.T) {
		t.Setenv("NUMEROUS_GRAPHQL_WS_URL", "wss://test-graphql-url")

		assert.Equal(t, "wss://test-graphql-url", GetWSURL())
	})

	t.Run("uses default variable if environment is not set", func(t *testing.T) {
		t.Setenv("NUMEROUS_GRAPHQL_WS_URL", "")

		assert.Equal(t, wsURL, GetWSURL())
	})
}
