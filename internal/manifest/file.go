package manifest

import (
	"errors"
	"os"
	"path/filepath"
)

const ManifestFileName string = "numerous.toml"

var ManifestPath string = filepath.Join(".", ManifestFileName)

func Exists(appDir string) (bool, error) {
	manifestPath := filepath.Join(appDir, ManifestFileName)
	_, err := os.Stat(manifestPath)
	exists := !errors.Is(err, os.ErrNotExist)

	return exists, err
}
