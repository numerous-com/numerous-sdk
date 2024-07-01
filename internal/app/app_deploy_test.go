package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeployApp(t *testing.T) {
	t.Run("returns expected output", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"appDeploy": {
						"id": "some-deploy-version-id"
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := DeployAppInput{
			AppVersionID: "some-app-version-id",
			Secrets:      map[string]string{"SECRET1": "some secret value"},
		}
		output, err := s.DeployApp(context.TODO(), input)

		expected := DeployAppOutput{
			DeploymentVersionID: "some-deploy-version-id",
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
					"path": ["appDeploy"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)
		input := DeployAppInput{
			AppVersionID: "some-app-version-id",
		}
		output, err := s.DeployApp(context.TODO(), input)

		expected := DeployAppOutput{}
		assert.ErrorIs(t, err, ErrAccesDenied)
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
					"path": ["appDeploy"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)
		input := DeployAppInput{
			AppVersionID: "some-app-version-id",
		}
		output, err := s.DeployApp(context.TODO(), input)

		expected := DeployAppOutput{}
		assert.ErrorContains(t, err, "expected error message")
		assert.Equal(t, expected, output)
	})
}
