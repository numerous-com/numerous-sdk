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

const v1TOMLStreamlitWithSize string = `name = "Streamlit App Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = 80
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
size = "small"

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`

const v1JSONStreamlit string = `{"name":"Streamlit App Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"],"Size":null,"deploy":{"organization":"organization-slug","app":"app-slug"}}`

const v1JSONStreamlitWithSize string = `{"name":"Streamlit App Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"],"Size":"small","deploy":{"organization":"organization-slug","app":"app-slug"}}`

const v1JSONStreamlitNoDeploy string = `{"name":"Streamlit App Name","description":"A description","library":"streamlit","python":"3.11","app_file":"app.py","requirements_file":"requirements.txt","port":80,"cover_image":"cover.png","exclude":["*venv","venv*"],"Size":null}`

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

var v1ManifestStreamlitWithSize = ManifestV1{
	Name:             "Streamlit App Name",
	Description:      "A description",
	Library:          LibraryStreamlit,
	Python:           "3.11",
	Port:             80,
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
	Size:             ref("small"),
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
	App: App{
		Name:        "Marimo App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        80,
	},
	Python: &Python{
		Library:          LibraryMarimo,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
	},
	Deployment: &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

var manifestStreamlitV1WithSize = Manifest{
	App: App{
		Name:        "Streamlit App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        80,
		Size:        ref("small"),
	},
	Python: &Python{
		Library:          LibraryStreamlit,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
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
			name:         "streamlit app with size",
			manifest:     v1ManifestStreamlitWithSize,
			expectedTOML: v1TOMLStreamlitWithSize,
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
			name:         "streamlit app with size",
			manifest:     v1ManifestStreamlitWithSize,
			expectedJSON: v1JSONStreamlitWithSize,
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
			{name: "streamlit app with size", toml: v1TOMLStreamlitWithSize, expected: v1ManifestStreamlitWithSize},
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
				name:     "streamlit with deployment and size",
				expected: manifestStreamlitV1WithSize,
				v1:       v1ManifestStreamlitWithSize,
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
