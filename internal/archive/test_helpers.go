package archive

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()

	files := make(map[string][]byte, 0)
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return err
		}

		if relpath, err := filepath.Rel(root, path); !assert.NoError(t, err) {
			return err
		} else if data, err := os.ReadFile(path); assert.NoError(t, err) {
			files[relpath] = data
			return err
		}

		return err
	})
	assert.NoError(t, err)

	return files
}
