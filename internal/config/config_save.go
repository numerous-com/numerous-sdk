package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

const configPerm os.FileMode = 0o640

func (c *Config) Save() error {
	path, err := c.configFilePath()
	if err != nil {
		return err
	}

	w, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, configPerm)
	if err != nil {
		return err
	}

	return toml.NewEncoder(w).Encode(c)
}
