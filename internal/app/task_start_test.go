package app

import (
	"context"
	"math"
	"strings"
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

	t.Run("returns expected result when input is provided", func(t *testing.T) {
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

		inputData := "test input data"
		input := StartTaskInput{
			DeployID: "test-deploy-id",
			TaskName: "worker",
			Input:    &inputData,
		}
		result, err := s.StartTask(context.TODO(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		doer.AssertCalled(t, "Do", mock.Anything)
	})

	t.Run("returns error when input exceeds size limit", func(t *testing.T) {
		maxRawDataSize := 3 * int(math.Ceil(float64(MaxTaskInputSize/4)))
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := New(c, nil, nil)

		largeInputData := strings.Repeat("a", maxRawDataSize+1)
		input := StartTaskInput{
			DeployID: "test-deploy-id",
			TaskName: "worker",
			Input:    &largeInputData,
		}
		result, err := s.StartTask(context.TODO(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrTaskInputTooLarge, err)
		assert.Nil(t, result)
	})

	t.Run("returns expected result when input is nil", func(t *testing.T) {
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
			Input:    nil,
		}
		result, err := s.StartTask(context.TODO(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("returns expected result when input is JSON", func(t *testing.T) {
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

		jsonInput := `{"user_id": 123, "action": "process"}`
		input := StartTaskInput{
			DeployID: "test-deploy-id",
			TaskName: "worker",
			Input:    &jsonInput,
		}
		result, err := s.StartTask(context.TODO(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
