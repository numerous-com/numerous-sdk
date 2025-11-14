package get

import (
	"context"
	"errors"
	"testing"
	"time"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetInstance(t *testing.T) {
	const (
		taskInstanceID = "test-instance-id"
	)
	testError := errors.New("test error")

	t.Run("calls service with expected parameters", func(t *testing.T) {
		service := &TaskGetServiceMock{}

		cpuUsage := &app.WorkloadResourceUsage{Current: 1.5}
		memoryUsage := &app.WorkloadResourceUsage{Current: 512.0}
		exitCode := 0
		inputStr := "test input"
		outputStr := "test output"

		expectedResult := &app.TaskInstance{
			ID:        taskInstanceID,
			CreatedAt: time.Now(),
			Task: app.Task{
				ID:      "task-id",
				Command: []string{"python", "main.py"},
			},
			Workload: app.TaskInstanceWorkload{
				Status:        "running",
				StartedAt:     time.Now(),
				CPUUsage:      cpuUsage,
				MemoryUsageMB: memoryUsage,
				ExitCode:      &exitCode,
				Input:         &inputStr,
				Output:        &outputStr,
			},
		}

		service.On("GetTaskInstance", mock.Anything, app.GetTaskInstanceInput{
			TaskInstanceID: taskInstanceID,
		}).Return(expectedResult, nil)

		input := TaskGetInput{
			TaskInstanceID: taskInstanceID,
		}
		err := getInstance(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error if GetTaskInstance fails", func(t *testing.T) {
		service := &TaskGetServiceMock{}

		service.On("GetTaskInstance", mock.Anything, app.GetTaskInstanceInput{
			TaskInstanceID: taskInstanceID,
		}).Return(nil, testError)

		input := TaskGetInput{
			TaskInstanceID: taskInstanceID,
		}
		err := getInstance(context.TODO(), service, input)

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})
}
