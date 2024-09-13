package manifest

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
)

type ManifestV1 struct {
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

func (m *ManifestV1) ToTOML() (string, error) {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(m)

	return buf.String(), err
}

func (m *ManifestV1) ToJSON() (string, error) {
	manifest, err := json.Marshal(m)

	return string(manifest), err
}

func (m *ManifestV1) ToManifest() Manifest {
	return Manifest{
		App: App{
			Name:        m.Name,
			Description: m.Description,
			CoverImage:  m.CoverImage,
			Exclude:     m.Exclude,
		},
		Python: &Python{
			Library:          m.Library,
			Version:          m.Python,
			AppFile:          m.AppFile,
			RequirementsFile: m.RequirementsFile,
			Port:             m.Port,
		},
		Deployment: m.Deployment,
	}
}

func loadV1(filePath string) (*ManifestV1, error) {
	var m ManifestV1

	_, err := toml.DecodeFile(filePath, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
