package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

func (c *Config) Print() {
	toml.NewEncoder(os.Stdout).Encode(c) // nolint:errcheck
}
