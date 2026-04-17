package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskInstance(t *testing.T) {
	t.Run("returns expected result", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"taskInstance": {
						"id": "test-instance-id",
						"createdAt": "2025-01-01T00:00:00Z",
						"input": null,
						"output": null,
						"progress": {
							"value": 50.0,
							"message": "downloading"
						},
						"task": {
							"id": "test-task-id",
							"command": ["python", "main.py"]
						},
						"workload": {
							"status": "RUNNING",
							"startedAt": "2025-01-01T00:00:00Z",
							"cpuUsage": null,
							"memoryUsageMB": null,
							"exitCode": null
						}
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		result, err := s.GetTaskInstance(context.TODO(), GetTaskInstanceInput{
			TaskInstanceID: "test-instance-id",
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-instance-id", result.ID)
		assert.Equal(t, "test-task-id", result.Task.ID)
		assert.Equal(t, "RUNNING", result.Workload.Status)
	})

	t.Run("returns ErrTaskInstanceNotFound when server returns null", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"taskInstance": null
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		result, err := s.GetTaskInstance(context.TODO(), GetTaskInstanceInput{
			TaskInstanceID: "nonexistent-instance",
		})

		assert.ErrorIs(t, err, ErrTaskInstanceNotFound)
		assert.Nil(t, result)
	})

	t.Run("returns error if query fails", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [{
					"message": "access denied",
					"location": [{"line": 1, "column": 1}],
					"path": ["taskInstance"]
				}]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		result, err := s.GetTaskInstance(context.TODO(), GetTaskInstanceInput{
			TaskInstanceID: "test-instance-id",
		})

		assert.ErrorIs(t, err, ErrAccessDenied)
		assert.Nil(t, result)
	})
}
