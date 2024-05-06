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
	WriteFile(t, filePath, content)

	return filePath
}

func WriteFile(t *testing.T, filePath string, content []byte) {
	t.Helper()

	err := os.WriteFile(filePath, content, tempFileMode)
	require.NoError(t, err)
}
