package config

var configBaseDir string

// Set the config base directory, and return the old value.
func OverrideConfigBaseDir(newValue string) string {
	oldValue := configBaseDir
	configBaseDir = newValue

	return oldValue
}
