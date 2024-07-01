package manifest

import (
	"os"
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const streamlitTOML string = `name = "Tool Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = 80
cover_image = "cover.png"
exclude = ["*venv", "venv*"]

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`
const streamlitJSON string = `{"name":"Tool Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"],"deploy":{"organization":"organization-slug","app":"app-slug"}}`

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

var streamlitManifest = Manifest{
	Name:             "Tool Name",
	Description:      "A description",
	Library:          LibraryStreamlit,
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
	Deployment:       &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

const noDeployJSON string = `{"name":"Tool Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"]}`

const noDeployTOML string = `name = "Tool Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = 80
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`

var noDeployManifest = Manifest{
	Name:             "Tool Name",
	Description:      "A description",
	Library:          LibraryStreamlit,
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
}

func TestTOMLEncoding(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     Manifest
		expectedTOML string
	}{
		{
			name:         "streamlit app",
			manifest:     streamlitManifest,
			expectedTOML: streamlitTOML,
		},
		{
			name:         "without default deployment",
			manifest:     noDeployManifest,
			expectedTOML: noDeployTOML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.manifest.ToToml()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTOML, actual)
		})
	}
}

func TestJSONEncoding(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     Manifest
		expectedJSON string
	}{
		{
			name:         "streamlit app",
			manifest:     streamlitManifest,
			expectedJSON: streamlitJSON,
		},
		{
			name:         "without default deployment",
			manifest:     noDeployManifest,
			expectedJSON: noDeployJSON,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.manifest.ToJSON()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedJSON, actual)
		})
	}
}

func TestManifestDecodeTOML(t *testing.T) {
	testCases := []struct {
		name        string
		tomlContent string
		expected    Manifest
	}{
		{
			name:        "streamlit with deprecated string port",
			tomlContent: deprecatedTOML,
			expected:    noDeployManifest,
		},
		{
			name:        "streamlit",
			tomlContent: streamlitTOML,
			expected:    streamlitManifest,
		},
		{
			name:        "without default deployment",
			tomlContent: noDeployTOML,
			expected:    noDeployManifest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save Manifest
			filePath := test.WriteTempFile(t, ManifestFileName, []byte(tc.tomlContent))

			defer os.Remove(filePath)

			// Decode file
			actual, err := LoadManifest(filePath)
			require.NoError(t, err)

			if assert.NotNil(t, actual) {
				assert.Equal(t, tc.expected, *actual)
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
		expected       error
	}{
		{
			name:           "numerous app with appdef definition succeeds",
			library:        "numerous",
			appfileContent: numerousAppContentAppdefDefinition,
			expected:       nil,
		},
		{
			name:           "numerous app with appdef assignment succeeds",
			library:        "numerous",
			appfileContent: numerousAppContentAppdefAssignment,
			expected:       nil,
		},
		{
			name:           "numerous app without appdef fails",
			library:        "numerous",
			appfileContent: numerousAppContentWithoutAppdef,
			expected:       ErrValidateNumerousApp,
		},
		{
			name:           "non-numerous app without appdef succeeds",
			library:        "streamlit",
			appfileContent: `the_content = "does not matter here"`,
			expected:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appfile := test.WriteTempFile(t, "appfile.py", []byte(tc.appfileContent))
			l, err := GetLibraryByKey(tc.library)
			require.NoError(t, err)

			m := Manifest{Library: l, AppFile: appfile}
			err = m.ValidateApp()

			assert.ErrorIs(t, err, tc.expected)
		})
	}
}
