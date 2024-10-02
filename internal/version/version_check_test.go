package version

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"numerous.com/cli/internal/test"
)

func TestCheck(t *testing.T) {
	t.Run("given response error then it returns empty result and non-empty error", func(t *testing.T) {
		doer := test.MockDoer{}
		c := test.CreateTestGQLClient(t, &doer)
		s := NewService(c)
		respBody := `
		{
			"errors": [{
				"message": "internal error",
				"location": [{"line": 1, "column": 1}],
				"path": ["checkVersion"]
			}]
		}`
		resp := test.JSONResponse(respBody)
		doer.On("Do", mock.Anything).Return(resp, nil)

		actual, err := s.Check(context.TODO())

		assert.Empty(t, actual)
		assert.Error(t, err, "internal error")
	})

	t.Run("returns expected result with no errors", func(t *testing.T) {
		testCases := []struct {
			name         string
			responseBody string
			expected     CheckVersionOutput
		}{
			{
				name: "if version check is OK",
				responseBody: `
				{
					"data": {
						"checkVersion": {
							"__typename": "VersionCheckOK",
							"version": "0.20.1"
						}
					}
				}`,
				expected: CheckVersionOutput{Result: VersionCheckResultOK, Message: "Version is actual"},
			},
			{
				name: "if version check warning",
				responseBody: `
				{
					"data": {
						"checkVersion": {
							"__typename": "VersionCheckWarning",
							"message": "Test warning"
						}
					}
				}`,
				expected: CheckVersionOutput{Result: VersionCheckResultWarning, Message: "Test warning"},
			},
			{
				name: "if version check critical",
				responseBody: `
				{
					"data": {
						"checkVersion": {
							"__typename": "VersionCheckCritical",
							"message": "Test critical"
						}
					}
				}`,
				expected: CheckVersionOutput{Result: VersionCheckResultCritical, Message: "Test critical"},
			},
			{
				name: "if version is unknown",
				responseBody: `
				{
					"data": {
						"checkVersion": {
							"__typename": "VersionUnknown",
							"version": "beta"
						}
					}
				}`,
				expected: CheckVersionOutput{Result: VersionCheckResultUnknown, Message: "Unknown version 'beta'"},
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				doer := test.MockDoer{}
				c := test.CreateTestGQLClient(t, &doer)
				s := NewService(c)
				resp := test.JSONResponse(testCase.responseBody)
				doer.On("Do", mock.Anything).Return(resp, nil)

				actual, err := s.Check(context.TODO())

				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, actual)
			})
		}
	})
}
