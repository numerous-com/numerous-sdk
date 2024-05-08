package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"numerous/cli/cmd/output"
	"numerous/cli/tool"

	"github.com/BurntSushi/toml"
)

const ManifestFileName string = "numerous.toml"

var ManifestPath string = filepath.Join(".", ManifestFileName)

type Manifest struct {
	Name             string   `toml:"name" json:"name"`
	Description      string   `toml:"description" json:"description"`
	Library          string   `toml:"library" json:"library"`
	Python           string   `toml:"python" json:"python"`
	AppFile          string   `toml:"app_file" json:"app_file"`
	RequirementsFile string   `toml:"requirements_file" json:"requirements_file"`
	Port             uint     `toml:"port" json:"port"`
	CoverImage       string   `toml:"cover_image" json:"cover_image"`
	Exclude          []string `toml:"exclude" json:"exclude"`
}

type DeprecatedManifest struct {
	Name             string   `toml:"name" json:"name"`
	Description      string   `toml:"description" json:"description"`
	Library          string   `toml:"library" json:"library"`
	Python           string   `toml:"python" json:"python"`
	AppFile          string   `toml:"app_file" json:"app_file"`
	RequirementsFile string   `toml:"requirements_file" json:"requirements_file"`
	Port             string   `toml:"port" json:"port"`
	CoverImage       string   `toml:"cover_image" json:"cover_image"`
	Exclude          []string `toml:"exclude" json:"exclude"`
}

func LoadManifest(filePath string) (*Manifest, error) {
	var m Manifest

	if _, err := toml.DecodeFile(filePath, &m); err != nil {
		return loadDeprecatedManifest(filePath)
	}

	return &m, nil
}

func loadDeprecatedManifest(filePath string) (*Manifest, error) {
	var m DeprecatedManifest

	_, err := toml.DecodeFile(filePath, &m)
	if err != nil {
		return nil, err
	}

	return m.ToManifest()
}

func (d *DeprecatedManifest) ToManifest() (*Manifest, error) {
	port, err := strconv.ParseUint(d.Port, 10, 64)
	if err != nil {
		return nil, err
	}

	m := Manifest{
		Name:             d.Name,
		Description:      d.Description,
		Library:          d.Library,
		Python:           d.Python,
		AppFile:          d.AppFile,
		RequirementsFile: d.RequirementsFile,
		Port:             uint(port),
		CoverImage:       d.CoverImage,
		Exclude:          d.Exclude,
	}

	return &m, nil
}

func FromTool(t tool.Tool) *Manifest {
	return &Manifest{
		Name:             t.Name,
		Description:      t.Description,
		Library:          t.Library.Key,
		Python:           t.Python,
		AppFile:          t.AppFile,
		RequirementsFile: t.RequirementsFile,
		Port:             t.Library.Port,
		CoverImage:       t.CoverImage,
		Exclude:          []string{"*venv", "venv*", ".git"},
	}
}

func ManifestExistsInCurrentDir() (bool, error) {
	_, err := os.Stat(ManifestPath)
	exists := !errors.Is(err, os.ErrNotExist)

	return exists, err
}

func (m *Manifest) ToToml() (string, error) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(m)

	return buf.String(), err
}

func (m *Manifest) ToJSON() (string, error) {
	manifest, err := json.Marshal(m)

	return string(manifest), err
}

// Validates that the app defined in the manifest is valid. Returns false, if
// the app is in a state, where it does not make sense, to be able to push the
// app.
func (m *Manifest) ValidateApp() (bool, error) {
	switch m.Library {
	case "numerous":
		return m.validateNumerousApp()
	default:
		return true, nil
	}
}

func (m *Manifest) validateNumerousApp() (bool, error) {
	data, err := os.ReadFile(m.AppFile)
	if err != nil {
		return false, err
	}

	filecontent := string(data)
	if strings.Contains(filecontent, "appdef =") || strings.Contains(filecontent, "class appdef") {
		return true, nil
	} else {
		output.PrintError("Your app file must have an app definition called 'appdef'", strings.Join(
			[]string{
				"You can solve this by assigning your app definition to this name, for example:",
				"",
				"@app",
				"class MyApp:",
				"    my_field: str",
				"",
				"appdef = MyApp",
			}, "\n"),
		)

		return false, nil
	}
}
