package manifest

import (
	"strconv"

	"github.com/BurntSushi/toml"
)

type ManifestV0 struct {
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

func (d *ManifestV0) ToManifest() (*Manifest, error) {
	port, err := strconv.ParseUint(d.Port, 10, 64)
	if err != nil {
		return nil, err
	}

	m := Manifest{
		App: App{
			Name:        d.Name,
			Description: d.Description,
			CoverImage:  d.CoverImage,
			Exclude:     d.Exclude,
		},
		Python: &Python{
			Library:          d.Library,
			Version:          d.Python,
			AppFile:          d.AppFile,
			RequirementsFile: d.RequirementsFile,
			Port:             uint(port),
		},
		Deployment: nil,
	}

	return &m, nil
}

func loadV0(filePath string) (*ManifestV0, error) {
	var m ManifestV0

	_, err := toml.DecodeFile(filePath, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
