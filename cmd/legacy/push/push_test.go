package push

import (
	"path/filepath"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

var verbosePrintTestCases = []struct {
	name     string
	expected []string
	verbose  bool
}{
	{
		name: "Verbose flag is false", expected: []string{"some message 1", "some message 2", "some message 3"},
		verbose: true,
	},
	{name: "Verbose flag is true", expected: nil, verbose: false},
}

type recordingWriter struct {
	writes []string
}

// Write implements io.Writer.
func (r *recordingWriter) Write(p []byte) (n int, err error) {
	r.writes = append(r.writes, string(p))
	return len(p), nil
}

func TestPrintVerbose(t *testing.T) {
	for _, test := range verbosePrintTestCases {
		t.Run(test.name, func(t *testing.T) {
			buildEventMessages := []string{"some message 1", "some message 2", "some message 3"}

			w := &recordingWriter{}
			for _, elem := range buildEventMessages {
				printVerbose(w, elem, test.verbose)
			}

			assert.Equal(t, test.expected, w.writes)
		})
	}
}

type loadSecretsFromEnvTestCase struct {
	name     string
	content  string
	expected map[string]string
}

func TestLoadSecretsFromEnv(t *testing.T) {
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
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			appDir := t.TempDir()
			test.WriteFile(t, filepath.Join(appDir, ".env"), []byte(testCase.content))

			actual := loadSecretsFromEnv(appDir)

			assert.Equal(t, testCase.expected, actual)
		})
	}
}
