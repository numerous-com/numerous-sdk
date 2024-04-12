package auth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWaitUntilUserLogsIn(t *testing.T) {
	ticker := time.NewTicker(time.Millisecond)
	deviceCode := "1234"
	initialOauthTokenEndpoint := "https://test.com/token"
	testCredentials := Credentials{
		Audience:           "https://test.com/api/v2/",
		ClientID:           "client-id",
		DeviceCodeEndpoint: "https://test.com/oauth/device/code",
		OauthTokenEndpoint: initialOauthTokenEndpoint,
	}
	t.Run("successfully waits and handles response", func(t *testing.T) {
		validToken := test.GenerateJWT(t, "https://test.com/", time.Now()) // d
		tokenResponse := fmt.Sprintf(`{
			"access_token": "%s",
			"id_token": "id-token-here",
			"refresh_token": "refresh-token-here",
			"scope": "scope-here",
			"token_type": "token-type-here",
			"expires_in": 1000
		}`, validToken)

		pendingResponse := `{
			"error": "authorization_pending",
			"error_description": "still pending auth"
		}`

		ts := test.MockTransport{}
		ts.On("RoundTrip", mock.Anything).Once().Return(
			&http.Response{
				StatusCode: http.StatusForbidden,
				Body:       io.NopCloser(bytes.NewReader([]byte(pendingResponse))),
			},
			nil,
		)
		ts.On("RoundTrip", mock.Anything).Return(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(tokenResponse))),
			},
			nil,
		)
		client := &http.Client{Transport: &ts}

		actualResult, err := waitUntilUserLogsIn(context.Background(), client, ticker, deviceCode, testCredentials)

		expectedResult := Result{
			AccessToken:  validToken,
			RefreshToken: "refresh-token-here",
			IDToken:      "id-token-here",
		}

		require.NoError(t, err)
		assert.Equal(t, expectedResult.AccessToken, actualResult.AccessToken)
		assert.Equal(t, expectedResult.RefreshToken, actualResult.RefreshToken)
		assert.Equal(t, expectedResult.IDToken, actualResult.IDToken)
	})

	testCases := []struct {
		name       string
		httpStatus int
		response   string
		expect     string
	}{
		{
			name:       "handle malformed JSON",
			httpStatus: http.StatusOK,
			response:   "foo",
			expect:     "cannot decode response: invalid character 'o' in literal false (expecting 'a')",
		},
		{
			name:       "should pass through authorization server errors",
			httpStatus: http.StatusOK,
			response:   "{\"error\": \"slow_down\", \"error_description\": \"slow down!\"}",
			expect:     "slow down!",
		},
		{
			name:       "should error if can't parse as JWT",
			httpStatus: http.StatusOK,
			response:   "{\"access_token\": \"bad.token\"}",
			expect:     "failed to parse token: invalid character 'b' looking for beginning of value",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			transport := &test.TestTransport{
				WithResponse: &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(testCase.response))),
				},
			}
			client := &http.Client{Transport: transport}

			_, err := waitUntilUserLogsIn(context.Background(), client, ticker, deviceCode, testCredentials)

			assert.EqualError(t, err, testCase.expect)
		})
	}
}
