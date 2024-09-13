package wizard

import (
	"fmt"
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseOrCreateAppFolder(t *testing.T) {
	t.Run("Creates a folder structure when it doesn't exist", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{createFolderQuestion(nonexistingPath): true}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, nonexistingPath)
	})

	t.Run("Identify the current user folder structure", func(t *testing.T) {
		path := t.TempDir()
		asker := StubAsker{useFolderQuestion(path): true}

		shouldContinue, err := UseOrCreateAppFolder(&asker, path)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, path)
	})

	t.Run("User rejects folder creation", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{createFolderQuestion(nonexistingPath): false}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
		assert.NoDirExists(t, nonexistingPath)
	})

	t.Run("User rejects existing folder choice", func(t *testing.T) {
		existingPath := t.TempDir()
		require.DirExists(t, existingPath)
		asker := StubAsker{useFolderQuestion(existingPath): false}

		shouldContinue, err := UseOrCreateAppFolder(&asker, existingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
	})

	t.Run("User interrupts folder creation", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{createFolderQuestion(nonexistingPath): terminal.InterruptErr}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
		assert.NoDirExists(t, nonexistingPath)
	})

	t.Run("User interrupts existing folder choice", func(t *testing.T) {
		existingPath := t.TempDir()
		require.DirExists(t, existingPath)
		asker := StubAsker{useFolderQuestion(existingPath): terminal.InterruptErr}

		shouldContinue, err := UseOrCreateAppFolder(&asker, existingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
	})
}

func useFolderQuestion(path string) string {
	return fmt.Sprintf("Use the existing folder %s for your app? (default: yes)", path)
}

func createFolderQuestion(path string) string {
	return fmt.Sprintf("Create new folder '%s'? (default: yes)", path)
}
