package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAppShare(t *testing.T) {
	t.Run("given app shared URL response it returns expected output", func(t *testing.T) {
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

		appShareRespBody := `
			{
				"data": {
					"appDeployShare": {
						"sharedURL": "https://test-numerous.com/share/123"
					}
				}
			}
		`
		resp = test.JSONResponse(appShareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ShareApp(context.TODO(), input)

		expected := ShareAppOutput{
			SharedURL: ref("https://test-numerous.com/share/123"),
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, output)
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

		appShareRespBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["appDeployShare"]
				}]
			}
		`
		resp = test.JSONResponse(appShareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ShareApp(context.TODO(), input)

		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Equal(t, ShareAppOutput{}, output)
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

		appShareRespBody := `
			{
				"errors": [{
					"message": "expected error message",
					"location": [{"line": 1, "column": 1}],
					"path": ["appDeployShare"]
				}]
			}
		`
		resp = test.JSONResponse(appShareRespBody)
		doer.On("Do", mock.Anything).Return(resp, nil).Once()

		input := appident.AppIdentifier{
			OrganizationSlug: "organization-slug",
			AppSlug:          "app-slug",
		}
		output, err := s.ShareApp(context.TODO(), input)

		assert.ErrorContains(t, err, "expected error message")
		assert.Equal(t, ShareAppOutput{}, output)
	})
}
