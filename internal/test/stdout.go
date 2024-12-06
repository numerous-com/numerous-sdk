package test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Patches Stdout, runs the function fn, closes the stdout writer so that
// the stdout reader can be read until the end, and returns a stdout reader and
// any error returned by fn.
func RunEWithPatchedStdout(t *testing.T, fn func() error) (io.Reader, error) {
	t.Helper()

	r, w := patchStdout(t)
	err := fn()
	w.Close()

	return r, err
}

// Patches Stdout, runs the function fn, and closes the stdout writer so that
// the stdout reader can be read until the end, and returns a stdout reader.
func RunWithPatchedStdout(t *testing.T, fn func()) io.Reader {
	t.Helper()

	r, w := patchStdout(t)
	fn()
	w.Close()

	return r
}

func patchStdout(t *testing.T) (io.ReadCloser, io.WriteCloser) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	realStdout := os.Stdout
	t.Cleanup(func() {
		os.Stdout = realStdout
		w.Close()
		r.Close()
	})
	os.Stdout = w

	return r, w
}
