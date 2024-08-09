package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	t.Run("returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"appCreate": {
						"id": "some-app-id"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := CreateAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
			DisplayName:      "App Name",
			Description:      "App description",
		}
		output, err := s.Create(context.TODO(), input)

		expected := CreateAppOutput{
			AppID: "some-app-id",
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	})

	t.Run("returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["appCreate"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)
		input := CreateAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
			DisplayName:      "App Name",
			Description:      "App description",
		}
		output, err := s.Create(context.TODO(), input)

		expected := CreateAppOutput{}
		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Equal(t, expected, output)
	})

	t.Run("returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "expected error message",
					"location": [{"line": 1, "column": 1}],
					"path": ["appCreate"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := CreateAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
			DisplayName:      "App Name",
			Description:      "App description",
		}
		output, err := s.Create(context.TODO(), input)

		expected := CreateAppOutput{}
		assert.ErrorContains(t, err, "expected error message")
		assert.Equal(t, expected, output)
	})
}
