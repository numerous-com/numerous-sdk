package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/test"
)

const v0TOMLStreamlit string = `name = "Streamlit App Name"
description = "A description"
library = "streamlit"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = "80"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`

var v0ManifestStreamlit = ManifestV0{
	Name:             "Streamlit App Name",
	Description:      "A description",
	Library:          LibraryStreamlit,
	Python:           "3.11",
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	Port:             "80",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
}

const v0TOMLMarimo string = `name = "Marimo App Name"
description = "A description"
library = "marimo"
python = "3.11"
app_file = "app.py"
requirements_file = "requirements.txt"
port = "80"
cover_image = "cover.png"
exclude = ["*venv", "venv*"]
`

var v0ManifestMarimo = ManifestV0{
	Name:             "Marimo App Name",
	Description:      "A description",
	Library:          LibraryMarimo,
	Python:           "3.11",
	AppFile:          "app.py",
	RequirementsFile: "requirements.txt",
	Port:             "80",
	CoverImage:       "cover.png",
	Exclude:          []string{"*venv", "venv*"},
}

var manifestMarimoNoDeploy = Manifest{
	App: App{
		Name:        "Marimo App Name",
		Description: "A description",
		CoverImage:  "cover.png",
		Exclude:     []string{"*venv", "venv*"},
	},
	Python: &Python{
		Library:          LibraryMarimo,
		Version:          "3.11",
		AppFile:          "app.py",
		RequirementsFile: "requirements.txt",
		Port:             80,
	},
}

func TestLoadV0(t *testing.T) {
	t.Run("returns expected v0 manifest", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			toml     string
			expected ManifestV0
		}{
			{name: "streamlit app", toml: v0TOMLStreamlit, expected: v0ManifestStreamlit},
			{name: "marimo app", toml: v0TOMLMarimo, expected: v0ManifestMarimo},
		} {
			t.Run(tc.name, func(t *testing.T) {
				filepath := test.WriteTempFile(t, "numerous.toml", []byte(tc.toml))

				actual, err := loadV0(filepath)

				assert.NoError(t, err)
				if assert.NotNil(t, actual) {
					assert.Equal(t, tc.expected, *actual)
				}
			})
		}
	})
}

func TestManifestV0ToManifest(t *testing.T) {
	t.Run("returns expected manifest", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			expected Manifest
			v0       ManifestV0
		}{
			{
				name:     "marimo",
				expected: manifestMarimoNoDeploy,
				v0:       v0ManifestMarimo,
			},
			{
				name:     "streamlit",
				expected: manifestStreamlitNoDeploy,
				v0:       v0ManifestStreamlit,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				actual, err := tc.v0.ToManifest()

				assert.NoError(t, err)
				if assert.NotNil(t, actual) {
					assert.Equal(t, tc.expected, *actual)
				}
			})
		}
	})
}
