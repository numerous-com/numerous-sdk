package args

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionalAppDir(t *testing.T) {
	t.Run("given single argument then it updates the variable", func(t *testing.T) {
		wd := tempChdir(t)
		type testCase struct {
			arg      string
			expected string
		}

		for _, tc := range []testCase{
			{
				arg:      "relative-dir",
				expected: filepath.Join(wd, "relative-dir"),
			},
			{
				arg:      "relative/nested/dir",
				expected: filepath.Join(wd, "relative", "nested", "dir"),
			},
			{
				arg:      "/some/absolute/path",
				expected: "/some/absolute/path",
			},
		} {
			var appDir string

			err := OptionalAppDir(&appDir)(nil, []string{tc.arg})

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, appDir)
		}
	})

	t.Run("given zero arguments then it updates to the current directory", func(t *testing.T) {
		wd := tempChdir(t)
		appDir := "some-preexisting-value"

		err := OptionalAppDir(&appDir)(nil, []string{})

		assert.NoError(t, err)
		assert.Equal(t, wd, appDir)
	})

	t.Run("given more than one argument then it returns an error", func(t *testing.T) {
		var appDir string

		err := OptionalAppDir(&appDir)(&cobra.Command{}, []string{"arg1", "arg2", "arg3"})

		assert.ErrorIs(t, err, ErrOptionalAppDirArgCount)
	})
}

// Create a temporary directory, change directory to it, return to the original
// working directory on test cleanup. Returns the temporary directory that the
// current directory is changed to.
func tempChdir(t *testing.T) string {
	t.Helper()

	testWd := t.TempDir()
	prevWd, err := os.Getwd()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Chdir(prevWd)
	})

	require.NoError(t, os.Chdir(testWd))

	return testWd
}
