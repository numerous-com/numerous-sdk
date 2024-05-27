package archive

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarCreate(t *testing.T) {
	tarDir := t.TempDir()
	tarFilePath := tarDir + "/test.tar"

	err := TarCreate("testdata/testfolder/", tarFilePath)
	assert.NoError(t, err)
	actual, err := readTarFile(tarFilePath)
	assert.NoError(t, err)

	expected := readFiles(t, "testdata/testfolder")
	assert.Equal(t, expected, actual)
}

func TestTarExtract(t *testing.T) {
	untar := t.TempDir()
	tarFile, err := os.Open("testdata/testfolder.tar")
	require.NoError(t, err)

	err = TarExtract(tarFile, untar)
	assert.NoError(t, err)

	expected := readFiles(t, "testdata/testfolder")
	actual := readFiles(t, untar)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestTarCreateTarExtract(t *testing.T) {
	tarDir := t.TempDir()
	tarFilePath := tarDir + "/test.tar"

	err := TarCreate("testdata/testfolder/", tarFilePath)
	assert.NoError(t, err)
	created, err := os.Open(tarFilePath)
	assert.NoError(t, err)
	err = TarExtract(created, tarDir+"/extracted")
	assert.NoError(t, err)

	expected := readFiles(t, "testdata/testfolder")
	actual := readFiles(t, tarDir+"/extracted")
	assert.Equal(t, expected, actual)
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
