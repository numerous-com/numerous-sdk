package test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func PatchStdout(t *testing.T) (io.ReadCloser, io.WriteCloser) {
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

// Patches Stdout, runs the function fn, and closes the stdout writer so that
// the stdout reader can be read until the end.
func RunWithPatchedStdout(t *testing.T, fn func() error) (io.Reader, error) {
	t.Helper()

	r, w := PatchStdout(t)
	t.Cleanup(func() { r.Close() })
	err := fn()
	w.Close()

	return r, err
}
