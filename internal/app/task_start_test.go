package app

import (
	"context"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartTask(t *testing.T) {
	t.Run("returns expected result", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"data": {
					"taskStart": {
						"id": "test-instance-id",
						"task": {
							"id": "test-task-id",
							"command": ["python", "worker.py"]
						}
					}
				}
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := StartTaskInput{
			DeployID: "test-deploy-id",
			TaskName: "worker",
		}
		result, err := s.StartTask(context.TODO(), input)

		expected := &TaskStartResult{
			TaskInstanceID: "test-instance-id",
			TaskID:         "test-task-id",
			Command:        []string{"python", "worker.py"},
		}
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns error if mutation fails", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		respBody := `
			{
				"errors": [
					{
						"message": "Task not found",
						"extensions": {
							"code": "TASK_NOT_FOUND"
						}
					}
				]
			}
		`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		input := StartTaskInput{
			DeployID: "test-deploy-id",
			TaskName: "nonexistent-task",
		}
		result, err := s.StartTask(context.TODO(), input)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
