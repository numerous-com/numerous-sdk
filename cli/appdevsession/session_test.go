package appdevsession

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"testing"
	"time"

	"numerous/cli/appdev"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockCommandExecutor struct{ mock.Mock }

func (e *mockCommandExecutor) Create(name string, args ...string) command {
	var callArgs []interface{}
	callArgs = append(callArgs, name)
	for _, a := range args {
		callArgs = append(callArgs, a)
	}

	mockArgs := e.Called(callArgs...)

	return mockArgs.Get(0).(command)
}

type mockCommand struct {
	mock.Mock
	name string
}

func (e *mockCommand) Kill() error {
	args := e.Called()
	return args.Error(0)
}

func (e *mockCommand) Start() error {
	args := e.Called()
	return args.Error(0)
}

func (e *mockCommand) SetEnv(key string, value string) {
	e.Called()
}

func (e *mockCommand) StdoutPipe() (io.ReadCloser, error) {
	args := e.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (e *mockCommand) StderrPipe() (io.ReadCloser, error) {
	args := e.Called()
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (e *mockCommand) Output() ([]byte, error) {
	args := e.Called()
	return args.Get(0).([]byte), args.Error(1)
}

type MockFileWatcherFactory struct{ mock.Mock }

func (f *MockFileWatcherFactory) Create() (FileWatcher, error) {
	args := f.Called()
	return args.Get(0).(FileWatcher), args.Error(1)
}

type mockFileWatcher struct{ mock.Mock }

func (w *mockFileWatcher) Close() error {
	args := w.Called()
	return args.Error(0)
}

func (w *mockFileWatcher) Add(name string) error {
	args := w.Called(name)
	return args.Error(0)
}

func (w *mockFileWatcher) GetEvents() chan fsnotify.Event {
	args := w.Called()
	return args.Get(0).(chan fsnotify.Event)
}

func (w *mockFileWatcher) GetErrors() chan error {
	args := w.Called()
	return args.Get(0).(chan error)
}

type mockClock struct{ mock.Mock }

func (c *mockClock) Now() time.Time {
	args := c.Called()
	return args.Get(0).(time.Time)
}

func (c *mockClock) Since(t time.Time) time.Duration {
	args := c.Called(t)
	return args.Get(0).(time.Duration)
}

func TestRunDevSession(t *testing.T) {
	t.Run("calls expected commands", func(t *testing.T) {
		appFile, err := os.CreateTemp("", "*-app.py")
		require.NoError(t, err)
		appFileName := appFile.Name()
		appClassName := "App"

		readCmd := mockCommand{}
		runCmd := mockCommand{}
		executor := mockCommandExecutor{}
		sessions := appdev.NewMockAppSessionRepository()
		fileWatcher := &mockFileWatcher{}
		fileWatcher.On("Add", mock.Anything).Return(nil)
		fileWatcher.On("Close").Return(nil)
		fileWatcher.On("GetEvents").Return(make(chan fsnotify.Event), nil)
		fileWatcher.On("GetErrors").Return(make(chan error), nil)

		fileWatcherFactory := &MockFileWatcherFactory{}
		fileWatcherFactory.On("Create").Return(fileWatcher, nil)
		session := devSession{
			executor:             &executor,
			fileWatcherFactory:   fileWatcherFactory,
			clock:                &timeclock{},
			appSessions:          sessions,
			appSessionService:    appdev.NewAppSessionService(sessions),
			port:                 "7001",
			appModulePath:        appFile.Name(),
			appClassName:         "App",
			pythonInterpeterPath: "python",
			exit:                 make(chan struct{}),
			output:               &appdev.FmtOutput{},
		}

		output := `{"app": {"title": "app", "elements": [{"name": "text", "type": "string", "label": "Text", "default": "default"}]}}`
		readCmd.On("Output").Return([]byte(output), nil)
		runCmd.On("SetEnv", mock.Anything).Return()
		runCmd.On("StdoutPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		runCmd.On("StderrPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		runCmd.On("Start").Return(nil)
		executor.On("Create", "python", "-m", "numerous", "read", appFileName, appClassName).Return(&readCmd, nil)
		executor.On("Create", "python", "-m", "numerous", "run", "--graphql-url", "http://localhost:7001/query", "--graphql-ws-url", "ws://localhost:7001/query", appFileName, appClassName, "0").Return(&runCmd, nil).Run(func(args mock.Arguments) {
			session.signalExit()
		})

		session.run()

		executor.AssertExpectations(t)
		readCmd.AssertExpectations(t)
		runCmd.AssertExpectations(t)
	})

	t.Run("updates session and restarts app when file is updated", func(t *testing.T) {
		appFile, err := os.CreateTemp("", "*-app.py")
		require.NoError(t, err)
		appFileName := appFile.Name()
		appClassName := "App"
		fileEvents := make(chan fsnotify.Event)

		initialReadCmd := mockCommand{name: "initial read"}
		initialRunCmd := mockCommand{name: "initial run"}
		updateReadCmd := mockCommand{name: "updated read"}
		updateRunCmd := mockCommand{name: "updated run"}
		executor := mockCommandExecutor{}
		sessions := appdev.NewMockAppSessionRepository()
		fileWatcher := &mockFileWatcher{}
		fileWatcher.On("Add", mock.Anything).Return(nil)
		fileWatcher.On("Close").Return(nil)
		fileWatcher.On("GetEvents").Return(fileEvents, nil)
		fileWatcher.On("GetErrors").Return(make(chan error), nil)
		mockClock := mockClock{}

		fileWatcherFactory := &MockFileWatcherFactory{}
		fileWatcherFactory.On("Create").Return(fileWatcher, nil)
		session := devSession{
			executor:             &executor,
			fileWatcherFactory:   fileWatcherFactory,
			clock:                &mockClock,
			appSessions:          sessions,
			appSessionService:    appdev.NewAppSessionService(sessions),
			port:                 "7001",
			appModulePath:        appFile.Name(),
			appClassName:         "App",
			pythonInterpeterPath: "python",
			exit:                 make(chan struct{}),
			minUpdateInterval:    time.Second,
			output:               &appdev.FmtOutput{},
		}

		initialDef := `
			{
				"app": {
					"title": "app",
					"elements": [{"name": "text", "type": "string", "label": "Text", "default": "default"}]
				}
			}`
		initialReadCmd.On("Output").Return([]byte(initialDef), nil)
		initialRunCmd.On("SetEnv", mock.Anything).Return()
		initialRunCmd.On("StdoutPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		initialRunCmd.On("StderrPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		initialRunCmd.On("Start").Return(nil)
		initialRunCmd.On("Kill").Return(nil)

		updatedDef := `
			{
				"app": {
					"title": "app",
					"elements": [{"name": "number", "type": "number", "label": "Number", "default": 12.34}]
				}
			}`
		updateReadCmd.On("Output").Return([]byte(updatedDef), nil)
		updateRunCmd.On("SetEnv", mock.Anything).Return()
		updateRunCmd.On("StdoutPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		updateRunCmd.On("StderrPipe", mock.Anything).Return(io.NopCloser(bytes.NewBuffer(nil)), nil)
		updateRunCmd.On("Start").Return(nil)

		mockClock.On("Now", mock.Anything).Return(time.Time{})
		mockClock.On("Since", mock.Anything).Return(2 * session.minUpdateInterval)

		executor.On("Create", "python", "-m", "numerous", "read", appFileName, appClassName).Return(&initialReadCmd, nil).Once()
		executor.On("Create", "python", "-m", "numerous", "read", appFileName, appClassName).Return(&updateReadCmd, nil).Once()
		executor.On("Create", "python", "-m", "numerous", "run", "--graphql-url", "http://localhost:7001/query", "--graphql-ws-url", "ws://localhost:7001/query", appFileName, appClassName, "0").
			Return(&initialRunCmd, nil).
			Once().
			Run(func(args mock.Arguments) {
				println("Creating a run command first time!")
				fileEvents <- fsnotify.Event{Name: appFileName, Op: fsnotify.Write}
			})
		executor.On("Create", "python", "-m", "numerous", "run", "--graphql-url", "http://localhost:7001/query", "--graphql-ws-url", "ws://localhost:7001/query", appFileName, appClassName, "0").
			Return(&updateRunCmd, nil).
			Once().
			Run(func(args mock.Arguments) {
				println("Creating a run command second time!")
				session.signalExit()
			})

		session.run()
		close(fileEvents)

		expectedElements := []appdev.AppSessionElement{
			{Model: gorm.Model{ID: 1}, Name: "number", Label: "Number", Type: "number", NumberValue: sql.NullFloat64{Valid: true, Float64: 12.34}, Elements: []appdev.AppSessionElement{}},
		}
		if appSession, err := sessions.Read(0); assert.NoError(t, err) {
			assert.Equal(t, expectedElements, appSession.Elements)
		}
		executor.AssertExpectations(t)
		initialReadCmd.AssertExpectations(t)
		initialRunCmd.AssertExpectations(t)
		updateReadCmd.AssertExpectations(t)
		updateRunCmd.AssertExpectations(t)
	})
}
