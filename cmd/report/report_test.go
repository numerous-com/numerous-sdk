package report

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOpenURL(t *testing.T) {
	testURL := "numerous.test.com"

	testCases := []struct {
		name            string
		goos            string
		expectedCommand string
		expectedArgs    []string
		output          []byte
		errorContent    error
	}{
		{"opens URL - windows", "windows", "cmd", []string{"/c", "start"}, nil, nil},
		{"opens URL - Darwin", "darwin", "open", nil, nil, nil},
		{"opens URL - Freebsd", "freebsd", "xdg-open", nil, nil, nil},
		{"opens URL - WSL", "linux", "sensible-browser", nil, []byte("it will return a line that contains the windows word"), nil},
		{"opens URL - Linux", "linux", "xdg-open", nil, []byte(nil), errors.New("exit status 1")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec := new(FakeCommandExecutor)
			exec.On("Output", "sh", "-c", "grep -i Windows /proc/version").Return(tc.output, tc.errorContent)
			mockArgs := aggregateArgs(tc.expectedCommand, &testURL, tc.expectedArgs...)
			exec.On("Start", mockArgs...).Return(nil)
			_ = openURL(testURL, tc.goos, exec)
			if tc.goos == "linux" {
				exec.AssertNumberOfCalls(t, "Output", 1)
			}
			exec.AssertNumberOfCalls(t, "Start", 1)
		})
	}

	testErrorCases := []struct {
		name            string
		goos            string
		expectedCommand string
		errorContent    error
		errorType       error
	}{
		{"tests error checking the OS environment with GOOS returns OS check error", "unknown", "", nil, ErrOSCheck},
		{"tests error propagation with a grep error returns WSL check error", "linux", "", nil, ErrWSLCheck},
	}

	for _, tc := range testErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			exec := new(FakeCommandExecutor)
			exec.On("Output", "sh", "-c", "grep -i Windows /proc/version").Return([]byte(nil), ErrGrepBashCommand)
			mockArgs := aggregateArgs(tc.expectedCommand, &testURL)
			exec.On("Start", mockArgs...).Return(tc.errorType)
			err := openURL(testURL, tc.goos, exec)
			if tc.goos == "linux" {
				exec.AssertNumberOfCalls(t, "Output", 1)
			} else {
				exec.AssertNumberOfCalls(t, "Output", 0)
			}
			exec.AssertNumberOfCalls(t, "Start", 0)
			assert.ErrorIs(t, err, tc.errorType)
		})
	}
}

func TestCmdByOS(t *testing.T) {
	testCases := []struct {
		name     string
		goos     string
		wsl      string
		expected string
		args     []string
	}{
		{"returns command for Windows environment plus extra arguments", "windows", "", "cmd", []string{"/c", "start"}},
		{"returns the right command for a Darwin environment", "darwin", "", "open", nil},
		{"returns the right command for a Freebsd environment", "freebsd", "", "xdg-open", nil},
		{"returns the right command for a Linux environment", "linux", "os", "xdg-open", nil},
		{"returns the right command for a WSL environment", "linux", "wsl", "sensible-browser", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isWSLfake := func(exec CommandExecutor) (string, error) {
				return tc.wsl, nil
			}
			cmd, args, err := cmdByOS(tc.goos, nil, isWSLfake)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, cmd)
			assert.Equal(t, tc.args, args)
		})
	}

	testErrorCases := []struct {
		name        string
		goos        string
		wslErr      bool
		errExpected error
	}{
		{"returns OS check error", "unknown", false, ErrOSCheck},
		{"returns WSL check error", "linux", true, ErrWSLCheck},
	}

	for _, tc := range testErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			isWSLfake := func(exec CommandExecutor) (string, error) {
				var err error = nil
				if tc.wslErr {
					err = errors.New("any error from isWSL function call")
				}

				return "", err
			}
			cmd, args, err := cmdByOS(tc.goos, nil, isWSLfake)
			assert.Equal(t, "", cmd)
			assert.Nil(t, args)
			assert.ErrorIs(t, err, tc.errExpected)
		})
	}
}

func TestOSOrWSL(t *testing.T) {
	testCases := []struct {
		name           string
		output         []byte
		errorOut       error
		resultExpected string
		errExpected    error
	}{
		{"wsl environment identification", []byte("it will return a line that contains the windows word"), nil, "wsl", nil},
		{"linux distro smooth - grep exit status 0", []byte(nil), nil, "os", nil},
		{"linux distro with error - grep exit status 1", []byte(nil), errors.New("exit status 1"), "os", nil},
		{"grep any other error", []byte(nil), errors.New("any other error"), "", ErrGrepBashCommand},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exec := new(FakeCommandExecutor)
			exec.On("Output", "sh", "-c", "grep -i Windows /proc/version").Return(tc.output, tc.errorOut)
			result, err := osOrWSL(exec)
			if tc.errExpected == nil {
				assert.Equal(t, tc.resultExpected, result)
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tc.errExpected)
			}
		})
	}
}

type FakeCommandExecutor struct {
	mock.Mock
}

func (exec *FakeCommandExecutor) Output(command string, args ...string) ([]byte, error) {
	var err error
	allArgs := aggregateArgs(command, nil, args...)
	mockArgs := exec.Called(allArgs...)
	result := mockArgs.Get(0).([]byte)
	if mockArgs.Get(1) == nil {
		err = nil
	} else {
		err = mockArgs.Get(1).(error)
	}

	return result, err
}

func (exec *FakeCommandExecutor) Start(command string, args ...string) error {
	var err error
	allArgs := aggregateArgs(command, nil, args...)
	mockArgs := exec.Called(allArgs...)
	if mockArgs.Get(0) == nil {
		err = nil
	} else {
		err = mockArgs.Get(0).(error)
	}

	return err
}

func aggregateArgs(functionMocked string, url *string, args ...string) []interface{} {
	var argsPack []interface{}
	argsPack = append(argsPack, functionMocked)
	for _, a := range args {
		argsPack = append(argsPack, a)
	}
	if url != nil {
		argsPack = append(argsPack, *url)
	}

	return argsPack
}
