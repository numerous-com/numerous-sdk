package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTasks(t *testing.T) {
	newTestService := func(t *testing.T) (*test.MockDoer, *Service) {
		t.Helper()

		doer := &test.MockDoer{}
		c := test.CreateTestGQLClient(t, doer)

		return doer, New(c, nil, nil)
	}

	t.Run("returns expected tasks", func(t *testing.T) {
		doer, s := newTestService(t)

		respBody := `
			{
				"data": {
					"app": {
						"id": "test-app-id",
						"defaultDeployment": {
							"current": {
								"appVersion": {
									"tasks": [
										{
											"id": "task-1",
											"command": ["python", "worker.py"]
										},
										{
											"id": "task-2",
											"command": ["python", "process.py"]
										}
									]
								}
							}
						}
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		tasks, err := s.GetTasks(context.TODO(), GetTasksInput{
			OrganizationSlug: "test-org",
			AppSlug:          "test-app",
		})

		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "task-1", tasks[0].ID)
		assert.Equal(t, "task-2", tasks[1].ID)
	})

	t.Run("returns ErrAppNotFound when app is null", func(t *testing.T) {
		doer, s := newTestService(t)

		respBody := `
			{
				"data": {
					"app": null
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		tasks, err := s.GetTasks(context.TODO(), GetTasksInput{
			OrganizationSlug: "test-org",
			AppSlug:          "nonexistent-app",
		})

		assert.ErrorIs(t, err, ErrAppNotFound)
		assert.Nil(t, tasks)
	})

	t.Run("returns ErrDeploymentNotFound when deployment is nil", func(t *testing.T) {
		doer, s := newTestService(t)

		respBody := `
			{
				"data": {
					"app": {
						"id": "test-app-id",
						"defaultDeployment": null
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		tasks, err := s.GetTasks(context.TODO(), GetTasksInput{
			OrganizationSlug: "test-org",
			AppSlug:          "test-app",
		})

		assert.ErrorIs(t, err, ErrDeploymentNotFound)
		assert.Nil(t, tasks)
	})

	t.Run("returns ErrDeploymentNotFound when current is nil", func(t *testing.T) {
		doer, s := newTestService(t)

		respBody := `
			{
				"data": {
					"app": {
						"id": "test-app-id",
						"defaultDeployment": {
							"current": null
						}
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		tasks, err := s.GetTasks(context.TODO(), GetTasksInput{
			OrganizationSlug: "test-org",
			AppSlug:          "test-app",
		})

		assert.ErrorIs(t, err, ErrDeploymentNotFound)
		assert.Nil(t, tasks)
	})

	t.Run("returns error if query fails", func(t *testing.T) {
		doer, s := newTestService(t)

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

		tasks, err := s.GetTasks(context.TODO(), GetTasksInput{
			OrganizationSlug: "test-org",
			AppSlug:          "test-app",
		})

		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, tasks)
	})
}
