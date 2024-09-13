package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

const v1TOMLStreamlit string = `name = "Streamlit App Name"
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

const v1JSONStreamlit string = `{"name":"Streamlit App Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"],"deploy":{"organization":"organization-slug","app":"app-slug"}}`

const v1JSONStreamlitNoDeploy string = `{"name":"Streamlit App Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"]}`

const v1TOMLStreamlitNoDeploy string = `name = "Streamlit App Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = 80
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`

var v1ManifestStreamlit = ManifestV1{
	Name:             "Streamlit App Name",
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

var v1ManifestStreamlitNoDeploy = ManifestV1{
	Name:             "Streamlit App Name",
	Description:      "A description",
	Library:          LibraryStreamlit,
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
}

const v1TOMLMarimo string = `name = "Marimo App Name"
description = "A description"
library = "marimo"
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

var v1ManifestMarimo = ManifestV1{
	Name:             "Marimo App Name",
	Description:      "A description",
	Library:          LibraryMarimo,
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
	Deployment:       &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

var manifestMarimo = Manifest{
	ManifestApp: ManifestApp{
		Name:        "Marimo App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
	},
	Python: &ManifestPython{
		Library:          LibraryMarimo,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
		Port:             80,
	},
	Deployment: &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

func TestV1ToTOML(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     ManifestV1
		expectedTOML string
	}{
		{
			name:         "streamlit app",
			manifest:     v1ManifestStreamlit,
			expectedTOML: v1TOMLStreamlit,
		},
		{
			name:         "without default deployment",
			manifest:     v1ManifestStreamlitNoDeploy,
			expectedTOML: v1TOMLStreamlitNoDeploy,
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

func TestV1ToJSON(t *testing.T) {
	testCases := []struct {
		name         string
		manifest     ManifestV1
		expectedJSON string
	}{
		{
			name:         "streamlit app",
			manifest:     v1ManifestStreamlit,
			expectedJSON: v1JSONStreamlit,
		},
		{
			name:         "without default deployment",
			manifest:     v1ManifestStreamlitNoDeploy,
			expectedJSON: v1JSONStreamlitNoDeploy,
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

func TestLoadV1(t *testing.T) {
	t.Run("returns expected v1 manifest", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			toml     string
			expected ManifestV1
		}{
			{name: "streamlit app", toml: v1TOMLStreamlit, expected: v1ManifestStreamlit},
			{name: "marimo app", toml: v1TOMLMarimo, expected: v1ManifestMarimo},
		} {
			t.Run(tc.name, func(t *testing.T) {
				filepath := test.WriteTempFile(t, "numerous.toml", []byte(tc.toml))

				actual, err := loadV1(filepath)

				assert.NoError(t, err)
				if assert.NotNil(t, actual) {
					assert.Equal(t, tc.expected, *actual)
				}
			})
		}
	})
}

func TestManifestV1ToManifest(t *testing.T) {
	t.Run("returns expected manifest", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			expected Manifest
			v1       ManifestV1
		}{
			{
				name:     "marimo with deployment",
				expected: manifestMarimo,
				v1:       v1ManifestMarimo,
			},
			{
				name:     "streamlit with deployment",
				expected: manifestStreamlit,
				v1:       v1ManifestStreamlit,
			},
			{
				name:     "streamlit without deployment",
				expected: manifestStreamlitNoDeploy,
				v1:       v1ManifestStreamlitNoDeploy,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				actual := tc.v1.ToManifest()

				assert.Equal(t, tc.expected, actual)
			})
		}
	})
}
