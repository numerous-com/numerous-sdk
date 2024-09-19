package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const ManifestFileName string = "numerous.toml"

type Manifest struct {
	App
	Python     *Python     `toml:"python,omitempty" json:"python,omitempty"`
	Docker     *Docker     `toml:"docker,omitempty" json:"docker,omitempty"`
	Deployment *Deployment `toml:"deploy,omitempty" json:"deploy,omitempty"`
}

type Docker struct {
	Dockerfile string `toml:"dockerfile,omitempty" json:"dockerfile,omitempty"`
	Context    string `toml:"context,omitempty" json:"context,omitempty"`
}

type Python struct {
	Library          Library `toml:"library" json:"library"`
	Version          string  `toml:"version" json:"version"`
	AppFile          string  `toml:"app_file" json:"app_file"`
	RequirementsFile string  `toml:"requirements_file" json:"requirements_file"`
}

type App struct {
	Name        string   `toml:"name" json:"name"`
	Description string   `toml:"description" json:"description"`
	CoverImage  string   `toml:"cover_image" json:"cover_image"`
	Exclude     []string `toml:"exclude" json:"exclude"`
	Port        uint     `toml:"port" json:"port"`
}

type Deployment struct {
	OrganizationSlug string `toml:"organization" json:"organization"`
	AppSlug          string `toml:"app" json:"app"`
}

func load(filePath string) (*Manifest, error) {
	var m Manifest

	_, err := toml.DecodeFile(filePath, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func Load(filePath string) (*Manifest, error) {
	m, err := load(filePath)
	if err == nil {
		return m, nil
	}

	v1, v1Err := loadV1(filePath)
	if v1Err == nil {
		m := v1.ToManifest()
		return &m, nil
	}

	v0, v0Err := loadV0(filePath)
	if v0Err == nil {
		return v0.ToManifest()
	}

	return nil, err
}

func NewApp(name, desc string, port uint) App {
	return App{
		Name:        name,
		Description: desc,
		CoverImage:  "app_cover.jpg",
		Exclude:     []string{"*venv", "venv*", ".git", ".env"},
		Port:        port,
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

func (m *Manifest) ToTOML() (string, error) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(m)

	return buf.String(), err
}

func (m *Manifest) ToJSON() (string, error) {
	manifest, err := json.Marshal(m)

	return string(manifest), err
}
