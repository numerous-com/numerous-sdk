package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAppUnshare(t *testing.T) {
	t.Run("given successful response it returns no error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appDeploymentIDRespBody := `
			{
				"data": {
					"app": {
						"defaultDeployment": {
							"id": "some-deployment-id"
						}
					}
				}
			}
		`
		resp := test.JSONResponse(appDeploymentIDRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		appUnshareRespBody := `
			{
				"data": {
					"appDeployUnshare": {
						"__typename": "AppDeployment"
					}
				}
			}
		`
		resp = test.JSONResponse(appUnshareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.UnshareApp(context.TODO(), input)

		assert.NoError(t, err)
	})

	t.Run("given access denied error then it returns access denied error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appDeploymentIDRespBody := `
			{
				"data": {
					"app": {
						"defaultDeployment": {
							"id": "some-deployment-id"
						}
					}
				}
			}
		`
		resp := test.JSONResponse(appDeploymentIDRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		appUnshareRespBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["appDeployUnshare"]
				}]
			}
		`
		resp = test.JSONResponse(appUnshareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.UnshareApp(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAccessDenied)
	})

	t.Run("given graphql error then it returns expected error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		appDeploymentIDRespBody := `
			{
				"data": {
					"app": {
						"defaultDeployment": {
							"id": "some-deployment-id"
						}
					}
				}
			}
		`
		resp := test.JSONResponse(appDeploymentIDRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		appUnshareRespBody := `
			{
				"errors": [{
					"message": "expected error message",
					"location": [{"line": 1, "column": 1}],
					"path": ["appDeployUnshare"]
				}]
			}
		`
		resp = test.JSONResponse(appUnshareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		err := s.UnshareApp(context.TODO(), input)

		assert.ErrorContains(t, err, "expected error message")
	})
}
