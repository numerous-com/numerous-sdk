package initialize

import (
	"os"
	"path/filepath"
	"testing"

	"numerous/cli/tool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expectedAddedGitIgnorePattern = "expected-added-gitignore-pattern"

const mockGitIgnore string = `
some.txt
*.go
another.yaml`

const expectedGitIgnore string = `
some.txt
*.go
another.yaml
` + expectedAddedGitIgnorePattern

func TestCreateAppIdFile(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.Chdir(tmpDir)
	require.NoError(t, err)
	appIDPath := filepath.Join(tmpDir, tool.AppIDFileName)
	appID := "some-id"

	err = createAppIDFile(tmpDir, appID)

	if assert.NoError(t, err) {
		toolIDContent, err := os.ReadFile(appIDPath)
		if assert.NoError(t, err) {
			actualID := string(toolIDContent)
			assert.Equal(t, appID, actualID)
		}
	}
}

func TestAddToGitIgnore(t *testing.T) {
	t.Run("Adds if .gitignore exists", func(t *testing.T) {
		testPath := t.TempDir()
		err := os.Chdir(testPath)
		require.NoError(t, err)
		gitignorePath := filepath.Join(testPath, ".gitignore")

		file, err := os.Create(gitignorePath)
		require.NoError(t, err)
		_, err = file.WriteString(mockGitIgnore)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		err = addToGitIgnore(testPath, expectedAddedGitIgnorePattern)

		if assert.NoError(t, err) {
			fileContentInBytes, err := os.ReadFile(gitignorePath)
			if assert.NoError(t, err) {
				assert.Equal(t, expectedGitIgnore, string(fileContentInBytes))
			}
		}
	})

	t.Run("Creates and adds if .gitignore does not exists", func(t *testing.T) {
		testPath := t.TempDir()
		err := os.Chdir(testPath)
		require.NoError(t, err)
		gitignorePath := filepath.Join(testPath, ".gitignore")
		require.NoFileExists(t, gitignorePath)

		err = addToGitIgnore(testPath, expectedAddedGitIgnorePattern)

		if assert.NoError(t, err) {
			fileContentInBytes, err := os.ReadFile(gitignorePath)
			if assert.NoError(t, err) {
				assert.Equal(t, expectedAddedGitIgnorePattern, string(fileContentInBytes))
			}
		}
	})
}

func TestWriteFiles(t *testing.T) {
	expectedContent := "some text input"

	t.Run("Can write to existing file", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "test_file.txt")
		file, err := os.Create(filePath)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		err = writeOrAppendFile(filePath, expectedContent)

		if assert.NoError(t, err) {
			actualContent, err := os.ReadFile(filePath)
			if assert.NoError(t, err) {
				assert.Equal(t, expectedContent, string(actualContent))
			}
		}
	})

	t.Run("Returns error if file does not exist", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "test_file.txt")
		file, err := os.Create(filePath)
		require.NoError(t, err)
		require.NoError(t, file.Close())

		err = writeOrAppendFile(filePath, expectedContent)

		assert.NoError(t, err)
	})
}

func TestCreateFile(t *testing.T) {
	t.Run("Can create file if it does not exist", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "test_file.txt")
		require.NoFileExists(t, filePath)

		err := createFile(filePath)

		if assert.NoError(t, err) {
			assert.FileExists(t, filePath)
		}
	})

	t.Run("Does nothing if file exists", func(t *testing.T) {
		filePath := filepath.Join(t.TempDir(), "test_file.txt")
		expectedContent := "some content"
		f, err := os.Create(filePath)
		require.NoError(t, err)
		_, err = f.WriteString(expectedContent)
		require.NoError(t, err)

		err = createFile(filePath)

		if assert.NoError(t, err) {
			actualContent, err := os.ReadFile(filePath)
			if assert.NoError(t, err) {
				assert.Equal(t, expectedContent, string(actualContent))
			}
		}
	})
}
