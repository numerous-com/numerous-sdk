package push

import (
	"bytes"
	"path"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

var verbosePrintTestCases = []struct {
	name     string
	expected string
	verbose  bool
}{
	{name: "Verbose flag is false", expected: "     Build: Building app.......\n     Build: ---> Using cache\n     Build: ---> 5dee4a5dd2e0\n     Build: Step 4/7 : RUN pip install -r /app/requirements.txt\n     Build: \n     Build: Step 1/7 : FROM python:3.11-slim\n     Build: \n     Build: ---> f8e98f0336d5\n     Build: Step 2/7 : EXPOSE 80\n     Build: \n     Build: ---> Using cache\n     Build: ---> 70cbb8104f4d\n     Build: Step 3/7 : COPY ./requirements.txt /app/requirements.txt\n     Build: \n     Build: ---> Using cache\n     Build: ---> dc49b771620c\n     Build: Step 5/7 : COPY . /app\n     Build: \n     Build: ---> Using cache\n     Build: ---> 5f020e3f851a\n     Build: Step 6/7 : WORKDIR /app\n     Build: \n     Build: ---> Using cache\n     Build: ---> 36a177af492a\n     Build: Step 7/7 : CMD [\"streamlit\", \"run\", \"app.py\", \"--server.port\", \"80\"]\n     Build: \n     Build: ---> Using cache\n     Build: ---> f5b4a0f5108b\n     Build: Successfully built f5b4a0f5108b\n     Build: Successfully tagged 00a66264-651b-43ec-b46d-a176a653657c:latest\n", verbose: true},
	{name: "Verbose flag is true", expected: "", verbose: false},
}

func TestPrintVerbose(t *testing.T) {
	for _, test := range verbosePrintTestCases {
		t.Run(test.name, func(t *testing.T) {
			buildEventMessages := []string{
				"Building app.......",
				"---> Using cache",
				"---> 5dee4a5dd2e0",
				"Step 4/7 : RUN pip install -r /app/requirements.txt",
				"",
				"Step 1/7 : FROM python:3.11-slim",
				"",
				"---> f8e98f0336d5",
				"Step 2/7 : EXPOSE 80",
				"",
				"---> Using cache",
				"---> 70cbb8104f4d",
				"Step 3/7 : COPY ./requirements.txt /app/requirements.txt",
				"",
				"---> Using cache",
				"---> dc49b771620c",
				"Step 5/7 : COPY . /app",
				"",
				"---> Using cache",
				"---> 5f020e3f851a",
				"Step 6/7 : WORKDIR /app",
				"",
				"---> Using cache",
				"---> 36a177af492a",
				"Step 7/7 : CMD [\"streamlit\", \"run\", \"app.py\", \"--server.port\", \"80\"]",
				"",
				"---> Using cache",
				"---> f5b4a0f5108b",
				"Successfully built f5b4a0f5108b",
				"Successfully tagged 00a66264-651b-43ec-b46d-a176a653657c:latest",
			}

			out := new(bytes.Buffer)

			for _, elem := range buildEventMessages {
				printVerbose(out, elem, test.verbose)
			}
			outContent := (*bytes.Buffer).String(out)

			assert.Equal(t, test.expected, outContent)
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
			test.WriteFile(t, path.Join(appDir, ".env"), []byte(testCase.content))

			actual := loadSecretsFromEnv(appDir)

			assert.Equal(t, testCase.expected, actual)
		})
	}
}
