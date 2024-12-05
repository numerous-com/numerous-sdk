package config

import (
	"os"
)

func init() {
	configBaseDir = os.Getenv("APPDATA")
}
