package test

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var testFileMode fs.FileMode = 0o644

func WriteTempFile(t *testing.T, filename string, content []byte) string {
	t.Helper()

	filePath := filepath.Join(t.TempDir(), filename)
	WriteFile(t, filePath, content)

	return filePath
}

func WriteFile(t *testing.T, filePath string, content []byte) {
	t.Helper()

	err := os.WriteFile(filePath, content, testFileMode)
	require.NoError(t, err)
}

func CopyFile(t *testing.T, source string, dest string) {
	t.Helper()

	f, err := os.Open(source)
	require.NoError(t, err)

	data, err := io.ReadAll(f)
	require.NoError(t, err)

	err = os.WriteFile(dest, data, testFileMode)
	require.NoError(t, err)
}
