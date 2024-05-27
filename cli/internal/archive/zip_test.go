package archive

import (
	"archive/zip"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZipCreate(t *testing.T) {
	t.Run("creates zip with all files", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/test.zip"
		err := ZipCreate("testdata/testfolder/", path, nil)
		assert.NoError(t, err)
		actual, err := readZipFile(t, path)
		assert.NoError(t, err)

		expected := readFiles(t, "testdata/testfolder")
		assert.Equal(t, expected, actual)
	})

	t.Run("creates zip without ignored file path", func(t *testing.T) {
		dir := t.TempDir()
		path := dir + "/test.zip"

		err := ZipCreate("testdata/testfolder/", path, []string{"dir/*"})
		assert.NoError(t, err)
		actual, err := readZipFile(t, path)
		assert.NoError(t, err)

		expected := readFiles(t, "testdata/testfolder")
		delete(expected, "dir/nested_file.txt")
		assert.Equal(t, expected, actual)
	})
}

func readZipFile(t *testing.T, path string) (map[string][]byte, error) {
	t.Helper()

	stat, err := os.Stat(path)
	require.NoError(t, err)

	result := make(map[string][]byte)
	file, err := os.Open(path)
	require.NoError(t, err)

	r, err := zip.NewReader(file, stat.Size())
	require.NoError(t, err)

	for _, file := range r.File {
		if file.Mode().IsDir() {
			continue
		}

		f, err := r.Open(file.Name)
		require.NoError(t, err)

		content, err := io.ReadAll(f)
		require.NoError(t, err)

		result[file.Name] = content
	}

	return result, nil
}
