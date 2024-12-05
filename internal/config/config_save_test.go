package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("truncates configuration file before saving", func(t *testing.T) {
		configBaseDir = t.TempDir()
		cfg := Config{OrganizationSlug: "test-very-long-organization-slug"}
		err := cfg.Save()
		require.NoError(t, err)

		cfg.OrganizationSlug = "test-shorter-org-slug"
		err = cfg.Save()

		assert.NoError(t, err)
		actual, err := os.ReadFile(filepath.Join(configBaseDir, "numerous", "config.toml"))
		if assert.NoError(t, err) {
			expected := "organization = \"test-shorter-org-slug\"\n"
			assert.Equal(t, expected, string(actual))
		}
	})
}
