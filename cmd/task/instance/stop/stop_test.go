package stop

import (
	"context"
	"errors"
	"testing"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStopTask(t *testing.T) {
	const (
		taskInstanceID = "test-instance-id"
	)
	testError := errors.New("test error")

	newTestStopInput := func() TaskStopInput {
		return TaskStopInput{TaskInstanceID: taskInstanceID}
	}

	newTestStopResult := func() *app.TaskStopResult {
		return &app.TaskStopResult{TaskInstanceID: taskInstanceID}
	}

	newTestService := func(instanceID string) *TaskStopServiceMock {
		service := &TaskStopServiceMock{}
		service.On("StopTask", mock.Anything, instanceID).Return(newTestStopResult(), nil)

		return service
	}

	t.Run("calls service with expected parameters and return no error on successful stop", func(t *testing.T) {
		service := newTestService(taskInstanceID)

		err := stopTask(context.TODO(), service, newTestStopInput())

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error if StopTask fails", func(t *testing.T) {
		service := &TaskStopServiceMock{}
		service.On("StopTask", mock.Anything, taskInstanceID).Return(nil, testError)

		err := stopTask(context.TODO(), service, newTestStopInput())

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("handles different task instance IDs", func(t *testing.T) {
		differentInstanceID := "ce5aba38-842d-4ee0-877b-4af9d426c848"
		service := newTestService(differentInstanceID)

		err := stopTask(context.TODO(), service, TaskStopInput{TaskInstanceID: differentInstanceID})

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}
