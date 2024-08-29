package token

import (
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/token/create"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "token",
	Short:   "Manage personal access tokens for the Numerous API",
	Args:    args.SubCommandRequired,
	GroupID: group.AdditionalCommandsGroupID,
}

func init() {
	Cmd.AddCommand(create.Cmd)
}
