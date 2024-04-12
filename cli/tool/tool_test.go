package tool

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTool(t *testing.T) {
	t.Run("ReadToolID returns toolID if it exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, ToolIDFileName)
		expectedToolID := "tool-id-goes-here"
		require.NoError(t, os.WriteFile(path, []byte(expectedToolID), 0o644))

		actualToolID, err := ReadToolID(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, expectedToolID, actualToolID)
	})

	t.Run("ReadToolID returns error if tool id document does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		actualToolID, err := ReadToolID(tmpDir)
		require.ErrorIs(t, err, ErrToolIDNotFound)
		assert.Equal(t, "", actualToolID)
	})
}
