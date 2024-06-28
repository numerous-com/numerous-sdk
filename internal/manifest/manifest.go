package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
)

const ManifestFileName string = "numerous.toml"

var ManifestPath string = filepath.Join(".", ManifestFileName)

type Manifest struct {
	Name             string      `toml:"name" json:"name"`
	Description      string      `toml:"description" json:"description"`
	Library          Library     `toml:"library" json:"library"`
	Python           string      `toml:"python" json:"python"`
	AppFile          string      `toml:"app_file" json:"app_file"`
	RequirementsFile string      `toml:"requirements_file" json:"requirements_file"`
	Port             uint        `toml:"port" json:"port"`
	CoverImage       string      `toml:"cover_image" json:"cover_image"`
	Exclude          []string    `toml:"exclude" json:"exclude"`
	Deployment       *Deployment `toml:"deploy,omitempty" json:"deploy,omitempty"`
}

type DeprecatedManifest struct {
	Name             string   `toml:"name" json:"name"`
	Description      string   `toml:"description" json:"description"`
	Library          Library  `toml:"library" json:"library"`
	Python           string   `toml:"python" json:"python"`
	AppFile          string   `toml:"app_file" json:"app_file"`
	RequirementsFile string   `toml:"requirements_file" json:"requirements_file"`
	Port             string   `toml:"port" json:"port"`
	CoverImage       string   `toml:"cover_image" json:"cover_image"`
	Exclude          []string `toml:"exclude" json:"exclude"`
}

type Deployment struct {
	OrganizationSlug string `toml:"organization" json:"organization"`
	AppSlug          string `toml:"app" json:"app"`
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

func New(lib Library, name string, description string, python string, appFile string, requirementsFile string) *Manifest {
	return &Manifest{
		Name:             name,
		Description:      description,
		Library:          lib,
		Python:           python,
		AppFile:          appFile,
		RequirementsFile: requirementsFile,
		Port:             lib.Port,
		CoverImage:       "app_cover.jpg",
		Exclude:          []string{"*venv", "venv*", ".git", ".env"},
	}
}

func ManifestExistsInCurrentDir() (bool, error) {
	return ManifestExists(".")
}

func ManifestExists(appDir string) (bool, error) {
	manifestPath := filepath.Join(appDir, ManifestFileName)
	_, err := os.Stat(manifestPath)
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
