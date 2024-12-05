package config

import "errors"

var ErrNoConfigDirDefined = errors.New("no config directory is defined")

type Config struct {
	OrganizationSlug string `toml:"organization,omitempty"`
}
