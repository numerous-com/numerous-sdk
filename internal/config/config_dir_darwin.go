package config

import (
	"os"
	"path/filepath"
)

func init() {
	configBaseDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
}
