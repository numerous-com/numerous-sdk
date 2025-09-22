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
		organizationSlug = "test-org"
		appSlug          = "test-app"
		taskInstanceID   = "test-instance-id"
	)
	testError := errors.New("test error")

	t.Run("calls service with expected parameters", func(t *testing.T) {
		service := &TaskStopServiceMock{}

		expectedResult := &app.TaskStopResult{
			TaskInstanceID: taskInstanceID,
		}
		service.On("StopTask", mock.Anything, taskInstanceID).Return(expectedResult, nil)

		input := TaskStopInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskInstanceID:   taskInstanceID,
		}
		err := stopTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error if StopTask fails", func(t *testing.T) {
		service := &TaskStopServiceMock{}

		var nilResult *app.TaskStopResult = nil
		service.On("StopTask", mock.Anything, taskInstanceID).Return(nilResult, testError)

		input := TaskStopInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskInstanceID:   taskInstanceID,
		}
		err := stopTask(context.TODO(), service, input)

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("returns no error if successful task stop", func(t *testing.T) {
		service := &TaskStopServiceMock{}

		expectedResult := &app.TaskStopResult{
			TaskInstanceID: taskInstanceID,
		}
		service.On("StopTask", mock.Anything, taskInstanceID).Return(expectedResult, nil)

		input := TaskStopInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskInstanceID:   taskInstanceID,
		}
		err := stopTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("handles different task instance IDs", func(t *testing.T) {
		service := &TaskStopServiceMock{}

		differentInstanceID := "ce5aba38-842d-4ee0-877b-4af9d426c848"
		expectedResult := &app.TaskStopResult{
			TaskInstanceID: differentInstanceID,
		}
		service.On("StopTask", mock.Anything, differentInstanceID).Return(expectedResult, nil)

		input := TaskStopInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskInstanceID:   differentInstanceID,
		}
		err := stopTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}
