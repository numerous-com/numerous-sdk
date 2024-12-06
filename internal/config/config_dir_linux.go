package config

import (
	"os"
	"path/filepath"
)

func init() {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configBaseDir = xdgConfigHome
	} else {
		configBaseDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
}
