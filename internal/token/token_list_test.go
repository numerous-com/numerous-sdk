package token

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

func TestList(t *testing.T) {
	t.Run("given access denied then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"errors": [{
				"message": "access denied",
				"location": [{"line": 1, "column": 1}],
				"path": ["me"]
			}]
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.List(context.TODO())

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("returns expected tokens", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"data": {
				"me": {
					"personalAccessTokens": [
						{
							"id": "first-token-id",
							"name": "First token name",
							"description": "first token description",
							"expiresAt": "2026-09-27T11:12:13.123456Z"
						},
						{
							"id": "second-token-id",
							"name": "Second token name",
							"description": "second token description",
							"expiresAt": null
						}
					]
				}
			}
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.List(context.TODO())

		require.NoError(t, err)
		expectedTime, err := time.Parse(time.RFC3339, "2026-09-27T11:12:13.123456Z")
		require.NoError(t, err)
		expected := ListTokenOutput{
			{
				ID:          "first-token-id",
				Name:        "First token name",
				Description: "first token description",
				ExpiresAt:   &expectedTime,
			},
			{
				ID:          "second-token-id",
				Name:        "Second token name",
				Description: "second token description",
				ExpiresAt:   nil,
			},
		}
		assert.Equal(t, expected, actual)
	})
}
