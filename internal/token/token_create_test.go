package token

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/test"
)

func TestCreate(t *testing.T) {
	t.Run("given access denied then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"errors": [{
				"message": "access denied",
				"location": [{"line": 1, "column": 1}],
				"path": ["personalAccessTokenCreate"]
			}]
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Create(context.TODO(), CreateTokenInput{})

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("returns expected created token", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"data": {
				"personalAccessTokenCreate": {
					"__typename": "PersonalAccessTokenCreated",
					"entry": {
						"name": "token name",
						"description": "token description"
					},
					"token": "some token value"
				}
			}
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Create(context.TODO(), CreateTokenInput{})

		if assert.NoError(t, err) {
			expected := CreateTokenOutput{Name: "token name", Description: "token description", Token: "some token value"}
			assert.Equal(t, expected, actual)
		}
	})

	t.Run("returns expected already exists error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"data": {
				"personalAccessTokenCreate": {
					"__typename": "PersonalAccessTokenAlreadyExists",
					"name": "some already existing name"
				}
			}
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Create(context.TODO(), CreateTokenInput{})

		assert.Empty(t, actual)
		if assert.ErrorIs(t, err, ErrPersonalAccessTokenAlreadyExists) {
			assert.ErrorContains(t, err, "some already existing name")
		}
	})

	t.Run("returns expected name invalid error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"data": {
				"personalAccessTokenCreate": {
					"__typename": "PersonalAccessTokenInvalidName",
					"name": "some invalid name",
					"reason": "some reason"
				}
			}
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Create(context.TODO(), CreateTokenInput{})

		assert.Empty(t, actual)
		if assert.ErrorIs(t, err, ErrPersonalAccessTokenNameInvalid) {
			assert.ErrorContains(t, err, "some reason")
		}
	})
}
