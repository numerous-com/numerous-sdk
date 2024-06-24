package wizard

import (
	"io"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockFileReader struct {
	index   int
	returns []struct {
		b   []byte
		err error
	}
}

func (f *MockFileReader) Read(p []byte) (n int, err error) {
	if f.index > len(f.returns) {
		return 0, io.EOF
	} else {
		ret := f.returns[f.index]
		f.index++
		copy(p, ret.b)

		return len(ret.b), ret.err
	}
}

func (f *MockFileReader) Fd() uintptr {
	return 0
}

func (f *MockFileReader) Reset() {
	f.index = 0
}

func TestUseOrCreateAppFolder(t *testing.T) {
	yesReader := MockFileReader{
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
		filePathTest := t.TempDir() + "/test/folder"

		shouldContinue, err := UseOrCreateAppFolder(filePathTest, &yesReader)

		assert.True(t, shouldContinue)
		require.NoError(t, err)
		_, err = os.Stat(filePathTest)
		require.NoError(t, err)
	})

	t.Run("Identify the current user folder structure", func(t *testing.T) {
		yesReader.Reset()
		filePathTest := t.TempDir()

		shouldContinue, err := UseOrCreateAppFolder(filePathTest, &yesReader)

		assert.True(t, shouldContinue)
		require.NoError(t, err)
		_, err = os.Stat(filePathTest)
		require.NoError(t, err)
	})

	noReader := MockFileReader{
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
		require.NoError(t, err)
		_, err = os.Stat(nonExistingFolder)
		os.IsNotExist(err)
		assert.Equal(t, syscall.ENOENT, err.(*os.PathError).Err)
	})

	t.Run("User rejects existing folder", func(t *testing.T) {
		noReader.Reset()
		existingPath := t.TempDir()
		_, err := os.Stat(existingPath)
		require.NoError(t, err)

		shouldContinue, err := UseOrCreateAppFolder(existingPath, &noReader)

		assert.False(t, shouldContinue)
		require.NoError(t, err)
	})
}
