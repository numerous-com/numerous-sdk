package manifest

import (
	"os"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const manifestTOML string = `name = "Tool Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = 80
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`

const deprecatedTOML string = `name = "Tool Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = "80"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`
const manifestJSON string = `{"name":"Tool Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"]}`

type encodeTOMLTestCase struct {
	name         string
	manifest     Manifest
	expectedTOML string
}

var streamlitManifest = Manifest{
	Name:             "Tool Name",
	Description:      "A description",
	Library:          "streamlit",
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
}

var encodeTOMLTestCases = []encodeTOMLTestCase{
	{
		name:         "streamlit app",
		manifest:     streamlitManifest,
		expectedTOML: manifestTOML,
	},
}

func TestTOMLEncoding(t *testing.T) {
	for _, testcase := range encodeTOMLTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			actual, err := testcase.manifest.ToToml()
			require.NoError(t, err)
			assert.Equal(t, testcase.expectedTOML, actual)
		})
	}
}

type encodeJSONTestCase struct {
	name         string
	manifest     Manifest
	expectedJSON string
}

var encodeJSONTestCases = []encodeJSONTestCase{
	{
		name:         "streamlit app",
		manifest:     streamlitManifest,
		expectedJSON: manifestJSON,
	},
}

func TestJSONEncoding(t *testing.T) {
	for _, testcase := range encodeJSONTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			actual, err := testcase.manifest.ToJSON()
			require.NoError(t, err)
			assert.Equal(t, testcase.expectedJSON, actual)
		})
	}
}

type decodeTOMLTestCase struct {
	name        string
	tomlContent string
	expected    Manifest
}

var decodeTOMLTestCases = []decodeTOMLTestCase{
	{
		name:        "streamlit with deprecated string port",
		tomlContent: deprecatedTOML,
		expected:    streamlitManifest,
	},
	{
		name:        "streamlit",
		tomlContent: manifestTOML,
		expected:    streamlitManifest,
	},
}

func TestManifestDecodeTOML(t *testing.T) {
	for _, testcase := range decodeTOMLTestCases {
		t.Run(testcase.name, func(t *testing.T) {
			// Save Manifest
			filePath := test.WriteTempFile(t, ManifestFileName, []byte(testcase.tomlContent))

			defer os.Remove(filePath)

			// Decode file
			actual, err := LoadManifest(filePath)
			require.NoError(t, err)

			if assert.NotNil(t, actual) {
				assert.Equal(t, testcase.expected, *actual)
			}
		})
	}
}

var numerousAppContentAppdefAssignment string = `
from numerous import app

@app
class MyApp:
	field: str

appdef = MyApp
`

var numerousAppContentAppdefDefinition string = `
from numerous import app

@app
class appdef:
	field: str
`

var numerousAppContentWithoutAppdef string = `
from numerous import app

@app
class MyApp:
	field: str
`

func TestManifestValidateApp(t *testing.T) {
	testCases := []struct {
		name           string
		library        string
		appfileContent string
		expected       bool
	}{
		{
			name:           "numerous app with appdef definition succeeds",
			library:        "numerous",
			appfileContent: numerousAppContentAppdefDefinition,
			expected:       true,
		},
		{
			name:           "numerous app with appdef assignment succeeds",
			library:        "numerous",
			appfileContent: numerousAppContentAppdefAssignment,
			expected:       true,
		},
		{
			name:           "numerous app without appdef fails",
			library:        "numerous",
			appfileContent: numerousAppContentWithoutAppdef,
			expected:       false,
		},
		{
			name:           "non-numerous app without appdef succeeds",
			library:        "streamlit",
			appfileContent: `the_content = "does not matter here"`,
			expected:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			appfile := test.WriteTempFile(t, "appfile.py", []byte(testCase.appfileContent))
			m := Manifest{Library: testCase.library, AppFile: appfile}
			validated, err := m.ValidateApp()
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, validated)
		})
	}
}
