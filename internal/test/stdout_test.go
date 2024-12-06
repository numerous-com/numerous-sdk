package test

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("test error")

func TestRunWithPatchedStdout(t *testing.T) {
	t.Run("captures writes to stdout", func(t *testing.T) {
		expected := "some printed text"

		stdout := RunWithPatchedStdout(t, func() {
			os.Stdout.WriteString(expected)
		})

		actual, err := io.ReadAll(stdout)

		assert.NoError(t, err)
		assert.Equal(t, expected, string(actual))
	})
}

func TestRunEWithPatchedStdout(t *testing.T) {
	t.Run("captures writes to stdout", func(t *testing.T) {
		expected := "some printed text"

		stdout, err := RunEWithPatchedStdout(t, func() error {
			os.Stdout.WriteString(expected)
			return nil
		})

		if assert.NoError(t, err) {
			actual, err := io.ReadAll(stdout)
			assert.NoError(t, err)
			assert.Equal(t, expected, string(actual))
		}
	})

	t.Run("returns expected error", func(t *testing.T) {
		_, err := RunEWithPatchedStdout(t, func() error {
			return errTest
		})

		assert.ErrorIs(t, err, errTest)
	})
}
