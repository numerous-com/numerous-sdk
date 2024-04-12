package test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var tempFileMode fs.FileMode = 0o644

func WriteTempFile(t *testing.T, filename string, content []byte) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), filename)
	err := os.WriteFile(filePath, content, tempFileMode)
	require.NoError(t, err)

	return filePath
}
