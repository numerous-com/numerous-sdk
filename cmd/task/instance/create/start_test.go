package create

import (
	"context"
	"errors"
	"testing"

	"numerous.com/cli/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartTask(t *testing.T) {
	const (
		organizationSlug = "test-org"
		appSlug          = "test-app"
		taskName         = "test-task"
		deployID         = "test-deploy-id"
		taskInstanceID   = "test-instance-id"
		taskID           = "test-task-id"
	)
	testError := errors.New("test error")

	t.Run("calls service with expected parameters", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return(deployID, nil)

		expectedStartInput := app.StartTaskInput{
			DeployID: deployID,
			TaskName: taskName,
		}
		expectedResult := &app.TaskStartResult{
			TaskInstanceID: taskInstanceID,
			TaskID:         taskID,
			Command:        []string{"echo", "hello"},
		}
		service.On("StartTask", mock.Anything, expectedStartInput).Return(expectedResult, nil)

		input := TaskStartInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskName:         taskName,
		}
		err := startTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error if GetAppDeploymentID fails", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return("", testError)

		input := TaskStartInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskName:         taskName,
		}
		err := startTask(context.TODO(), service, input)

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("returns error if StartTask fails", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return(deployID, nil)

		expectedStartInput := app.StartTaskInput{
			DeployID: deployID,
			TaskName: taskName,
		}
		var nilResult *app.TaskStartResult = nil
		service.On("StartTask", mock.Anything, expectedStartInput).Return(nilResult, testError)

		input := TaskStartInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskName:         taskName,
		}
		err := startTask(context.TODO(), service, input)

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("returns no error if successful task start", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return(deployID, nil)

		expectedStartInput := app.StartTaskInput{
			DeployID: deployID,
			TaskName: taskName,
		}
		expectedResult := &app.TaskStartResult{
			TaskInstanceID: taskInstanceID,
			TaskID:         taskID,
			Command:        []string{"python", "worker.py"},
		}
		service.On("StartTask", mock.Anything, expectedStartInput).Return(expectedResult, nil)

		input := TaskStartInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskName:         taskName,
		}
		err := startTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}
