package init

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/test"
	"numerous.com/cli/internal/wizard"
)

func TestInitialize(t *testing.T) {
	appName := "App name"
	appDesc := "App description"
	appLibraryKey := "streamlit"
	appFile := "app.py"
	appRequirementsFile := "requirements.txt"

	t.Run("initializes python app", func(t *testing.T) {
		appdir := t.TempDir()

		asker := wizard.StubAsker{wizard.UseFolderQuestion(appdir): true}
		params := InitializeParams{
			AppDir:           appdir,
			Name:             appName,
			Desc:             appDesc,
			LibraryKey:       appLibraryKey,
			AppFile:          appFile,
			RequirementsFile: appRequirementsFile,
		}
		_, err := Initialize(asker, params)

		assert.NoError(t, err)
		assert.FileExists(t, filepath.Join(appdir, "numerous.toml"))
		assert.FileExists(t, filepath.Join(appdir, appFile))
		assert.FileExists(t, filepath.Join(appdir, appRequirementsFile))
		assert.FileExists(t, filepath.Join(appdir, ".gitignore"))
		m, err := manifest.Load(filepath.Join(appdir, "numerous.toml"))
		if assert.NoError(t, err) {
			assert.Equal(t, appName, m.Name)
			assert.Equal(t, appDesc, m.Description)
			assert.Equal(t, appLibraryKey, m.Python.Library.Key)
		}
	})

	t.Run("given appdir with manifest file returns error", func(t *testing.T) {
		appdir := t.TempDir()
		test.WriteFile(t, filepath.Join(appdir, "numerous.toml"), []byte("manifest contet"))

		asker := wizard.StubAsker{wizard.UseFolderQuestion(appdir): true}
		params := InitializeParams{AppDir: appdir}
		_, err := Initialize(asker, params)

		assert.ErrorIs(t, err, ErrAppAlreadyInitialized)
	})

	t.Run("given invalid library key it returns error", func(t *testing.T) {
		appdir := t.TempDir()

		asker := wizard.StubAsker{wizard.UseFolderQuestion(appdir): true}
		params := InitializeParams{
			AppDir:           appdir,
			Name:             appName,
			Desc:             appDesc,
			LibraryKey:       "some-unsupported-library",
			AppFile:          appFile,
			RequirementsFile: appRequirementsFile,
		}
		_, err := Initialize(asker, params)

		assert.ErrorIs(t, err, manifest.ErrUnsupportedLibrary)
	})
}
