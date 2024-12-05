package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	t.Run("saves expected file", func(t *testing.T) {
		configBaseDir = t.TempDir()

		cfg := Config{OrganizationSlug: "test-organization-slug"}
		err := cfg.Save()

		assert.NoError(t, err)
		actual, err := os.ReadFile(filepath.Join(configBaseDir, "numerous", "config.toml"))
		if assert.NoError(t, err) {
			expected := "organization = \"test-organization-slug\"\n"
			assert.Equal(t, expected, string(actual))
		}
	})
}
