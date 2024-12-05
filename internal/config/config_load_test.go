package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("returns empty config if file does not exist", func(t *testing.T) {
		configBaseDir = t.TempDir()
		cfg := Config{}

		err := cfg.Load()

		assert.NoError(t, err)
		assert.Empty(t, cfg)
	})

	t.Run("returns empty config if directory does not exist", func(t *testing.T) {
		configBaseDir = t.TempDir()
		cfg := Config{}

		err := cfg.Load()

		assert.NoError(t, err)
		assert.Empty(t, cfg)
	})

	t.Run("returns expected config", func(t *testing.T) {
		configBaseDir = t.TempDir()
		cfg := Config{}
		require.NoError(t, os.Mkdir(filepath.Join(configBaseDir, "numerous"), configDirPerm))
		require.NoError(t, os.WriteFile(filepath.Join(configBaseDir, "numerous", "config.toml"), []byte("organization = \"organization-slug\""), configPerm))

		err := cfg.Load()

		expected := Config{OrganizationSlug: "organization-slug"}
		assert.NoError(t, err)
		assert.Equal(t, expected, cfg)
	})
}
