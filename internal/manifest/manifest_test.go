package manifest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

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

const tomlStreamlit string = `name = "Streamlit App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"
  port = 80

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`

const jsonStreamlit string = `{"name":"Streamlit App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"python":{"library":"streamlit","version":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80},"deploy":{"organization":"organization-slug","app":"app-slug"}}`

const jsonStreamlitNoDeploy string = `{"name":"Streamlit App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"python":{"library":"streamlit","version":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80}}`

const tomlStreamlitNoDeploy string = `name = "Streamlit App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"
  port = 80
`

var manifestStreamlit Manifest = Manifest{
	ManifestApp: ManifestApp{
		Name:        "Streamlit App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
	},
	Python: &ManifestPython{
		Library:          LibraryStreamlit,
		Version:          "3.11",
		Port:             80,
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
	},
	Deployment: &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

var manifestStreamlitNoDeploy Manifest = Manifest{
	ManifestApp: ManifestApp{
		Name:        "Streamlit App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
	},
	Python: &ManifestPython{
		Library:          LibraryStreamlit,
		Version:          "3.11",
		Port:             80,
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
	},
	Deployment: nil,
}

func TestValidateApp(t *testing.T) {
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

			m := Manifest{Python: &ManifestPython{Library: l, AppFile: appfile}}

			err = m.ValidateApp()

			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestLoad(t *testing.T) {
	t.Run("returns expected manifest", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			tomlContent string
			expected    Manifest
		}{
			{
				name:        "v0 streamlit with deprecated string port",
				tomlContent: v0TOMLStreamlit,
				expected:    manifestStreamlitNoDeploy,
			},
			{
				name:        "v1 streamlit",
				tomlContent: v1TOMLStreamlit,
				expected:    manifestStreamlit,
			},
			{
				name:        "v1 streamlit without default deployment",
				tomlContent: v1TOMLStreamlitNoDeploy,
				expected:    manifestStreamlitNoDeploy,
			},
			{
				name:        "streamlit with python section",
				tomlContent: tomlStreamlit,
				expected:    manifestStreamlit,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				filePath := test.WriteTempFile(t, ManifestFileName, []byte(tc.tomlContent))
				defer os.Remove(filePath)

				actual, err := Load(filePath)
				require.NoError(t, err)

				if assert.NotNil(t, actual) {
					assert.Equal(t, tc.expected, *actual)
				}
			})
		}
	})

	t.Run("returns expected error", func(t *testing.T) {
		for _, tc := range []struct {
			name             string
			tomlContent      string
			expectedContains string
		}{
			{
				name:             "empty file",
				tomlContent:      "name = \"app name\"\nname = \"app name\"",
				expectedContains: "toml:",
			},
			{
				name:             "repeated value",
				tomlContent:      "name = \"app name\"\nname = \"app name\"",
				expectedContains: "toml:",
			},
			{
				name:             "unclosed table name",
				tomlContent:      "[python",
				expectedContains: "toml:",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				filePath := test.WriteTempFile(t, ManifestFileName, []byte(tc.tomlContent))
				defer os.Remove(filePath)

				actual, err := Load(filePath)

				assert.ErrorContains(t, err, tc.expectedContains)
				assert.Nil(t, actual)
			})
		}
	})
}

func TestToTOML(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     Manifest
		expectedTOML string
	}{
		{
			name:         "streamlit app",
			manifest:     manifestStreamlit,
			expectedTOML: tomlStreamlit,
		},
		{
			name:         "without default deployment",
			manifest:     manifestStreamlitNoDeploy,
			expectedTOML: tomlStreamlitNoDeploy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := tc.manifest.ToTOML()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTOML, actual)
		})
	}
}

func TestToJSON(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     Manifest
		expectedJSON string
	}{
		{
			name:         "streamlit app",
			manifest:     manifestStreamlit,
			expectedJSON: jsonStreamlit,
		},
		{
			name:         "without default deployment",
			manifest:     manifestStreamlitNoDeploy,
			expectedJSON: jsonStreamlitNoDeploy,
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
