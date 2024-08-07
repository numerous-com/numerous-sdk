package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCurrentAppVersioon(t *testing.T) {
	t.Run("given app with default deployment response, then it returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"app": {
						"defaultDeployment": {
							"current": {
								"appVersion": {
									"id": "some-app-version-id"
								}
							}
						}
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := CurrentAppVersionInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.CurrentAppVersion(context.TODO(), input)

		expected := CurrentAppVersionOutput{
			AppVersionID: "some-app-version-id",
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
	})

	t.Run("given app with no default deployment, then it returns app not deployed error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"app": {
						"defaultDeployment": null
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := CurrentAppVersionInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.CurrentAppVersion(context.TODO(), input)

		assert.ErrorIs(t, err, ErrNotDeployed)
		assert.Empty(t, output)
	})

	t.Run("given app not found error, then it returns not found error", func(t *testing.T) {
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

		input := CurrentAppVersionInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.CurrentAppVersion(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAppNotFound)
		assert.Empty(t, output)
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

		input := CurrentAppVersionInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.CurrentAppVersion(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAccesDenied)
		assert.Equal(t, CurrentAppVersionOutput{}, output)
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

		input := CurrentAppVersionInput{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.CurrentAppVersion(context.TODO(), input)

		assert.ErrorContains(t, err, "expected error message")
		assert.Empty(t, output)
	})
}
