package auth

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDeviceCode(t *testing.T) {
	testTenant := NewTenantAuthenticator("numerous-test.com", "test-client-id")
	t.Run("successfully retrieve state from response", func(t *testing.T) {
		transport := &test.TestTransport{
			WithResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"device_code": "device-code-here",
					"user_code": "user-code-here",
					"verification_uri_complete": "verification-uri-here",
					"expires_in": 1000,
					"interval": 1
				}`))),
			},
		}
		client := &http.Client{Transport: transport}

		state, err := testTenant.GetDeviceCode(context.Background(), client)

		require.NoError(t, err)
		assert.Equal(t, "device-code-here", state.DeviceCode)
		assert.Equal(t, "user-code-here", state.UserCode)
		assert.Equal(t, "verification-uri-here", state.VerificationURI)
		assert.Equal(t, 1000, state.ExpiresIn)
		assert.Equal(t, 1, state.Interval)
		assert.Equal(t, time.Duration(4000000000), state.IntervalDuration())
	})

	testCases := []struct {
		name       string
		httpStatus int
		response   string
		expect     string
	}{
		{
			name:       "handle HTTP status errors",
			httpStatus: http.StatusNotFound,
			response:   "Test response return",
			expect:     "received a 404 response: Test response return",
		},
		{
			name:       "handle bad JSON response",
			httpStatus: http.StatusOK,
			response:   "foo",
			expect:     "failed to decode the response: invalid character 'o' in literal false (expecting 'a')",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			transport := &test.TestTransport{
				WithResponse: &http.Response{
					StatusCode: testCase.httpStatus,
					Body:       io.NopCloser(bytes.NewReader([]byte(testCase.response))),
				},
			}
			client := &http.Client{Transport: transport}

			_, err := testTenant.GetDeviceCode(context.Background(), client)

			assert.EqualError(t, err, testCase.expect)
		})
	}
}
