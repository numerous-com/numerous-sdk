package wizard

import (
	"testing"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseOrCreateAppFolder(t *testing.T) {
	t.Run("Creates a folder structure when it doesn't exist", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{CreateFolderQuestion(nonexistingPath): true}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, nonexistingPath)
	})

	t.Run("Identify the current user folder structure", func(t *testing.T) {
		path := t.TempDir()
		asker := StubAsker{UseFolderQuestion(path): true}

		shouldContinue, err := UseOrCreateAppFolder(&asker, path)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, path)
	})

	t.Run("User rejects folder creation", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{CreateFolderQuestion(nonexistingPath): false}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
		assert.NoDirExists(t, nonexistingPath)
	})

	t.Run("User rejects existing folder choice", func(t *testing.T) {
		existingPath := t.TempDir()
		require.DirExists(t, existingPath)
		asker := StubAsker{UseFolderQuestion(existingPath): false}

		shouldContinue, err := UseOrCreateAppFolder(&asker, existingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
	})

	t.Run("User interrupts folder creation", func(t *testing.T) {
		nonexistingPath := t.TempDir() + "/test/folder"
		asker := StubAsker{CreateFolderQuestion(nonexistingPath): terminal.InterruptErr}

		shouldContinue, err := UseOrCreateAppFolder(&asker, nonexistingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
		assert.NoDirExists(t, nonexistingPath)
	})

	t.Run("User interrupts existing folder choice", func(t *testing.T) {
		existingPath := t.TempDir()
		require.DirExists(t, existingPath)
		asker := StubAsker{UseFolderQuestion(existingPath): terminal.InterruptErr}

		shouldContinue, err := UseOrCreateAppFolder(&asker, existingPath)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
	})
}
