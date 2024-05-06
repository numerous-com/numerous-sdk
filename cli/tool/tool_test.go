package tool

import (
	"fmt"
	"path/filepath"
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAppID(t *testing.T) {
	expectedAppID := "app-id-goes-here"
	expectedToolD := "test-tool-id"

	t.Run(fmt.Sprintf("returns app ID if only %s exists", AppIDFileName), func(t *testing.T) {
		tmpDir := t.TempDir()
		test.WriteFile(t, filepath.Join(tmpDir, AppIDFileName), []byte(expectedAppID))

		actualAppID, err := ReadAppID(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, expectedAppID, actualAppID)
	})

	t.Run(fmt.Sprintf("returns tool ID if only %s exists", ToolIDFileName), func(t *testing.T) {
		tmpDir := t.TempDir()
		test.WriteFile(t, filepath.Join(tmpDir, ToolIDFileName), []byte(expectedToolD))

		actualToolID, err := ReadAppID(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, expectedToolD, actualToolID)
	})

	t.Run(fmt.Sprintf("returns app ID from %q even if %q also exists", AppIDFileName, ToolIDFileName), func(t *testing.T) {
		tmpDir := t.TempDir()
		test.WriteFile(t, filepath.Join(tmpDir, AppIDFileName), []byte(expectedAppID))
		test.WriteFile(t, filepath.Join(tmpDir, ToolIDFileName), []byte("some-tool-id"))

		actualAppID, err := ReadAppID(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, expectedAppID, actualAppID)
	})

	t.Run(fmt.Sprintf("returns error if neither %q or %q exist", AppIDFileName, ToolIDFileName), func(t *testing.T) {
		tmpDir := t.TempDir()

		actualAppID, err := ReadAppID(tmpDir)
		require.ErrorIs(t, err, ErrAppIDNotFound)
		assert.Equal(t, "", actualAppID)
	})
}
