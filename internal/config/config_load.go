package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

func (c *Config) Load() error {
	path, err := c.configFilePath()
	if err != nil {
		return err
	}

	r, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	_, err = toml.NewDecoder(r).Decode(c)

	return err
}
