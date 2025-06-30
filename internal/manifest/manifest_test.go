package manifest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/test"
)

const tomlStreamlit string = `name = "Streamlit App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`

const jsonStreamlit string = `{"name":"Streamlit App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"port":80,"python":{"library":"streamlit","version":"3.11","app_file":"app.py","requirements_file":"requirements.txt"},"deploy":{"organization":"organization-slug","app":"app-slug"}}`

const jsonStreamlitNoDeploy string = `{"name":"Streamlit App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"port":80,"python":{"library":"streamlit","version":"3.11","app_file":"app.py","requirements_file":"requirements.txt"}}`

const tomlStreamlitNoDeploy string = `name = "Streamlit App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 80

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"
`

const tomlStreamlitWithSize string = `name = "Streamlit App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 80
size = "small"

[python]
  library = "streamlit"
  version = "3.11"
  app_file = "app.py"
  requirements_file = "requirements.txt"

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`

const jsonStreamlitWithSize string = `{"name":"Streamlit App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"port":80,"size":"small","python":{"library":"streamlit","version":"3.11","app_file":"app.py","requirements_file":"requirements.txt"},"deploy":{"organization":"organization-slug","app":"app-slug"}}`

var manifestStreamlit Manifest = Manifest{
	App: App{
		Name:        "Streamlit App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        80,
	},
	Python: &Python{
		Library:          LibraryStreamlit,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
	},
	Deployment: &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

var manifestStreamlitNoDeploy Manifest = Manifest{
	App: App{
		Name:        "Streamlit App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        80,
	},
	Python: &Python{
		Library:          LibraryStreamlit,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
	},
	Deployment: nil,
}

var manifestStreamlitWithSize Manifest = Manifest{
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

const tomlDockerNoDeploy string = `name = "Docker App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 1234

[docker]
  dockerfile = "Dockerfile"
  context = "."
`

const tomlDocker string = `name = "Docker App Name"
description = "A description"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
port = 1234

[docker]
  dockerfile = "Dockerfile"
  context = "."

[deploy]
  organization = "organization-slug"
  app = "app-slug"
`

const jsonDocker string = `{"name":"Docker App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"port":1234,"docker":{"dockerfile":"Dockerfile","context":"."},"deploy":{"organization":"organization-slug","app":"app-slug"}}`

const jsonDockerNoDeploy string = `{"name":"Docker App Name","description":"A description","cover_image":"cover.png","exclude":["*venv","venv*"],"port":1234,"docker":{"dockerfile":"Dockerfile","context":"."}}`

var manifestDocker Manifest = Manifest{
	App: App{
		Name:        "Docker App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        1234,
	},
	Python:     nil,
	Docker:     &Docker{Dockerfile: "Dockerfile", Context: "."},
	Deployment: &Deployment{OrganizationSlug: "organization-slug", AppSlug: "app-slug"},
}

var manifestDockerNoDeploy Manifest = Manifest{
	App: App{
		Name:        "Docker App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
		Port:        1234,
	},
	Python:     nil,
	Docker:     &Docker{Dockerfile: "Dockerfile", Context: "."},
	Deployment: nil,
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
			{
				name:        "streamlit with size",
				tomlContent: tomlStreamlitWithSize,
				expected:    manifestStreamlitWithSize,
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
			name:         "streamlit",
			manifest:     manifestStreamlit,
			expectedTOML: tomlStreamlit,
		},
		{
			name:         "streamlit without default deployment",
			manifest:     manifestStreamlitNoDeploy,
			expectedTOML: tomlStreamlitNoDeploy,
		},
		{
			name:         "streamlit with size",
			manifest:     manifestStreamlitWithSize,
			expectedTOML: tomlStreamlitWithSize,
		},
		{
			name:         "docker",
			manifest:     manifestDocker,
			expectedTOML: tomlDocker,
		},
		{
			name:         "docker without default deployment",
			manifest:     manifestDockerNoDeploy,
			expectedTOML: tomlDockerNoDeploy,
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
			name:         "streamlit",
			manifest:     manifestStreamlit,
			expectedJSON: jsonStreamlit,
		},
		{
			name:         "streamlit without default deployment",
			manifest:     manifestStreamlitNoDeploy,
			expectedJSON: jsonStreamlitNoDeploy,
		},
		{
			name:         "streamlit with size",
			manifest:     manifestStreamlitWithSize,
			expectedJSON: jsonStreamlitWithSize,
		},
		{
			name:         "docker",
			manifest:     manifestDocker,
			expectedJSON: jsonDocker,
		},
		{
			name:         "docker without default deployment",
			manifest:     manifestDockerNoDeploy,
			expectedJSON: jsonDockerNoDeploy,
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
