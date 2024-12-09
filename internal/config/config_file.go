package config

import (
	"os"
	"path/filepath"
)

const configDirPerm os.FileMode = 0o755

func (c *Config) configFilePath() (string, error) {
	numerousConfigDir := filepath.Join(configBaseDir, "numerous")
	if err := os.MkdirAll(numerousConfigDir, configDirPerm); err != nil {
		return "", err
	}

	return filepath.Join(numerousConfigDir, "config.toml"), nil
}
