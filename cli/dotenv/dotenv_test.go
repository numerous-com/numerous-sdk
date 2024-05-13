package dotenv

import (
	"os"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

type loadSecretsFromEnvTestCase struct {
	name     string
	content  string
	expected map[string]string
}

func TestLoad(t *testing.T) {
	testCases := []loadSecretsFromEnvTestCase{
		{
			name:    "reads all key value pairs",
			content: "VAR_1=var 1 value\nVAR_2=var 2 value\nVAR_3=var 3 value",
			expected: map[string]string{
				"VAR_1": "var 1 value",
				"VAR_2": "var 2 value",
				"VAR_3": "var 3 value",
			},
		},
		{
			name:    "ignores empty lines",
			content: "VAR_1=var 1 value\n\n\nVAR_2=var 2 value",
			expected: map[string]string{
				"VAR_1": "var 1 value",
				"VAR_2": "var 2 value",
			},
		},
		{
			name:    "ignores content on line after '#' comment symbol",
			content: "VAR_1=var 1 value# my first comment\n#VAR_2=var 2 value\nVAR_3=#var 3 value",
			expected: map[string]string{
				"VAR_1": "var 1 value",
				"VAR_3": "",
			},
		},
		{
			name:    "trims all whitespace",
			content: "   VAR_1=var 1 value\nVAR_2   =var 2 value\nVAR_3=    var 3 value\nVAR_4=var 4 value    # some comment",
			expected: map[string]string{
				"VAR_1": "var 1 value",
				"VAR_2": "var 2 value",
				"VAR_3": "var 3 value",
				"VAR_4": "var 4 value",
			},
		},
		{
			name:    "trims matching quotes",
			content: "VAR_1=\"var 1 value\"\nVAR_2 = \"var 2 value   \"\nVAR_3='var 3 value'\nVAR_4='var 4 value    # some comment'",
			expected: map[string]string{
				"VAR_1": "var 1 value",
				"VAR_2": "var 2 value   ",
				"VAR_3": "var 3 value",
				"VAR_4": "'var 4 value",
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path := test.WriteTempFile(t, ".env", []byte(testCase.content))

			env, err := Load(path)

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, env)
		})
	}

	t.Run("returns error reading non-existing file", func(t *testing.T) {
		env, err := Load("some/non-existing/file")
		assert.Nil(t, env)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
