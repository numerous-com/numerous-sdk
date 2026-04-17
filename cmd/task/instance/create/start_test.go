package create

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

	newTestInput := func() TaskStartInput {
		return TaskStartInput{
			AppDir:           "",
			OrganizationSlug: organizationSlug,
			AppSlug:          appSlug,
			TaskName:         taskName,
		}
	}

	newExpectedTestInput := func() app.StartTaskInput {
		return app.StartTaskInput{
			OrganizationSlug: organizationSlug,
			DeployID:         deployID,
			TaskName:         taskName,
		}
	}

	newSuccessResult := func() *app.TaskStartResult {
		return &app.TaskStartResult{
			TaskInstanceID: taskInstanceID,
			TaskID:         taskID,
			Command:        []string{"python", "worker.py"},
		}
	}

	setupDeployID := func(service *TaskStartServiceMock) {
		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return(deployID, nil)
	}

	t.Run("calls service with expected parameters", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)
		service.On("StartTask", mock.Anything, newExpectedTestInput()).Return(newSuccessResult(), nil)

		err := startTask(context.TODO(), service, newTestInput())

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error if GetAppDeploymentID fails", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		service.On("GetAppDeploymentID", mock.Anything, organizationSlug, appSlug).Return("", testError)

		err := startTask(context.TODO(), service, newTestInput())

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("returns error if StartTask fails", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)
		var nilResult *app.TaskStartResult = nil
		service.On("StartTask", mock.Anything, newExpectedTestInput()).Return(nilResult, testError)

		err := startTask(context.TODO(), service, newTestInput())

		assert.ErrorIs(t, err, testError)
		service.AssertExpectations(t)
	})

	t.Run("returns no error if successful task start", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)
		service.On("StartTask", mock.Anything, newExpectedTestInput()).Return(newSuccessResult(), nil)

		err := startTask(context.TODO(), service, newTestInput())

		assert.NoError(t, err)
	})

	t.Run("returns no error if input is provided", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)

		testInput := "test input data"
		expectedInput := newExpectedTestInput()
		expectedInput.Input = &testInput
		service.On("StartTask", mock.Anything, expectedInput).Return(newSuccessResult(), nil)

		input := newTestInput()
		input.Input = testInput
		err := startTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("reads input from file when input file is provided", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.txt")
		fileContent := "file input data"
		err := os.WriteFile(inputFile, []byte(fileContent), 0o644)
		assert.NoError(t, err)

		expectedInput := newExpectedTestInput()
		expectedInput.Input = &fileContent
		service.On("StartTask", mock.Anything, expectedInput).Return(newSuccessResult(), nil)

		input := newTestInput()
		input.InputFile = inputFile
		err = startTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})

	t.Run("returns error when both input and input file are provided", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		input := newTestInput()
		input.Input = "direct input"
		input.InputFile = "/path/to/file"
		err := startTask(context.TODO(), service, input)

		assert.ErrorIs(t, err, ErrConflictingInputFlags)
	})

	t.Run("returns error when input file does not exist", func(t *testing.T) {
		service := &TaskStartServiceMock{}

		input := newTestInput()
		input.InputFile = "/nonexistent/file.txt"
		err := startTask(context.TODO(), service, input)

		assert.Error(t, err)
	})

	t.Run("returns no error if input file contains JSON", func(t *testing.T) {
		service := &TaskStartServiceMock{}
		setupDeployID(service)

		tmpDir := t.TempDir()
		inputFile := filepath.Join(tmpDir, "input.json")
		jsonContent := `{"user_id": 123, "action": "process"}`
		err := os.WriteFile(inputFile, []byte(jsonContent), 0o644)
		assert.NoError(t, err)

		expectedInput := newExpectedTestInput()
		expectedInput.Input = &jsonContent
		service.On("StartTask", mock.Anything, expectedInput).Return(newSuccessResult(), nil)

		input := newTestInput()
		input.InputFile = inputFile
		err = startTask(context.TODO(), service, input)

		assert.NoError(t, err)
		service.AssertExpectations(t)
	})
}
