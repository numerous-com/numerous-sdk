package args

import (
	"flag"

	"github.com/spf13/cobra"
)

func SubCommandRequired(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	return nil
}
