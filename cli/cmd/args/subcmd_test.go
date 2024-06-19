package args

import (
	"flag"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSubCommandRequired(t *testing.T) {
	t.Run("given no arguments then it returns ErrHelp", func(t *testing.T) {
		err := SubCommandRequired(&cobra.Command{}, []string{})

		assert.ErrorIs(t, err, flag.ErrHelp)
	})

	t.Run("given an argument it returns no error", func(t *testing.T) {
		err := SubCommandRequired(&cobra.Command{}, []string{"subcmd"})

		assert.NoError(t, err)
	})
}
