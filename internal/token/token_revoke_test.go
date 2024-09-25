package token

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

func TestRevoke(t *testing.T) {
	t.Run("given access denied then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"errors": [{
				"message": "access denied",
				"location": [{"line": 1, "column": 1}],
				"path": ["personalAccessTokenRevoke"]
			}]
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Revoke(context.TODO(), "some-token-id")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("returns expected revoked token", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"data": {
				"personalAccessTokenRevoke": {
					"name": "token name",
					"description": "token description"
				}
			}
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Revoke(context.TODO(), "some-token-id")

		require.NoError(t, err)
		expected := RevokeTokenOutput{Name: "token name", Description: "token description"}
		assert.Equal(t, expected, actual)
	})
}
