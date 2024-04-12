package auth

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefeshAccessToken(t *testing.T) {
	testCredentials := Credentials{
		Audience:           "https://test.com/api/v2/",
		ClientID:           "client-id",
		DeviceCodeEndpoint: "https://test.com/oauth/device/code",
		OauthTokenEndpoint: "https://test.com/token",
	}
	t.Run("happy path", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
						"access_token": "access-token-here",
						"id_token": "id-token-here",
						"token_type": "token-type-here",
						"expires_in": 1000
					}`))),
			},
		}

		client := &http.Client{Transport: transport}

		actualResponse, err := refreshAccessToken(client, "refresh-token-here", testCredentials)
		require.NoError(t, err)

		expectedResponse := TokenResponse{
			AccessToken: "access-token-here",
			IDToken:     "id-token-here",
			TokenType:   "token-type-here",
			ExpiresIn:   1000,
		}

		assert.Equal(t, expectedResponse, actualResponse)

		req := transport.Requests[0]
		err = req.ParseForm()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "https://test.com/token", req.URL.String())
		assert.Equal(t, "refresh_token", req.Form["grant_type"][0])
		assert.Equal(t, "client-id", req.Form["client_id"][0])
		assert.Equal(t, "refresh-token-here", req.Form["refresh_token"][0])
	})

	t.Run("Fails if empty refreshToken", func(t *testing.T) {
		client := &http.Client{Transport: &test.TestTransport{}}
		response, err := refreshAccessToken(client, "", testCredentials)

		require.Error(t, err)
		assert.Equal(t, TokenResponse{}, response)
	})

	t.Run("Fails if post request returns error", func(t *testing.T) {
		client := &http.Client{Transport: &test.TestTransport{WithError: http.ErrHandlerTimeout}}
		response, err := refreshAccessToken(client, "refresh-token-here", testCredentials)

		require.Error(t, err)
		assert.Equal(t, TokenResponse{}, response)
	})

	t.Run("Fails if http status is not http.OK", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewReader([]byte("Bad request"))),
			},
		}
		client := &http.Client{Transport: transport}

		response, err := refreshAccessToken(client, "refresh-token-here", testCredentials)

		require.Error(t, err)
		assert.Equal(t, TokenResponse{}, response)
	})

	t.Run("Fails if body can't be decoded", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte("Bad response"))),
			},
		}
		client := &http.Client{Transport: transport}

		response, err := refreshAccessToken(client, "refresh-token-here", testCredentials)

		require.Error(t, err)
		assert.Equal(t, TokenResponse{}, response)
	})
}
