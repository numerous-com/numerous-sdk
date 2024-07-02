package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAppDelete(t *testing.T) {
	t.Run("given access denied then it returns access denied error", func(t *testing.T) {
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

		input := DeleteAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.Delete(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAccesDenied)
	})

	t.Run("given app not found then it returns app not found error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "app not found",
					"location": [{"line": 1, "column": 1}],
					"path": ["app"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := DeleteAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.Delete(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAppNotFound)
	})

	t.Run("given graphql error then it returns given graphql error", func(t *testing.T) {
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

		input := DeleteAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.Delete(context.TODO(), input)

		assert.ErrorContains(t, err, "expected error message")
	})

	t.Run("given successful response it returns no error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"appDelete": {
						"__typename": "AppDeleted"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := DeleteAppInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.Delete(context.TODO(), input)

		assert.NoError(t, err)
	})
}
