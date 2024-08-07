package args

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestOptionalAppDir(t *testing.T) {
	t.Run("given single argument then it updates the variable", func(t *testing.T) {
		var appDir string

		err := OptionalAppDir(&appDir)(nil, []string{"some-app-dir"})

		assert.NoError(t, err)
		assert.Equal(t, "some-app-dir", appDir)
	})

	t.Run("given zero arguments then it does not update the variable", func(t *testing.T) {
		appDir := "some-preexisting-value"

		err := OptionalAppDir(&appDir)(nil, []string{})

		assert.NoError(t, err)
		assert.Equal(t, "some-preexisting-value", appDir)
	})

	t.Run("given more than one argument then it returns an error", func(t *testing.T) {
		var appDir string

		err := OptionalAppDir(&appDir)(&cobra.Command{}, []string{"arg1", "arg2", "arg3"})

		assert.ErrorIs(t, err, ErrOptionalAppDirArgCount)
	})
}
