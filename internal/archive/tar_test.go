package archive

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarCreate(t *testing.T) {
	t.Run("creates tar with all files", func(t *testing.T) {
		tarDir := t.TempDir()
		tarFilePath := tarDir + "/test.tar"

		err := TarCreate("testdata/testfolder/", tarFilePath, nil)
		assert.NoError(t, err)
		actual, err := readTarFile(tarFilePath)
		assert.NoError(t, err)

		expected := readFiles(t, "testdata/testfolder")
		assert.Equal(t, expected, actual)
	})

	t.Run("creates tar without ignored file path", func(t *testing.T) {
		tarDir := t.TempDir()
		tarFilePath := tarDir + "/test.tar"

		err := TarCreate("testdata/testfolder/", tarFilePath, []string{"dir/*"})
		assert.NoError(t, err)
		actual, err := readTarFile(tarFilePath)
		assert.NoError(t, err)

		expected := readFiles(t, "testdata/testfolder")
		delete(expected, "dir/nested_file.txt")
		assert.Equal(t, expected, actual)
	})
}

func readTarFile(tarFilePath string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(tarFile)

ReadTar:
	for {
		var b []byte

		h, err := tr.Next()

		switch {
		case (errors.Is(err, io.EOF)):
			break ReadTar
		case (err != nil):
			return nil, err
		case (h.Typeflag == tar.TypeDir):
			continue
		default:
			if b, err = io.ReadAll(tr); err != nil {
				return nil, err
			}
			result[h.Name] = b
		}
	}

	return result, nil
}
