package wizard

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubFileReader struct {
	index   int
	returns []struct {
		b   []byte
		err error
	}
}

func (f *stubFileReader) Read(p []byte) (n int, err error) {
	if f.index > len(f.returns) {
		return 0, io.EOF
	} else {
		ret := f.returns[f.index]
		f.index++
		copy(p, ret.b)

		return len(ret.b), ret.err
	}
}

func (f *stubFileReader) Fd() uintptr {
	return 0
}

func (f *stubFileReader) Reset() {
	f.index = 0
}

func TestUseOrCreateAppFolder(t *testing.T) {
	yesReader := stubFileReader{
		returns: []struct {
			b   []byte
			err error
		}{
			{b: []byte("\x1b[10;10R"), err: nil},
			{b: []byte("\x1b[10;10R"), err: nil},
			{b: []byte("y\n"), err: nil},
			{b: []byte(""), err: io.EOF},
		},
	}

	t.Run("Creates a folder structure when it doesn't exist", func(t *testing.T) {
		yesReader.Reset()
		path := t.TempDir() + "/test/folder"

		shouldContinue, err := UseOrCreateAppFolder(path, &yesReader)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, path)
	})

	t.Run("Identify the current user folder structure", func(t *testing.T) {
		yesReader.Reset()
		path := t.TempDir()

		shouldContinue, err := UseOrCreateAppFolder(path, &yesReader)

		assert.True(t, shouldContinue)
		assert.NoError(t, err)
		assert.DirExists(t, path)
	})

	noReader := stubFileReader{
		returns: []struct {
			b   []byte
			err error
		}{
			{b: []byte("\x1b[10;10R"), err: nil},
			{b: []byte("\x1b[10;10R"), err: nil},
			{b: []byte("n\n"), err: nil},
			{b: []byte(""), err: io.EOF},
		},
	}

	t.Run("User cancels folder creation", func(t *testing.T) {
		noReader.Reset()
		nonExistingFolder := t.TempDir() + "/test/folder"

		shouldContinue, err := UseOrCreateAppFolder(nonExistingFolder, &noReader)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
		assert.NoDirExists(t, nonExistingFolder)
	})

	t.Run("User rejects existing folder", func(t *testing.T) {
		noReader.Reset()
		existingPath := t.TempDir()
		_, err := os.Stat(existingPath)
		require.NoError(t, err)

		shouldContinue, err := UseOrCreateAppFolder(existingPath, &noReader)

		assert.False(t, shouldContinue)
		assert.NoError(t, err)
	})
}
