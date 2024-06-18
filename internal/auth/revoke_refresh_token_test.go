package auth

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevokeRefreshToken(t *testing.T) {
	testTenant := NewTenantAuthenticator("numerous-test.com", "test-client-id", "numerous-test.com/api/v2/")
	t.Run("successfully revoke token", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(""))),
			},
		}
		client := &http.Client{Transport: transport}

		err := testTenant.RevokeRefreshToken(client, "some-token")

		require.NoError(t, err)
	})

	t.Run("handles invalid request", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_request","error_description":"request was bad"}`))),
			},
		}
		client := &http.Client{Transport: transport}

		err := testTenant.RevokeRefreshToken(client, "some-token")

		assert.ErrorIs(t, err, ErrInvalidRequest)
	})

	t.Run("handles invalid client", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"invalid_client","error_description":"client was bad"}`))),
			},
		}
		client := &http.Client{Transport: transport}

		err := testTenant.RevokeRefreshToken(client, "some-token")

		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("handles unexpected error", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusBadGateway,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":"something","error_description":"something unexpected was bad"}`))),
			},
		}
		client := &http.Client{Transport: transport}

		err := testTenant.RevokeRefreshToken(client, "some-token")

		assert.ErrorIs(t, err, ErrUnexpected)
	})
}
