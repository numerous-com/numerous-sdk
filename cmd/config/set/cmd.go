package set

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
)

var Cmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	return errorhandling.ErrorAlreadyPrinted(configSet(args))
}
