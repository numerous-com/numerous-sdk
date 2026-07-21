package archive

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
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

func TestTarExtract(t *testing.T) {
	t.Run("extracts files into dest", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		writeTarEntry(t, tw, "file.txt", []byte("hello"))
		writeTarEntry(t, tw, "dir/nested.txt", []byte("nested"))
		assert.NoError(t, tw.Close())

		dest := t.TempDir()
		err := TarExtract(&buf, dest)
		assert.NoError(t, err)

		assert.Equal(t, []byte("hello"), readFile(t, filepath.Join(dest, "file.txt")))
		assert.Equal(t, []byte("nested"), readFile(t, filepath.Join(dest, "dir", "nested.txt")))
	})

	t.Run("rejects entries that escape dest", func(t *testing.T) {
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)
		writeTarEntry(t, tw, "../escape.txt", []byte("evil"))
		assert.NoError(t, tw.Close())

		parent := t.TempDir()
		dest := filepath.Join(parent, "dest")
		assert.NoError(t, os.MkdirAll(dest, mkdirPerms))

		err := TarExtract(&buf, dest)
		assert.ErrorContains(t, err, "illegal file path in archive")

		_, statErr := os.Stat(filepath.Join(parent, "escape.txt"))
		assert.True(t, os.IsNotExist(statErr), "traversal target must not be written outside dest")
	})
}

func writeTarEntry(t *testing.T, tw *tar.Writer, name string, content []byte) {
	t.Helper()

	err := tw.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Mode:     0o644,
		Size:     int64(len(content)),
	})
	assert.NoError(t, err)

	_, err = tw.Write(content)
	assert.NoError(t, err)
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	assert.NoError(t, err)

	return data
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
		case errors.Is(err, io.EOF):
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
