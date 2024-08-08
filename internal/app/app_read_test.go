package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAppRead(t *testing.T) {
	t.Run("given app response it returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"app": {
						"id": "some-app-id"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := ReadAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ReadApp(context.TODO(), input)

		expected := ReadAppOutput{
			AppID: "some-app-id",
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	})

	t.Run("given app not found error then it returns not found error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"app": null
				},
				"errors": [{
					"message": "app not found",
					"location": [{"line": 1, "column": 1}],
					"path": ["app"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := ReadAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ReadApp(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAppNotFound)
		assert.Equal(t, ReadAppOutput{}, output)
	})

	t.Run("given access denied error then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["app"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := ReadAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ReadApp(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Equal(t, ReadAppOutput{}, output)
	})

	t.Run("given graphql error then it returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "expected error message",
					"location": [{"line": 1, "column": 1}],
					"path": ["app"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := ReadAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ReadApp(context.TODO(), input)

		expected := ReadAppOutput{}
		assert.ErrorContains(t, err, "expected error message")
		assert.Equal(t, expected, output)
	})
}
