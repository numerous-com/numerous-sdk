package version

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/version"
)

func TestCheck(t *testing.T) {
	testErr := errors.New("test error")

	t.Run("returns true on error", func(t *testing.T) {
		checker := MockVersionChecker{}
		checker.On("Check", mock.Anything).Return(version.CheckVersionOutput{}, testErr)

		result := Check(&checker)

		assert.True(t, result)
		checker.AssertExpectations(t)
	})

	t.Run("returns", func(t *testing.T) {
		testCases := []struct {
			output   version.CheckVersionOutput
			expected bool
		}{
			{
				output:   version.CheckVersionOutput{Result: version.VersionCheckResultOK, Message: "Version is actual"},
				expected: true,
			},
			{
				output:   version.CheckVersionOutput{Result: version.VersionCheckResultWarning, Message: "Test warning"},
				expected: true,
			},
			{
				output:   version.CheckVersionOutput{Result: version.VersionCheckResultCritical, Message: "Test critical"},
				expected: false,
			},
			{
				output:   version.CheckVersionOutput{Result: version.VersionCheckResultUnknown, Message: "Unknown version 'beta'"},
				expected: true,
			},
		}
		for _, testCase := range testCases {
			t.Run(string(fmt.Sprintf("%v if %v", testCase.expected, testCase.output.Result)), func(t *testing.T) {
				checker := MockVersionChecker{}
				checker.On("Check", mock.Anything).Return(testCase.output, nil)

				result := Check(&checker)

				assert.Equal(t, testCase.expected, result)
				checker.AssertExpectations(t)
			})
		}
	})
}
